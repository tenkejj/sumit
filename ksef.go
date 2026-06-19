package main

// UWAGA PRAWNA: Plik XML generowany przez SumIt jest dokumentem pomocniczym
// dla użytkownika. SumIt nie wysyła faktur do KSeF automatycznie i nie ponosi
// odpowiedzialności za poprawność ani kompletność danych. Obowiązek weryfikacji
// zgodności z przepisami ustawy o VAT oraz wysyłkę do systemu KSeF
// (https://ksef.podatki.gov.pl) ponosi wyłącznie użytkownik.

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ksefNS = "http://crd.gov.pl/wzor/2025/06/25/13775/"

// nipRe dopasowuje dokładnie 10 kolejnych cyfr.
var nipRe = regexp.MustCompile(`\b(\d{10})\b`)

// ─── Struktury XML FA(3) ──────────────────────────────────────────────────────

type ksefFaktura struct {
	XMLName  xml.Name     `xml:"Faktura"`
	Xmlns    string       `xml:"xmlns,attr"`
	Naglowek ksefNaglowek `xml:"Naglowek"`
	Podmiot1 ksefPodmiot1 `xml:"Podmiot1"`
	Podmiot2 ksefPodmiot2 `xml:"Podmiot2"`
	Fa       ksefFa       `xml:"Fa"`
}

type ksefNaglowek struct {
	KodFormularza     ksefKodFormularza `xml:"KodFormularza"`
	WariantFormularza int               `xml:"WariantFormularza"`
	DataWytworzeniaFa string            `xml:"DataWytworzeniaFa"`
	SystemInfo        string            `xml:"SystemInfo,omitempty"`
}

type ksefKodFormularza struct {
	KodSystemowy string `xml:"kodSystemowy,attr"`
	WersjaSchemy string `xml:"wersjaSchemy,attr"`
	Wartosc      string `xml:",chardata"`
}

// ─── Podmiot1 (sprzedawca) ────────────────────────────────────────────────────

type ksefPodmiot1DaneId struct {
	NIP   string `xml:"NIP"`
	Nazwa string `xml:"Nazwa"`
}

type ksefAdres struct {
	KodKraju string `xml:"KodKraju"`
	AdresL1  string `xml:"AdresL1"`
	AdresL2  string `xml:"AdresL2,omitempty"`
}

type ksefPodmiot1 struct {
	DaneIdentyfikacyjne ksefPodmiot1DaneId `xml:"DaneIdentyfikacyjne"`
	Adres               ksefAdres          `xml:"Adres"`
}

// ─── Podmiot2 (nabywca) ───────────────────────────────────────────────────────

type ksefPodmiot2DaneId struct {
	NIP    string `xml:"NIP,omitempty"`    // opcja A: polski podatnik VAT
	BrakID int    `xml:"BrakID,omitempty"` // opcja D = 1: konsument bez NIP (B2C)
	Nazwa  string `xml:"Nazwa"`
}

type ksefPodmiot2 struct {
	DaneIdentyfikacyjne ksefPodmiot2DaneId `xml:"DaneIdentyfikacyjne"`
	Adres               ksefAdres          `xml:"Adres"`
	JST                 int                `xml:"JST"` // 2 = nie dotyczy
	GV                  int                `xml:"GV"`  // 2 = nie dotyczy
}

// ─── Sekcja Fa ────────────────────────────────────────────────────────────────

type ksefAdnotacje struct {
	P_16                int                      `xml:"P_16"`
	P_17                int                      `xml:"P_17"`
	P_18                int                      `xml:"P_18"`
	P_18A               int                      `xml:"P_18A"`
	Zwolnienie          ksefZwolnienie           `xml:"Zwolnienie"`
	NoweSrodkiTransportu ksefNoweSrodkiTransportu `xml:"NoweSrodkiTransportu"`
	P_23                int                      `xml:"P_23"`
	PMarzy              ksefPMarzy               `xml:"PMarzy"`
}

type ksefZwolnienie struct {
	P_19N int `xml:"P_19N,omitempty"` // 1 = brak zwolnienia
}

type ksefNoweSrodkiTransportu struct {
	P_22N int `xml:"P_22N,omitempty"` // 1 = brak nowych środków transportu
}

type ksefPMarzy struct {
	P_PMarzyN int `xml:"P_PMarzyN,omitempty"` // 1 = brak procedury marży
}

