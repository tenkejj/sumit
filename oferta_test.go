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

// validnaOferta zwraca minimalną Ofertę, która przechodzi Waliduj() bez błędu.
// Pojedyncze testy mutują kopię i sprawdzają wybrane reguły walidacji.
func validnaOferta() Oferta {
	return Oferta{
		Firma: FirmaDane{
			Nazwa: "ACME sp. z o.o.",
			NIP:   "1234567890",
		},
		Klient: "Jan Kowalski",
		Pozycje: []Pozycja{
			{Nazwa: "Usługa montażowa", Ilosc: 1, CenaJednostkowa: 100},
		},
	}
}

func TestOferta_Waliduj(t *testing.T) {
	cases := []struct {
		nazwa       string
		mutator     func(o *Oferta)
		oczekBlad   bool
		fragmentBl  string // opcjonalny fragment, który musi znaleźć się w komunikacie błędu
	}{
		{
			nazwa:     "poprawne dane",
			mutator:   func(o *Oferta) {},
			oczekBlad: false,
		},
		{
			nazwa:      "pusta nazwa firmy",
			mutator:    func(o *Oferta) { o.Firma.Nazwa = "" },
			oczekBlad:  true,
			fragmentBl: "nazwa_firmy",
		},
		{
			nazwa:      "nazwa firmy z samych białych znaków",
			mutator:    func(o *Oferta) { o.Firma.Nazwa = "   \t " },
			oczekBlad:  true,
			fragmentBl: "nazwa_firmy",
		},
		{
			nazwa:      "pusty klient",
			mutator:    func(o *Oferta) { o.Klient = "" },
			oczekBlad:  true,
			fragmentBl: "klient",
		},
		{
			nazwa:      "klient z samych białych znaków",
			mutator:    func(o *Oferta) { o.Klient = "   " },
			oczekBlad:  true,
			fragmentBl: "klient",
		},
		{
			nazwa:      "pusta lista pozycji (nil)",
			mutator:    func(o *Oferta) { o.Pozycje = nil },
			oczekBlad:  true,
			fragmentBl: "pozycji",
		},
		{
			nazwa:      "pusta lista pozycji (slice zerowy)",
			mutator:    func(o *Oferta) { o.Pozycje = []Pozycja{} },
			oczekBlad:  true,
			fragmentBl: "pozycji",
		},
		{
			nazwa:      "ilość równa zero",
			mutator:    func(o *Oferta) { o.Pozycje[0].Ilosc = 0 },
			oczekBlad:  true,
			fragmentBl: "ilość",
		},
		{
			nazwa:      "ilość ujemna",
			mutator:    func(o *Oferta) { o.Pozycje[0].Ilosc = -1 },
			oczekBlad:  true,
			fragmentBl: "ilość",
		},
		{
			nazwa:      "cena jednostkowa ujemna",
			mutator:    func(o *Oferta) { o.Pozycje[0].CenaJednostkowa = -0.01 },
			oczekBlad:  true,
			fragmentBl: "cena_jednostkowa",
		},
		{
			nazwa:     "cena jednostkowa równa zero jest dozwolona",
			mutator:   func(o *Oferta) { o.Pozycje[0].CenaJednostkowa = 0 },
			oczekBlad: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			o := validnaOferta()
			tc.mutator(&o)
			err := o.Waliduj()

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

func TestOferta_UnmarshalJSON_PlaskiJSON(t *testing.T) {
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

	var o Oferta
	if err := json.Unmarshal(plaski, &o); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	t.Run("pola firmy", func(t *testing.T) {
		cases := []struct {
			pole, got, want string
		}{
			{"Nazwa", o.Firma.Nazwa, "ACME sp. z o.o."},
			{"NIP", o.Firma.NIP, "1234567890"},
			{"Adres", o.Firma.Adres, "ul. Przykładowa 1"},
			{"Miasto", o.Firma.Miasto, "00-001 Warszawa"},
			{"Telefon", o.Firma.Telefon, "+48 600 700 800"},
			{"Email", o.Firma.Email, "biuro@acme.pl"},
			{"LogoBase64", o.Firma.LogoBase64, "data:image/png;base64,AAAA"},
		}
		for _, c := range cases {
			if c.got != c.want {
				t.Errorf("Firma.%s = %q, want %q", c.pole, c.got, c.want)
			}
		}
	})

	t.Run("pola oferty", func(t *testing.T) {
		if o.Klient != "Jan Kowalski" {
			t.Errorf("Klient = %q", o.Klient)
		}
		if o.NumerOferty != "2026/06/001" {
			t.Errorf("NumerOferty = %q", o.NumerOferty)
		}
		if o.DataWaznosci != "2026-07-12" {
			t.Errorf("DataWaznosci = %q", o.DataWaznosci)
		}
		if o.Uwagi != "Płatność 14 dni" {
			t.Errorf("Uwagi = %q", o.Uwagi)
		}
	})

	t.Run("pozycje", func(t *testing.T) {
		if len(o.Pozycje) != 2 {
			t.Fatalf("len(Pozycje) = %d, want 2", len(o.Pozycje))
		}
		if o.Pozycje[0].Nazwa != "Usługa A" || o.Pozycje[0].Ilosc != 2 || o.Pozycje[0].CenaJednostkowa != 150.5 {
			t.Errorf("Pozycje[0] = %+v", o.Pozycje[0])
		}
		if o.Pozycje[1].Nazwa != "Usługa B" || o.Pozycje[1].Ilosc != 1 || o.Pozycje[1].CenaJednostkowa != 49.5 {
			t.Errorf("Pozycje[1] = %+v", o.Pozycje[1])
		}
	})
}

func TestOferta_UnmarshalJSON_BrakOpcjonalnychPol(t *testing.T) {
	// Tylko pola wymagane przez Waliduj(): nazwa_firmy, klient, pozycje.
	// Pozostałe (logo, email, telefon, NIP, adres, numer, data, uwagi) są
	// opcjonalne — Unmarshal nie może panikować i ma zostawić zero values.
	minimalny := []byte(`{
		"nazwa_firmy": "ACME",
		"klient": "Jan Kowalski",
		"pozycje": [{"nazwa": "X", "ilosc": 1, "cena_jednostkowa": 10}]
	}`)

	var o Oferta
	if err := json.Unmarshal(minimalny, &o); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if o.Firma.Nazwa != "ACME" {
		t.Errorf("Firma.Nazwa = %q", o.Firma.Nazwa)
	}
	if o.Klient != "Jan Kowalski" {
		t.Errorf("Klient = %q", o.Klient)
	}

	t.Run("opcjonalne pola firmy są puste", func(t *testing.T) {
		cases := []struct {
			pole, got string
		}{
			{"NIP", o.Firma.NIP},
			{"Adres", o.Firma.Adres},
			{"Miasto", o.Firma.Miasto},
			{"Telefon", o.Firma.Telefon},
			{"Email", o.Firma.Email},
			{"LogoBase64", o.Firma.LogoBase64},
		}
		for _, c := range cases {
			if c.got != "" {
				t.Errorf("Firma.%s = %q, oczekiwano \"\"", c.pole, c.got)
			}
		}
	})

	t.Run("opcjonalne pola oferty są puste", func(t *testing.T) {
		if o.NumerOferty != "" {
			t.Errorf("NumerOferty = %q", o.NumerOferty)
		}
		if o.DataWaznosci != "" {
			t.Errorf("DataWaznosci = %q", o.DataWaznosci)
		}
		if o.Uwagi != "" {
			t.Errorf("Uwagi = %q", o.Uwagi)
		}
	})

	if len(o.Pozycje) != 1 {
		t.Errorf("len(Pozycje) = %d, want 1", len(o.Pozycje))
	}

	// Brakująca lista pozycji w ogóle — Unmarshal też nie może panikować.
	bezPozycji := []byte(`{"nazwa_firmy": "ACME", "klient": "Jan"}`)
	var o2 Oferta
	if err := json.Unmarshal(bezPozycji, &o2); err != nil {
		t.Fatalf("Unmarshal bezPozycji: %v", err)
	}
	if o2.Pozycje != nil {
		t.Errorf("Pozycje = %v, oczekiwano nil", o2.Pozycje)
	}
}

func TestOferta_MarshalUnmarshal_Roundtrip(t *testing.T) {
	// MarshalJSON musi emitować JSON PŁASKI (pola firmy obok klient/pozycji),
	// a UnmarshalJSON musi go odczytać z powrotem do zagnieżdżonej struktury.
	orig := validnaOferta()
	orig.Firma.Email = "biuro@acme.pl"
	orig.NumerOferty = "2026/06/001"

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

	var roundtrip Oferta
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("Unmarshal roundtrip: %v", err)
	}
	if roundtrip.Firma.Nazwa != orig.Firma.Nazwa {
		t.Errorf("Firma.Nazwa po roundtrip = %q, want %q", roundtrip.Firma.Nazwa, orig.Firma.Nazwa)
	}
	if roundtrip.Firma.Email != orig.Firma.Email {
		t.Errorf("Firma.Email po roundtrip = %q, want %q", roundtrip.Firma.Email, orig.Firma.Email)
	}
	if roundtrip.Klient != orig.Klient {
		t.Errorf("Klient po roundtrip = %q, want %q", roundtrip.Klient, orig.Klient)
	}
	if roundtrip.NumerOferty != orig.NumerOferty {
		t.Errorf("NumerOferty po roundtrip = %q, want %q", roundtrip.NumerOferty, orig.NumerOferty)
	}
	if len(roundtrip.Pozycje) != len(orig.Pozycje) {
		t.Fatalf("len(Pozycje) po roundtrip = %d, want %d", len(roundtrip.Pozycje), len(orig.Pozycje))
	}
}

func TestPozycja_Wartosc(t *testing.T) {
	cases := []struct {
		nazwa string
		p     Pozycja
		oczek float64
	}{
		{"liczby całkowite", Pozycja{Ilosc: 5, CenaJednostkowa: 20}, 100},
		{"grosze 3 x 33.33", Pozycja{Ilosc: 3, CenaJednostkowa: 33.33}, 99.99},
		{"zero ilości", Pozycja{Ilosc: 0, CenaJednostkowa: 100}, 0},
		{"zero ceny", Pozycja{Ilosc: 10, CenaJednostkowa: 0}, 0},
		{"ułamkowa ilość", Pozycja{Ilosc: 2.5, CenaJednostkowa: 4}, 10},
	}
	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			got := tc.p.Wartosc()
			if math.Abs(got-tc.oczek) > 0.005 {
				t.Errorf("Wartosc() = %v, want %v (różnica %v)", got, tc.oczek, math.Abs(got-tc.oczek))
			}
		})
	}
}

