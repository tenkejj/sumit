<p align="center">
  <img src="static/logoo.png" alt="SumIt Logo" width="320" style="background-color: #ffffff; padding: 24px; border-radius: 12px;">
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
