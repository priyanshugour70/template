# internal/

Private application code for this service. Everything under `internal/` is private to this Go module.

## Directory structure

```
internal/
├── bootstrap/      → Application composition root (DI wiring)
├── cache/          → Cache abstraction (Redis / memory / noop) with stampede protection
├── clients/        → External service HTTP clients (one folder per partner — empty by default)
├── config/         → Static config (Viper) + dynamic business config interface
├── health/         → Health endpoints (Postgres + Redis + system info)
├── jobs/           → Background job implementations (cron-like)
├── middleware/     → Global HTTP middleware (CORS, security, rate-limit, logging)
├── modules/        → Domain business modules (empty by default)
├── pkg/            → Shared infrastructure utilities (errors, logger, response, mail, …)
├── queue/          → Redis Pub/Sub Producer + Consumer
├── repository/     → Shared repository utilities (GORM PostgreSQL connection pool)
└── server/         → HTTP server wrapper with timeouts and graceful shutdown
```

## Architectural principles

1. **Layered modules**: handler → service → repository → model.
2. **Interface boundaries**: Modules communicate through interfaces, not concrete types.
3. **Single composition root**: `bootstrap/` is the only package that constructs and wires all dependencies.
4. **Client isolation**: External API logic lives in `clients/`, never in modules.
5. **Config centralization**: Static config in `config/config.go`, dynamic business config via `config.BizConfigReader`.

## Dependency flow

```
cmd/ → bootstrap/ → modules/ → pkg/
                  → clients/ → config/
                  → cache/
                  → middleware/
```

- `bootstrap/` imports everything (composition root).
- `modules/` import `pkg/`, `config/`, and client interfaces.
- `clients/` import only `config/` and `pkg/`.
- `pkg/` imports nothing from `internal/` (except `config/` when a struct type is needed).