type ksefFaWiersz struct {
	NrWierszaFa int    `xml:"NrWierszaFa"`
	P_7         string `xml:"P_7,omitempty"` // nazwa towaru/usługi
	P_8A        string `xml:"P_8A,omitempty"` // jednostka miary
	P_8B        string `xml:"P_8B"`           // ilość
	P_9A        string `xml:"P_9A"`           // cena jedn. netto
	P_11        string `xml:"P_11"`           // wartość netto
	P_12        string `xml:"P_12"`           // stawka VAT (enum TStawkaPodatku)
}

type ksefFaWiersze struct {
	LiczbaWierszy       int            `xml:"LiczbaWierszy"`
	WartoscWierszyNetto string         `xml:"WartoscWierszyNetto,omitempty"`
	FaWiersz            []ksefFaWiersz `xml:"FaWiersz"`
}

// ksefPlatnosc reprezentuje sekcję płatności FA(3).
// Wypełniamy ją tylko gdy termin płatności jest datą ISO (YYYY-MM-DD).
type ksefTerminPlatnosci struct {
	Termin string `xml:"Termin"` // data ISO YYYY-MM-DD
}

type ksefPlatnosc struct {
	TerminPlatnosci ksefTerminPlatnosci `xml:"TerminPlatnosci"`
}

// Kolejność pól ksefFa odzwierciedla sekwencję zdefiniowaną w XSD FA(3).
type ksefFa struct {
	KodWaluty string `xml:"KodWaluty"`
	P_1       string `xml:"P_1"`                // data wystawienia
	P_2       string `xml:"P_2"`                // numer faktury
	P_6       string `xml:"P_6,omitempty"`      // data dostawy/sprzedaży (pojedyncza)
	// sumy netto / VAT per stawka
	P_13_1   string `xml:"P_13_1,omitempty"`   // netto 23%
	P_14_1   string `xml:"P_14_1,omitempty"`   // VAT  23%
	P_13_2   string `xml:"P_13_2,omitempty"`   // netto 8%
	P_14_2   string `xml:"P_14_2,omitempty"`   // VAT  8%
	P_13_3   string `xml:"P_13_3,omitempty"`   // netto 5%
	P_14_3   string `xml:"P_14_3,omitempty"`   // VAT  5%
	P_13_6_1 string `xml:"P_13_6_1,omitempty"` // netto 0% sprzedaż krajowa
	P_13_7   string `xml:"P_13_7,omitempty"`   // zwolnione (ZW)
	P_15     string `xml:"P_15"`               // należność ogółem (brutto)
	Adnotacje     ksefAdnotacje  `xml:"Adnotacje"`
	RodzajFaktury string         `xml:"RodzajFaktury"`
	Platnosc      *ksefPlatnosc  `xml:"Platnosc,omitempty"` // opcjonalna; tylko gdy data ISO
	FaWiersze     ksefFaWiersze  `xml:"FaWiersze"`
}

// ─── Walidacja NIP ────────────────────────────────────────────────────────────

// nipDigitsRe pasuje do ciągu dokładnie 10 cyfr (po sanityzacji PL-prefix/spacji).
var nipDigitsRe = regexp.MustCompile(`^\d{10}$`)

// ValidateNIP sprawdza sumę kontrolną polskiego NIP (10 cyfr).
// Zwraca nil gdy NIP jest pusty (pole opcjonalne), lub błąd gdy forma/checksum nieprawidłowa.
func ValidateNIP(nip string) error {
	nip = strings.TrimSpace(nip)
	// Usuń opcjonalny prefix "PL" i nieznaczące separatory
	nip = strings.TrimPrefix(strings.ToUpper(nip), "PL")
	nip = strings.NewReplacer("-", "", " ", "").Replace(nip)
	if nip == "" {
		return nil
	}
	if !nipDigitsRe.MatchString(nip) {
		return fmt.Errorf("NIP musi składać się z dokładnie 10 cyfr (podano: %q)", nip)
	}
	weights := [9]int{6, 5, 7, 2, 3, 4, 5, 6, 7}
	sum := 0
	for i, w := range weights {
		sum += w * int(nip[i]-'0')
	}
	if sum%11 != int(nip[9]-'0') {
		return fmt.Errorf("NIP %s ma nieprawidłową cyfrę kontrolną", nip)
	}
	return nil
}

