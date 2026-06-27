<p align="center">
  <img src="static/images/readme-banner.png" alt="SumIt" width="100%">
</p>

<h3 align="center">PDF estimates for tradespeople. No signup.</h3>

<p align="center">
  <a href="https://www.sum-it.app/">https://www.sum-it.app/</a>
</p>

<p align="center">
  <a href="#english">English</a> · <a href="#polski">Polski</a>
</p>

---

<a id="english"></a>

## English

Fill in a form, get a PDF estimate. SumIt is built for Polish electricians, plumbers, and small service crews who need a proper estimate on site — not another SaaS with a 20-step onboarding.

No account. No database. Your drafts and client history stay in the browser (`localStorage`). The Go server renders PDFs, proxies the MF White List and Groq AI APIs, and can export KSeF helper XML — no CORS headaches on NIP lookup.

**Live:** [https://www.sum-it.app/](https://www.sum-it.app/) — Oracle Cloud (Ubuntu 24.04), nginx + Let's Encrypt, `./deploy.sh` + systemd.

### What you get

- PDF estimates with seller/client blocks, line items, totals, notes, optional logo
- **VAT invoices** — separate PDF template (9-column table, VAT breakdown, amount in words, KSeF verification QR)
- Bank transfer QR code (ZBP standard) when you set an account number; toggle on/off in Company settings
- NIP lookup → company name and address from the Ministry of Finance White List
- **AI assistant** — paste a client message, dictate, or photograph a note → suggested line items (Groq)
- Client autocomplete from your estimate history
- Service catalog, CSV import/export, profit estimate per line item
- Stats dashboard (KPI, chart, top clients, activity heatmap), full estimate history
- **Mobile app UX** — home screen, 3-step wizard, bottom tabs (Estimate / History / Company / Stats), PWA install
- Share PDF with the client (Web Share API), copy a review link, client acceptance flow
- **KSeF helper** — download FA(3) XML for manual upload to KSeF (not e-invoicing itself)
- App settings: theme, default validity, default document type & VAT rate, local data reset, contact
- Light/dark mode, live PDF preview on desktop

### Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.22+, `net/http`, [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | HTML, CSS, vanilla JS — no npm, no bundler |

### Project layout

```
sumit/
├── *.go                 # Go backend (HTTP, PDF, API)
├── assets/              # Embedded fonts, KSeF XSD
├── static/
│   ├── index.html       # App entry
│   ├── css/             # style.css, style-mobile.css
│   ├── js/              # app.js, sw.js
│   ├── images/          # logos, readme banner
│   └── icons/           # favicon, PWA icons
├── deploy/              # deploy.sh, DEPLOY.md, sumit.service
├── scripts/             # maintainer utilities (PWA icons)
└── testdata/            # test fixtures
```

### Run locally

```bash
git clone https://github.com/tenkejj/sumit.git
cd sumit
go mod tidy
go run .
```

Open http://localhost:8080

Maintainers: see [deploy/DEPLOY.md](deploy/DEPLOY.md) for VPS setup.

---

<a id="polski"></a>

## Polski

Wypełniasz formularz, dostajesz PDF z wyceną. SumIt jest pod hydraulików, elektryków i małe ekipy remontowe — ludzi, którzy potrzebują wyceny u klienta, a nie kolejnej aplikacji z rejestracją i panelem admina.

Bez konta. Bez bazy danych. Szkice i historia klientów siedzą w przeglądarce (`localStorage`). Serwer Go generuje PDF, proxy do Białej Listy MF i API Groq (AI) oraz eksportuje pomocniczy XML KSeF — bez problemów z CORS przy NIP.

**Produkcja:** [https://www.sum-it.app/](https://www.sum-it.app/) — Oracle Cloud (Ubuntu 24.04), nginx + Let's Encrypt, wdrożenie przez `./deploy.sh` + systemd.

### Co potrafi

- PDF z danymi sprzedawcy i klienta, pozycjami, sumą, uwagami, opcjonalnym logo
- **Faktura VAT** — osobny szablon PDF (tabela 9 kolumn, rozbicie VAT, kwota słownie, QR weryfikacji KSeF)
- Kod QR do przelewu (standard ZBP), gdy ustawisz numer konta; włącz/wyłącz w Moja firma
- Pobieranie firmy po NIP z Białej Listy Ministerstwa Finansów
- **Asystent AI** — wklej wiadomość od klienta, dyktuj lub sfotografuj notatkę → propozycja pozycji (Groq)
- Autouzupełnianie klientów z historii wycen
- Katalog usług, import/eksport CSV, szacowany zysk na pozycji
- Statystyki (KPI, wykres, top klienci, heatmapa aktywności), pełna historia wycen
- **Mobile** — ekran startowy, kreator 3 kroków, dolne zakładki (Wycena / Historia / Firma / Statystyki), instalacja PWA
- Udostępnianie PDF klientowi (Web Share API), link do podglądu, akceptacja wyceny przez klienta
- **Pomocnik KSeF** — pobieranie XML FA(3) do ręcznego wgrania w KSeF (to nie jest samo wysyłanie e-faktury)
- Ustawienia aplikacji: motyw, domyślna ważność, typ dokumentu i stawka VAT, czyszczenie danych lokalnych, kontakt
- Tryb jasny/ciemny, podgląd PDF na żywo na desktopie

### Stack

| Warstwa | Technologie |
|---|---|
| Backend | Go 1.22+, `net/http`, [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | HTML, CSS, vanilla JS — bez npm, bez bundlera |

### Struktura projektu

```
sumit/
├── *.go                 # backend Go (HTTP, PDF, API)
├── assets/              # czcionki embed, schematy KSeF
├── static/
│   ├── index.html       # entry aplikacji
│   ├── css/             # style.css, style-mobile.css
│   ├── js/              # app.js, sw.js
│   ├── images/          # logo, banner README
│   └── icons/           # favicon, ikony PWA
├── deploy/              # deploy.sh, DEPLOY.md, sumit.service
├── scripts/             # narzędzia maintainera (ikony PWA)
└── testdata/            # pliki testowe
```

### Uruchomienie lokalne

```bash
git clone https://github.com/tenkejj/sumit.git
cd sumit
go mod tidy
go run .
```

Wejdź na http://localhost:8080

Wdrożenie na VPS: [deploy/DEPLOY.md](deploy/DEPLOY.md).
