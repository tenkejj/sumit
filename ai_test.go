package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGroqModelForParse(t *testing.T) {
	t.Parallel()

	text := groqModelForParse(parseRequest{Tekst: "montaż 100 zł"})
	if text != groqModelText {
		t.Fatalf("tekst: model=%q, chciano %q", text, groqModelText)
	}

	vision := groqModelForParse(parseRequest{
		Obraz:    "aGVsbG8=",
		MimeType: "image/png",
	})
	if vision != groqModelVision {
		t.Fatalf("obraz: model=%q, chciano %q", vision, groqModelVision)
	}

	both := groqModelForParse(parseRequest{
		Tekst:    "dopisek",
		Obraz:    "aGVsbG8=",
		MimeType: "image/jpeg",
	})
	if both != groqModelVision {
		t.Fatalf("tekst+obraz: model=%q, chciano %q", both, groqModelVision)
	}
}

func TestPostProcessParseItems(t *testing.T) {
	t.Parallel()

	in := []parseItem{
		{Nazwa: "  montaz m2  ", Ilosc: 0, Cena: 50},
		{Nazwa: "", Ilosc: 1, Cena: 10},
		{Nazwa: "ok", Ilosc: 2, Cena: -1},
		{Nazwa: "druga", Ilosc: 3, Cena: 0},
	}
	out := postProcessParseItems(in)
	if len(out) != 2 {
		t.Fatalf("len=%d, oczekiwano 2", len(out))
	}
	if out[0].Nazwa != "Montaz m²" || out[0].Ilosc != 1 || out[0].Cena != 50 {
		t.Fatalf("pierwsza pozycja: %+v", out[0])
	}
	if out[1].Nazwa != "Druga" || out[1].Ilosc != 3 || out[1].Cena != 0 {
		t.Fatalf("druga pozycja: %+v", out[1])
	}
}

func TestStripMarkdownJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in  string
		out string
	}{
		{
			in:  "```json\n[{\"nazwa\":\"A\",\"ilosc\":1,\"cena\":1}]\n```",
			out: "[{\"nazwa\":\"A\",\"ilosc\":1,\"cena\":1}]",
		},
		{
			in:  "[]\n```",
			out: "[]",
		},
	}
	for _, tc := range cases {
		raw := stripMarkdownJSON(tc.in)
		if raw != tc.out {
			t.Fatalf("stripMarkdownJSON(%q) = %q, chciano %q", tc.in, raw, tc.out)
		}
		var items []parseItem
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			t.Fatalf("json.Unmarshal po strip: %v (raw=%q)", err, raw)
		}
	}
}

func TestCallGroqParse_modelSelection(t *testing.T) {
	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req groqChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		gotModel = req.Model
		_ = json.NewEncoder(w).Encode(groqChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `[{"nazwa":"Test","ilosc":1,"cena":10}]`}},
			},
		})
	}))
	defer srv.Close()

	prevURL := groqAPIEndpoint
	prevClient := httpClientGroq
	groqAPIEndpoint = srv.URL
	httpClientGroq = srv.Client()
	t.Cleanup(func() {
		groqAPIEndpoint = prevURL
		httpClientGroq = prevClient
	})

	png1x1 := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	t.Run("tekst", func(t *testing.T) {
		gotModel = ""
		_, err := callGroqParse(context.Background(), "test-key", parseRequest{Tekst: "usługa 10 zł"})
		if err != nil {
			t.Fatalf("callGroqParse: %v", err)
		}
		if gotModel != groqModelText {
			t.Fatalf("model=%q, chciano %q", gotModel, groqModelText)
		}
	})

	t.Run("obraz", func(t *testing.T) {
		gotModel = ""
		_, err := callGroqParse(context.Background(), "test-key", parseRequest{
			Obraz:    base64.StdEncoding.EncodeToString(png1x1),
			MimeType: "image/png",
		})
		if err != nil {
			t.Fatalf("callGroqParse: %v", err)
		}
		if gotModel != groqModelVision {
			t.Fatalf("model=%q, chciano %q", gotModel, groqModelVision)
		}
	})
}

