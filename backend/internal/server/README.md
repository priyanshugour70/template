# internal/server/

HTTP server lifecycle management.

- Wraps `net/http.Server` with a Gin engine (or any `http.Handler`) and configurable read/write/idle timeouts.
- `Run()` starts the server on the configured port.
- `Shutdown(ctx)` provides graceful shutdown bounded by the supplied context.
