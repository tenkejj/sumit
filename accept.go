package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Acceptance przechowuje dane o zaakceptowaniu wyceny przez klienta.
type Acceptance struct {
	Token      string `json:"token"`
	AcceptedAt string `json:"acceptedAt"` // RFC3339
	IP         string `json:"ip"`
	Imie       string `json:"imie"`
}

const acceptancesFile = "acceptances.json"

var acceptanceMu sync.Mutex

func loadAcceptances() (map[string]Acceptance, error) {
	data, err := os.ReadFile(acceptancesFile)
	if os.IsNotExist(err) {
		return make(map[string]Acceptance), nil
	}
	if err != nil {
		return nil, err
	}
	var m map[string]Acceptance
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]Acceptance), nil
	}
	return m, nil
}

func saveAcceptances(m map[string]Acceptance) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(acceptancesFile, data, 0600)
}

// handleAccept obsługuje POST /api/accept
// Body: {"token":"<uuid>","imie":"<opcjonalne>"}
// Idempotentne — pierwsza akceptacja jest utrwalana, kolejne są ignorowane.
func handleAccept(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req struct {
		Token string `json:"token"`
		Imie  string `json:"imie"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "nieprawidłowy JSON", http.StatusBadRequest)
		return
	}
	if !uuidRe.MatchString(req.Token) {
		http.Error(w, "nieprawidłowy token", http.StatusBadRequest)
		return
	}
	req.Imie = strings.TrimSpace(req.Imie)
	if len(req.Imie) > 60 {
		req.Imie = req.Imie[:60]
	}

	acceptanceMu.Lock()
	defer acceptanceMu.Unlock()

	m, err := loadAcceptances()
	if err != nil {
		http.Error(w, fmt.Sprintf("błąd odczytu danych: %v", err), http.StatusInternalServerError)
		return
	}

	existing, alreadyAccepted := m[req.Token]
	if alreadyAccepted {
		// Idempotentne — zwróć istniejącą akceptację
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accepted":   true,
			"acceptedAt": existing.AcceptedAt,
			"imie":       existing.Imie,
		})
		return
	}

	acc := Acceptance{
		Token:      req.Token,
		AcceptedAt: time.Now().UTC().Format(time.RFC3339),
		IP:         clientIP(r),
		Imie:       req.Imie,
	}
	m[req.Token] = acc

	if err := saveAcceptances(m); err != nil {
		http.Error(w, fmt.Sprintf("błąd zapisu: %v", err), http.StatusInternalServerError)
		return
	}

	_ = recordEventFromRequest("client_accepted", "", r)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accepted":   true,
		"acceptedAt": acc.AcceptedAt,
		"imie":       acc.Imie,
	})
}

// handleAcceptStatus obsługuje GET /api/accept?token=<uuid>
func handleAcceptStatus(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if !uuidRe.MatchString(token) {
		http.Error(w, "nieprawidłowy token", http.StatusBadRequest)
		return
	}

	acceptanceMu.Lock()
	defer acceptanceMu.Unlock()

	m, err := loadAcceptances()
	if err != nil {
		http.Error(w, fmt.Sprintf("błąd odczytu danych: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if acc, ok := m[token]; ok {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accepted":   true,
			"acceptedAt": acc.AcceptedAt,
			"imie":       acc.Imie,
		})
	} else {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accepted": false,
		})
	}
}
