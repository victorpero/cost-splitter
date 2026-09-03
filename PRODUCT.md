# Product

## Platform

web

## Users

The primary user imports one or more American Express CSV exports and needs to identify grocery purchases, adjust which transactions count, and calculate how much each person owes.

## Product Purpose

Cost Splitter turns Amex transaction exports into a reviewable grocery-cost split. Success means the user can reproduce the existing matching and splitting workflow through a browser while the calculation remains available to other clients through a versioned API.

## Positioning

The product combines explainable, editable grocery matching with exact per-transaction allocation. Imported charges remain visible and reversible instead of being reduced to an opaque total.

## Operating Context

Users work from CSV files exported by American Express. They may import several files, review automatically matched and unmatched transactions, override split amounts, allocate individual charges, include or remove transactions, and copy the resulting totals into their household settlement workflow.

## Capabilities and Constraints

- Preserve the established CSV parsing, grocery matching, currency, allocation, and half-cent behavior.
- Keep the Next.js frontend and Go API independently deployable.
- Expose reusable, versioned JSON endpoints under `/api/v1` with structured errors.
- Configure network and deployment concerns through environment variables.
- Do not fabricate transactions, merchant evidence, or deployment status.

## Engineering Basis

- Existing parser, matcher, and split-domain tests define calculation behavior.
- The frontend/API separation and operational requirements are documented in
  [the architecture guide](docs/architecture.md) and the API contract.
- No testimonials, customer logos, or usage benchmarks are available and none should be invented.

## Product Principles

- Make every included amount inspectable and reversible.
- Preserve exact monetary behavior across interfaces.
- Keep imports local to the requested calculation flow; do not imply storage without evidence.
- Let the API remain useful independently of the browser interface.

## Accessibility & Inclusion

The web interface must support keyboard operation, visible focus, semantic controls, sufficient text contrast, reduced motion preferences, and a usable single-column mobile layout.