// validateKSeFData sprawdza wymagania specyficzne dla XML KSeF — ponad Validate().
// Wywołać przed BuildKSeFXML.
func validateKSeFData(q Quote) error {
	// NIP sprzedawcy: wymagany i musi mieć poprawną sumę kontrolną
	nipSprzedawcy := strings.TrimSpace(q.Company.NIP)
	if nipSprzedawcy == "" {
		return fmt.Errorf("NIP sprzedawcy jest wymagany do generowania XML KSeF — uzupełnij go w zakładce Moja firma")
	}
	if err := ValidateNIP(nipSprzedawcy); err != nil {
		return fmt.Errorf("NIP sprzedawcy: %w", err)
	}

	// Adres sprzedawcy: przynajmniej jedna niepusta linia
	if strings.TrimSpace(q.Company.Address) == "" && strings.TrimSpace(q.Company.City) == "" {
		return fmt.Errorf("adres sprzedawcy jest wymagany do generowania XML KSeF — uzupełnij go w zakładce Moja firma")
	}

	// NIP nabywcy: gdy wykryty w polu klient, musi mieć poprawną sumę kontrolną
	nipNabywcy := extractNIPFromClient(q.Client)
	if nipNabywcy != "" {
		if err := ValidateNIP(nipNabywcy); err != nil {
			return fmt.Errorf("NIP nabywcy: %w", err)
		}
	}

	return nil
}

// extractNIPFromClient wyciąga NIP z wieloliniowego pola klient (ten sam regex co buildPodmiot2).
func extractNIPFromClient(client string) string {
	for _, line := range strings.Split(strings.TrimSpace(client), "\n") {
		if m := nipRe.FindString(strings.TrimSpace(line)); m != "" {
			return m
		}
	}
	return ""
}

// isISODate sprawdza czy string jest datą w formacie YYYY-MM-DD (bez parsowania kalendarza).
var isoDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func isISODate(s string) bool {
	return isoDateRe.MatchString(strings.TrimSpace(s))
}

// ─── Budowanie XML ────────────────────────────────────────────────────────────

