# internal/bootstrap/

Application composition root — wires modules, clients, middleware, and infrastructure together.

| File | Purpose |
|------|---------|
| `bootstrap.go` | `BootstrapAPI(ctx, cfg, log)` builds the Gin engine, opens DB/Redis/Meilisearch, instantiates every module, mounts middleware, and registers all routes. The single source of truth for dependency injection. |

## Responsibilities

1. **Infrastructure**: MariaDB (GORM), Redis (cache + queue DBs), optional Meilisearch, optional OTLP tracing.
2. **Clients**: One factory call per external partner.
3. **Modules**: For each module, create `Repository → Service → Handler` and register routes on the protected/public groups.
4. **Middleware chain**: CORS, security headers, correlation ID, request logging, panic recovery, tracing.
5. **Health checks**: `/healthz`, `/health`, `/health/live`, `/health/ready`.
6. **Graceful degradation**: When MariaDB is unavailable at startup, return 503 on `/api/v1/*` and keep health endpoints live.

## Startup flow

```
cmd/api/main.go
    │
    ├─ config.Load()
    ├─ logger.New()
    ├─ gin stub (binds port immediately)
    │
    └─ bootstrap.BootstrapAPI(ctx, cfg, log)
        ├─ Connect DB, Redis, Meilisearch (optional)
        ├─ Construct clients
        ├─ Construct modules: repo → service → handler
        ├─ Wire cross-module hooks
        ├─ Mount middleware
        └─ Register routes  ──► live router atomically swapped in
```

## Dependency rules

- This is the ONLY package allowed to import concrete module and client types together.
- All cross-module dependencies are wired here through interfaces.
- No business logic in bootstrap — only construction and wiring.
