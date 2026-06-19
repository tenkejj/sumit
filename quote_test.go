package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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

func TestQuote_ValidateFaktura(t *testing.T) {
	base := validQuote()
	base.Items[0].VatRate = 23

	cases := []struct {
		nazwa     string
		mutator   func(q *Quote)
		oczekBlad bool
		fragment  string
	}{
		{
			nazwa: "faktura_vat z wymaganymi polami",
			mutator: func(q *Quote) {
				q.DocType = DocTypeFakturaVAT
				q.InvoiceNumber = "FV/2026/001"
				q.SaleDate = "2026-06-18"
			},
			oczekBlad: false,
		},
		{
			nazwa: "faktura_vat bez numeru faktury",
			mutator: func(q *Quote) {
				q.DocType = DocTypeFakturaVAT
				q.SaleDate = "2026-06-18"
			},
			oczekBlad: true,
			fragment:  "numer_faktury",
		},
		{
			nazwa: "faktura_vat bez daty sprzedaży",
			mutator: func(q *Quote) {
				q.DocType = DocTypeFakturaVAT
				q.InvoiceNumber = "FV/2026/001"
			},
			oczekBlad: true,
			fragment:  "data_sprzedazy",
		},
		{
			nazwa: "faktura_vat z nieprawidłową stawką VAT",
			mutator: func(q *Quote) {
				q.DocType = DocTypeFakturaVAT
				q.InvoiceNumber = "FV/2026/001"
				q.SaleDate = "2026-06-18"
				q.Items[0].VatRate = 7 // nieprawidłowa
			},
			oczekBlad: true,
			fragment:  "stawka_vat",
		},
		{
			nazwa: "faktura_proforma bez numeru",
			mutator: func(q *Quote) {
				q.DocType = DocTypeProforma
				q.SaleDate = "2026-06-18"
			},
			oczekBlad: true,
			fragment:  "numer_faktury",
		},
		{
			nazwa: "oferta bez pól faktury — nie wymaga VAT ani numeru",
			mutator: func(q *Quote) {
				q.DocType = ""
				q.Items[0].VatRate = 0
			},
			oczekBlad: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			q := base
			q.Items = []LineItem{{Name: base.Items[0].Name, Quantity: base.Items[0].Quantity, UnitPrice: base.Items[0].UnitPrice, VatRate: base.Items[0].VatRate}}
			tc.mutator(&q)
			err := q.Validate()
			if tc.oczekBlad && err == nil {
				t.Fatalf("oczekiwano błędu, brak błędu")
			}
			if !tc.oczekBlad && err != nil {
				t.Fatalf("nieoczekiwany błąd: %v", err)
			}
			if tc.oczekBlad && tc.fragment != "" && err != nil && !strings.Contains(err.Error(), tc.fragment) {
				t.Errorf("błąd %q nie zawiera %q", err.Error(), tc.fragment)
			}
		})
	}
}

// ─── Testy BuildKSeFXML ───────────────────────────────────────────────────────

// validFakturaVAT zwraca Quote, który przechodzi wszystkie walidacje KSeF (Validate + validateKSeFData).
// NIP sprzedawcy "1111111111" i nabywcy "9876543210" mają poprawną sumę kontrolną.
func validFakturaVAT() Quote {
	return Quote{
		Company: Company{
			Name:    "ACME sp. z o.o.",
			NIP:     "1111111111", // poprawna suma kontrolna
			Address: "ul. Testowa 1",
			City:    "00-001 Warszawa",
		},
		Client:        "Firma Klient Sp. z o.o.\nul. Klienta 5, Kraków\nNIP: 9876543210",
		DocType:       DocTypeFakturaVAT,
		InvoiceNumber: "FV/2026/001",
		SaleDate:      "2026-06-18",
		PaymentDue:    "2026-07-02",
		Items: []LineItem{
			{Name: "Montaż instalacji", Quantity: 1, UnitPrice: 1000, VatRate: 23},
			{Name: "Materiały", Quantity: 2, UnitPrice: 250, VatRate: 8},
		},
	}
}

