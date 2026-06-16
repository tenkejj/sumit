<p align="center">
  <img src="static/readme-banner.png" alt="SumIt" width="100%">
</p>

<h3 align="center">PDF estimates for tradespeople. No signup.</h3>

<p align="center">
  <a href="#english">English</a> · <a href="#polski">Polski</a>
</p>

---

<a id="english"></a>

## English

Fill in a form, get a PDF estimate. SumIt is built for Polish electricians, plumbers, and small service crews who need a proper estimate on site — not another SaaS with a 20-step onboarding.

No account. No database. Your drafts and client history stay in the browser (`localStorage`). The Go server only renders PDFs and proxies the MF White List API so NIP lookup works without CORS headaches.

**Live:** [http://130.61.35.204:8080](http://130.61.35.204:8080) — Oracle Cloud (Ubuntu 24.04), deployed via `deploy.sh` + systemd.

### What you get

- PDF estimates with seller/client blocks, line items, totals, notes, optional logo
- Bank transfer QR code (ZBP standard) when you set an account number
- NIP lookup → company name and address from the Ministry of Finance White List
- Client autocomplete from your estimate history
- Service catalog, CSV import/export, profit estimate per line item
- Stats dashboard, estimate history, light/dark mode, PWA install on mobile

### Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.22+, `net/http`, [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | HTML, CSS, vanilla JS — no npm, no bundler |

### Run locally

```bash
git clone https://github.com/tenkejj/sumit.git
cd sumit
go mod tidy
go run .
```

Open http://localhost:8080

Maintainers: see [DEPLOY.md](DEPLOY.md) for VPS setup.

---

<a id="polski"></a>

## Polski

Wypełniasz formularz, dostajesz PDF z wyceną. SumIt jest pod hydraulików, elektryków i małe ekipy remontowe — ludzi, którzy potrzebują wyceny u klienta, a nie kolejnej aplikacji z rejestracją i panelem admina.

Bez konta. Bez bazy danych. Szkice i historia klientów siedzą w przeglądarce (`localStorage`). Serwer Go robi dwie rzeczy: generuje PDF i proxy do Białej Listy MF, żeby pobieranie danych po NIP działało bez problemów z CORS.

**Produkcja:** [http://130.61.35.204:8080](http://130.61.35.204:8080) — Oracle Cloud (Ubuntu 24.04), wdrożenie przez `deploy.sh` + systemd.

### Co potrafi

- PDF z danymi sprzedawcy i klienta, pozycjami, sumą, uwagami, opcjonalnym logo
- Kod QR do przelewu (standard ZBP), gdy ustawisz numer konta
- Pobieranie firmy po NIP z Białej Listy Ministerstwa Finansów
- Autouzupełnianie klientów z historii wycen
- Katalog usług, import/eksport CSV, szacowany zysk na pozycji
- Statystyki, historia wycen, tryb jasny/ciemny, instalacja PWA na telefonie

### Stack

| Warstwa | Technologie |
|---|---|
| Backend | Go 1.22+, `net/http`, [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | HTML, CSS, vanilla JS — bez npm, bez bundlera |

### Uruchomienie lokalne

```bash
git clone https://github.com/tenkejj/sumit.git
cd sumit
go mod tidy
go run .
```

Wejdź na http://localhost:8080

Wdrożenie na VPS: [DEPLOY.md](DEPLOY.md).
