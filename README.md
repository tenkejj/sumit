<p align="center">
  <img src="static/readme-banner.png" alt="SumIt Logo" width="100%">
</p>

# SumIt - Your Valuation Standard

SumIt is a modern, high-performance web application designed for instantaneous generation of commercial offers and valuations in PDF format. Built with a stateless Go backend and a pure Vanilla JS fat client, the application delivers exceptional performance, zero-overhead user onboarding, and seamless cross-device compatibility, operating completely within the browser environment as a Progressive Web App (PWA).

## Target Audience and Market

SumIt is specifically engineered for the Polish B2B and B2C services market. The primary user base includes independent contractors, craftsmen such as electricians, plumbers, and HVAC technicians, as well as field service professionals who require the capability to generate professional cost estimates immediately on-site. The solution also serves small to medium enterprise service providers looking for an agile alternative to traditional, complex accounting software or error-prone spreadsheet templates. By eliminating mandatory user registration and complex configuration pipelines, SumIt addresses the critical market need for rapid administrative workflows among trade professionals.

## Key Features

The system offers real-time PDF generation, which provides high-fidelity, mathematically precise layout compilation executed entirely on the server side in milliseconds. Integration with the Polish Ministry of Finance White List API is handled via a dedicated Go proxy, which successfully bypasses browser CORS restrictions to automate company data retrieval. The embedded Mini-CRM manages client metadata indexing and historical deduplication through local browser storage, featuring an advanced, keyboard-navigable custom auto-complete interface. The adaptive visual engine ensures full compliance with modern interface design benchmarks by supporting light and dark modes through CSS custom properties and dynamic real-time graphics filtering for asset localization. The entire architecture relies on zero infrastructure overhead, remaining independent of traditional database engines by offloading historical state management to client-side sandboxes.

## Architecture Overview

SumIt operates as a decentralized, stateless system. The frontend functions as a fat client built entirely using HTML5, CSS3, and modern ECMAScript Vanilla JavaScript, maintaining the application state, drafts, and transaction history within the client's localized security sandbox. The backend acts as a stateless engine, consisting of a lightweight server compiled in Go using standard library components and the performance-optimized go-pdf/fpdf library. This engine handles immutable processing operations, including PDF rendering and upstream API routing, maintaining zero local persistence to guarantee horizontal scalability.

## Technical Stack

The backend engine utilizes Go 1.22+, net/http, and go-pdf/fpdf. The frontend framework consists of native Vanilla JavaScript, HTML5, and CSS3, maintaining complete independence from third-party framework dependencies.

## Quick Start

Execute the following commands in your terminal environment to initialize the application locally:

```bash
git clone https://github.com/tenkejj/sumit.git
cd sumit
go mod tidy
go run .
```

Once initialized, navigate to the local environment instance at: http://localhost:8080

## Wdrożenie na VPS (Linux x86_64)

Poniższa instrukcja opisuje **jednorazową** konfigurację serwera oraz powtarzalne wdrożenia z maszyny deweloperskiej.

Aplikacja nasłuchuje na porcie **8080** (`http://130.61.35.204:8080` po wdrożeniu).

### Wymagania

| Środowisko | Wymagania |
|---|---|
| Maszyna deweloperska | Go 1.22+, `bash`, `ssh`, `scp` (np. WSL lub Git Bash na Windows) |
| Serwer VPS | Ubuntu 24.04 (Oracle Cloud Free Tier), użytkownik `ubuntu`, dostęp SSH |

Go na serwerze **nie jest wymagane** — binarka jest kompilowana lokalnie i przesyłana przez `deploy.sh`.

### Klucz SSH (`SSH_KEY`)

W pliku `deploy.sh` ustaw zmienną `SSH_KEY` na ścieżkę do klucza prywatnego SSH. Skrypt używa jej we wszystkich wywołaniach `scp` i `ssh` (flaga `-i`):

```bash
readonly SSH_KEY="/ścieżka/do/twojego/klucza.key"
```

Na Windows (Git Bash) ścieżka w stylu `/c/Users/...` jest poprawna. Przykładowe komendy ręczne poniżej zakładają tę samą zmienną:

```bash
export SSH_KEY="/ścieżka/do/twojego/klucza.key"
```

### Jednorazowa konfiguracja serwera

Zaloguj się na serwer:

```bash
ssh -i "${SSH_KEY}" ubuntu@130.61.35.204
```

Utwórz katalog aplikacji i nadaj uprawnienia:

```bash
sudo mkdir -p /opt/sumit/static /opt/sumit/assets
sudo chown -R ubuntu:ubuntu /opt/sumit
```

Skopiuj plik usługi systemd (z maszyny lokalnej, w katalogu projektu):

```bash
scp -i "${SSH_KEY}" sumit.service ubuntu@130.61.35.204:/tmp/sumit.service
```

Na serwerze zainstaluj i włącz usługę:

```bash
sudo mv /tmp/sumit.service /etc/systemd/system/sumit.service
sudo systemctl daemon-reload
sudo systemctl enable sumit
sudo systemctl start sumit
```

Opcjonalnie otwórz port HTTP w firewallu:

```bash
sudo ufw allow 8080/tcp
```

### Wdrożenie aplikacji

Upewnij się, że `SSH_KEY` w `deploy.sh` wskazuje na właściwy klucz. Z katalogu głównego projektu na maszynie deweloperskiej:

```bash
chmod +x deploy.sh
./deploy.sh
```

Skrypt:

1. Kompiluje binarkę pod Linux (`GOOS=linux GOARCH=amd64`).
2. Przesyła ją na serwer przez `scp -i "${SSH_KEY}"` do `/usr/local/bin/sumit`.
3. Aktualizuje katalogi `/opt/sumit/static` i `/opt/sumit/assets`.
4. Restartuje usługę `sumit` (jeśli jest już włączona).

### Weryfikacja

```bash
# status usługi
ssh -i "${SSH_KEY}" ubuntu@130.61.35.204 'sudo systemctl status sumit'

# logi
ssh -i "${SSH_KEY}" ubuntu@130.61.35.204 'journalctl -u sumit -f'
```

Aplikacja dostępna pod adresem: **http://130.61.35.204:8080**
