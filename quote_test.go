package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// validQuote zwraca minimalny Quote, który przechodzi Validate() bez błędu.
// Pojedyncze testy mutują kopię i sprawdzają wybrane reguły walidacji.
func validQuote() Quote {
	return Quote{
		Company: Company{
			Name: "ACME sp. z o.o.",
			NIP:  "1234567890",
		},
		Client: "Jan Kowalski",
		Items: []LineItem{
			{Name: "Usługa montażowa", Quantity: 1, UnitPrice: 100},
		},
	}
}

func TestQuote_Validate(t *testing.T) {
	cases := []struct {
		nazwa      string
		mutator    func(q *Quote)
		oczekBlad  bool
		fragmentBl string // opcjonalny fragment, który musi znaleźć się w komunikacie błędu
	}{
		{
			nazwa:     "poprawne dane",
			mutator:   func(q *Quote) {},
			oczekBlad: false,
		},
		{
			nazwa:      "pusta nazwa firmy",
			mutator:    func(q *Quote) { q.Company.Name = "" },
			oczekBlad:  true,
			fragmentBl: "nazwa_firmy",
		},
		{
			nazwa:      "nazwa firmy z samych białych znaków",
			mutator:    func(q *Quote) { q.Company.Name = "   \t " },
			oczekBlad:  true,
			fragmentBl: "nazwa_firmy",
		},
		{
			nazwa:      "pusty klient",
			mutator:    func(q *Quote) { q.Client = "" },
			oczekBlad:  true,
			fragmentBl: "klient",
		},
		{
			nazwa:      "klient z samych białych znaków",
			mutator:    func(q *Quote) { q.Client = "   " },
			oczekBlad:  true,
			fragmentBl: "klient",
		},
		{
			nazwa:      "pusta lista pozycji (nil)",
			mutator:    func(q *Quote) { q.Items = nil },
			oczekBlad:  true,
			fragmentBl: "pozycji",
		},
		{
			nazwa:      "pusta lista pozycji (slice zerowy)",
			mutator:    func(q *Quote) { q.Items = []LineItem{} },
			oczekBlad:  true,
			fragmentBl: "pozycji",
		},
		{
			nazwa:      "ilość równa zero",
			mutator:    func(q *Quote) { q.Items[0].Quantity = 0 },
			oczekBlad:  true,
			fragmentBl: "ilość",
		},
		{
			nazwa:      "ilość ujemna",
			mutator:    func(q *Quote) { q.Items[0].Quantity = -1 },
			oczekBlad:  true,
			fragmentBl: "ilość",
		},
		{
			nazwa:      "cena jednostkowa ujemna",
			mutator:    func(q *Quote) { q.Items[0].UnitPrice = -0.01 },
			oczekBlad:  true,
			fragmentBl: "cena_jednostkowa",
		},
		{
			nazwa:     "cena jednostkowa równa zero jest dozwolona",
			mutator:   func(q *Quote) { q.Items[0].UnitPrice = 0 },
			oczekBlad: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			q := validQuote()
			tc.mutator(&q)
			err := q.Validate()

			if tc.oczekBlad && err == nil {
				t.Fatalf("oczekiwano błędu, otrzymano nil")
			}
			if !tc.oczekBlad && err != nil {
				t.Fatalf("nie oczekiwano błędu, otrzymano: %v", err)
			}
			if tc.oczekBlad && tc.fragmentBl != "" && !strings.Contains(err.Error(), tc.fragmentBl) {
				t.Errorf("komunikat %q nie zawiera fragmentu %q", err.Error(), tc.fragmentBl)
			}
		})
	}
}

func TestQuote_UnmarshalJSON_PlaskiJSON(t *testing.T) {
	plaski := []byte(`{
		"nazwa_firmy": "ACME sp. z o.o.",
		"nip": "1234567890",
		"adres": "ul. Przykładowa 1",
		"miasto": "00-001 Warszawa",
		"telefon": "+48 600 700 800",
		"email": "biuro@acme.pl",
		"logo_base64": "data:image/png;base64,AAAA",
		"klient": "Jan Kowalski",
		"numer_oferty": "2026/06/001",
		"data_waznosci": "2026-07-12",
		"uwagi": "Płatność 14 dni",
		"pozycje": [
			{"nazwa": "Usługa A", "ilosc": 2, "cena_jednostkowa": 150.5},
			{"nazwa": "Usługa B", "ilosc": 1, "cena_jednostkowa": 49.5}
		]
	}`)

	var q Quote
	if err := json.Unmarshal(plaski, &q); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	t.Run("pola firmy", func(t *testing.T) {
		cases := []struct {
			pole, got, want string
		}{
			{"Name", q.Company.Name, "ACME sp. z o.o."},
			{"NIP", q.Company.NIP, "1234567890"},
			{"Address", q.Company.Address, "ul. Przykładowa 1"},
			{"City", q.Company.City, "00-001 Warszawa"},
			{"Phone", q.Company.Phone, "+48 600 700 800"},
			{"Email", q.Company.Email, "biuro@acme.pl"},
			{"LogoBase64", q.Company.LogoBase64, "data:image/png;base64,AAAA"},
		}
		for _, c := range cases {
			if c.got != c.want {
				t.Errorf("Company.%s = %q, want %q", c.pole, c.got, c.want)
			}
		}
	})

	t.Run("pola oferty", func(t *testing.T) {
		if q.Client != "Jan Kowalski" {
			t.Errorf("Client = %q", q.Client)
		}
		if q.Number != "2026/06/001" {
			t.Errorf("Number = %q", q.Number)
		}
		if q.ValidUntil != "2026-07-12" {
			t.Errorf("ValidUntil = %q", q.ValidUntil)
		}
		if q.Notes != "Płatność 14 dni" {
			t.Errorf("Notes = %q", q.Notes)
		}
	})

	t.Run("pozycje", func(t *testing.T) {
		if len(q.Items) != 2 {
			t.Fatalf("len(Items) = %d, want 2", len(q.Items))
		}
		if q.Items[0].Name != "Usługa A" || q.Items[0].Quantity != 2 || q.Items[0].UnitPrice != 150.5 {
			t.Errorf("Items[0] = %+v", q.Items[0])
		}
		if q.Items[1].Name != "Usługa B" || q.Items[1].Quantity != 1 || q.Items[1].UnitPrice != 49.5 {
			t.Errorf("Items[1] = %+v", q.Items[1])
		}
	})
}

