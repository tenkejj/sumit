package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// Liberation Sans (SIL Open Font License) — metryczny zamiennik Arial
// z pełnym wsparciem polskich znaków. Wbudowany w binarkę przez //go:embed,
// dzięki czemu generator PDF nie zależy od czcionek systemowych.
//
//go:embed assets/fonts/LiberationSans-Regular.ttf
var fontRegular []byte

//go:embed assets/fonts/LiberationSans-Bold.ttf
var fontBold []byte

func GeneratePDF(q Quote, w io.Writer) error {
	const family = "LiberationSans"

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(family, "", fontRegular)
	pdf.AddUTF8FontFromBytes(family, "B", fontBold)
	pdf.SetFont(family, "", 11)
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()

	const (
		marginL = 15.0
		marginR = 15.0
		pageW   = 210.0
		rightX  = pageW - marginR
		usableW = pageW - marginL - marginR
	)

	colorDark := func() { pdf.SetTextColor(10, 10, 10) }
	colorMuted := func() { pdf.SetTextColor(107, 114, 128) }
	colorLine := func() { pdf.SetDrawColor(220, 220, 220) }
	colorLineStrong := func() { pdf.SetDrawColor(10, 10, 10) }

	// Nagłówek: logo po lewej, blok "WYCENA / nr / data" po prawej.
	const logoY = 12.0
	const logoH = 16.0
	if strings.TrimSpace(q.Company.LogoBase64) != "" {
		if data, imgType, err := decodeLogoBase64(q.Company.LogoBase64); err == nil {
			const logoName = "oferta_logo"
			opts := fpdf.ImageOptions{ImageType: imgType, ReadDpi: false}
			pdf.RegisterImageOptionsReader(logoName, opts, bytes.NewReader(data))
			pdf.ImageOptions(logoName, marginL, logoY, 0, logoH, false, opts, 0, "")
		}
	}

	// Tytuł dokumentu zależny od DocType
	docTitle := "WYCENA"
	switch q.DocType {
	case DocTypeFakturaVAT:
		docTitle = "FAKTURA VAT"
	case DocTypeProforma:
		docTitle = "FAKTURA PRO FORMA"
	}

	pdf.SetXY(marginL, logoY)
	pdf.SetFont(family, "B", 26)
	colorDark()
	pdf.CellFormat(usableW, 10, docTitle, "", 1, "R", false, 0, "")

	pdf.SetFont(family, "", 10)
	colorMuted()
	// Numer dokumentu
	numDoc := strings.TrimSpace(q.InvoiceNumber)
	if numDoc == "" {
		numDoc = strings.TrimSpace(q.Number)
	}
	if numDoc != "" {
		pdf.SetX(marginL)
		pdf.CellFormat(usableW, 5, "Nr "+numDoc, "", 1, "R", false, 0, "")
	}
	pdf.SetX(marginL)
	pdf.CellFormat(usableW, 5, "Data wystawienia: "+time.Now().Format("2006-01-02"), "", 1, "R", false, 0, "")
	if q.IsInvoice() {
		if s := strings.TrimSpace(q.SaleDate); s != "" {
			pdf.SetX(marginL)
			pdf.CellFormat(usableW, 5, "Data sprzedaży: "+s, "", 1, "R", false, 0, "")
		}
		if s := strings.TrimSpace(q.PaymentDue); s != "" {
			pdf.SetX(marginL)
			pdf.CellFormat(usableW, 5, "Termin płatności: "+s, "", 1, "R", false, 0, "")
		}
	}

	headerEnd := maxF(logoY+logoH, pdf.GetY())
	pdf.SetY(headerEnd)
	pdf.Ln(12)

	// Sprzedawca / Klient — duże marginesy, etykiety mniejsze i bold.
	startY := pdf.GetY()

	pdf.SetXY(marginL, startY)
	pdf.SetFont(family, "B", 9)
	colorMuted()
	pdf.CellFormat(85, 5, "Sprzedawca:", "", 1, "L", false, 0, "")
	pdf.SetX(marginL)
	pdf.SetFont(family, "", 11)
	colorDark()
	pdf.MultiCell(85, 6, q.Company.Name, "", "L", false)

	var companyDetails []string
	if s := strings.TrimSpace(q.Company.NIP); s != "" {
		companyDetails = append(companyDetails, "NIP: "+s)
	}
	if s := strings.TrimSpace(q.Company.Address); s != "" {
		companyDetails = append(companyDetails, s)
	}
	if s := strings.TrimSpace(q.Company.City); s != "" {
		companyDetails = append(companyDetails, s)
	}
	phone := strings.TrimSpace(q.Company.Phone)
	email := strings.TrimSpace(q.Company.Email)
	if len(companyDetails) > 0 || phone != "" || email != "" {
		pdf.SetX(marginL)
		pdf.SetFont(family, "", 8)
		colorMuted()
		if len(companyDetails) > 0 {
			pdf.MultiCell(85, 4, strings.Join(companyDetails, "\n"), "", "L", false)
		}
		if phone != "" {
			pdf.SetX(marginL)
			pdf.Write(4, "tel. ")
			pdf.WriteLinkString(4, phone, "tel:"+sanitizePhoneLink(phone))
			pdf.Ln(4)
		}
		if email != "" {
			pdf.SetX(marginL)
			pdf.Write(4, "e-mail: ")
			pdf.WriteLinkString(4, email, "mailto:"+email)
			pdf.Ln(4)
		}
	}
	leftEnd := pdf.GetY()

	pdf.SetXY(110, startY)
	pdf.SetFont(family, "B", 9)
	colorMuted()
	pdf.CellFormat(85, 5, "Klient:", "", 1, "L", false, 0, "")
	pdf.SetX(110)
	pdf.SetFont(family, "", 11)
	colorDark()
	pdf.MultiCell(85, 6, q.Client, "", "L", false)
	rightEnd := pdf.GetY()

	pdf.SetY(maxF(leftEnd, rightEnd))
	pdf.Ln(12)

	// Tabela pozycji — dwa tryby: oferta/pro forma (bez VAT) i faktura VAT (z VAT).
	pdf.SetLineWidth(0.2)
	colorLine()

	if q.DocType == DocTypeFakturaVAT {
		// ── Kolumny faktury VAT ──────────────────────────────────────────
		const (
			fLp    = 8.0
			fNazwa = 50.0
			fIl    = 14.0
			fNetJ  = 22.0
			fNetW  = 24.0
			fVatP  = 13.0
			fVatK  = 22.0
			fBrut  = 27.0
		)

		pdf.SetFont(family, "B", 8)
		pdf.SetFillColor(245, 245, 245)
		colorDark()
		pdf.CellFormat(fLp, 7, "Lp.", "", 0, "C", true, 0, "")
		pdf.CellFormat(fNazwa, 7, "Nazwa", "", 0, "L", true, 0, "")
		pdf.CellFormat(fIl, 7, "Ilość", "", 0, "R", true, 0, "")
		pdf.CellFormat(fNetJ, 7, "Cena netto", "", 0, "R", true, 0, "")
		pdf.CellFormat(fNetW, 7, "Wart. netto", "", 0, "R", true, 0, "")
		pdf.CellFormat(fVatP, 7, "VAT%", "", 0, "R", true, 0, "")
		pdf.CellFormat(fVatK, 7, "Kwota VAT", "", 0, "R", true, 0, "")
		pdf.CellFormat(fBrut, 7, "Brutto", "", 1, "R", true, 0, "")

		yLine := pdf.GetY()
		pdf.Line(marginL, yLine, rightX, yLine)

		pdf.SetFont(family, "", 9)
		colorDark()

		type vatGroup struct{ netto, vat, brutto float64 }
		vatGroups := make(map[float64]*vatGroup)

		for i, li := range q.Items {
			netto := li.Quantity * li.UnitPrice
			vatAmt := roundCents(netto * li.VatRate / 100)
			brutto := netto + vatAmt

			pdf.CellFormat(fLp, 6, strconv.Itoa(i+1), "", 0, "C", false, 0, "")
			pdf.CellFormat(fNazwa, 6, li.Name, "", 0, "L", false, 0, "")
			pdf.CellFormat(fIl, 6, formatNumber(li.Quantity), "", 0, "R", false, 0, "")
			pdf.CellFormat(fNetJ, 6, formatPLN(li.UnitPrice), "", 0, "R", false, 0, "")
			pdf.CellFormat(fNetW, 6, formatPLN(netto), "", 0, "R", false, 0, "")
			pdf.CellFormat(fVatP, 6, strconv.FormatFloat(li.VatRate, 'f', 0, 64)+"%", "", 0, "R", false, 0, "")
			pdf.CellFormat(fVatK, 6, formatPLN(vatAmt), "", 0, "R", false, 0, "")
			pdf.CellFormat(fBrut, 6, formatPLN(brutto), "", 1, "R", false, 0, "")

			if g, ok := vatGroups[li.VatRate]; ok {
				g.netto += netto; g.vat += vatAmt; g.brutto += brutto
			} else {
				vatGroups[li.VatRate] = &vatGroup{netto, vatAmt, brutto}
			}

			yRow := pdf.GetY()
			pdf.Line(marginL, yRow, rightX, yRow)
		}

		// Gruba linia + razem brutto
		yBefore := pdf.GetY()
		pdf.SetLineWidth(0.5)
		colorLineStrong()
		pdf.Line(marginL, yBefore, rightX, yBefore)

		var totalNetto, totalVAT, totalBrutto float64
		for _, g := range vatGroups {
			totalNetto += g.netto
			totalVAT += g.vat
			totalBrutto += g.brutto
		}
		pdf.SetFont(family, "B", 10)
		colorDark()
		pdf.CellFormat(fLp+fNazwa+fIl+fNetJ+fNetW, 8, "Razem:", "", 0, "R", false, 0, "")
		pdf.CellFormat(fVatP+fVatK, 8, formatPLN(totalVAT), "", 0, "R", false, 0, "")
		pdf.CellFormat(fBrut, 8, formatPLN(totalBrutto), "", 1, "R", false, 0, "")

		yAfter := pdf.GetY()
		pdf.Line(marginL, yAfter, rightX, yAfter)
		pdf.SetLineWidth(0.2)
		colorLine()

		// Rozbicie VAT wg stawek
		if len(vatGroups) > 0 {
			pdf.Ln(6)
			pdf.SetFont(family, "B", 8)
			colorMuted()
			pdf.CellFormat(usableW, 4, "ZESTAWIENIE PODATKU VAT", "", 1, "L", false, 0, "")
			pdf.Ln(2)
			pdf.SetFont(family, "B", 8)
			colorDark()
			pdf.CellFormat(20, 6, "Stawka", "", 0, "C", false, 0, "")
			pdf.CellFormat(50, 6, "Wartość netto", "", 0, "R", false, 0, "")
			pdf.CellFormat(50, 6, "Kwota VAT", "", 0, "R", false, 0, "")
			pdf.CellFormat(50, 6, "Wartość brutto", "", 1, "R", false, 0, "")
			colorLine()
			pdf.Line(marginL, pdf.GetY(), rightX, pdf.GetY())

			// Sortowanie stawek
			rates := make([]float64, 0, len(vatGroups))
			for r := range vatGroups { rates = append(rates, r) }
			for i := 0; i < len(rates)-1; i++ {
				for j := i + 1; j < len(rates); j++ {
					if rates[j] < rates[i] { rates[i], rates[j] = rates[j], rates[i] }
				}
			}
			pdf.SetFont(family, "", 8)
			colorDark()
			for _, r := range rates {
				g := vatGroups[r]
				label := strconv.FormatFloat(r, 'f', 0, 64) + "%"
				pdf.CellFormat(20, 5, label, "", 0, "C", false, 0, "")
				pdf.CellFormat(50, 5, formatPLN(g.netto), "", 0, "R", false, 0, "")
				pdf.CellFormat(50, 5, formatPLN(g.vat), "", 0, "R", false, 0, "")
				pdf.CellFormat(50, 5, formatPLN(g.brutto), "", 1, "R", false, 0, "")
			}
			pdf.SetFont(family, "B", 8)
			pdf.CellFormat(20, 5, "Razem", "", 0, "C", false, 0, "")
			pdf.CellFormat(50, 5, formatPLN(totalNetto), "", 0, "R", false, 0, "")
			pdf.CellFormat(50, 5, formatPLN(totalVAT), "", 0, "R", false, 0, "")
			pdf.CellFormat(50, 5, formatPLN(totalBrutto), "", 1, "R", false, 0, "")
		}
	} else {
		// ── Kolumny oferty / pro formy (bez VAT) ─────────────────────────
		const (
			colLp    = 10.0
			colNazwa = 85.0
			colIlosc = 20.0
			colCena  = 30.0
			colWart  = 35.0
		)

		pdf.SetFont(family, "B", 10)
		pdf.SetFillColor(245, 245, 245)
		colorDark()
		pdf.CellFormat(colLp, 8, "Lp.", "", 0, "C", true, 0, "")
		pdf.CellFormat(colNazwa, 8, "Nazwa", "", 0, "L", true, 0, "")
		pdf.CellFormat(colIlosc, 8, "Ilość", "", 0, "R", true, 0, "")
		pdf.CellFormat(colCena, 8, "Cena jedn.", "", 0, "R", true, 0, "")
		pdf.CellFormat(colWart, 8, "Wartość", "", 1, "R", true, 0, "")

		yLine := pdf.GetY()
		pdf.Line(marginL, yLine, rightX, yLine)

		pdf.SetFont(family, "", 10)
		colorDark()
		for i, li := range q.Items {
			pdf.CellFormat(colLp, 7, strconv.Itoa(i+1), "", 0, "C", false, 0, "")
			pdf.CellFormat(colNazwa, 7, li.Name, "", 0, "L", false, 0, "")
			pdf.CellFormat(colIlosc, 7, formatNumber(li.Quantity), "", 0, "R", false, 0, "")
			pdf.CellFormat(colCena, 7, formatPLN(li.UnitPrice), "", 0, "R", false, 0, "")
			pdf.CellFormat(colWart, 7, formatPLN(li.Total()), "", 1, "R", false, 0, "")
			yRow := pdf.GetY()
			pdf.Line(marginL, yRow, rightX, yRow)
		}

		yBeforeRazem := pdf.GetY()
		pdf.SetLineWidth(0.5)
		colorLineStrong()
		pdf.Line(marginL, yBeforeRazem, rightX, yBeforeRazem)

		pdf.SetFont(family, "B", 12)
		colorDark()
		pdf.CellFormat(colLp+colNazwa+colIlosc+colCena, 9, "Razem:", "", 0, "R", false, 0, "")
		pdf.CellFormat(colWart, 9, formatPLN(q.Total()), "", 1, "R", false, 0, "")

		yAfterRazem := pdf.GetY()
		pdf.SetLineWidth(0.5)
		colorLineStrong()
		pdf.Line(marginL, yAfterRazem, rightX, yAfterRazem)

		pdf.SetLineWidth(0.2)
		colorLine()
	}

	// QR + dane do przelewu pod tabelą.
	if nrb := sanitizeNRB(q.Company.BankAccount); len(nrb) == 26 {
		var paymentTitle string
		switch q.DocType {
		case DocTypeFakturaVAT:
			paymentTitle = "Faktura"
			if s := strings.TrimSpace(q.InvoiceNumber); s != "" {
				paymentTitle = "Faktura " + s
			}
		case DocTypeProforma:
			paymentTitle = "Pro forma"
			if s := strings.TrimSpace(q.InvoiceNumber); s != "" {
				paymentTitle = "Pro forma " + s
			}
		default:
			paymentTitle = "Oferta"
			if s := strings.TrimSpace(q.Number); s != "" {
				paymentTitle = "Oferta " + s
			}
		}
		qrContent := formatPolishPaymentQR(nrb, q.Total(), q.Company.Name, paymentTitle)
		if png, err := generateQRPNG(qrContent); err == nil {
			pdf.Ln(10)
			yQR := pdf.GetY()
			const qrSize = 35.0

			pdf.SetXY(marginL, yQR)
			pdf.SetFont(family, "B", 8)
			colorMuted()
			pdf.CellFormat(0, 4, "ZESKANUJ, ABY ZAPŁACIĆ", "", 1, "L", false, 0, "")
			pdf.Ln(1)
			pdf.SetX(marginL)
			pdf.SetFont(family, "", 10)
			colorDark()
			pdf.MultiCell(usableW-qrSize-5, 5,
				"Numer konta: "+formatNRBWithSpaces(nrb)+"\n"+
					"Odbiorca: "+q.Company.Name+"\n"+
					"Tytuł: "+paymentTitle+"\n"+
					"Kwota: "+formatPLN(q.Total()),
				"", "L", false)

			const qrName = "oferta_qr"
			opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
			pdf.RegisterImageOptionsReader(qrName, opts, bytes.NewReader(png))
			pdf.ImageOptions(qrName, rightX-qrSize, yQR, qrSize, qrSize, false, opts, 0, "")

			qrEnd := yQR + qrSize + 2
			if pdf.GetY() < qrEnd {
				pdf.SetY(qrEnd)
			}
		}
	}

	// Uwagi do oferty.
	if notes := strings.TrimSpace(q.Notes); notes != "" {
		pdf.Ln(10)
		pdf.SetX(marginL)
		pdf.SetFont(family, "B", 8)
		colorMuted()
		pdf.CellFormat(0, 5, "UWAGI DO OFERTY", "", 1, "L", false, 0, "")
		pdf.Ln(1)
		pdf.SetX(marginL)
		pdf.SetFont(family, "", 10)
		colorDark()
		pdf.MultiCell(0, 5, notes, "", "L", false)
	}

	// Stopka — klauzula zależna od typu dokumentu.
	pdf.Ln(10)
	pdf.SetFont(family, "", 9)
	colorMuted()
	switch q.DocType {
	case DocTypeFakturaVAT:
		paymentDueLabel := strings.TrimSpace(q.PaymentDue)
		if paymentDueLabel == "" {
			paymentDueLabel = "zgodnie z ustaleniami"
		}
		pdf.MultiCell(0, 5,
			"Faktura VAT wystawiona zgodnie z ustawą o podatku od towarów i usług. "+
				"Termin płatności: "+paymentDueLabel+". Dziękujemy za zainteresowanie naszą ofertą.",
			"", "L", false)
	case DocTypeProforma:
		pdf.MultiCell(0, 5,
			"Faktura pro forma nie jest dokumentem księgowym. Stanowi jedynie zapowiedź faktury właściwej. "+
				"Po dokonaniu wpłaty zostanie wystawiona faktura VAT. Dziękujemy za zainteresowanie naszą ofertą.",
			"", "L", false)
	default:
		validityPeriod := "14 dni od daty wystawienia"
		if s := strings.TrimSpace(q.ValidUntil); s != "" {
			validityPeriod = "do " + s
		}
		pdf.MultiCell(0, 5,
			"Oferta ważna "+validityPeriod+". Ceny są cenami netto, do których należy doliczyć podatek VAT zgodnie z obowiązującymi przepisami. "+
				"Dziękujemy za zainteresowanie naszą ofertą.",
			"", "L", false)
	}

	return pdf.Output(w)
}

func formatPLN(v float64) string {
	return fmt.Sprintf("%.2f zł", v)
}

// roundCents zaokrągla do 2 miejsc po przecinku (grosze).
func roundCents(v float64) float64 {
	return math.Round(v*100) / 100
}

func formatNumber(v float64) string {
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

// decodeLogoBase64 przyjmuje data URL (np. "data:image/png;base64,...") lub
// surowy base64 i zwraca zdekodowane bajty obrazu wraz z typem ("PNG" / "JPG")
// rozpoznanym z prefiksu MIME lub magicznych bajtów obrazu.
func decodeLogoBase64(s string) ([]byte, string, error) {
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

// sanitizePhoneLink przygotowuje wartość dla schematu URI tel:.
// Usuwa wyłącznie spacje i myślniki — pozostałe znaki (m.in. wiodące "+"
// dla numerów międzynarodowych) muszą zostać zachowane.
func sanitizePhoneLink(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}
