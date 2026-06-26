package main

import (
	"bytes"
	"compress/zlib"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestAmountToWordsPL(t *testing.T) {
	cases := []struct {
		amount float64
		want   string
	}{
		{0, "(zero złotych PLN 00/100)"},
		{1, "(jeden złoty PLN 00/100)"},
		{2, "(dwa złote PLN 00/100)"},
		{14, "(czternaście złotych PLN 00/100)"},
		{23, "(dwadzieścia trzy złote PLN 00/100)"},
		{100, "(sto złotych PLN 00/100)"},
		{999, "(dziewięćset dziewięćdziesiąt dziewięć złotych PLN 00/100)"},
		{1000, "(tysiąc złotych PLN 00/100)"},
		{1234.56, "(tysiąc dwieście trzydzieści cztery złote PLN 56/100)"},
		{1000000, "(jeden milion złotych PLN 00/100)"},
	}
	for _, tc := range cases {
		got := amountToWordsPL(tc.amount)
		if got != tc.want {
			t.Errorf("amountToWordsPL(%v) = %q, want %q", tc.amount, got, tc.want)
		}
	}
}

func TestFormatKSeFVerifyURL(t *testing.T) {
	got := formatKSeFVerifyURL("PL123-456-78-90", "FV/2026/001")
	want := "https://ksef.podatki.gov.pl/web/verify?nip=1234567890&numer=FV%2F2026%2F001"
	if got != want {
		t.Errorf("formatKSeFVerifyURL() = %q, want %q", got, want)
	}
}

func TestFormatDatePL(t *testing.T) {
	got := formatDatePL("2026-06-25")
	if got != "25-06-2026" {
		t.Errorf("formatDatePL() = %q, want 25-06-2026", got)
	}
}

func TestParseBuyerForPDF(t *testing.T) {
	name, addr, nip := parseBuyerForPDF("ACME Sp. z o.o.\nul. Testowa 1\n00-001 Warszawa\nNIP: 9876543210")
	if name != "ACME Sp. z o.o." {
		t.Errorf("name = %q", name)
	}
	if nip != "9876543210" {
		t.Errorf("nip = %q", nip)
	}
	if !strings.Contains(addr, "Testowa") {
		t.Errorf("addr = %q", addr)
	}
}

func TestFakturaVATColWidths(t *testing.T) {
	const usableW = 180.0
	cols := fakturaVATColWidths(usableW)
	var sum float64
	for _, w := range cols {
		if w <= 0 {
			t.Fatalf("column width must be positive, got %v", cols)
		}
		sum += w
	}
	if sum < usableW-0.01 || sum > usableW+0.01 {
		t.Fatalf("column widths sum = %v, want %v", sum, usableW)
	}
}

func decompressPDFStreams(pdf []byte) string {
	var out strings.Builder
	needle := []byte("stream\n")
	endMark := []byte("\nendstream")
	for i := 0; i < len(pdf); {
		idx := bytes.Index(pdf[i:], needle)
		if idx < 0 {
			break
		}
		start := i + idx + len(needle)
		end := bytes.Index(pdf[start:], endMark)
		if end < 0 {
			break
		}
		data := pdf[start : start+end]
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			i = start + end
			continue
		}
		dec, err := io.ReadAll(r)
		r.Close()
		if err == nil {
			out.Write(dec)
			out.WriteByte('\n')
		}
		i = start + end
	}
	return out.String()
}

func TestFakturaVATPDF_TableRightEdgeMM(t *testing.T) {
	q := validFakturaVAT()
	var buf bytes.Buffer
	if err := GeneratePDF(q, &buf); err != nil {
		t.Fatalf("GeneratePDF: %v", err)
	}
	content := decompressPDFStreams(buf.Bytes())
	re := regexp.MustCompile(`([\d.]+) ([\d.]+) ([\d.]+) ([\d.-]+) re`)
	var maxRightPt float64
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		x, err1 := strconv.ParseFloat(m[1], 64)
		w, err2 := strconv.ParseFloat(m[3], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if right := x + w; right > maxRightPt {
			maxRightPt = right
		}
	}
	maxRightMM := maxRightPt * 25.4 / 72
	wantRight := pdfPageW - pdfMarginR
	t.Logf("max cell right edge: %.2f mm (want %.2f mm)", maxRightMM, wantRight)
	if maxRightMM < wantRight-1 {
		t.Fatalf("table does not reach right margin: %.2f mm < %.2f mm", maxRightMM, wantRight)
	}
}

func TestGeneratePDF_FakturaVAT_Smoke(t *testing.T) {
	q := validFakturaVAT()
	var buf bytes.Buffer
	if err := GeneratePDF(q, &buf); err != nil {
		t.Fatalf("GeneratePDF faktura_vat: %v", err)
	}
	if buf.Len() < 1000 {
		t.Fatalf("PDF too small: %d bytes", buf.Len())
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Fatal("output is not a PDF")
	}
}

func TestGeneratePDF_Oferta_HideQR(t *testing.T) {
	q := validQuote()
	q.Company.BankAccount = "12 3456 7890 1234 5678 9012 3456"
	q.HideQR = true

	var withQR, withoutQR bytes.Buffer
	qHide := q
	qHide.HideQR = true
	qShow := q
	qShow.HideQR = false

	if err := GeneratePDF(qHide, &withoutQR); err != nil {
		t.Fatalf("GeneratePDF hide_qr: %v", err)
	}
	if err := GeneratePDF(qShow, &withQR); err != nil {
		t.Fatalf("GeneratePDF show qr: %v", err)
	}
	if withoutQR.Len() >= withQR.Len() {
		t.Errorf("expected smaller PDF without QR: hide=%d show=%d", withoutQR.Len(), withQR.Len())
	}
}

func TestShouldShowTransferQR(t *testing.T) {
	if !(Quote{}).ShouldShowTransferQR() {
		t.Error("default should show QR")
	}
	q := Quote{HideQR: true}
	if q.ShouldShowTransferQR() {
		t.Error("HideQR=true should hide QR")
	}
}
