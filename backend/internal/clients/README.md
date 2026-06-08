# internal/clients/

External-service HTTP clients. One folder per third-party API (Shopify, ERP, OMS, payment gateway, …).

## Why isolate clients

- Keeps **retry/backoff/circuit-breaker** logic out of services.
- Lets you swap real clients with mocks/fakes during tests by injecting interfaces.
- Concentrates **auth flows** (API keys, JWT, OAuth) and **request signing** in one place.
- Failure of an external partner doesn't crash unrelated modules — clients return typed errors.

## Standard client layout

```
internal/clients/<partner>/
├── README.md          # What this partner does + auth/credentials
├── client.go          # Constructor, `Enabled()`, HTTP wiring (timeouts, base URL)
├── types.go           # Request/response DTOs
├── <feature>.go       # One file per API surface (e.g. orders.go, articles.go)
└── *_test.go
```

## Rules

- Clients import only `internal/config` and `internal/pkg`. They MUST NOT import `internal/modules/`.
- The constructor returns a no-op-friendly value when configuration is incomplete (e.g. `Enabled()=false`). Bootstrap logs a warning at startup.
- Every external call goes through the package's `http.Client` with explicit timeouts.
- Errors from this layer should be wrapped with `errors.AppError` codes so the upstream service can decide how to react.

## Reference client

[`sample/`](./sample) is a minimal token-auth REST client. Copy it as a starter for new partners.
