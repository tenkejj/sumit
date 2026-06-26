package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"regexp"
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

//go:embed assets/fonts/LiberationSans-Italic.ttf
var fontItalic []byte

const pdfFamily = "LiberationSans"

const (
	pdfMarginL = 15.0
	pdfMarginR = 15.0
	pdfPageW   = 210.0
)

type pdfDoc struct {
	pdf     *fpdf.Fpdf
	rightX  float64
	usableW float64
}

func newPDFDoc() *pdfDoc {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(pdfFamily, "", fontRegular)
	pdf.AddUTF8FontFromBytes(pdfFamily, "B", fontBold)
	pdf.AddUTF8FontFromBytes(pdfFamily, "I", fontItalic)
	pdf.SetFont(pdfFamily, "", 11)
	pdf.SetMargins(pdfMarginL, 15, pdfMarginR)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()
	return &pdfDoc{
		pdf:     pdf,
		rightX:  pdfPageW - pdfMarginR,
		usableW: pdfPageW - pdfMarginL - pdfMarginR,
	}
}

func (d *pdfDoc) colorDark()  { d.pdf.SetTextColor(10, 10, 10) }
func (d *pdfDoc) colorMuted() { d.pdf.SetTextColor(107, 114, 128) }
func (d *pdfDoc) colorLine()  { d.pdf.SetDrawColor(220, 220, 220) }
func (d *pdfDoc) colorLineStrong() {
	d.pdf.SetDrawColor(10, 10, 10)
}

func GeneratePDF(q Quote, w io.Writer) error {
	doc := newPDFDoc()
	if q.DocType == DocTypeFakturaVAT {
		renderFakturaVAT(doc, q)
	} else {
		renderOferta(doc, q)
	}
	return doc.pdf.Output(w)
}