func TestOferta_Suma(t *testing.T) {
	cases := []struct {
		nazwa   string
		pozycje []Pozycja
		oczek   float64
	}{
		{
			nazwa:   "brak pozycji",
			pozycje: nil,
			oczek:   0,
		},
		{
			nazwa: "jedna pozycja",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 3, CenaJednostkowa: 100},
			},
			oczek: 300,
		},
		{
			nazwa: "wiele pozycji bez ułamków",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 2, CenaJednostkowa: 50},
				{Nazwa: "B", Ilosc: 1, CenaJednostkowa: 150},
				{Nazwa: "C", Ilosc: 4, CenaJednostkowa: 25},
			},
			oczek: 350,
		},
		{
			nazwa: "grosze - 3 x 33.33",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 3, CenaJednostkowa: 33.33},
			},
			oczek: 99.99,
		},
		{
			nazwa: "klasyczny float 0.1 + 0.2",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 1, CenaJednostkowa: 0.1},
				{Nazwa: "B", Ilosc: 1, CenaJednostkowa: 0.2},
			},
			oczek: 0.30,
		},
		{
			nazwa: "wiele pozycji z groszami",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 7, CenaJednostkowa: 12.34},
				{Nazwa: "B", Ilosc: 3, CenaJednostkowa: 0.07},
				{Nazwa: "C", Ilosc: 1, CenaJednostkowa: 999.99},
			},
			oczek: 7*12.34 + 3*0.07 + 999.99,
		},
		{
			nazwa: "ułamkowa ilość",
			pozycje: []Pozycja{
				{Nazwa: "A", Ilosc: 2.5, CenaJednostkowa: 80},
				{Nazwa: "B", Ilosc: 0.5, CenaJednostkowa: 100},
			},
			oczek: 250,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nazwa, func(t *testing.T) {
			o := Oferta{Pozycje: tc.pozycje}
			got := o.Suma()

			// Tolerancja pół grosza — pokrywa szum arytmetyki zmiennoprzecinkowej,
			// a jednocześnie gwarantuje, że wartość zaokrąglona do groszy jest poprawna.
			if math.Abs(got-tc.oczek) > 0.005 {
				t.Errorf("Suma() = %v, want %v (różnica %v)", got, tc.oczek, math.Abs(got-tc.oczek))
			}

			// Sprawdzamy też prezentację — to, co użytkownik widzi w PDF jako „Razem".
			if g, w := formatPLN(got), formatPLN(tc.oczek); g != w {
				t.Errorf("formatPLN(Suma()) = %q, want %q", g, w)
			}
		})
	}
}

func TestHandleOferta_BodyPowyzej1MB(t *testing.T) {
	// handleOferta opakowuje body w http.MaxBytesReader(w, r.Body, 1<<20).
	// Body większy niż 1 MB musi skutkować odpowiedzią 400 Bad Request,
	// zanim cokolwiek dotrze do generatora PDF.
	const limit = 1 << 20
	wypelniacz := strings.Repeat("a", limit+1024)

	o := Oferta{
		Firma:  FirmaDane{Nazwa: "ACME"},
		Klient: "Jan Kowalski",
		Uwagi:  wypelniacz,
		Pozycje: []Pozycja{
			{Nazwa: "Usługa A", Ilosc: 1, CenaJednostkowa: 100},
		},
	}
	body, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(body) <= limit {
		t.Fatalf("przygotowany body ma %d B, oczekiwano > %d B", len(body), limit)
	}

	req := httptest.NewRequest(http.MethodPost, "/oferta", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handleOferta(rec, req)

	if rec.Code != http.StatusBadRequest {
		respBody, _ := io.ReadAll(rec.Body)
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, respBody)
	}
}
