# internal/bootstrap/

Application composition root — wires modules, clients, middleware, and infrastructure together.

| File | Purpose |
|------|---------|
| `bootstrap.go` | `BootstrapAPI(ctx, cfg, log)` builds the Gin engine, opens PostgreSQL + Redis, instantiates every module, mounts middleware, and registers all routes. The single source of truth for dependency injection. |

## Responsibilities

1. **Infrastructure**: PostgreSQL (GORM via pgx), Redis (cache + queue DBs).
2. **Modules**: For each module, create `Repository → Service → Handler` and register routes on the protected/public groups. (Currently empty — add yours.)
3. **Middleware chain**: CORS, security headers, correlation ID, request logging, panic recovery.
4. **Health checks**: `/health`, `/health/live`, `/health/ready`.
5. **Graceful degradation**: When PostgreSQL is unavailable at startup, return 503 on `/api/v1/*` and keep health endpoints live.

## Startup flow

```
cmd/api/main.go
    │
    ├─ config.Load()
    ├─ logger.New()
    ├─ gin stub (binds port immediately)
    │
    └─ bootstrap.BootstrapAPI(ctx, cfg, log)
        ├─ Connect PostgreSQL + Redis
        ├─ Construct modules: repo → service → handler
        ├─ Mount middleware
        └─ Register routes  ──► live router atomically swapped in
```

## Dependency rules

- This is the ONLY package allowed to import concrete module and client types together.
- All cross-module dependencies are wired here through interfaces.
- No business logic in bootstrap — only construction and wiring.