// renderOferta — dotychczasowy układ wyceny / faktury pro forma (bez zmian w logice).
func renderOferta(d *pdfDoc, q Quote) {
	pdf := d.pdf
	const logoY = 12.0
	const logoH = 16.0
	if strings.TrimSpace(q.Company.LogoBase64) != "" {
		if data, imgType, err := decodeLogoBase64(q.Company.LogoBase64); err == nil {
			const logoName = "oferta_logo"
			opts := fpdf.ImageOptions{ImageType: imgType, ReadDpi: false}
			pdf.RegisterImageOptionsReader(logoName, opts, bytes.NewReader(data))
			pdf.ImageOptions(logoName, pdfMarginL, logoY, 0, logoH, false, opts, 0, "")
		}
	}

	docTitle := "WYCENA"
	switch q.DocType {
	case DocTypeProforma:
		docTitle = "FAKTURA PRO FORMA"
	}

	pdf.SetXY(pdfMarginL, logoY)
	pdf.SetFont(pdfFamily, "B", 26)
	d.colorDark()
	pdf.CellFormat(d.usableW, 10, docTitle, "", 1, "R", false, 0, "")

	pdf.SetFont(pdfFamily, "", 10)
	d.colorMuted()
	numDoc := strings.TrimSpace(q.InvoiceNumber)
	if numDoc == "" {
		numDoc = strings.TrimSpace(q.Number)
	}
	if numDoc != "" {
		pdf.SetX(pdfMarginL)
		pdf.CellFormat(d.usableW, 5, "Nr "+numDoc, "", 1, "R", false, 0, "")
	}
	pdf.SetX(pdfMarginL)
	pdf.CellFormat(d.usableW, 5, "Data wystawienia: "+time.Now().Format("2006-01-02"), "", 1, "R", false, 0, "")
	if q.IsInvoice() {
		if s := strings.TrimSpace(q.SaleDate); s != "" {
			pdf.SetX(pdfMarginL)
			pdf.CellFormat(d.usableW, 5, "Data sprzedaży: "+s, "", 1, "R", false, 0, "")
		}
		if s := strings.TrimSpace(q.PaymentDue); s != "" {
			pdf.SetX(pdfMarginL)
			pdf.CellFormat(d.usableW, 5, "Termin płatności: "+s, "", 1, "R", false, 0, "")
		}
	}

	headerEnd := maxF(logoY+logoH, pdf.GetY())
	pdf.SetY(headerEnd)
	pdf.Ln(12)

	startY := pdf.GetY()
	pdf.SetXY(pdfMarginL, startY)
	pdf.SetFont(pdfFamily, "B", 9)
	d.colorMuted()
	pdf.CellFormat(85, 5, "Sprzedawca:", "", 1, "L", false, 0, "")
	pdf.SetX(pdfMarginL)
	pdf.SetFont(pdfFamily, "", 11)
	d.colorDark()
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
		pdf.SetX(pdfMarginL)
		pdf.SetFont(pdfFamily, "", 8)
		d.colorMuted()
		if len(companyDetails) > 0 {
			pdf.MultiCell(85, 4, strings.Join(companyDetails, "\n"), "", "L", false)
		}
		if phone != "" {
			pdf.SetX(pdfMarginL)
			pdf.Write(4, "tel. ")
			pdf.WriteLinkString(4, phone, "tel:"+sanitizePhoneLink(phone))
			pdf.Ln(4)
		}
		if email != "" {
			pdf.SetX(pdfMarginL)
			pdf.Write(4, "e-mail: ")
			pdf.WriteLinkString(4, email, "mailto:"+email)
			pdf.Ln(4)
		}
	}
	leftEnd := pdf.GetY()

	pdf.SetXY(110, startY)
	pdf.SetFont(pdfFamily, "B", 9)
	d.colorMuted()
	pdf.CellFormat(85, 5, "Klient:", "", 1, "L", false, 0, "")
	pdf.SetX(110)
	pdf.SetFont(pdfFamily, "", 11)
	d.colorDark()
	pdf.MultiCell(85, 6, q.Client, "", "L", false)
	rightEnd := pdf.GetY()

	pdf.SetY(maxF(leftEnd, rightEnd))
	pdf.Ln(12)

	pdf.SetLineWidth(0.2)
	d.colorLine()

	const (
		colLp    = 10.0
		colNazwa = 85.0
		colIlosc = 20.0
		colCena  = 30.0
		colWart  = 35.0
	)

	pdf.SetFont(pdfFamily, "B", 10)
	pdf.SetFillColor(245, 245, 245)
	d.colorDark()
	pdf.CellFormat(colLp, 8, "Lp.", "", 0, "C", true, 0, "")
	pdf.CellFormat(colNazwa, 8, "Nazwa", "", 0, "L", true, 0, "")
	pdf.CellFormat(colIlosc, 8, "Ilość", "", 0, "R", true, 0, "")
	pdf.CellFormat(colCena, 8, "Cena jedn.", "", 0, "R", true, 0, "")
	pdf.CellFormat(colWart, 8, "Wartość", "", 1, "R", true, 0, "")

	yLine := pdf.GetY()
	pdf.Line(pdfMarginL, yLine, d.rightX, yLine)

	pdf.SetFont(pdfFamily, "", 10)
	d.colorDark()
	for i, li := range q.Items {
		pdf.CellFormat(colLp, 7, strconv.Itoa(i+1), "", 0, "C", false, 0, "")
		pdf.CellFormat(colNazwa, 7, li.Name, "", 0, "L", false, 0, "")
		pdf.CellFormat(colIlosc, 7, formatNumber(li.Quantity), "", 0, "R", false, 0, "")
		pdf.CellFormat(colCena, 7, formatPLN(li.UnitPrice), "", 0, "R", false, 0, "")
		pdf.CellFormat(colWart, 7, formatPLN(li.Total()), "", 1, "R", false, 0, "")
		yRow := pdf.GetY()
		pdf.Line(pdfMarginL, yRow, d.rightX, yRow)
	}

	yBeforeRazem := pdf.GetY()
	pdf.SetLineWidth(0.5)
	d.colorLineStrong()
	pdf.Line(pdfMarginL, yBeforeRazem, d.rightX, yBeforeRazem)

	pdf.SetFont(pdfFamily, "B", 12)
	d.colorDark()
	pdf.CellFormat(colLp+colNazwa+colIlosc+colCena, 9, "Razem:", "", 0, "R", false, 0, "")
	pdf.CellFormat(colWart, 9, formatPLN(q.Total()), "", 1, "R", false, 0, "")

	yAfterRazem := pdf.GetY()
	pdf.SetLineWidth(0.5)
	d.colorLineStrong()
	pdf.Line(pdfMarginL, yAfterRazem, d.rightX, yAfterRazem)
	pdf.SetLineWidth(0.2)
	d.colorLine()

	if q.ShouldShowTransferQR() {
		renderTransferQR(d, q, q.Total())
	}

	if notes := strings.TrimSpace(q.Notes); notes != "" {
		pdf.Ln(10)
		pdf.SetX(pdfMarginL)
		pdf.SetFont(pdfFamily, "B", 8)
		d.colorMuted()
		pdf.CellFormat(0, 5, "UWAGI DO OFERTY", "", 1, "L", false, 0, "")
		pdf.Ln(1)
		pdf.SetX(pdfMarginL)
		pdf.SetFont(pdfFamily, "", 10)
		d.colorDark()
		pdf.MultiCell(0, 5, notes, "", "L", false)
	}

	pdf.Ln(10)
	pdf.SetFont(pdfFamily, "", 9)
	d.colorMuted()
	switch q.DocType {
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
}