func TestBuildKSeFXML_PoprawnaStruktura(t *testing.T) {
	q := validFakturaVAT()
	data, err := BuildKSeFXML(q)
	if err != nil {
		t.Fatalf("BuildKSeFXML: %v", err)
	}

	xmlStr := string(data)

	t.Run("nagłówek XML", func(t *testing.T) {
		if !strings.HasPrefix(xmlStr, "<?xml version=") {
			t.Error("brak nagłówka XML")
		}
	})

	t.Run("namespace FA(3)", func(t *testing.T) {
		if !strings.Contains(xmlStr, ksefNS) {
			t.Errorf("brak namespace FA(3) %q", ksefNS)
		}
	})

	t.Run("KodFormularza", func(t *testing.T) {
		if !strings.Contains(xmlStr, `kodSystemowy="FA (3)"`) {
			t.Error("brak atrybutu kodSystemowy")
		}
		if !strings.Contains(xmlStr, `wersjaSchemy="1-0E"`) {
			t.Error("brak atrybutu wersjaSchemy")
		}
	})

	t.Run("NIP sprzedawcy", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<NIP>1111111111</NIP>") {
			t.Error("brak NIP sprzedawcy")
		}
	})

	t.Run("NIP nabywcy", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<NIP>9876543210</NIP>") {
			t.Error("brak NIP nabywcy (powinien być wykryty z linii 'NIP: 9876543210')")
		}
	})

	t.Run("numer faktury", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_2>FV/2026/001</P_2>") {
			t.Error("brak numeru faktury P_2")
		}
	})

	t.Run("data sprzedaży", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_6>2026-06-18</P_6>") {
			t.Error("brak daty sprzedaży P_6")
		}
	})

	t.Run("poprawny XML", func(t *testing.T) {
		if err := xml.Unmarshal(data, new(interface{})); err != nil {
			t.Errorf("XML nie parsuje się poprawnie: %v", err)
		}
	})

	t.Run("Adnotacje wymagane", func(t *testing.T) {
		for _, pole := range []string{"<P_16>2</P_16>", "<P_17>2</P_17>", "<P_18>2</P_18>", "<P_18A>2</P_18A>",
			"<P_19N>1</P_19N>", "<P_22N>1</P_22N>", "<P_23>2</P_23>", "<P_PMarzyN>1</P_PMarzyN>"} {
			if !strings.Contains(xmlStr, pole) {
				t.Errorf("brak wymaganego pola adnotacji: %s", pole)
			}
		}
	})

	t.Run("JST i GV = 2", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<JST>2</JST>") {
			t.Error("brak JST=2 w Podmiot2")
		}
		if !strings.Contains(xmlStr, "<GV>2</GV>") {
			t.Error("brak GV=2 w Podmiot2")
		}
	})
}