// BuildKSeFXML buduje plik XML w strukturze FA(3) na podstawie zatwierdzonej faktury VAT.
// Zwraca kompletny dokument z nagłówkiem <?xml ...?>.
// Wywołaj validateKSeFData przed BuildKSeFXML, aby mieć pewność co do danych wejściowych.
func BuildKSeFXML(q Quote) ([]byte, error) {
	now := time.Now().UTC()

	podmiot1 := buildPodmiot1(q.Company)
	podmiot2 := buildPodmiot2(q.Client)

	type vatGroup struct{ netto, vat float64 }
	groups := make(map[float64]*vatGroup)
	var totalNetto, totalBrutto float64

	rows := make([]ksefFaWiersz, 0, len(q.Items))
	for i, li := range q.Items {
		netto := roundCents(li.Quantity * li.UnitPrice)
		vatAmt := roundCents(netto * li.VatRate / 100)
		brutto := netto + vatAmt

		if g, ok := groups[li.VatRate]; ok {
			g.netto += netto
			g.vat += vatAmt
		} else {
			groups[li.VatRate] = &vatGroup{netto, vatAmt}
		}
		totalNetto += netto
		totalBrutto += brutto

		rows = append(rows, ksefFaWiersz{
			NrWierszaFa: i + 1,
			P_7:         li.Name,
			P_8A:        "szt.",
			P_8B:        formatNumber(li.Quantity),
			P_9A:        fmt.Sprintf("%.2f", li.UnitPrice),
			P_11:        fmt.Sprintf("%.2f", netto),
			P_12:        vatRateToP12(li.VatRate),
		})
	}

	// Sumy zaokrąglone per grupa — zapobiegają rozjazdom groszowym
	sumaNetto := roundCents(totalNetto)
	sumaBrutto := roundCents(totalBrutto)

	// Zbuduj sekcję Platnosc gdy termin płatności to data ISO
	var platnosc *ksefPlatnosc
	if pd := strings.TrimSpace(q.PaymentDue); isISODate(pd) {
		platnosc = &ksefPlatnosc{
			TerminPlatnosci: ksefTerminPlatnosci{Termin: pd},
		}
	}

	// P_18A: mechanizm podzielonej płatności wymagany gdy brutto > 15 000 zł
	// Uproszczenie: sygnalizujemy MPP dla faktur powyżej progu; kwalifikacja towarów z zał. 15
	// leży po stronie użytkownika.
	p18a := 2
	if sumaBrutto > 15000 {
		p18a = 1
	}

	fa := ksefFa{
		KodWaluty: "PLN",
		P_1:       now.Format("2006-01-02"),
		P_2:       q.InvoiceNumber,
		P_6:       strings.TrimSpace(q.SaleDate),
		P_15:      fmt.Sprintf("%.2f", sumaBrutto),
		Adnotacje: ksefAdnotacje{
			P_16:                 2,
			P_17:                 2,
			P_18:                 2,
			P_18A:                p18a,
			Zwolnienie:           ksefZwolnienie{P_19N: 1},
			NoweSrodkiTransportu: ksefNoweSrodkiTransportu{P_22N: 1},
			P_23:                 2,
			PMarzy:               ksefPMarzy{P_PMarzyN: 1},
		},
		RodzajFaktury: "VAT",
		Platnosc:      platnosc,
		FaWiersze: ksefFaWiersze{
			LiczbaWierszy:       len(q.Items),
			WartoscWierszyNetto: fmt.Sprintf("%.2f", sumaNetto),
			FaWiersz:            rows,
		},
	}

	if g, ok := groups[23]; ok {
		fa.P_13_1 = fmt.Sprintf("%.2f", roundCents(g.netto))
		fa.P_14_1 = fmt.Sprintf("%.2f", roundCents(g.vat))
	}
	if g, ok := groups[8]; ok {
		fa.P_13_2 = fmt.Sprintf("%.2f", roundCents(g.netto))
		fa.P_14_2 = fmt.Sprintf("%.2f", roundCents(g.vat))
	}
	if g, ok := groups[5]; ok {
		fa.P_13_3 = fmt.Sprintf("%.2f", roundCents(g.netto))
		fa.P_14_3 = fmt.Sprintf("%.2f", roundCents(g.vat))
	}
	if g, ok := groups[0]; ok {
		fa.P_13_6_1 = fmt.Sprintf("%.2f", roundCents(g.netto))
		// stawka 0% — brak pola P_14_x (XSD nie przewiduje VAT=0 od strony kwotowej)
	}

	// Sprawdzenie spójności sum przed serializacją
	if err := checkKSeFSums(fa); err != nil {
		return nil, fmt.Errorf("niespójne sumy w XML KSeF: %w", err)
	}

	faktura := ksefFaktura{
		Xmlns: ksefNS,
		Naglowek: ksefNaglowek{
			KodFormularza: ksefKodFormularza{
				KodSystemowy: "FA (3)",
				WersjaSchemy: "1-0E",
				Wartosc:      "FA",
			},
			WariantFormularza: 3,
			DataWytworzeniaFa: now.Format("2006-01-02T15:04:05Z07:00"),
			SystemInfo:        "SumIt",
		},
		Podmiot1: podmiot1,
		Podmiot2: podmiot2,
		Fa:       fa,
	}

	out, err := xml.MarshalIndent(faktura, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}

// checkKSeFSums weryfikuje wewnętrzną spójność sum FA(3):
//   - suma wartości netto wierszy (WartoscWierszyNetto) = suma P_13_x
//   - P_15 (brutto) = WartoscWierszyNetto + suma P_14_x
//
// Tolerancja 1 grosz na grupę z powodu kolejnych zaokrągleń float64.
func checkKSeFSums(fa ksefFa) error {
	parseF := func(s string) float64 {
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	wartoscWierszyNetto := parseF(fa.FaWiersze.WartoscWierszyNetto)

	// Suma P_11 wierszy powinna = WartoscWierszyNetto
	var sumaP11 float64
	for _, w := range fa.FaWiersze.FaWiersz {
		sumaP11 += parseF(w.P_11)
	}
	sumaP11 = roundCents(sumaP11)
	if math.Abs(sumaP11-wartoscWierszyNetto) > 0.015 {
		return fmt.Errorf("suma P_11 wierszy (%.2f) ≠ WartoscWierszyNetto (%.2f)", sumaP11, wartoscWierszyNetto)
	}

	// Suma P_13_x = WartoscWierszyNetto
	sumaP13 := roundCents(parseF(fa.P_13_1) + parseF(fa.P_13_2) + parseF(fa.P_13_3) +
		parseF(fa.P_13_6_1) + parseF(fa.P_13_7))
	if math.Abs(sumaP13-wartoscWierszyNetto) > 0.015 {
		return fmt.Errorf("suma P_13_x (%.2f) ≠ WartoscWierszyNetto (%.2f)", sumaP13, wartoscWierszyNetto)
	}

	// P_15 = WartoscWierszyNetto + suma P_14_x
	sumaVAT := roundCents(parseF(fa.P_14_1) + parseF(fa.P_14_2) + parseF(fa.P_14_3))
	oczekiwaneBrutto := roundCents(wartoscWierszyNetto + sumaVAT)
	p15 := parseF(fa.P_15)
	if math.Abs(p15-oczekiwaneBrutto) > 0.015 {
		return fmt.Errorf("P_15 (%.2f) ≠ WartoscWierszyNetto + VAT (%.2f)", p15, oczekiwaneBrutto)
	}

	return nil
}

// buildPodmiot1 mapuje Company (sprzedawca) na sekcję Podmiot1.
// Zakłada, że validateKSeFData już sprawdziła obecność adresu i NIP.
func buildPodmiot1(c Company) ksefPodmiot1 {
	adresL1 := strings.TrimSpace(c.Address)
	adresL2 := strings.TrimSpace(c.City)
	// Gdy brak ulicy, miasto idzie jako AdresL1 (XSD wymaga AdresL1)
	if adresL1 == "" {
		adresL1 = adresL2
		adresL2 = ""
	}
	return ksefPodmiot1{
		DaneIdentyfikacyjne: ksefPodmiot1DaneId{
			NIP:   strings.TrimSpace(c.NIP),
			Nazwa: strings.TrimSpace(c.Name),
		},
		Adres: ksefAdres{
			KodKraju: "PL",
			AdresL1:  adresL1,
			AdresL2:  adresL2,
		},
	}
}

// buildPodmiot2 parsuje wieloliniowe pole klient na sekcję Podmiot2.
// Linia 1 → Nazwa; linia z 10 cyframi → NIP (brak → BrakID=1);
// pozostałe linie → AdresL1.
func buildPodmiot2(client string) ksefPodmiot2 {
	lines := strings.Split(strings.TrimSpace(client), "\n")
	var nazwa, adresL1, nipNabywcy string

	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if i == 0 {
			nazwa = line
			continue
		}
		if nipNabywcy == "" {
			if m := nipRe.FindString(line); m != "" {
				nipNabywcy = m
				continue
			}
		}
		if adresL1 == "" {
			adresL1 = line
		}
	}

	if nazwa == "" {
		nazwa = strings.TrimSpace(client)
	}
	// Brak adresu nabywcy to częsty przypadek B2C — dopuszczamy pusty AdresL1
	// (pole opcjonalne w FA(3) dla BrakID=1; dla NIP-owego nabywcy XSD wymaga adresu)
	if adresL1 == "" && nipNabywcy != "" {
		adresL1 = "-"
	}

	daneId := ksefPodmiot2DaneId{Nazwa: nazwa}
	if nipNabywcy != "" {
		daneId.NIP = nipNabywcy
	} else {
		daneId.BrakID = 1
	}

	return ksefPodmiot2{
		DaneIdentyfikacyjne: daneId,
		Adres: ksefAdres{
			KodKraju: "PL",
			AdresL1:  adresL1,
		},
		JST: 2,
		GV:  2,
	}
}

// vatRateToP12 mapuje stawkę VAT (0/5/8/23) na enum TStawkaPodatku ze schematu FA(3).
func vatRateToP12(rate float64) string {
	switch rate {
	case 23:
		return "23"
	case 8:
		return "8"
	case 5:
		return "5"
	case 0:
		return "0 KR" // sprzedaż krajowa 0%
	default:
		return "zw"
	}
}

// ─── Handler HTTP ─────────────────────────────────────────────────────────────

// handleXML obsługuje POST /api/xml.
// Przyjmuje ten sam payload JSON co POST /quote i zwraca plik XML FA(3).
// Endpoint działa wyłącznie dla typ_dokumentu = "faktura_vat".
func handleXML(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var q Quote
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		http.Error(w, "nieprawidłowy JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if q.DocType != DocTypeFakturaVAT {
		http.Error(w, "endpoint /api/xml obsługuje tylko typ_dokumentu = faktura_vat", http.StatusBadRequest)
		return
	}

	if err := q.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateKSeFData(q); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := BuildKSeFXML(q)
	if err != nil {
		http.Error(w, "błąd generowania XML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Bezpieczna nazwa pliku: faktura-FV_2026_001-Klient.xml
	invoiceSlug := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, q.InvoiceNumber)
	clientSlug := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, strings.SplitN(q.Client, "\n", 2)[0])
	filename := fmt.Sprintf("faktura-%s-%s.xml", invoiceSlug, clientSlug)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, _ = io.Copy(w, strings.NewReader(string(data)))
}