// renderFakturaVAT — klasyczny polski układ faktury VAT.

// fakturaVATColWidths — proporcje kolumn z oryginału (152 mm), skalowane do pełnej szerokości content area.
func fakturaVATColWidths(usableW float64) [9]float64 {
	base := [9]float64{7, 38, 10, 11, 18, 20, 11, 18, 19}
	var baseSum float64
	for _, w := range base {
		baseSum += w
	}
	var out [9]float64
	var used float64
	for i, w := range base {
		if i == len(base)-1 {
			out[i] = usableW - used
			break
		}
		out[i] = w / baseSum * usableW
		used += out[i]
	}
	return out
}

// fakturaVATDrawRow rysuje pełny wiersz tabeli (9 kolumn) od lewego marginesu do prawego.
func (d *pdfDoc) fakturaVATDrawRow(widths [9]float64, rowH float64, texts [9]string, aligns [9]string, fill bool) {
	pdf := d.pdf
	y := pdf.GetY()
	x := pdfMarginL
	for i := 0; i < 9; i++ {
		ln := 0
		if i == 8 {
			ln = 1
		}
		pdf.SetXY(x, y)
		pdf.CellFormat(widths[i], rowH, texts[i], "1", ln, aligns[i], fill, 0, "")
		x += widths[i]
	}
}

// fakturaVATDrawSummaryRow — wiersz „Razem” / „W tym” (6 komórek, od Lp. do brutto).
func (d *pdfDoc) fakturaVATDrawSummaryRow(sumLabelW, cNetJ, cNetW, cVatP, cVatK, cBrut float64, cells [6]string, bold bool) {
	pdf := d.pdf
	if bold {
		pdf.SetFont(pdfFamily, "B", 7)
	} else {
		pdf.SetFont(pdfFamily, "", 7)
	}
	d.colorDark()
	y := pdf.GetY()
	x := pdfMarginL
	widths := [6]float64{sumLabelW, cNetJ, cNetW, cVatP, cVatK, cBrut}
	for i, w := range widths {
		ln := 0
		if i == 5 {
			ln = 1
		}
		pdf.SetXY(x, y)
		pdf.CellFormat(w, 6, cells[i], "1", ln, "R", false, 0, "")
		x += w
	}
}

