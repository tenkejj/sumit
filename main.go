package main

import (
	"log"
	"net/http"
	"strings"
)

// utf8Middleware wymusza Content-Type z charset=utf-8 dla strony głównej
// i plików .html, żeby przeglądarki w sieci lokalnej (np. starsze Android WebView)
// nie próbowały zgadywać kodowania na podstawie heurystyk Windows-1252.
func utf8Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	indexHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})
	mux.Handle("GET /", utf8Middleware(indexHandler))

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.Handle("GET /static/", utf8Middleware(staticHandler))

	mux.HandleFunc("POST /quote", handleQuote)
	mux.HandleFunc("GET /api/nip", handleNIP)
	mux.HandleFunc("POST /api/parse", rateLimitParse(handleParse))
	mux.HandleFunc("POST /api/accept", handleAccept)
	mux.HandleFunc("GET /api/accept", handleAcceptStatus)
	mux.HandleFunc("POST /api/xml", handleXML)
	mux.HandleFunc("POST /api/track", handleTrack)
	mux.Handle("GET /stats", utf8Middleware(http.HandlerFunc(handleStatsPage)))
	mux.HandleFunc("GET /api/stats-data", handleStatsData)

	const addr = ":8080"
	log.Printf("SumIt nasłuchuje na http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("błąd serwera: %v", err)
	}
}
