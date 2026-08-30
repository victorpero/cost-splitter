# Architecture

Cost Splitter is split into two independently runnable services:

- `frontend/` is a Next.js application responsible for presentation, routing,
  interaction state, and proxying browser requests to the configured API URL.
- `cmd/cost-splitter-api` is a standalone Go HTTP API. It owns CSV normalization,
  merchant matching, selection rules, and allocation calculations.

The frontend proxy reads `API_BASE_URL` at runtime and forwards requests without
embedding domain behavior. Other browser, mobile, CLI, or automation clients can
call the Go API directly. Direct browser clients can be admitted with the
comma-separated `API_ALLOWED_ORIGINS` configuration.

Within the Go service, `internal/api` owns HTTP transport concerns,
`internal/application` coordinates use cases, and the existing parser, matcher,
split, and transaction packages contain domain behavior. No persistence is needed:
uploaded files and calculation state remain in memory on the client/request path.

The versioned contract is documented in [openapi.yaml](openapi.yaml). Monetary
values are signed integer cents. Participant results are integer half-cents so a
single odd cent can be divided without rounding or floating-point arithmetic.