func renderFakturaVAT(d *pdfDoc, q Quote) {
	pdf := d.pdf
	const maxLogoH = 15.0
	const maxLogoW = 45.0
	startY := pdf.GetY()

	// Lewa strona: logo + nazwa firmy
	leftW := 95.0
	logoBottom := startY
	if strings.TrimSpace(q.Company.LogoBase64) != "" {
		if data, imgType, err := decodeLogoBase64(q.Company.LogoBase64); err == nil {
			const logoName = "faktura_logo"
			opts := fpdf.ImageOptions{ImageType: imgType, ReadDpi: false}
			pdf.RegisterImageOptionsReader(logoName, opts, bytes.NewReader(data))
			logoW, logoH := fitLogoSize(pdf, logoName, maxLogoW, maxLogoH)
			pdf.ImageOptions(logoName, pdfMarginL, startY, logoW, logoH, false, opts, 0, "")
			logoBottom = startY + logoH + 2
		}
	}
	pdf.SetXY(pdfMarginL, logoBottom)
	pdf.SetFont(pdfFamily, "B", 11)
	d.colorDark()
	pdf.MultiCell(leftW, 5, strings.TrimSpace(q.Company.Name), "", "L", false)
	leftEnd := pdf.GetY()

	// Prawa strona: Faktura + numer (wyśrodkowane w prawej połowie)
	rightColX := pdfMarginL + leftW + 5
	rightColW := d.usableW - leftW - 5
	pdf.SetXY(rightColX, startY)
	pdf.SetFont(pdfFamily, "I", 22)
	d.colorDark()
	pdf.CellFormat(rightColW, 12, "Faktura", "", 1, "C", false, 0, "")

	invNum := strings.TrimSpace(q.InvoiceNumber)
	pdf.SetX(rightColX)
	pdf.SetFont(pdfFamily, "B", 14)
	pdf.CellFormat(rightColW, 8, "nr: "+invNum, "", 1, "C", false, 0, "")
	rightEnd := pdf.GetY()

	pdf.SetY(maxF(leftEnd, rightEnd) + 4)

	// Daty — wyrównane do prawej
	issueDate := formatDatePL(time.Now().Format("2006-01-02"))
	saleDate := formatDatePL(strings.TrimSpace(q.SaleDate))
	city := strings.TrimSpace(q.Company.City)
	issueLine := "Wystawiona w dniu: " + issueDate
	if city != "" {
		issueLine += ", " + city
	}
	pdf.SetFont(pdfFamily, "", 9)
	d.colorDark()
	pdf.SetX(pdfMarginL)
	pdf.CellFormat(d.usableW, 5, issueLine, "", 1, "R", false, 0, "")
	if saleDate != "" {
		pdf.SetX(pdfMarginL)
		pdf.CellFormat(d.usableW, 5, "Data zakończenia dostawy/usługi: "+saleDate, "", 1, "R", false, 0, "")
	}
	pdf.Ln(8)

	// Sprzedawca / Nabywca
	partyY := pdf.GetY()
	colW := 88.0

	pdf.SetXY(pdfMarginL, partyY)
	pdf.SetFont(pdfFamily, "B", 10)
	d.colorDark()
	pdf.CellFormat(colW, 5, "Sprzedawca", "", 1, "L", false, 0, "")
	pdf.SetX(pdfMarginL)
	pdf.SetFont(pdfFamily, "", 9)
	pdf.MultiCell(colW, 4.5, formatSellerBlock(q.Company), "", "L", false)
	sellerEnd := pdf.GetY()

	buyerName, buyerAddr, buyerNIP := parseBuyerForPDF(q.Client)
	pdf.SetXY(pdfMarginL+colW+4, partyY)
	pdf.SetFont(pdfFamily, "B", 10)
	d.colorDark()
	pdf.CellFormat(colW, 5, "Nabywca", "", 1, "L", false, 0, "")
	var buyerLines []string
	if buyerName != "" {
		buyerLines = append(buyerLines, buyerName)
	}
	if buyerAddr != "" {
		buyerLines = append(buyerLines, buyerAddr)
	}
	if buyerNIP != "" {
		buyerLines = append(buyerLines, "NIP: "+buyerNIP)
	}
	if len(buyerLines) == 0 {
		buyerLines = []string{strings.TrimSpace(q.Client)}
	}
	pdf.SetX(pdfMarginL + colW + 4)
	pdf.SetFont(pdfFamily, "", 9)
	pdf.MultiCell(colW, 4.5, strings.Join(buyerLines, "\n"), "", "L", false)
	buyerEnd := pdf.GetY()

	pdf.SetY(maxF(sellerEnd, buyerEnd) + 6)

	// Sposób zapłaty
	paymentMethod := "gotówka"
	if nrb := sanitizeNRB(q.Company.BankAccount); len(nrb) == 26 {
		paymentMethod = "Przelew bankowy"
	}
	pdf.SetX(pdfMarginL)
	pdf.SetFont(pdfFamily, "", 9)
	d.colorDark()
	pdf.CellFormat(d.usableW, 5, "Sposób zapłaty: "+paymentMethod, "", 1, "L", false, 0, "")
	pdf.Ln(11) // +5 mm przerwy między sekcją stron a tabelą (łącznie z poprzednim odstępem)

	// Tabela pozycji
	vatGroups := make(map[float64]*vatGroup)
	var totalNetto, totalVAT, totalBrutto float64

	cols := fakturaVATColWidths(d.usableW)
	cLp, cNazwa, cJm, cIl := cols[0], cols[1], cols[2], cols[3]
	cNetJ, cNetW, cVatP, cVatK, cBrut := cols[4], cols[5], cols[6], cols[7], cols[8]

	pdf.SetX(pdfMarginL)
	pdf.SetLineWidth(0.2)
	d.colorLine()
	pdf.SetFont(pdfFamily, "B", 7)
	pdf.SetFillColor(245, 245, 245)
	d.colorDark()
	d.fakturaVATDrawRow(cols, 7, [9]string{
		"Lp.", "Nazwa towaru lub usługi", "Jm", "Ilość", "Cena netto", "Wartość netto", "VAT %", "Kwota VAT", "Wartość brutto",
	}, [9]string{"C", "L", "C", "R", "R", "R", "R", "R", "R"}, true)

	pdf.SetFont(pdfFamily, "", 7)
	d.colorDark()
	for i, li := range q.Items {
		netto := roundCents(li.Quantity * li.UnitPrice)
		vatAmt := roundCents(netto * li.VatRate / 100)
		brutto := roundCents(netto + vatAmt)

		d.fakturaVATDrawRow(cols, 6, [9]string{
			strconv.Itoa(i + 1),
			li.Name,
			"szt.",
			formatNumber(li.Quantity),
			formatAmountComma(li.UnitPrice),
			formatAmountComma(netto),
			strconv.FormatFloat(li.VatRate, 'f', 0, 64) + "%",
			formatAmountComma(vatAmt),
			formatAmountComma(brutto),
		}, [9]string{"C", "L", "C", "R", "R", "R", "R", "R", "R"}, false)

		if g, ok := vatGroups[li.VatRate]; ok {
			g.netto += netto
			g.vat += vatAmt
			g.brutto += brutto
		} else {
			vatGroups[li.VatRate] = &vatGroup{netto, vatAmt, brutto}
		}
		totalNetto += netto
		totalVAT += vatAmt
		totalBrutto += brutto
	}
	totalNetto = roundCents(totalNetto)
	totalVAT = roundCents(totalVAT)
	totalBrutto = roundCents(totalBrutto)

	// Podsumowanie: Razem
	sumLabelW := cLp + cNazwa + cJm + cIl
	d.fakturaVATDrawSummaryRow(sumLabelW, cNetJ, cNetW, cVatP, cVatK, cBrut, [6]string{
		"Razem:", "", formatAmountComma(totalNetto), "", formatAmountComma(totalVAT), formatAmountComma(totalBrutto),
	}, true)

	// W tym: per stawka VAT
	rates := sortVatRates(vatGroups)
	for _, r := range rates {
		g := vatGroups[r]
		g.netto = roundCents(g.netto)
		g.vat = roundCents(g.vat)
		g.brutto = roundCents(g.brutto)
		d.fakturaVATDrawSummaryRow(sumLabelW, cNetJ, cNetW, cVatP, cVatK, cBrut, [6]string{
			"W tym:", "", formatAmountComma(g.netto), strconv.FormatFloat(r, 'f', 0, 64) + "%", formatAmountComma(g.vat), formatAmountComma(g.brutto),
		}, false)
	}

	pdf.SetX(pdfMarginL)
	pdf.Ln(8)

	// Razem do zapłaty + słownie
	pdf.SetX(pdfMarginL)
	pdf.SetFont(pdfFamily, "B", 12)
	d.colorDark()
	pdf.CellFormat(d.usableW, 7, "Razem do zapłaty: "+formatPLNPolish(totalBrutto), "", 1, "L", false, 0, "")
	pdf.SetFont(pdfFamily, "", 9)
	pdf.CellFormat(d.usableW, 5, "Słownie złotych: "+amountToWordsPL(totalBrutto), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Uwagi — prostokąt po prawej (stała wysokość 25 mm)
	notesY := pdf.GetY()
	const notesBoxW = 85.0
	const notesBoxH = 25.0
	notesX := d.rightX - notesBoxW
	d.colorLine()
	pdf.Rect(notesX, notesY, notesBoxW, notesBoxH, "D")
	pdf.SetXY(notesX+2, notesY+2)
	pdf.SetFont(pdfFamily, "B", 8)
	d.colorDark()
	pdf.CellFormat(notesBoxW-4, 4, "Uwagi:", "", 1, "L", false, 0, "")
	pdf.SetX(notesX + 2)
	pdf.SetFont(pdfFamily, "", 8)
	d.colorDark()
	notesText := strings.TrimSpace(q.Notes)
	if notesText == "" {
		notesText = " "
	}
	pdf.SetAutoPageBreak(false, 0)
	pdf.ClipRect(notesX+2, notesY+6, notesBoxW-4, notesBoxH-8, false)
	pdf.MultiCell(notesBoxW-4, 3.5, notesText, "", "L", false)
	pdf.ClipEnd()
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetY(notesY + notesBoxH)
	pdf.Ln(8)

	// Stopka prawna
	pdf.SetFont(pdfFamily, "", 8)
	d.colorMuted()
	pdf.MultiCell(d.usableW, 4,
		"Towar (usługa) odebrana bez zastrzeżeń ilościowych i jakościowych.",
		"", "L", false)
	pdf.Ln(6)

	// QR KSeF — po lewej pod stopką prawną
	renderKSeFQRAt(d, q, pdf.GetY())
}

func renderTransferQR(d *pdfDoc, q Quote, amount float64) {
	pdf := d.pdf
	nrb := sanitizeNRB(q.Company.BankAccount)
	if len(nrb) != 26 {
		return
	}
	var paymentTitle string
	switch q.DocType {
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
	qrContent := formatPolishPaymentQR(nrb, amount, q.Company.Name, paymentTitle)
	if png, err := generateQRPNG(qrContent); err == nil {
		pdf.Ln(10)
		yQR := pdf.GetY()
		const qrSize = 35.0

		pdf.SetXY(pdfMarginL, yQR)
		pdf.SetFont(pdfFamily, "B", 8)
		d.colorMuted()
		pdf.CellFormat(0, 4, "ZESKANUJ, ABY ZAPŁACIĆ", "", 1, "L", false, 0, "")
		pdf.Ln(1)
		pdf.SetX(pdfMarginL)
		pdf.SetFont(pdfFamily, "", 10)
		d.colorDark()
		pdf.MultiCell(d.usableW-qrSize-5, 5,
			"Numer konta: "+formatNRBWithSpaces(nrb)+"\n"+
				"Odbiorca: "+q.Company.Name+"\n"+
				"Tytuł: "+paymentTitle+"\n"+
				"Kwota: "+formatPLN(amount),
			"", "L", false)

		const qrName = "oferta_qr"
		opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(qrName, opts, bytes.NewReader(png))
		pdf.ImageOptions(qrName, d.rightX-qrSize, yQR, qrSize, qrSize, false, opts, 0, "")

		qrEnd := yQR + qrSize + 2
		if pdf.GetY() < qrEnd {
			pdf.SetY(qrEnd)
		}
	}
}

func renderKSeFQRAt(d *pdfDoc, q Quote, yStart float64) float64 {
	pdf := d.pdf
	nip := extractNIPDigits(q.Company.NIP)
	invNum := strings.TrimSpace(q.InvoiceNumber)
	if nip == "" || invNum == "" {
		return yStart
	}
	url := formatKSeFVerifyURL(nip, invNum)
	png, err := generateQRPNG(url)
	if err != nil {
		return yStart
	}
	const qrSize = 21.0 // ~60 px
	opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	const qrName = "ksef_qr"
	pdf.RegisterImageOptionsReader(qrName, opts, bytes.NewReader(png))
	pdf.ImageOptions(qrName, pdfMarginL, yStart, qrSize, qrSize, false, opts, 0, "")
	pdf.SetXY(pdfMarginL, yStart+qrSize+1)
	pdf.SetFont(pdfFamily, "", 7)
	d.colorMuted()
	pdf.CellFormat(qrSize, 4, "Zweryfikuj fakturę w KSeF", "", 0, "C", false, 0, "")
	return yStart + qrSize + 5
}

func fitLogoSize(pdf *fpdf.Fpdf, name string, maxW, maxH float64) (w, h float64) {
	info := pdf.GetImageInfo(name)
	if info == nil {
		return maxW, maxH
	}
	iw := info.Width()
	ih := info.Height()
	if iw <= 0 || ih <= 0 {
		return maxW, maxH
	}
	scale := maxH / ih
	w = iw * scale
	h = maxH
	if w > maxW {
		scale = maxW / iw
		w = maxW
		h = ih * scale
	}
	return w, h
}

func renderKSeFQR(d *pdfDoc, q Quote, _ float64) {
	renderKSeFQRAt(d, q, d.pdf.GetY()+4)
}

func formatSellerBlock(c Company) string {
	var lines []string
	if s := strings.TrimSpace(c.Name); s != "" {
		lines = append(lines, s)
	}
	if s := strings.TrimSpace(c.Address); s != "" {
		lines = append(lines, s)
	}
	if s := strings.TrimSpace(c.City); s != "" {
		lines = append(lines, s)
	}
	if s := strings.TrimSpace(c.NIP); s != "" {
		lines = append(lines, "NIP: "+s)
	}
	if s := strings.TrimSpace(c.Phone); s != "" {
		lines = append(lines, "Telefon: "+s)
	}
	if s := strings.TrimSpace(c.Email); s != "" {
		lines = append(lines, "Email: "+s)
	}
	return strings.Join(lines, "\n")
}

var nipLineRe = regexp.MustCompile(`\b\d{10}\b`)

func parseBuyerForPDF(client string) (name, addr, nip string) {
	lines := strings.Split(strings.TrimSpace(client), "\n")
	var addrParts []string
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if i == 0 {
			name = line
			continue
		}
		if nip == "" {
			if m := nipLineRe.FindString(line); m != "" {
				nip = m
				if strings.TrimSpace(strings.ReplaceAll(line, m, "")) != "" {
					addrParts = append(addrParts, line)
				}
				continue
			}
		}
		addrParts = append(addrParts, line)
	}
	addr = strings.Join(addrParts, "\n")
	return name, addr, nip
}

func extractNIPDigits(nip string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(strings.ToUpper(nip)) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatDatePL(iso string) string {
	iso = strings.TrimSpace(iso)
	if len(iso) == 10 && iso[4] == '-' && iso[7] == '-' {
		return iso[8:10] + "-" + iso[5:7] + "-" + iso[0:4]
	}
	return iso
}

func formatAmountComma(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.Replace(s, ".", ",", 1)
	return s
}

func formatPLNPolish(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	zl := int64(v)
	gr := int64(math.Round((v - float64(zl)) * 100))
	if gr >= 100 {
		zl++
		gr = 0
	}
	s := formatThousands(zl) + "," + fmt.Sprintf("%02d", gr) + " zł"
	if neg {
		return "-" + s
	}
	return s
}

func formatThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, " ")
}

