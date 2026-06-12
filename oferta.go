package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"
)

type Pozycja struct {
	Nazwa           string  `json:"nazwa"`
	Ilosc           float64 `json:"ilosc"`
	CenaJednostkowa float64 `json:"cena_jednostkowa"`
}

func (p Pozycja) Wartosc() float64 {
	return p.Ilosc * p.CenaJednostkowa
}

// FirmaDane zawiera dane sprzedawcy renderowane w nagłówku PDF.
// Pola są deserializowane z płaskiego JSON-a wysyłanego przez frontend
// (patrz Oferta.UnmarshalJSON).
type FirmaDane struct {
	Nazwa      string `json:"nazwa_firmy"`
	NIP        string `json:"nip"`
	Adres      string `json:"adres"`
	Miasto     string `json:"miasto"`
	Telefon    string `json:"telefon"`
	Email      string `json:"email"`
	LogoBase64 string `json:"logo_base64"`
	NumerKonta string `json:"numer_konta"`
}

type Oferta struct {
	Firma        FirmaDane `json:"-"`
	Klient       string    `json:"klient"`
	NumerOferty  string    `json:"numer_oferty"`
	DataWaznosci string    `json:"data_waznosci"`
	Uwagi        string    `json:"uwagi"`
	Pozycje      []Pozycja `json:"pozycje"`
}

// ofertaJSON to wewnętrzna reprezentacja serializacji przyjmująca płaski
// JSON, jaki wysyła frontend (pola firmy obok klient/pozycji na jednym
// poziomie). Mapowanie do zagnieżdżonej Oferta.Firma odbywa się w
// (Un)MarshalJSON, dzięki czemu kontrakt JSON pozostaje niezmieniony.
type ofertaJSON struct {
	FirmaDane
	Klient       string    `json:"klient"`
	NumerOferty  string    `json:"numer_oferty"`
	DataWaznosci string    `json:"data_waznosci"`
	Uwagi        string    `json:"uwagi"`
	Pozycje      []Pozycja `json:"pozycje"`
}

func (o *Oferta) UnmarshalJSON(data []byte) error {
	var raw ofertaJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*o = Oferta{
		Firma:        raw.FirmaDane,
		Klient:       raw.Klient,
		NumerOferty:  raw.NumerOferty,
		DataWaznosci: raw.DataWaznosci,
		Uwagi:        raw.Uwagi,
		Pozycje:      raw.Pozycje,
	}
	return nil
}

func (o Oferta) MarshalJSON() ([]byte, error) {
	return json.Marshal(ofertaJSON{
		FirmaDane:    o.Firma,
		Klient:       o.Klient,
		NumerOferty:  o.NumerOferty,
		DataWaznosci: o.DataWaznosci,
		Uwagi:        o.Uwagi,
		Pozycje:      o.Pozycje,
	})
}

func (o Oferta) Suma() float64 {
	var s float64
	for _, p := range o.Pozycje {
		s += p.Wartosc()
	}
	return s
}