func TestQuote_UnmarshalJSON_BrakOpcjonalnychPol(t *testing.T) {
	// Tylko pola wymagane przez Validate(): nazwa_firmy, klient, pozycje.
	// Pozostałe (logo, email, telefon, NIP, adres, numer, data, uwagi) są
	// opcjonalne — Unmarshal nie może panikować i ma zostawić zero values.
	minimalny := []byte(`{
		"nazwa_firmy": "ACME",
		"klient": "Jan Kowalski",
		"pozycje": [{"nazwa": "X", "ilosc": 1, "cena_jednostkowa": 10}]
	}`)

	var q Quote
	if err := json.Unmarshal(minimalny, &q); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if q.Company.Name != "ACME" {
		t.Errorf("Company.Name = %q", q.Company.Name)
	}
	if q.Client != "Jan Kowalski" {
		t.Errorf("Client = %q", q.Client)
	}

	t.Run("opcjonalne pola firmy są puste", func(t *testing.T) {
		cases := []struct {
			pole, got string
		}{
			{"NIP", q.Company.NIP},
			{"Address", q.Company.Address},
			{"City", q.Company.City},
			{"Phone", q.Company.Phone},
			{"Email", q.Company.Email},
			{"LogoBase64", q.Company.LogoBase64},
		}
		for _, c := range cases {
			if c.got != "" {
				t.Errorf("Company.%s = %q, oczekiwano \"\"", c.pole, c.got)
			}
		}
	})

	t.Run("opcjonalne pola oferty są puste", func(t *testing.T) {
		if q.Number != "" {
			t.Errorf("Number = %q", q.Number)
		}
		if q.ValidUntil != "" {
			t.Errorf("ValidUntil = %q", q.ValidUntil)
		}
		if q.Notes != "" {
			t.Errorf("Notes = %q", q.Notes)
		}
	})

	if len(q.Items) != 1 {
		t.Errorf("len(Items) = %d, want 1", len(q.Items))
	}

	// Brakująca lista pozycji w ogóle — Unmarshal też nie może panikować.
	bezPozycji := []byte(`{"nazwa_firmy": "ACME", "klient": "Jan"}`)
	var q2 Quote
	if err := json.Unmarshal(bezPozycji, &q2); err != nil {
		t.Fatalf("Unmarshal bezPozycji: %v", err)
	}
	if q2.Items != nil {
		t.Errorf("Items = %v, oczekiwano nil", q2.Items)
	}
}

func TestQuote_MarshalUnmarshal_Roundtrip(t *testing.T) {
	// MarshalJSON musi emitować JSON PŁASKI (pola firmy obok klient/pozycji),
	// a UnmarshalJSON musi go odczytać z powrotem do zagnieżdżonej struktury.
	orig := validQuote()
	orig.Company.Email = "biuro@acme.pl"
	orig.Number = "2026/06/001"

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal do mapy: %v", err)
	}
	if _, ok := m["nazwa_firmy"]; !ok {
		t.Errorf("brak pola nazwa_firmy na najwyższym poziomie JSON: %s", data)
	}
	if _, ok := m["klient"]; !ok {
		t.Errorf("brak pola klient na najwyższym poziomie JSON: %s", data)
	}
	if _, ok := m["firma"]; ok {
		t.Errorf("nieoczekiwane zagnieżdżone pole \"firma\" w JSON: %s", data)
	}

	var roundtrip Quote
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("Unmarshal roundtrip: %v", err)
	}
	if roundtrip.Company.Name != orig.Company.Name {
		t.Errorf("Company.Name po roundtrip = %q, want %q", roundtrip.Company.Name, orig.Company.Name)
	}
	if roundtrip.Company.Email != orig.Company.Email {
		t.Errorf("Company.Email po roundtrip = %q, want %q", roundtrip.Company.Email, orig.Company.Email)
	}
	if roundtrip.Client != orig.Client {
		t.Errorf("Client po roundtrip = %q, want %q", roundtrip.Client, orig.Client)
	}
	if roundtrip.Number != orig.Number {
		t.Errorf("Number po roundtrip = %q, want %q", roundtrip.Number, orig.Number)
	}
	if len(roundtrip.Items) != len(orig.Items) {
		t.Fatalf("len(Items) po roundtrip = %d, want %d", len(roundtrip.Items), len(orig.Items))
	}
}