func sortVatRates(groups map[float64]*vatGroup) []float64 {
	rates := make([]float64, 0, len(groups))
	for r := range groups {
		rates = append(rates, r)
	}
	for i := 0; i < len(rates)-1; i++ {
		for j := i + 1; j < len(rates); j++ {
			if rates[j] > rates[i] {
				rates[i], rates[j] = rates[j], rates[i]
			}
		}
	}
	return rates
}

type vatGroup struct{ netto, vat, brutto float64 }

// amountToWordsPL zwraca kwotę słownie w formacie: (słownie PLN XX/100).
func amountToWordsPL(amount float64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	zlote := int64(math.Floor(amount + 1e-9))
	grosze := int64(math.Round((amount-float64(zlote))*100 + 1e-9))
	if grosze >= 100 {
		zlote++
		grosze = 0
	}
	words := polishNumberWords(zlote)
	if zlote == 1 {
		words += " złoty"
	} else if polishOnes(zlote%10) == 2 || polishOnes(zlote%10) == 3 || polishOnes(zlote%10) == 4 {
		if polishTeens(zlote%100) {
			words += " złotych"
		} else {
			words += " złote"
		}
	} else {
		words += " złotych"
	}
	out := fmt.Sprintf("(%s PLN %02d/100)", words, grosze)
	if neg {
		return "minus " + out
	}
	return out
}

