package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	parseBodyLimit      = 4 << 10       // 4 KiB — same tekst
	parseBodyLimitImage = 3 << 20       // 3 MiB — żądanie ze zdjęciem
	parseRatePerMin     = 10
	parseMaxKontekst    = 30
	parseMaxKontekstLen = 120
	groqModelText       = "llama-3.3-70b-versatile"
	groqModelVision     = "meta-llama/llama-4-scout-17b-16e-instruct"
	groqAPIURL          = "https://api.groq.com/openai/v1/chat/completions"
)

// groqAPIEndpoint pozwala podmienić endpoint Groq w testach.
var groqAPIEndpoint = groqAPIURL

var httpClientGroq = &http.Client{Timeout: 30 * time.Second}

const parseSystemPrompt = `Jesteś parserem pozycji wyceny/oferty. Wyodrębnij z tekstu lub zdjęcia notatki użytkownika WSZYSTKIE pozycje do wyceny — dowolne produkty, usługi, materiały lub inne wpisy, bez względu na branżę.

ZASADY:
- Każda linia lub fraza opisująca coś do wyceny to osobna pozycja.
- Nie odrzucaj pozycji tylko dlatego, że nie brzmią jak materiał budowlany — użytkownik może wpisać cokolwiek (np. własną nazwę usługi, produkt spoza katalogu).
- Opcjonalna lista kontekstu to tylko podpowiedź przy literówkach i brakujących cenach — NIE ogranicza tego, co można dodać.
- Przy zdjęciu notatki: odczytaj odręczny lub drukowany tekst (OCR), zignoruj tło.
- Nazwa: zachowaj sens użytkownika; popraw oczywiste literówki; pierwsza litera wielka. Jednostki (np. m, szt.) zostaw z małej litery.
- Ilość: gdy użytkownik nie poda ilości, ustaw 1.
- Cena jednostkowa w zł: rozpoznaj formy „za 50 zł”, „50zł”, „50 zl”; gdy brak ceny, ustaw 0.
- Zwróć WYŁĄCZNIE surową tablicę JSON bez żadnego dodatkowego tekstu przed ani po tablicy.
- ZAKAZ używania znaczników Markdown (np. ` + "```json" + ` lub ` + "```" + `).
- Każdy element tablicy musi mieć dokładnie pola: "nazwa" (string), "ilosc" (number), "cena" (number).
- Jeśli nie ma pozycji, zwróć pustą tablicę [].`

type parseRequest struct {
	Tekst    string   `json:"tekst"`
	Kontekst []string `json:"kontekst,omitempty"`
	Obraz    string   `json:"obraz,omitempty"`
	MimeType string   `json:"mime_type,omitempty"`
}

type parseItem struct {
	Nazwa string  `json:"nazwa"`
	Ilosc float64 `json:"ilosc"`
	Cena  float64 `json:"cena"`
}

