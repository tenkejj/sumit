package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// mfBaseURL pozwala podmienić endpoint Białej Listy MF w testach.
var mfBaseURL = "https://wl-api.mf.gov.pl"

// httpClientNIP ma sztywny timeout, żeby pojedynczy zwis MF nie zablokował serwera.
var httpClientNIP = &http.Client{Timeout: 8 * time.Second}

// nipResponse to płaski JSON oddawany frontendowi.
type nipResponse struct {
	Nazwa string `json:"nazwa"`
	Adres string `json:"adres"`
	NIP   string `json:"nip"`
}

// mfSubject odpowiada strukturze "subject" z odpowiedzi Białej Listy.
// Bierzemy tylko pola, których faktycznie używamy w UI.
type mfSubject struct {
	Name             string `json:"name"`
	NIP              string `json:"nip"`
	WorkingAddress   string `json:"workingAddress"`
	ResidenceAddress string `json:"residenceAddress"`
}

// mfEnvelope mapuje zewnętrzną kopertę odpowiedzi MF.
type mfEnvelope struct {
	Result struct {
		Subject *mfSubject `json:"subject"`
	} `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// sanitizeNIP usuwa wszystko poza cyframi (spacje, myślniki, prefiks PL itp.).
func sanitizeNIP(s string) string {
	s = strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(s)), "PL")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// writeNIPError serializuje błąd jako JSON z polem "error" i ustala status HTTP.
func writeNIPError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// handleNIP to proxy do API Białej Listy Ministerstwa Finansów.
// GET /api/nip?nip=1234567890 → {"nazwa": "...", "adres": "...", "nip": "..."}.
func handleNIP(w http.ResponseWriter, r *http.Request) {
	nip := sanitizeNIP(r.URL.Query().Get("nip"))
	if len(nip) != 10 {
		writeNIPError(w, http.StatusBadRequest, "NIP musi zawierać dokładnie 10 cyfr")
		return
	}

	date := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("%s/api/search/nip/%s?date=%s", mfBaseURL, nip, date)

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		writeNIPError(w, http.StatusInternalServerError, "nie udało się przygotować zapytania do MF")
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClientNIP.Do(req)
	if err != nil {
		writeNIPError(w, http.StatusBadGateway, "nie udało się połączyć z Białą Listą MF")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	if err != nil {
		writeNIPError(w, http.StatusBadGateway, "błąd odczytu odpowiedzi MF")
		return
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			writeNIPError(w, http.StatusBadRequest, "nieprawidłowy NIP")
		case http.StatusNotFound:
			writeNIPError(w, http.StatusNotFound, "nie znaleziono firmy o podanym NIP")
		default:
			writeNIPError(w, http.StatusBadGateway, fmt.Sprintf("MF odpowiedziało kodem %d", resp.StatusCode))
		}
		return
	}

	var env mfEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		writeNIPError(w, http.StatusBadGateway, "nie udało się zdekodować odpowiedzi MF")
		return
	}
	if env.Result.Subject == nil {
		writeNIPError(w, http.StatusNotFound, "nie znaleziono firmy o podanym NIP w rejestrze")
		return
	}

	subj := env.Result.Subject
	address := strings.TrimSpace(subj.WorkingAddress)
	if address == "" {
		address = strings.TrimSpace(subj.ResidenceAddress)
	}

	out := nipResponse{
		Nazwa: strings.TrimSpace(subj.Name),
		Adres: address,
		NIP:   strings.TrimSpace(subj.NIP),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(out)
}
