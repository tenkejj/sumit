package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	mux.HandleFunc("POST /oferta", handleOferta)

	const addr = ":8080"
	log.Printf("SumIt nasłuchuje na http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("błąd serwera: %v", err)
	}
}