func TestBuildKSeFXML_MieszaneStawkiVAT(t *testing.T) {
	q := Quote{
		Company:       Company{Name: "Firma", NIP: "1234567890"},
		Client:        "Klient",
		DocType:       DocTypeFakturaVAT,
		InvoiceNumber: "FV/2026/001",
		SaleDate:      "2026-06-18",
		Items: []LineItem{
			{Name: "A", Quantity: 1, UnitPrice: 1000, VatRate: 23},
			{Name: "B", Quantity: 1, UnitPrice: 500, VatRate: 8},
			{Name: "C", Quantity: 1, UnitPrice: 200, VatRate: 5},
			{Name: "D", Quantity: 1, UnitPrice: 100, VatRate: 0},
		},
	}

	data, err := BuildKSeFXML(q)
	if err != nil {
		t.Fatalf("BuildKSeFXML: %v", err)
	}
	xmlStr := string(data)

	t.Run("stawka 23%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_13_1>1000.00</P_13_1>") {
			t.Error("błędne P_13_1 (netto 23%)")
		}
		if !strings.Contains(xmlStr, "<P_14_1>230.00</P_14_1>") {
			t.Error("błędne P_14_1 (VAT 23%)")
		}
	})

	t.Run("stawka 8%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_13_2>500.00</P_13_2>") {
			t.Error("błędne P_13_2 (netto 8%)")
		}
		if !strings.Contains(xmlStr, "<P_14_2>40.00</P_14_2>") {
			t.Error("błędne P_14_2 (VAT 8%)")
		}
	})

	t.Run("stawka 5%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_13_3>200.00</P_13_3>") {
			t.Error("błędne P_13_3 (netto 5%)")
		}
		if !strings.Contains(xmlStr, "<P_14_3>10.00</P_14_3>") {
			t.Error("błędne P_14_3 (VAT 5%)")
		}
	})

	t.Run("stawka 0%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_13_6_1>100.00</P_13_6_1>") {
			t.Error("błędne P_13_6_1 (netto 0%)")
		}
		if strings.Contains(xmlStr, "<P_14_6") {
			t.Error("dla stawki 0% nie powinno być pola P_14_x")
		}
	})

	t.Run("P_15 brutto ogółem", func(t *testing.T) {
		// 1000+230 + 500+40 + 200+10 + 100+0 = 2080
		if !strings.Contains(xmlStr, "<P_15>2080.00</P_15>") {
			t.Errorf("błędne P_15 — oczekiwano 2080.00; XML: %.200s", xmlStr)
		}
	})

	t.Run("P_12 enum w wierszach", func(t *testing.T) {
		for _, stawka := range []string{"23", "8", "5", "0 KR"} {
			if !strings.Contains(xmlStr, "<P_12>"+stawka+"</P_12>") {
				t.Errorf("brak P_12=%s w wierszach", stawka)
			}
		}
	})

	t.Run("LiczbaWierszy", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<LiczbaWierszy>4</LiczbaWierszy>") {
			t.Error("błędna LiczbaWierszy")
		}
	})
}

func TestBuildKSeFXML_BrakNIPNabywcy(t *testing.T) {
	q := validFakturaVAT()
	q.Client = "Jan Kowalski\nul. Testowa 1\nWarszawa"

	data, err := BuildKSeFXML(q)
	if err != nil {
		t.Fatalf("BuildKSeFXML: %v", err)
	}
	xmlStr := string(data)

	if !strings.Contains(xmlStr, "<BrakID>1</BrakID>") {
		t.Error("brak nabywcy bez NIP powinien generować BrakID=1")
	}
	if strings.Contains(xmlStr, "<NIP>9876543210</NIP>") {
		t.Error("klient bez NIP nie powinien zawierać pola NIP w Podmiot2")
	}
	if !strings.Contains(xmlStr, "<Nazwa>Jan Kowalski</Nazwa>") {
		t.Error("brak Nazwa nabywcy")
	}
}

func TestBuildKSeFXML_JednaStawka(t *testing.T) {
	q := Quote{
		Company:       Company{Name: "Firma", NIP: "1111111111"},
		Client:        "Klient",
		DocType:       DocTypeFakturaVAT,
		InvoiceNumber: "FV/1",
		SaleDate:      "2026-06-18",
		Items: []LineItem{
			{Name: "Usługa A", Quantity: 3, UnitPrice: 100, VatRate: 23},
			{Name: "Usługa B", Quantity: 1, UnitPrice: 50, VatRate: 23},
		},
	}

	data, err := BuildKSeFXML(q)
	if err != nil {
		t.Fatalf("BuildKSeFXML: %v", err)
	}
	xmlStr := string(data)

	// netto: 3*100 + 1*50 = 350; VAT 23%: 80.50; brutto: 430.50
	t.Run("suma netto 23%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_13_1>350.00</P_13_1>") {
			t.Error("błędne P_13_1")
		}
	})
	t.Run("suma VAT 23%", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_14_1>80.50</P_14_1>") {
			t.Error("błędne P_14_1")
		}
	})
	t.Run("P_15 brutto", func(t *testing.T) {
		if !strings.Contains(xmlStr, "<P_15>430.50</P_15>") {
			t.Error("błędne P_15")
		}
	})
	t.Run("brak pól innych stawek", func(t *testing.T) {
		for _, f := range []string{"P_13_2", "P_13_3", "P_13_6_1", "P_13_7"} {
			if strings.Contains(xmlStr, "<"+f+">") {
				t.Errorf("nieoczekiwane pole <%s> — tylko stawka 23%%", f)
			}
		}
	})
}