func polishTeens(n int64) bool {
	return n >= 11 && n <= 19
}

func polishOnes(n int64) int64 {
	if n < 0 {
		n = -n
	}
	return n % 10
}

func polishNumberWords(n int64) string {
	if n == 0 {
		return "zero"
	}
	var parts []string
	if n >= 1_000_000 {
		m := n / 1_000_000
		parts = append(parts, polishNumberWords(m)+" "+polishMillionForm(m))
		n %= 1_000_000
	}
	if n >= 1000 {
		t := n / 1000
		if t == 1 {
			parts = append(parts, "tysiąc")
		} else {
			parts = append(parts, polishNumberWordsLess1000(t)+" "+polishThousandForm(t))
		}
		n %= 1000
	}
	if n > 0 {
		parts = append(parts, polishNumberWordsLess1000(n))
	}
	return strings.Join(parts, " ")
}

func polishMillionForm(n int64) string {
	if n == 1 {
		return "milion"
	}
	if !polishTeens(n%100) && polishOnes(n%10) >= 2 && polishOnes(n%10) <= 4 {
		return "miliony"
	}
	return "milionów"
}

func polishThousandForm(n int64) string {
	if !polishTeens(n%100) && polishOnes(n%10) >= 2 && polishOnes(n%10) <= 4 {
		return "tysiące"
	}
	return "tysięcy"
}