type groqChatRequest struct {
	Model       string         `json:"model"`
	Messages    []groqMessage  `json:"messages"`
	Temperature float64        `json:"temperature"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type groqContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *groqImageURL `json:"image_url,omitempty"`
}

type groqImageURL struct {
	URL string `json:"url"`
}

type groqChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ipRateEntry struct {
	mu          sync.Mutex
	windowStart time.Time
	count       int
}

var parseRateLimits sync.Map

func writeParseError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func allowParseRequest(ip string) bool {
	now := time.Now()
	v, _ := parseRateLimits.LoadOrStore(ip, &ipRateEntry{windowStart: now})
	entry := v.(*ipRateEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if now.Sub(entry.windowStart) >= time.Minute {
		entry.windowStart = now
		entry.count = 0
	}
	if entry.count >= parseRatePerMin {
		return false
	}
	entry.count++
	return true
}

// rateLimitParse ogranicza POST /api/parse do 10 żądań na minutę z jednego IP.
func rateLimitParse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowParseRequest(clientIP(r)) {
			writeParseError(w, http.StatusTooManyRequests, "Zbyt wiele żądań — spróbuj ponownie za chwilę")
			return
		}
		next(w, r)
	}
}

func stripMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
	}
	// Model czasem zwraca tablicę JSON z końcowym fence bez otwarcia (np. "[]\n```").
	if idx := strings.Index(s, "```"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	return s
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func capitalizeFirstLetter(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func normalizeNazwa(nazwa string) string {
	nazwa = collapseSpaces(strings.TrimSpace(nazwa))
	nazwa = strings.ReplaceAll(nazwa, "m2", "m²")
	nazwa = strings.ReplaceAll(nazwa, "M2", "m²")
	return capitalizeFirstLetter(nazwa)
}

func isValidParseItem(item parseItem) bool {
	return item.Nazwa != "" && item.Ilosc > 0 && item.Cena >= 0
}

func postProcessParseItems(items []parseItem) []parseItem {
	out := make([]parseItem, 0, len(items))
	for _, item := range items {
		item.Nazwa = normalizeNazwa(item.Nazwa)
		if item.Ilosc <= 0 && item.Nazwa != "" {
			item.Ilosc = 1
		}
		if !isValidParseItem(item) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func sanitizeKontekst(kontekst []string) []string {
	if len(kontekst) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, parseMaxKontekst)
	for _, raw := range kontekst {
		s := collapseSpaces(strings.TrimSpace(raw))
		if s == "" {
			continue
		}
		if len(s) > parseMaxKontekstLen {
			s = s[:parseMaxKontekstLen]
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
		if len(out) >= parseMaxKontekst {
			break
		}
	}
	return out
}

func buildUserMessage(in parseRequest) string {
	var b strings.Builder
	kontekst := sanitizeKontekst(in.Kontekst)
	if len(kontekst) > 0 {
		b.WriteString("ZNANE POZYCJE UŻYTKOWNIKA (opcjonalna podpowiedź — użyj do literówek i uzupełnienia brakującej ceny; NIE odrzucaj pozycji spoza tej listy):\n")
		for _, nazwa := range kontekst {
			b.WriteString("- ")
			b.WriteString(nazwa)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	tekst := strings.TrimSpace(in.Tekst)
	if tekst != "" {
		b.WriteString("NOTATKA UŻYTKOWNIKA:\n")
		b.WriteString(tekst)
	} else if in.Obraz != "" {
		b.WriteString("Wyodrębnij pozycje z załączonego zdjęcia notatki.")
	}
	return b.String()
}

func validateImageInput(obraz, mimeType string) error {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return fmt.Errorf("nieobsługiwany typ obrazu")
	}
	raw, err := base64.StdEncoding.DecodeString(obraz)
	if err != nil {
		return fmt.Errorf("nieprawidłowe kodowanie obrazu")
	}
	if len(raw) == 0 {
		return fmt.Errorf("pusty obraz")
	}
	if len(raw) > 2<<20 {
		return fmt.Errorf("obraz jest zbyt duży")
	}
	return nil
}

func buildGroqUserContent(in parseRequest) (any, error) {
	userMsg := buildUserMessage(in)
	hasObraz := strings.TrimSpace(in.Obraz) != ""
	if hasObraz {
		if err := validateImageInput(in.Obraz, in.MimeType); err != nil {
			return nil, err
		}
	}
	if !hasObraz {
		if userMsg == "" {
			return nil, fmt.Errorf("brak treści do przetworzenia")
		}
		return userMsg, nil
	}

	mimeType := strings.TrimSpace(strings.ToLower(in.MimeType))
	dataURI := "data:" + mimeType + ";base64," + strings.TrimSpace(in.Obraz)
	parts := make([]groqContentPart, 0, 2)
	if userMsg != "" {
		parts = append(parts, groqContentPart{Type: "text", Text: userMsg})
	}
	parts = append(parts, groqContentPart{
		Type:     "image_url",
		ImageURL: &groqImageURL{URL: dataURI},
	})
	return parts, nil
}

func groqModelForParse(in parseRequest) string {
	if strings.TrimSpace(in.Obraz) != "" {
		return groqModelVision
	}
	return groqModelText
}

func callGroqParse(ctx context.Context, apiKey string, in parseRequest) ([]parseItem, error) {
	userContent, err := buildGroqUserContent(in)
	if err != nil {
		return nil, err
	}

	reqBody := groqChatRequest{
		Model: groqModelForParse(in),
		Messages: []groqMessage{
			{Role: "system", Content: parseSystemPrompt},
			{Role: "user", Content: userContent},
		},
		Temperature: 0,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqAPIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClientGroq.Do(req)
	if err != nil {
		return nil, fmt.Errorf("groq request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var env groqChatResponse
		if json.Unmarshal(body, &env) == nil && env.Error != nil && env.Error.Message != "" {
			return nil, fmt.Errorf("groq status %d: %s", resp.StatusCode, env.Error.Message)
		}
		return nil, fmt.Errorf("groq status %d", resp.StatusCode)
	}

	var gen groqChatResponse
	if err := json.Unmarshal(body, &gen); err != nil {
		return nil, fmt.Errorf("decode groq: %w", err)
	}
	if len(gen.Choices) == 0 || gen.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("empty groq response")
	}

	raw := stripMarkdownJSON(gen.Choices[0].Message.Content)
	var items []parseItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("decode items: %w", err)
	}

	items = postProcessParseItems(items)
	if len(items) == 0 {
		return nil, fmt.Errorf("no valid items")
	}
	return items, nil
}

func sanitizeParseRequest(in *parseRequest) error {
	in.Tekst = strings.TrimSpace(in.Tekst)
	in.Obraz = strings.TrimSpace(in.Obraz)
	in.MimeType = strings.TrimSpace(strings.ToLower(in.MimeType))
	in.Kontekst = sanitizeKontekst(in.Kontekst)

	hasTekst := in.Tekst != ""
	hasObraz := in.Obraz != ""
	if !hasTekst && !hasObraz {
		return fmt.Errorf("wymagany tekst lub obraz")
	}
	if hasObraz {
		return validateImageInput(in.Obraz, in.MimeType)
	}
	return nil
}

// handleParse proxy do Groq API — parsuje notatkę tekstową lub zdjęcie na pozycje kosztorysu.
// POST /api/parse {"tekst": "...", "kontekst": [...], "obraz": "...", "mime_type": "image/jpeg"}
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeParseError(w, http.StatusMethodNotAllowed, "dozwolona jest tylko metoda POST")
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	if apiKey == "" {
		writeParseError(w, http.StatusServiceUnavailable, "Usługa parsowania AI jest tymczasowo niedostępna")
		return
	}

	peekLimit := parseBodyLimitImage + 1
	body, err := io.ReadAll(io.LimitReader(r.Body, int64(peekLimit)))
	if err != nil {
		writeParseError(w, http.StatusBadRequest, "nie udało się odczytać treści żądania")
		return
	}
	if len(body) >= peekLimit {
		writeParseError(w, http.StatusRequestEntityTooLarge, "żądanie jest zbyt duże")
		return
	}

	var in parseRequest
	if err := json.Unmarshal(body, &in); err != nil {
		writeParseError(w, http.StatusBadRequest, "nieprawidłowy format JSON")
		return
	}

	textOnlyLimit := parseBodyLimit + 1
	if strings.TrimSpace(in.Obraz) == "" && len(body) >= textOnlyLimit {
		writeParseError(w, http.StatusRequestEntityTooLarge, "notatka tekstowa może mieć maksymalnie 4 KB")
		return
	}

	if err := sanitizeParseRequest(&in); err != nil {
		switch {
		case strings.Contains(err.Error(), "wymagany"):
			writeParseError(w, http.StatusBadRequest, "podaj tekst notatki lub załącz zdjęcie")
		case strings.Contains(err.Error(), "obraz"):
			writeParseError(w, http.StatusBadRequest, "nieprawidłowy obraz — użyj JPG lub PNG (max 2 MB)")
		default:
			writeParseError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := callGroqParse(ctx, apiKey, in)
	if err != nil {
		if strings.Contains(err.Error(), "no valid items") {
			writeParseError(w, http.StatusBadRequest, "nie znaleziono poprawnych pozycji w notatce")
			return
		}
		writeParseError(w, http.StatusBadGateway, "nie udało się przetworzyć notatki — spróbuj ponownie")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(items)
}