func TestHandleXML_BrakNIPSprzedawcy(t *testing.T) {
	q := validFakturaVAT()
	q.Company.NIP = ""

	body, _ := json.Marshal(q)
	req := httptest.NewRequest(http.MethodPost, "/api/xml", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleXML(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("oczekiwano 400, otrzymano %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "NIP sprzedawcy") {
		t.Error("brak komunikatu o NIP sprzedawcy")
	}
}

func TestHandleXML_NieFakturaVAT(t *testing.T) {
	q := validFakturaVAT()
	q.DocType = DocTypeOferta

	body, _ := json.Marshal(q)
	req := httptest.NewRequest(http.MethodPost, "/api/xml", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleXML(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("oczekiwano 400, otrzymano %d", rec.Code)
	}
}

func TestHandleXML_PoprawnaFaktura(t *testing.T) {
	q := validFakturaVAT()
	body, _ := json.Marshal(q)
	req := httptest.NewRequest(http.MethodPost, "/api/xml", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleXML(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("oczekiwano 200, otrzymano %d; body: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/xml") {
		t.Errorf("Content-Type = %q, oczekiwano application/xml", ct)
	}
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, ".xml") {
		t.Errorf("Content-Disposition = %q, oczekiwano .xml", cd)
	}
	if !strings.Contains(rec.Body.String(), ksefNS) {
		t.Error("odpowiedź nie zawiera namespace FA(3)")
	}
}

// ─── Testy ValidateNIP ────────────────────────────────────────────────────────

func TestValidateNIP(t *testing.T) {
	cases := []struct {
		nip       string
		wantError bool
		fragment  string
	}{
		// Poprawne NIPs (suma kontrolna OK)
		{"1111111111", false, ""},  // suma 45, 45%11=1
		{"2222222222", false, ""},  // suma 90, 90%11=2
		{"9876543210", false, ""},  // suma 220, 220%11=0
		{"", false, ""},            // pusty → nil (pole opcjonalne)

		// Poprawne z prefiksem PL lub separatorami
		{"PL1111111111", false, ""},
		{"111-111-11-11", false, ""},
		{"111 111 11 11", false, ""},

		// Nieprawidłowe: błędna suma kontrolna
		{"1234567890", true, "kontrolną"}, // suma 230, 230%11=10, ostatnia=0
		{"0000000001", true, "kontrolną"}, // suma 0, 0%11=0, ostatnia=1

		// Nieprawidłowe: zła długość
		{"123456789", true, "10 cyfr"},   // 9 cyfr
		{"12345678901", true, "10 cyfr"}, // 11 cyfr
		{"abcdefghij", true, "10 cyfr"},  // litery
	}

	for _, tc := range cases {
		t.Run(tc.nip, func(t *testing.T) {
			err := ValidateNIP(tc.nip)
			if tc.wantError && err == nil {
				t.Fatalf("oczekiwano błędu dla NIP %q, otrzymano nil", tc.nip)
			}
			if !tc.wantError && err != nil {
				t.Fatalf("nie oczekiwano błędu dla NIP %q, otrzymano: %v", tc.nip, err)
			}
			if tc.wantError && tc.fragment != "" && err != nil && !strings.Contains(err.Error(), tc.fragment) {
				t.Errorf("komunikat %q nie zawiera fragmentu %q", err.Error(), tc.fragment)
			}
		})
	}
}

// ─── Testy validateKSeFData ───────────────────────────────────────────────────

func TestValidateKSeFData(t *testing.T) {
	cases := []struct {
		nazwa     string
		mutator   func(q *Quote)
		wantError bool
		fragment  string
	}{
		{
			nazwa:     "poprawne dane",
			mutator:   func(q *Quote) {},
			wantError: false,
		},
		{
			nazwa:     "brak NIP sprzedawcy",
			mutator:   func(q *Quote) { q.Company.NIP = "" },
			wantError: true,
			fragment:  "NIP sprzedawcy",
		},
		{
			nazwa:     "błędna suma kontrolna NIP sprzedawcy",
			mutator:   func(q *Quote) { q.Company.NIP = "1234567890" },
			wantError: true,
			fragment:  "NIP sprzedawcy",
		},
		{
			nazwa:     "brak adresu i miasta sprzedawcy",
			mutator:   func(q *Quote) { q.Company.Address = ""; q.Company.City = "" },
			wantError: true,
			fragment:  "adres sprzedawcy",
		},
		{
			nazwa:     "tylko miasto sprzedawcy — akceptowalne",
			mutator:   func(q *Quote) { q.Company.Address = "" },
			wantError: false,
		},
		{
			nazwa:     "NIP nabywcy z błędną sumą kontrolną",
			mutator:   func(q *Quote) { q.Client = "Firma\n1234567890" },
			wantError: true,
			fragment:  "NIP nabywcy",
		},
		{
			nazwa:     "nabywca bez NIP (B2C) — dozwolone",
			mutator:   func(q *Quote) { q.Client = "Jan Kowalski\nWarszawa" },
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			q := validFakturaVAT()
			tc.mutator(&q)
			err := validateKSeFData(q)
			if tc.wantError && err == nil {
				t.Fatalf("oczekiwano błędu, otrzymano nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("nieoczekiwany błąd: %v", err)
			}
			if tc.wantError && tc.fragment != "" && err != nil && !strings.Contains(err.Error(), tc.fragment) {
				t.Errorf("komunikat %q nie zawiera fragmentu %q", err.Error(), tc.fragment)
			}
		})
	}
}

// ─── Test sekcji Platnosc ─────────────────────────────────────────────────────

func TestBuildKSeFXML_Platnosc(t *testing.T) {
	t.Run("termin płatności jako data ISO", func(t *testing.T) {
		q := validFakturaVAT()
		q.PaymentDue = "2026-07-18"
		data, err := BuildKSeFXML(q)
		if err != nil {
			t.Fatalf("BuildKSeFXML: %v", err)
		}
		xmlStr := string(data)
		if !strings.Contains(xmlStr, "<Platnosc>") {
			t.Error("brak sekcji <Platnosc>")
		}
		if !strings.Contains(xmlStr, "<Termin>2026-07-18</Termin>") {
			t.Errorf("brak <Termin>2026-07-18</Termin>; fragment: %.300s", xmlStr)
		}
	})

	t.Run("brak terminu płatności — brak sekcji Platnosc", func(t *testing.T) {
		q := validFakturaVAT()
		q.PaymentDue = ""
		data, err := BuildKSeFXML(q)
		if err != nil {
			t.Fatalf("BuildKSeFXML: %v", err)
		}
		if strings.Contains(string(data), "<Platnosc>") {
			t.Error("nie powinno być sekcji <Platnosc> gdy brak terminu")
		}
	})

	t.Run("termin nieformatowy (słowny) — brak sekcji Platnosc", func(t *testing.T) {
		q := validFakturaVAT()
		q.PaymentDue = "14 dni"
		data, err := BuildKSeFXML(q)
		if err != nil {
			t.Fatalf("BuildKSeFXML: %v", err)
		}
		if strings.Contains(string(data), "<Platnosc>") {
			t.Error("termin słowny nie powinien generować sekcji <Platnosc>")
		}
	})
}

// ─── Test P_18A (mechanizm podzielonej płatności) ─────────────────────────────

func TestBuildKSeFXML_P18A(t *testing.T) {
	t.Run("brutto <= 15000 → P_18A=2", func(t *testing.T) {
		q := validFakturaVAT()
		// Montaż 1000 netto + VAT 230 + Materiały 500 netto + VAT 40 = 1770 brutto
		data, err := BuildKSeFXML(q)
		if err != nil {
			t.Fatalf("BuildKSeFXML: %v", err)
		}
		if !strings.Contains(string(data), "<P_18A>2</P_18A>") {
			t.Error("dla brutto ≤ 15 000 zł oczekiwano P_18A=2 (nie dotyczy MPP)")
		}
	})

	t.Run("brutto > 15000 → P_18A=1", func(t *testing.T) {
		q := validFakturaVAT()
		q.Items = []LineItem{
			{Name: "Duża usługa", Quantity: 1, UnitPrice: 15000, VatRate: 23},
		}
		data, err := BuildKSeFXML(q)
		if err != nil {
			t.Fatalf("BuildKSeFXML: %v", err)
		}
		if !strings.Contains(string(data), "<P_18A>1</P_18A>") {
			t.Error("dla brutto > 15 000 zł oczekiwano P_18A=1 (sygnalizacja MPP)")
		}
	})
}

// ─── Test handleXML z nieprawidłowym NIP (suma kontrolna) ─────────────────────

func TestHandleXML_NieprawidlowyNIPSprzedawcy(t *testing.T) {
	q := validFakturaVAT()
	q.Company.NIP = "1234567890" // nieprawidłowa suma kontrolna

	body, _ := json.Marshal(q)
	req := httptest.NewRequest(http.MethodPost, "/api/xml", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleXML(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("oczekiwano 400, otrzymano %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "NIP sprzedawcy") {
		t.Errorf("brak komunikatu o NIP sprzedawcy w: %s", rec.Body.String())
	}
}

// ─── Test XSD (Krok 2) ───────────────────────────────────────────────────────
//
// Walidacja wygenerowanego XML względem oficjalnego schematu FA(3) przez xmllint.
// Test pomijany gdy xmllint nie jest zainstalowany (lokalnie Windows) —
// aktywny na serwerze CI Linux (apt install libxml2-utils).

func TestBuildKSeFXML_XSDValidation(t *testing.T) {
	if _, err := exec.LookPath("xmllint"); err != nil {
		t.Skip("xmllint niedostępny — pomiń walidację XSD (zainstaluj: apt install libxml2-utils)")
	}

	// Znajdź katalog z plikami XSD względem katalogu roboczego testu
	xsdPath := filepath.Join("assets", "ksef", "FA3.xsd")
	if _, err := os.Stat(xsdPath); err != nil {
		t.Skipf("brak pliku %s — pomiń walidację XSD", xsdPath)
	}

	q := validFakturaVAT()
	data, err := BuildKSeFXML(q)
	if err != nil {
		t.Fatalf("BuildKSeFXML: %v", err)
	}

	// Zapisz XML do pliku tymczasowego
	tmp, err := os.CreateTemp("", "ksef-*.xml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("Write temp: %v", err)
	}
	tmp.Close()

	// Uruchom xmllint --schema FA3.xsd --noout plik.xml
	cmd := exec.Command("xmllint", "--schema", xsdPath, "--noout", tmp.Name())
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Errorf("xmllint: walidacja XSD nie powiodła się:\n%s", out)
	}
}

// ─── Istniejące testy ─────────────────────────────────────────────────────────

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
