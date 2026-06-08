# Backend – Directory Structure

Minimal **modular monolith** in Go with Clean Architecture, Gin, GORM (PostgreSQL), and Redis. API and Worker run as separate processes.

---

## Full folder tree

```
backend/
├── cmd/                              # Process entrypoints (one per binary)
│   ├── api/                          #   HTTP API server (Swagger annotations live here)
│   └── worker/                       #   Queue consumer / background jobs
├── internal/                         # Private application code
│   ├── bootstrap/                    #   Composition root (DI wiring of all modules)
│   ├── config/                       #   Static config (Viper) + dynamic biz config interface
│   ├── server/                       #   HTTP server wrapper (timeouts, graceful shutdown)
│   ├── health/                       #   Health endpoints (Postgres + Redis + system)
│   ├── middleware/                   #   Global middleware (CORS, logger, recovery, rate-limit)
│   ├── repository/                   #   GORM PostgreSQL connection pool
│   ├── cache/                        #   Cache abstraction (Redis / memory / noop)
│   ├── queue/                        #   Redis Pub/Sub Producer + Consumer
│   ├── jobs/                         #   Background job implementations (cron-like)
│   ├── clients/                      #   External service HTTP clients (one folder per partner — empty by default)
│   ├── modules/                      #   Domain modules — empty by default (add yours under here)
│   └── pkg/                          #   Shared infrastructure utilities
│       ├── logger/                   #     Zap with daily file rotation
│       ├── response/                 #     Gin JSON response helpers
│       ├── errors/                   #     AppError + PostgreSQL error codes
│       ├── correlation/              #     Correlation/request ID propagation
│       ├── mail/                     #     SMTP sender + transactional templates
│       ├── slug/                     #     URL-safe slug normalization
│       └── versioning/               #     Versioned-entity helpers
├── migrations/
│   └── postgres/                     # NNN_name.sql files applied by scripts/migrate.sh
├── deployments/
│   ├── docker/                       # Dockerfile.app/api/worker + docker-compose.yaml
│   └── k8s/                          # Kubernetes manifests
├── scripts/                          # Operational shell scripts (deploy, nginx, migrate, …)
├── docs/                             # Manual docs + generated Swagger
│   ├── swagger/                      #   `make swagger` output (docs.go, swagger.json/yaml)
│   └── samples/                      #   Sample payloads / import files
├── samples/                          # Example data (CSV / curl scripts)
├── test/
│   ├── unit/                         # `go test ./...`
│   └── integration/                  # `go test -tags=integration ./test/...`
├── logs/                             # Runtime log files (daily rotation, gitignored)
├── bin/                              # Built binaries (gitignored)
├── .github/workflows/                # CI/CD pipelines
├── .githooks/                        # Git hooks installed by `make install-hooks`
├── .env / .env.example
├── .gitignore
├── go.mod / go.sum
├── Makefile
├── README.md
├── STRUCTURE.md
└── DEPLOYMENT.md
```

---

## Directory overview

| Directory | Purpose |
|-----------|---------|
| **cmd/api** | HTTP API entrypoint. Binds the port immediately behind a startup stub, then swaps in the real router atomically once bootstrap finishes. |
| **cmd/worker** | Subscribes to Redis Pub/Sub channels and hosts scheduled jobs. |
| **internal/bootstrap** | Composition root. Wires config, GORM DB, Redis, cache, queue, health, swagger, and modules. |
| **internal/config** | Loads configuration via Viper. Reads `.env` → `.env.{APP_ENV}` → `.env.local`. Production validation enforces JWT length and CORS origins. |
| **internal/server** | HTTP server wrapper (Gin + timeouts, graceful shutdown). |
| **internal/health** | `/health` (full), `/health/live` (liveness), `/health/ready` (readiness). |
| **internal/middleware** | Global middleware: CORS, security headers, correlation ID, request logging, panic recovery, rate limiting. |
| **internal/modules** | Domain modules (empty by default — add yours as you build features). |
| **internal/queue** | Pub/Sub abstraction: Producer/Consumer interfaces + Redis Pub/Sub implementation + No-op. |
| **internal/cache** | Cache interface + Redis, in-memory, and no-op implementations. Includes stampede-protected `GetOrLoad[T]`. |
| **internal/repository** | GORM-based PostgreSQL connection pool (pgx driver). |
| **internal/jobs** | Cron-like background jobs run by the worker. |
| **internal/clients** | External HTTP API clients (one folder per partner — empty by default). |
| **internal/pkg** | Shared infrastructure: logger, errors, response, correlation, mail, slug, versioning. |
| **migrations/postgres** | Forward-only SQL migrations applied by `psql` via `scripts/migrate.sh`. |
| **deployments/** | Docker (compose + Dockerfiles) and Kubernetes manifests. |
| **scripts/** | Helper scripts for running, migrating, deploying. |
| **docs/** | Manual markdown + generated Swagger. |
| **test/** | Unit and integration tests. |

---

## Architectural principles

1. **Layered modules**: `handler → service → repository → model`
2. **Interface boundaries**: Modules communicate through interfaces, not concrete types
3. **Single composition root**: `bootstrap/` is the only package that constructs and wires all dependencies
4. **Client isolation**: External API logic lives in `clients/`, never in modules
5. **Config centralization**: Static config in `config/config.go`, dynamic business config via `config.BizConfigReader`
6. **Graceful degradation**: If PostgreSQL or Redis is unavailable at startup, the API still binds the port and returns 503 on `/api/v1/*`

## Dependency flow

```
cmd/ → bootstrap/ → modules/ → pkg/
                  → clients/ → config/
                  → cache/   → queue/
                  → middleware/
```

- `bootstrap/` imports everything (composition root)
- `modules/` import `pkg/`, `config/`, and client interfaces
- `clients/` import only `config/` and `pkg/`
- `pkg/` imports nothing from `internal/` (except `config/` when a struct type is needed)
