package main

import (
	"math"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

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

// truncateField przycina łańcuch do co najwyżej max znaków (runów),
// z uwzględnieniem polskich znaków diakrytycznych.
func truncateField(s string, max int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max])
	}
	return s
}

// formatPolishPaymentQR buduje payload QR zgodny z Rekomendacją Związku Banków
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
func formatPolishPaymentQR(nrb string, amountPLN float64, name, title string) string {
	amountGrosze := int64(math.Round(amountPLN * 100))
	if amountGrosze < 0 {
		amountGrosze = 0
	}
	return strings.Join([]string{
		"",
		"PL",
		nrb,
		strconv.FormatInt(amountGrosze, 10),
		truncateField(name, 20),
		truncateField(title, 32),
		"",
		"",
		"",
	}, "|")
}

// generateQRPNG zwraca obraz PNG kodu QR (256 px) z poziomem korekcji błędów M.
func generateQRPNG(content string) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, 256)
}

// formatNRBWithSpaces zwraca 26-cyfrowy NRB w czytelnej postaci
// "CC RRRR RRRR RRRR RRRR RRRR RRRR" (jak w drukowanym formacie polskim).
// Wejście, które nie jest dokładnie 26 cyfr, zwraca bez zmian.
func formatNRBWithSpaces(nrb string) string {
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
