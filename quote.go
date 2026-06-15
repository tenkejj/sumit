package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type LineItem struct {
	Name      string  `json:"nazwa"`
	Quantity  float64 `json:"ilosc"`
	UnitPrice float64 `json:"cena_jednostkowa"`
}

func (li LineItem) Total() float64 {
	return li.Quantity * li.UnitPrice
}

// Company zawiera dane sprzedawcy renderowane w nagłówku PDF.
// Pola są deserializowane z płaskiego JSON-a wysyłanego przez frontend
// (patrz Quote.UnmarshalJSON).
type Company struct {
	Name        string `json:"nazwa_firmy"`
	NIP         string `json:"nip"`
	Address     string `json:"adres"`
	City        string `json:"miasto"`
	Phone       string `json:"telefon"`
	Email       string `json:"email"`
	LogoBase64  string `json:"logo_base64"`
	BankAccount string `json:"numer_konta"`
}

type Quote struct {
	Company    Company    `json:"-"`
	Client     string     `json:"klient"`
	Number     string     `json:"numer_oferty"`
	ValidUntil string     `json:"data_waznosci"`
	Notes      string     `json:"uwagi"`
	Items      []LineItem `json:"pozycje"`
}

// quoteJSON to wewnętrzna reprezentacja serializacji przyjmująca płaski
// JSON, jaki wysyła frontend (pola firmy obok klient/pozycji na jednym
// poziomie). Mapowanie do zagnieżdżonej Quote.Company odbywa się w
// (Un)MarshalJSON, dzięki czemu kontrakt JSON pozostaje niezmieniony.
type quoteJSON struct {
	Company
	Client     string     `json:"klient"`
	Number     string     `json:"numer_oferty"`
	ValidUntil string     `json:"data_waznosci"`
	Notes      string     `json:"uwagi"`
	Items      []LineItem `json:"pozycje"`
}

func (q *Quote) UnmarshalJSON(data []byte) error {
	var raw quoteJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*q = Quote{
		Company:    raw.Company,
		Client:     raw.Client,
		Number:     raw.Number,
		ValidUntil: raw.ValidUntil,
		Notes:      raw.Notes,
		Items:      raw.Items,
	}
	return nil
}

func (q Quote) MarshalJSON() ([]byte, error) {
	return json.Marshal(quoteJSON{
		Company:    q.Company,
		Client:     q.Client,
		Number:     q.Number,
		ValidUntil: q.ValidUntil,
		Notes:      q.Notes,
		Items:      q.Items,
	})
}

func (q Quote) Total() float64 {
	var s float64
	for _, li := range q.Items {
		s += li.Total()
	}
	return s
}

func (q Quote) Validate() error {
	if strings.TrimSpace(q.Company.Name) == "" {
		return fmt.Errorf("pole nazwa_firmy jest wymagane")
	}
	if strings.TrimSpace(q.Client) == "" {
		return fmt.Errorf("pole klient jest wymagane")
	}
	if len(q.Items) == 0 {
		return fmt.Errorf("lista pozycji nie może być pusta")
	}
	for i, li := range q.Items {
		if strings.TrimSpace(li.Name) == "" {
			return fmt.Errorf("pozycja #%d: brak nazwy", i+1)
		}
		if li.Quantity <= 0 {
			return fmt.Errorf("pozycja #%d (%s): ilość musi być > 0", i+1, li.Name)
		}
		if li.UnitPrice < 0 {
			return fmt.Errorf("pozycja #%d (%s): cena_jednostkowa nie może być ujemna", i+1, li.Name)
		}
	}
	return nil
}

func handleQuote(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)

	var q Quote
	if err := dec.Decode(&q); err != nil {
		http.Error(w, "nieprawidłowy JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := q.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := GeneratePDF(q, &buf); err != nil {
		http.Error(w, "błąd generowania PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="oferta.pdf"`)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = io.Copy(w, &buf)
}
