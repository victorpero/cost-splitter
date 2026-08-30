# Cost Splitter

Cost Splitter is a local web app and command-line tool for splitting shared
expenses between two people. It currently includes an American Express CSV
import for Swedish grocery transactions, while keeping the core cost-splitting
workflow usable for additional sources and expense types.

## Features

- Local web UI for uploading and reviewing CSV files
- American Express CSV import with UTF-8, comma, and semicolon support
- Matching by configurable store or merchant prefixes
- Signed amounts that preserve refunds and credits
- Even, participant-one, or participant-two allocation per transaction
- CLI support for the same core CSV processing workflow

## Running the local web UI

Start the web UI:

```sh
go run ./cmd/cost-splitter-web
```

Then open <http://localhost:8080>. The web UI lets you upload one or more CSV
files, edit merchant prefixes, review matched and unmatched transactions, and
choose an allocation for each included transaction.

By default, the server listens only on your own machine at `127.0.0.1:8080`.
For a container or home-server deployment, bind it to all interfaces:

```sh
go run ./cmd/cost-splitter-web -addr 0.0.0.0:8080
```

Uploaded files are processed in memory and are not stored.

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
docker build --tag cost-splitter:ci .
```

Changes follow the repository promotion flow: create a feature branch from
`dev`, merge the feature branch into `dev` after checks pass, then promote the
verified `dev` branch to `main` through a pull request.

The source repository is <https://github.com/victorpero/cost-splitter>.