func TestLineItem_Total(t *testing.T) {
	cases := []struct {
		nazwa string
		li    LineItem
		oczek float64
	}{
		{"liczby całkowite", LineItem{Quantity: 5, UnitPrice: 20}, 100},
		{"grosze 3 x 33.33", LineItem{Quantity: 3, UnitPrice: 33.33}, 99.99},
		{"zero ilości", LineItem{Quantity: 0, UnitPrice: 100}, 0},
		{"zero ceny", LineItem{Quantity: 10, UnitPrice: 0}, 0},
		{"ułamkowa ilość", LineItem{Quantity: 2.5, UnitPrice: 4}, 10},
	}
	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			got := tc.li.Total()
			if math.Abs(got-tc.oczek) > 0.005 {
				t.Errorf("Total() = %v, want %v (różnica %v)", got, tc.oczek, math.Abs(got-tc.oczek))
			}
		})
	}
}

func TestQuote_Total(t *testing.T) {
	cases := []struct {
		nazwa   string
		pozycje []LineItem
		oczek   float64
	}{
		{
			nazwa:   "brak pozycji",
			pozycje: nil,
			oczek:   0,
		},
		{
			nazwa: "jedna pozycja",
			pozycje: []LineItem{
				{Name: "A", Quantity: 3, UnitPrice: 100},
			},
			oczek: 300,
		},
		{
			nazwa: "wiele pozycji bez ułamków",
			pozycje: []LineItem{
				{Name: "A", Quantity: 2, UnitPrice: 50},
				{Name: "B", Quantity: 1, UnitPrice: 150},
				{Name: "C", Quantity: 4, UnitPrice: 25},
			},
			oczek: 350,
		},
		{
			nazwa: "grosze - 3 x 33.33",
			pozycje: []LineItem{
				{Name: "A", Quantity: 3, UnitPrice: 33.33},
			},
			oczek: 99.99,
		},
		{
			nazwa: "klasyczny float 0.1 + 0.2",
			pozycje: []LineItem{
				{Name: "A", Quantity: 1, UnitPrice: 0.1},
				{Name: "B", Quantity: 1, UnitPrice: 0.2},
			},
			oczek: 0.30,
		},
		{
			nazwa: "wiele pozycji z groszami",
			pozycje: []LineItem{
				{Name: "A", Quantity: 7, UnitPrice: 12.34},
				{Name: "B", Quantity: 3, UnitPrice: 0.07},
				{Name: "C", Quantity: 1, UnitPrice: 999.99},
			},
			oczek: 7*12.34 + 3*0.07 + 999.99,
		},
		{
			nazwa: "ułamkowa ilość",
			pozycje: []LineItem{
				{Name: "A", Quantity: 2.5, UnitPrice: 80},
				{Name: "B", Quantity: 0.5, UnitPrice: 100},
			},
			oczek: 250,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			q := Quote{Items: tc.pozycje}
			got := q.Total()

			// Tolerancja pół grosza — pokrywa szum arytmetyki zmiennoprzecinkowej,
			// a jednocześnie gwarantuje, że wartość zaokrąglona do groszy jest poprawna.
			if math.Abs(got-tc.oczek) > 0.005 {
				t.Errorf("Total() = %v, want %v (różnica %v)", got, tc.oczek, math.Abs(got-tc.oczek))
			}

			// Sprawdzamy też prezentację — to, co użytkownik widzi w PDF jako „Razem".
			if g, w := formatPLN(got), formatPLN(tc.oczek); g != w {
				t.Errorf("formatPLN(Total()) = %q, want %q", g, w)
			}
		})
	}
}

func TestHandleQuote_BodyPowyzej1MB(t *testing.T) {
	// handleQuote opakowuje body w http.MaxBytesReader(w, r.Body, 1<<20).
	// Body większy niż 1 MB musi skutkować odpowiedzią 400 Bad Request,
	// zanim cokolwiek dotrze do generatora PDF.
	const limit = 1 << 20
	wypelniacz := strings.Repeat("a", limit+1024)

	q := Quote{
		Company: Company{Name: "ACME"},
		Client:  "Jan Kowalski",
		Notes:   wypelniacz,
		Items: []LineItem{
			{Name: "Usługa A", Quantity: 1, UnitPrice: 100},
		},
	}
	body, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(body) <= limit {
		t.Fatalf("przygotowany body ma %d B, oczekiwano > %d B", len(body), limit)
	}

	req := httptest.NewRequest(http.MethodPost, "/quote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handleQuote(rec, req)

	if rec.Code != http.StatusBadRequest {
		respBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, respBody)
	}
}