func TestHandleParse_brakKluczaAPI(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "")

	req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(`{"tekst":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleParse(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleParse_trybKlient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = json.NewEncoder(w).Encode(groqChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"nazwa":"ABC Sp. z o.o.","nip":"1234567890","adres":"ul. Testowa 1, 00-001 Warszawa"}`}},
			},
		})
	}))
	defer srv.Close()

	prevURL := groqAPIEndpoint
	prevClient := httpClientGroq
	groqAPIEndpoint = srv.URL
	httpClientGroq = srv.Client()
	t.Cleanup(func() {
		groqAPIEndpoint = prevURL
		httpClientGroq = prevClient
	})

	t.Setenv("GROQ_API_KEY", "test-key")

	body := `{"tekst":"ABC Sp. z o.o. NIP 1234567890","tryb":"klient"}`
	req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleParse(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}

	var resp parseClientResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v, body=%s", err, rec.Body.String())
	}
	if resp.Client.Nazwa != "ABC Sp. z o.o." {
		t.Fatalf("nazwa=%q, chciano ABC Sp. z o.o.", resp.Client.Nazwa)
	}
	if resp.Client.NIP != "1234567890" {
		t.Fatalf("nip=%q", resp.Client.NIP)
	}
}

func TestHandleParse_sukcesTekst(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = json.NewEncoder(w).Encode(groqChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `[{"nazwa":"Montaż","ilosc":2,"cena":99}]`}},
			},
		})
	}))
	defer srv.Close()

	prevURL := groqAPIEndpoint
	prevClient := httpClientGroq
	groqAPIEndpoint = srv.URL
	httpClientGroq = srv.Client()
	t.Cleanup(func() {
		groqAPIEndpoint = prevURL
		httpClientGroq = prevClient
	})

	t.Setenv("GROQ_API_KEY", "test-key")

	req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(`{"tekst":"montaż 99 zł x2"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleParse(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", rec.Code, rec.Body.String())
	}

	var items []parseItem
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 1 || items[0].Nazwa != "Montaż" {
		t.Fatalf("items=%+v", items)
	}
}

func TestLiveGroqIntegration(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	if apiKey == "" {
		t.Skip("pominięto: brak GROQ_API_KEY")
	}

	t.Run("pozycje_tekst", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(`{"tekst":"montaż hydrauliki 2 szt po 150 zł, materiały 500 zł"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handleParse(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		var items []parseItem
		if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(items) == 0 {
			t.Fatal("brak pozycji")
		}
		t.Logf("pozycje: %+v", items)
	})

	t.Run("klient_tekst", func(t *testing.T) {
		body := `{"tekst":"ABC Sp. z o.o., NIP 5261040828, ul. Marszałkowska 1, Warszawa","tryb":"klient"}`
		req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handleParse(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		var resp parseClientResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Client.Nazwa == "" && resp.Client.NIP == "" {
			t.Fatalf("pusty klient: %+v", resp.Client)
		}
		t.Logf("klient: %+v", resp.Client)
	})

	t.Run("klient_obraz", func(t *testing.T) {
		img, err := os.ReadFile("testdata/wizytowka.png")
		if err != nil {
			t.Skip("pominięto: brak testdata/wizytowka.png (wygeneruj skryptem testowym)")
		}
		body := fmt.Sprintf(`{"obraz":%q,"mime_type":"image/png","tryb":"klient"}`, base64.StdEncoding.EncodeToString(img))
		req := httptest.NewRequest(http.MethodPost, "/api/parse", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handleParse(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		var resp parseClientResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Client.NIP == "" && resp.Client.Nazwa == "" {
			t.Fatalf("pusty klient z obrazu: %+v", resp.Client)
		}
		t.Logf("klient z obrazu: %+v", resp.Client)
	})
}