func handleOferta(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)

	var o Oferta
	if err := dec.Decode(&o); err != nil {
		http.Error(w, "nieprawidłowy JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := o.Waliduj(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := generujOfertePDF(o, &buf); err != nil {
		http.Error(w, "błąd generowania PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="oferta.pdf"`)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = io.Copy(w, &buf)
}

func (o Oferta) Waliduj() error {
	if strings.TrimSpace(o.Firma.Nazwa) == "" {
		return fmt.Errorf("pole nazwa_firmy jest wymagane")
	}
	if strings.TrimSpace(o.Klient) == "" {
		return fmt.Errorf("pole klient jest wymagane")
	}
	if len(o.Pozycje) == 0 {
		return fmt.Errorf("lista pozycji nie może być pusta")
	}
	for i, p := range o.Pozycje {
		if strings.TrimSpace(p.Nazwa) == "" {
			return fmt.Errorf("pozycja #%d: brak nazwy", i+1)
		}
		if p.Ilosc <= 0 {
			return fmt.Errorf("pozycja #%d (%s): ilość musi być > 0", i+1, p.Nazwa)
		}
		if p.CenaJednostkowa < 0 {
			return fmt.Errorf("pozycja #%d (%s): cena_jednostkowa nie może być ujemna", i+1, p.Nazwa)
		}
	}
	return nil
}

// resolveFontPath zwraca ścieżkę do TTF obsługującego polskie znaki.
// Można nadpisać zmienną środowiskową OFEROWO_FONT_PATH.
func resolveFontPath() (string, error) {
	if p := os.Getenv("OFEROWO_FONT_PATH"); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("OFEROWO_FONT_PATH=%q nie istnieje: %w", p, err)
		}
		return p, nil
	}
	candidates := []string{
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\calibri.ttf`,
		`/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`,
		`/Library/Fonts/Arial.ttf`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("nie znaleziono czcionki TTF z polskimi znakami; ustaw OFEROWO_FONT_PATH")
}

func generujOfertePDF(o Oferta, w io.Writer) error {
	fontPath, err := resolveFontPath()
	if err != nil {
		return err
	}
	fontDir := filepath.Dir(fontPath)
	fontFile := filepath.Base(fontPath)
	family := strings.TrimSuffix(fontFile, filepath.Ext(fontFile))

	pdf := fpdf.New("P", "mm", "A4", fontDir)
	pdf.AddUTF8Font(family, "", fontFile)
	pdf.SetFont(family, "", 11)
	pdf.SetMargins(15, 18, 15)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()

	if strings.TrimSpace(o.Firma.LogoBase64) != "" {
		if data, imgType, err := dekodujLogoBase64(o.Firma.LogoBase64); err == nil {
			const nazwaLogo = "oferta_logo"
			opts := fpdf.ImageOptions{ImageType: imgType, ReadDpi: false}
			pdf.RegisterImageOptionsReader(nazwaLogo, opts, bytes.NewReader(data))
			pdf.ImageOptions(nazwaLogo, 15, 10, 0, 20, false, opts, 0, "")
		}
	}

	pdf.SetFont(family, "", 22)
	pdf.SetTextColor(30, 30, 30)
	yTytul := pdf.GetY()
	pdf.CellFormat(0, 12, "OFERTA HANDLOWA", "", 1, "C", false, 0, "")
	if s := strings.TrimSpace(o.NumerOferty); s != "" {
		yPoTytule := pdf.GetY()
		pdf.SetFont(family, "", 11)
		pdf.SetTextColor(120, 120, 120)
		pdf.SetXY(15, yTytul)
		pdf.CellFormat(0, 12, "Nr "+s, "", 0, "R", false, 0, "")
		pdf.SetFont(family, "", 22)
		pdf.SetTextColor(30, 30, 30)
		pdf.SetY(yPoTytule)
	}
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(15, pdf.GetY()+1, 195, pdf.GetY()+1)
	pdf.Ln(8)

	pdf.SetFont(family, "", 11)
	pdf.SetTextColor(0, 0, 0)

	startY := pdf.GetY()
	pdf.SetXY(15, startY)
	pdf.MultiCell(85, 6, "Sprzedawca:\n"+o.Firma.Nazwa, "", "L", false)

	var daneFirmy []string
	if s := strings.TrimSpace(o.Firma.NIP); s != "" {
		daneFirmy = append(daneFirmy, "NIP: "+s)
	}
	if s := strings.TrimSpace(o.Firma.Adres); s != "" {
		daneFirmy = append(daneFirmy, s)
	}
	if s := strings.TrimSpace(o.Firma.Miasto); s != "" {
		daneFirmy = append(daneFirmy, s)
	}
	if s := strings.TrimSpace(o.Firma.Telefon); s != "" {
		daneFirmy = append(daneFirmy, "tel. "+s)
	}
	if s := strings.TrimSpace(o.Firma.Email); s != "" {
		daneFirmy = append(daneFirmy, "e-mail: "+s)
	}
	if len(daneFirmy) > 0 {
		pdf.SetX(15)
		pdf.SetFont(family, "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.MultiCell(85, 4, strings.Join(daneFirmy, "\n"), "", "L", false)
		pdf.SetFont(family, "", 11)
		pdf.SetTextColor(0, 0, 0)
	}
	leftEnd := pdf.GetY()

	pdf.SetXY(110, startY)
	pdf.MultiCell(85, 6, "Klient:\n"+o.Klient, "", "L", false)
	rightEnd := pdf.GetY()

	pdf.SetY(maxF(leftEnd, rightEnd))
	pdf.Ln(4)

	pdf.SetFont(family, "", 10)
	pdf.CellFormat(0, 5, "Data wystawienia: "+time.Now().Format("2006-01-02"), "", 1, "R", false, 0, "")
	pdf.Ln(4)

	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont(family, "", 11)
	pdf.CellFormat(10, 8, "Lp.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(85, 8, "Nazwa", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Ilość", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Cena jedn.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Wartość", "1", 1, "C", true, 0, "")

	pdf.SetFont(family, "", 10)
	for i, p := range o.Pozycje {
		pdf.CellFormat(10, 7, strconv.Itoa(i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(85, 7, p.Nazwa, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 7, formatLiczba(p.Ilosc), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 7, formatPLN(p.CenaJednostkowa), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, formatPLN(p.Wartosc()), "1", 1, "R", false, 0, "")
	}

	pdf.SetFont(family, "", 12)
	pdf.SetFillColor(245, 245, 245)
	pdf.CellFormat(145, 9, "Razem:", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 9, formatPLN(o.Suma()), "1", 1, "R", true, 0, "")

	if nrb := sanitizeNRB(o.Firma.NumerKonta); len(nrb) == 26 {
		tytulPrzelewu := "Oferta"
		if s := strings.TrimSpace(o.NumerOferty); s != "" {
			tytulPrzelewu = "Oferta " + s
		}
		qrContent := formatPolskiQRPrzelew(nrb, o.Suma(), o.Firma.Nazwa, tytulPrzelewu)
		if png, err := generujQRPNG(qrContent); err == nil {
			pdf.Ln(8)
			startY := pdf.GetY()
			const qrRozmiar = 35.0

			pdf.SetXY(15, startY)
			pdf.SetFont(family, "", 8)
			pdf.SetTextColor(120, 120, 120)
			pdf.CellFormat(0, 4, "ZESKANUJ, ABY ZAPŁACIĆ", "", 1, "L", false, 0, "")
			pdf.SetX(15)
			pdf.SetFont(family, "", 10)
			pdf.SetTextColor(30, 30, 30)
			pdf.MultiCell(140, 5,
				"Numer konta: "+formatujNRBzeSpacjami(nrb)+"\n"+
					"Odbiorca: "+o.Firma.Nazwa+"\n"+
					"Tytuł: "+tytulPrzelewu+"\n"+
					"Kwota: "+formatPLN(o.Suma()),
				"", "L", false)

			const nazwaQR = "oferta_qr"
			opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
			pdf.RegisterImageOptionsReader(nazwaQR, opts, bytes.NewReader(png))
			pdf.ImageOptions(nazwaQR, 195-qrRozmiar, startY, qrRozmiar, qrRozmiar, false, opts, 0, "")

			koniecQR := startY + qrRozmiar + 2
			if pdf.GetY() < koniecQR {
				pdf.SetY(koniecQR)
			}
			pdf.SetTextColor(0, 0, 0)
		}
	}

	if uwagi := strings.TrimSpace(o.Uwagi); uwagi != "" {
		pdf.Ln(8)
		pdf.SetFont(family, "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, "UWAGI DO OFERTY", "", 1, "L", false, 0, "")
		pdf.Ln(1)
		pdf.SetFont(family, "", 10)
		pdf.SetTextColor(30, 30, 30)
		pdf.MultiCell(0, 5, uwagi, "", "L", false)
	}

	pdf.Ln(10)
	pdf.SetFont(family, "", 9)
	pdf.SetTextColor(120, 120, 120)
	terminWaznosci := "14 dni od daty wystawienia"
	if s := strings.TrimSpace(o.DataWaznosci); s != "" {
		terminWaznosci = "do " + s
	}
	pdf.MultiCell(0, 5,
		"Oferta ważna "+terminWaznosci+". Ceny są cenami netto, do których należy doliczyć podatek VAT zgodnie z obowiązującymi przepisami. "+
			"Dziękujemy za zainteresowanie naszą ofertą.",
		"", "L", false)

	return pdf.Output(w)
}

func formatPLN(v float64) string {
	return fmt.Sprintf("%.2f zł", v)
}

func formatLiczba(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// dekodujLogoBase64 przyjmuje data URL (np. "data:image/png;base64,...") lub
// surowy base64 i zwraca zdekodowane bajty obrazu wraz z typem ("PNG" / "JPG")
// rozpoznanym z prefiksu MIME lub magicznych bajtów obrazu.
func dekodujLogoBase64(s string) ([]byte, string, error) {
	s = strings.TrimSpace(s)
	raw := s
	imgType := ""

	if strings.HasPrefix(s, "data:") {
		idx := strings.Index(s, ";base64,")
		if idx == -1 {
			return nil, "", fmt.Errorf("nieprawidłowy data URL logo")
		}
		prefix := s[:idx]
		raw = s[idx+len(";base64,"):]
		switch {
		case strings.Contains(prefix, "image/png"):
			imgType = "PNG"
		case strings.Contains(prefix, "image/jpeg"), strings.Contains(prefix, "image/jpg"):
			imgType = "JPG"
		}
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", fmt.Errorf("dekodowanie base64 logo: %w", err)
	}

	if imgType == "" {
		switch {
		case len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
			imgType = "PNG"
		case len(data) >= 3 && bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
			imgType = "JPG"
		default:
			return nil, "", fmt.Errorf("nieobsługiwany typ obrazu logo (oczekiwano PNG lub JPG)")
		}
	}

	return data, imgType, nil
}

// sanitizeNRB normalizuje numer rachunku: usuwa białe znaki, myślniki, opcjonalny
// prefiks "PL" i zwraca surowy ciąg cyfr (oczekiwana długość: 26).
func sanitizeNRB(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	s = strings.TrimPrefix(s, "PL")
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// skrocPole przycina łańcuch do co najwyżej max znaków (runów),
// z uwzględnieniem polskich znaków diakrytycznych.
func skrocPole(s string, max int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max])
	}
	return s
}

// formatPolskiQRPrzelew buduje payload QR zgodny z Rekomendacją Związku Banków
// Polskich dla kodów dwuwymiarowych ("Standard 2D" / "Standard 2012"). Format
// pól rozdzielonych pionową kreską:
//
//	NIP|PL|NRB|KWOTA_W_GROSZACH|NAZWA_ODBIORCY|TYTUL|REZ1|REZ2|REZ3
//
// Pole NIP zostawiamy puste — nie jest wymagane do zwykłego przelewu.
// Nazwa odbiorcy jest ucinana do 20 znaków, tytuł do 32 znaków zgodnie ze
// specyfikacją (banki dłuższe pola potrafią odrzucać).
//
// TODO: zweryfikować z prawdziwą apką bankową (mBank/IKO/ING) po podpięciu
// pierwszego klienta — różne banki bywają wrażliwe na końcowe separatory.
func formatPolskiQRPrzelew(nrb string, kwotaPLN float64, nazwa, tytul string) string {
	kwotaGrosze := int64(math.Round(kwotaPLN * 100))
	if kwotaGrosze < 0 {
		kwotaGrosze = 0
	}
	return strings.Join([]string{
		"",
		"PL",
		nrb,
		strconv.FormatInt(kwotaGrosze, 10),
		skrocPole(nazwa, 20),
		skrocPole(tytul, 32),
		"",
		"",
		"",
	}, "|")
}

// generujQRPNG zwraca obraz PNG kodu QR (256 px) z poziomem korekcji błędów M.
func generujQRPNG(content string) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, 256)
}

// formatujNRBzeSpacjami zwraca 26-cyfrowy NRB w czytelnej postaci
// "CC RRRR RRRR RRRR RRRR RRRR RRRR" (jak w drukowanym formacie polskim).
// Wejście, które nie jest dokładnie 26 cyfr, zwraca bez zmian.
func formatujNRBzeSpacjami(nrb string) string {
	if len(nrb) != 26 {
		return nrb
	}
	var b strings.Builder
	b.Grow(26 + 6)
	b.WriteString(nrb[0:2])
	for i := 2; i < 26; i += 4 {
		b.WriteByte(' ')
		b.WriteString(nrb[i : i+4])
	}
	return b.String()
}
