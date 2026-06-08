# Backend – Directory Structure

Production-ready **modular monolith** in Go with Clean Architecture, Gin, GORM (MariaDB), and Redis. API and Worker run as separate processes. Designed for **high-request scale**.

---

## Full folder tree

```
backend/
├── cmd/                              # Process entrypoints (one per binary)
│   ├── api/                          #   HTTP API server (Swagger annotations live here)
│   ├── worker/                       #   Queue consumer / background jobs
│   ├── migrate/                      #   Apply SQL migrations
│   ├── seed/                         #   Apply idempotent seed data
│   ├── apply-sql/                    #   Apply a single SQL file (admin / ad-hoc)
│   ├── hashpass/                     #   Hash a password (admin)
│   ├── reindex-search/               #   Rebuild Meilisearch indices from DB
│   └── repair-data/                  #   One-off data repair commands
├── internal/                         # Private application code (Go internal)
│   ├── bootstrap/                    #   Composition root (DI wiring of all modules)
│   ├── config/                       #   Static config (Viper) + dynamic biz config interface
│   ├── server/                       #   HTTP server wrapper (timeouts, graceful shutdown)
│   ├── health/                       #   Health endpoints (DB, Redis, system)
│   ├── middleware/                   #   Global middleware (CORS, logger, recovery, rate-limit)
│   ├── repository/                   #   GORM MariaDB connection pool
│   ├── cache/                        #   Cache abstraction (Redis / memory / noop)
│   ├── queue/                        #   Redis Pub/Sub Producer + Consumer
│   ├── migrations/                   #   SQL migration runner
│   ├── meilisearch/                  #   Optional search client + index management
│   ├── tracing/                      #   OpenTelemetry setup
│   ├── jobs/                         #   Background job implementations (cron-like)
│   ├── clients/                      #   External service HTTP clients (Shopify, OMS, …)
│   ├── modules/                      #   Domain modules (handler → service → repository)
│   │   └── sample/                   #     Reference module showing the standard layout
│   └── pkg/                          #   Shared infrastructure utilities
│       ├── logger/                   #     Zap with daily file rotation
│       ├── response/                 #     Gin JSON response helpers
│       ├── errors/                   #     AppError + error codes
│       ├── correlation/              #     Correlation/request ID propagation
│       ├── mail/                     #     SMTP sender + transactional templates
│       ├── slug/                     #     URL-safe slug normalization
│       ├── sequence/                 #     Safe sequence/counter helpers
│       └── versioning/               #     Versioned-entity helpers
├── migrations/
│   ├── mariadb/                      # Versioned SQL DDL (NNN_name.sql)
│   └── seeds/                        # Idempotent seed SQL
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
├── .env / .env.example / .env.staging / .env.prod
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
| **cmd/** | Entrypoints. Each subdirectory builds an independent binary. |
| **internal/bootstrap** | Composition root. Wires config, GORM DB, Redis, cache, queue, health, swagger, and all modules. Single place for DI. |
| **internal/config** | Loads configuration from env vars via Viper. Reads `.env` → `.env.{APP_ENV}` → `.env.local`. Production validation enforces JWT length and CORS origins. |
| **internal/server** | HTTP server wrapper (Gin + timeouts, graceful shutdown). |
| **internal/health** | `/health` (full), `/health/live` (liveness), `/health/ready` (readiness). Checks MariaDB and Redis. |
| **internal/middleware** | Global middleware: CORS, security headers, correlation ID, request logging, panic recovery, rate limiting. |
| **internal/modules** | Domain modules (modular monolith). Each module follows `handler.go → service.go → repository.go → model.go`. |
| **internal/queue** | Pub/Sub abstraction: Producer/Consumer interfaces + Redis Pub/Sub implementation + No-op. |
| **internal/cache** | Cache interface + Redis, in-memory, and no-op implementations. Includes stampede-protected `GetOrLoad[T]`. |
| **internal/repository** | GORM-based MariaDB connection pool with tuned defaults. |
| **internal/jobs** | Cron-like background jobs run by the worker. |
| **internal/clients** | External HTTP API clients (one folder per partner). |
| **internal/meilisearch** | Optional Meilisearch client + index management. |
| **internal/tracing** | OpenTelemetry SDK setup, otelgin middleware. |
| **internal/pkg** | Shared infrastructure: logger, errors, response, correlation, mail, slug, sequence, versioning. |
| **migrations/** | Forward-only SQL migrations + idempotent seeds. |
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
6. **Graceful degradation**: If MariaDB or Redis is unavailable at startup, the API still binds the port and returns 503 on `/api/v1/*` so health probes report the actual reason

## Dependency flow

```
cmd/ → bootstrap/ → modules/ → pkg/
                  → clients/ → config/
                  → cache/   → queue/
                  → middleware/
```

- `bootstrap/` imports everything (composition root)
- `modules/` import `pkg/`, `config/`, client interfaces
- `clients/` import only `config/` and `pkg/`
- `pkg/` imports nothing from `internal/` (except `config/` where the struct type is needed)

---

## Clean architecture flow

- **Handlers** parse request → call service → write response (no business logic).
- **Services** contain use cases → call repositories / queue / cache.
- **Repositories** use GORM for DB ops; interfaces allow swapping implementations.

---

## Health endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Full health: DB + Redis + system info |
| `GET /health/live` | Kubernetes liveness probe (always 200) |
| `GET /health/ready` | Kubernetes readiness probe (200 only if DB + Redis up) |
