# internal/middleware/

Global HTTP middleware applied by the bootstrap composition root.

| File | Purpose |
|------|---------|
| `correlation.go` | Reads `X-Correlation-ID` / `X-Request-ID`, generates one if absent, stores a request-scoped Zap logger on `*gin.Context`, propagates the ID through `context.Context` and into the active OpenTelemetry span. |
| `logger.go` | Emits one structured access-log line per request with method, path (with sensitive query keys redacted), status, latency, IP, user agent, response size. Must run after `CorrelationID`. |
| `log_query_redact.go` | Redacts sensitive query keys (`password`, `token`, `api_key`, …) before paths are logged. |
| `recovery.go` | Recovers from panics, logs them with the request logger, and returns a structured `INTERNAL_ERROR` JSON. |
| `ratelimit.go` | In-memory IP-based rate limiter with configurable requests-per-minute window and automatic cleanup. |
| `security_headers.go` | Baseline hardening headers (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`). |

For per-route auth, RBAC, or audit middleware, keep those inside the relevant module (e.g., `internal/modules/auth/middleware.go`).
