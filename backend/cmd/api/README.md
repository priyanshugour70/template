# cmd/api/

HTTP API server entrypoint.

Responsibilities:

1. Load configuration (`config.Load()`).
2. Build Zap logger with daily file rotation.
3. Bind the HTTP port **immediately** behind a startup stub so Docker/k8s health probes never see "connection reset" while bootstrap runs.
4. Run `bootstrap.BootstrapAPI(ctx, cfg, log)` in a goroutine — once ready, swap the live router atomically.
5. Trap SIGINT/SIGTERM for graceful shutdown.

Swagger annotations (`@title`, `@host`, `@BasePath`, security definitions) live in this `main.go` and are picked up by `make swagger`.

## Run

```bash
make run-api
# or
go run ./cmd/api
```
