package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	parseBodyLimit   = 4 << 10 // 4 KiB
	parseRatePerMin  = 10
	geminiModel      = "gemini-2.5-flash"
	geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"
)

// geminiAPIBaseURL pozwala podmienić endpoint Gemini w testach.
var geminiAPIBase = geminiAPIBaseURL

var httpClientGemini = &http.Client{Timeout: 30 * time.Second}

const parseSystemPrompt = `Jesteś ścisłym parserem kosztorysów budowlanych. Twoim jedynym zadaniem jest wyodrębnienie pozycji z tekstu użytkownika.

ZASADY:
- Ignoruj tekst poboczny, komentarze, nagłówki, stopki i wszelkie treści niebędące pozycjami kosztorysu.
- Przed wygenerowaniem formatu JSON, dokonaj pełnej korekty ortograficznej wykrytych nazw materiałów i usług. Upewnij się, że każda nazwa pozycji zaczyna się od wielkiej litery i brzmi profesjonalnie w kontekście formalnego kosztorysu. Jednostki (np. m, szt.) pozostaw z małej litery.
- Zwróć WYŁĄCZNIE surową tablicę JSON bez żadnego dodatkowego tekstu przed ani po tablicy.
- ZAKAZ używania znaczników Markdown (np. ` + "```json" + ` lub ` + "```" + `).
- Każdy element tablicy musi mieć dokładnie pola: "nazwa" (string), "ilosc" (number), "cena" (number).
- "cena" to cena jednostkowa w złotych.
- Jeśli nie ma pozycji, zwróć pustą tablicę [].`

type parseRequest struct {
	Tekst string `json:"tekst"`
}

type parseItem struct {
	Nazwa string  `json:"nazwa"`
	Ilosc float64 `json:"ilosc"`
	Cena  float64 `json:"cena"`
}

type geminiGenerateRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature      float64              `json:"temperature"`
	ResponseMimeType string               `json:"responseMimeType"`
	ThinkingConfig   *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

type geminiThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
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
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func callGeminiParse(ctx context.Context, apiKey, userText string) ([]parseItem, error) {
	reqBody := geminiGenerateRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: parseSystemPrompt}},
		},
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: userText}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
			ThinkingConfig:   &geminiThinkingConfig{ThinkingBudget: 0},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBase, geminiModel)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := httpClientGemini.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var env geminiGenerateResponse
		if json.Unmarshal(body, &env) == nil && env.Error != nil && env.Error.Message != "" {
			return nil, fmt.Errorf("gemini status %d: %s", resp.StatusCode, env.Error.Message)
		}
		return nil, fmt.Errorf("gemini status %d", resp.StatusCode)
	}

	var gen geminiGenerateResponse
	if err := json.Unmarshal(body, &gen); err != nil {
		return nil, fmt.Errorf("decode gemini: %w", err)
	}
	if len(gen.Candidates) == 0 || len(gen.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}

	raw := stripMarkdownJSON(gen.Candidates[0].Content.Parts[0].Text)
	var items []parseItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("decode items: %w", err)
	}

	for i := range items {
		items[i].Nazwa = strings.TrimSpace(items[i].Nazwa)
		if items[i].Nazwa == "" {
			return nil, fmt.Errorf("item %d: empty nazwa", i)
		}
		if items[i].Ilosc <= 0 {
			return nil, fmt.Errorf("item %d: invalid ilosc", i)
		}
		if items[i].Cena < 0 {
			return nil, fmt.Errorf("item %d: invalid cena", i)
		}
	}

	return items, nil
}

// handleParse proxy do Gemini API — parsuje notatkę tekstową na pozycje kosztorysu.
// POST /api/parse {"tekst": "..."} → [{"nazwa": "...", "ilosc": 1, "cena": 100}]
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeParseError(w, http.StatusMethodNotAllowed, "dozwolona jest tylko metoda POST")
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		writeParseError(w, http.StatusServiceUnavailable, "Usługa parsowania AI jest tymczasowo niedostępna")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, parseBodyLimit))
	if err != nil {
		writeParseError(w, http.StatusBadRequest, "nie udało się odczytać treści żądania")
		return
	}

	var in parseRequest
	if err := json.Unmarshal(body, &in); err != nil {
		writeParseError(w, http.StatusBadRequest, "nieprawidłowy format JSON")
		return
	}

	tekst := strings.TrimSpace(in.Tekst)
	if tekst == "" {
		writeParseError(w, http.StatusBadRequest, "pole tekst nie może być puste")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := callGeminiParse(ctx, apiKey, tekst)
	if err != nil {
		writeParseError(w, http.StatusBadGateway, "nie udało się przetworzyć notatki — spróbuj ponownie")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(items)
}