func polishNumberWordsLess1000(n int64) string {
	if n == 0 {
		return ""
	}
	var ones = []string{"", "jeden", "dwa", "trzy", "cztery", "pięć", "sześć", "siedem", "osiem", "dziewięć"}
	var teens = []string{"dziesięć", "jedenaście", "dwanaście", "trzynaście", "czternaście", "piętnaście", "szesnaście", "siedemnaście", "osiemnaście", "dziewiętnaście"}
	var tens = []string{"", "", "dwadzieścia", "trzydzieści", "czterdzieści", "pięćdziesiąt", "sześćdziesiąt", "siedemdziesiąt", "osiemdziesiąt", "dziewięćdziesiąt"}
	var hundreds = []string{"", "sto", "dwieście", "trzysta", "czterysta", "pięćset", "sześćset", "siedemset", "osiemset", "dziewięćset"}

	var parts []string
	if n >= 100 {
		parts = append(parts, hundreds[n/100])
		n %= 100
	}
	if n >= 20 {
		parts = append(parts, tens[n/10])
		n %= 10
	}
	if n >= 10 {
		parts = append(parts, teens[n-10])
		n = 0
	}
	if n > 0 {
		if n == 1 {
			parts = append(parts, "jeden")
		} else {
			parts = append(parts, ones[n])
		}
	}
	return strings.Join(parts, " ")
}

func formatPLN(v float64) string {
	return fmt.Sprintf("%.2f zł", v)
}

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

func sanitizePhoneLink(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}
