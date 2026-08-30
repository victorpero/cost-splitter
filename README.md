# Cost Splitter

Cost Splitter is a Next.js web app, reusable Go HTTP API, and command-line tool
for splitting shared expenses between two people. It currently includes an
American Express CSV import for Swedish grocery transactions, while keeping the
API and core cost-splitting workflow reusable by additional clients, sources,
and expense types.

## Features

- Next.js web UI for uploading and reviewing CSV files
- Independently deployable, versioned Go API
- OpenAPI-documented import and split-calculation contract
- American Express CSV import with UTF-8, comma, and semicolon support
- Matching by configurable store or merchant prefixes
- Signed amounts that preserve refunds and credits
- Even, participant-one, or participant-two allocation per transaction
- CLI support for the same core CSV processing workflow

## Running locally

Start the Go API in one terminal:

```sh
API_ADDR=127.0.0.1:8080 go run ./cmd/cost-splitter-api
```

Start the Next.js frontend in another terminal:

```sh
cd frontend
npm ci
API_BASE_URL=http://127.0.0.1:8080 npm run dev
```

Open <http://localhost:3000>. The web UI lets you upload one or more CSV files,
edit merchant prefixes, review matched and unmatched transactions, adjust the
signed amount to split, and choose an allocation for every included row.

To run both independently built services with containers:

```sh
docker compose up --build
```

Uploaded files are processed in memory and are not stored.

## Go API

The standalone API exposes:

- `GET /healthz` and `GET /readyz` for deployment probes
- `GET /api/v1/defaults` for default currency and merchant prefixes
- `POST /api/v1/imports/amex` for multipart AmEx CSV normalization
- `POST /api/v1/split-calculations` for matching and allocation calculations

The complete contract is in [`docs/openapi.yaml`](docs/openapi.yaml), and the
service boundaries are described in [`docs/architecture.md`](docs/architecture.md).
All API errors use an `error.code` and `error.message` JSON object. Monetary
values are integer cents; participant totals are integer half-cents.

API configuration is environment-driven:

| Variable | Default | Purpose |
| --- | --- | --- |
| `API_ADDR` | `127.0.0.1:8080` | API listen address |
| `API_CURRENCY` | `SEK` | Default currency returned to clients |
| `API_ALLOWED_ORIGINS` | Local frontend origins | Comma-separated direct browser clients; `*` allows any origin |

The frontend reads `API_BASE_URL` at runtime and proxies browser requests to the
Go service. It defaults to `http://127.0.0.1:8080`.

## Running from the CLI

Process one or more CSV files:

```sh
go run ./cmd/cost-splitter transactions.csv
go run ./cmd/cost-splitter file1.csv file2.csv
```

Show transactions that did not match a merchant prefix:

```sh
go run ./cmd/cost-splitter --show-unmatched transactions.csv
```

Use a custom prefix file:

```sh
go run ./cmd/cost-splitter --stores config/grocery_stores.txt transactions.csv
```

The store file contains one prefix per line. Blank lines and lines beginning
with `#` are ignored. The default Swedish grocery prefixes are HEMKOP, ICA,
MAXI ICA, WILLYS, COOP, and PRESSBYRÅN.

## Amounts and matching

Imported amounts retain their positive or negative sign. In the web UI,
**Amount to split** can be edited to enter the signed remainder after a
repayment; the original imported amount remains visible for reference.

Matching is case-insensitive, trims whitespace, handles Swedish letters such as
Å, Ä, and Ö, and checks the beginning of each transaction description. The CLI
splits matched transactions evenly; the web UI also supports assigning a row
entirely to either participant.

## CSV support

The parser accepts common English and Swedish column names for transaction date,
description or name, and amount. It supports Swedish decimal formats such as
`123,45`, English formats such as `123.45`, and UTF-8 files with an optional
BOM.

## Development

Run the required local checks:

```sh
test -z "$(gofmt -l $(git ls-files '*.go'))"
go vet ./...
go test -run '^$' ./...
go test ./...
go build -trimpath ./...
cd frontend && npm ci && npm run typecheck && npm run build && cd ..
docker build --tag cost-splitter:ci .
docker build --file Dockerfile.api --tag cost-splitter-api:ci .
```

Changes follow the repository promotion flow: create a feature branch from
`dev`, merge the feature branch into `dev` after checks pass, then promote the
verified `dev` branch to `main` through a pull request.

Production images are published as `ghcr.io/victorpero/cost-splitter` for the
frontend and `ghcr.io/victorpero/cost-splitter-api` for the API. Both packages
link back to <https://github.com/victorpero/cost-splitter> through OCI metadata.
