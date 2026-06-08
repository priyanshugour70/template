# backend

Minimal Go backend template: **modular monolith** with Gin, **GORM** (PostgreSQL), and Redis. Separate **API** and **Worker** processes. Simple, low-dependency starting point.

> Replace `github.com/your-org/your-service` everywhere with the real module path before committing.

## Tech stack

- Go 1.25+
- Gin (HTTP framework)
- GORM (ORM for PostgreSQL via pgx)
- PostgreSQL 16 (primary database)
- Redis 7 (cache + pub/sub queue)
- Swagger (API documentation, generated from annotations)
- Zap (structured logging with daily file rotation)
- AWS S3 (optional, for asset uploads)
- SMTP (optional, for transactional email)

## Quick start

```bash
# 1. Copy env
cp .env.example .env

# 2. Start PostgreSQL + Redis (Docker)
make docker-deps

# 3. Install deps
make deps

# 4. Apply schema
make migrate

# 5. Run API
make run-api
```

API: `http://localhost:8080`
Health: `GET /health`
Readiness: `GET /health/ready`
Swagger UI: `http://localhost:8080/swagger/index.html`

## Full Docker Compose

```bash
make docker-up    # starts Postgres + Redis + API + Worker
make docker-down  # stops everything
```

## Deployment

GitHub Actions deploys `staging` to a staging EC2 instance and `main` to production. See [`DEPLOYMENT.md`](DEPLOYMENT.md) for the full SSM + Secrets Manager flow.

## Makefile targets

| Target | Description |
|--------|-------------|
| `make help` | List all targets |
| `make deps` | `go mod tidy` |
| `make build` | Build api and worker binaries |
| `make run-api` | Run API server |
| `make run-worker` | Run worker |
| `make swagger` | Generate Swagger docs |
| `make migrate` | Apply SQL migrations (uses `psql`) |
| `make docker-deps` | Start PostgreSQL + Redis only |
| `make docker-up` | Full docker-compose up |
| `make docker-down` | Stop docker-compose |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make lint` | Run golangci-lint |
| `make fmt` | gofmt + goimports |
| `make install-hooks` | Install git hooks from `.githooks/` |

## Health endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Full health (DB, Redis, system, memory, goroutines) |
| `GET /health/live` | Kubernetes liveness probe |
| `GET /health/ready` | Kubernetes readiness probe |

## Environment variables

See [`.env.example`](.env.example) for the full list. The template uses only what an application actually needs day-1:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `SERVER_PORT` | `8080` | HTTP port |
| `SERVER_MODE` | `debug` | Gin mode (`debug` / `release`) |
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `postgres` | DB user |
| `POSTGRES_PASSWORD` | `postgres` | DB password |
| `POSTGRES_DATABASE` | `app_db` | DB name |
| `POSTGRES_SSLMODE` | `disable` | `disable` / `require` / `verify-full` |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `AUTH_JWT_SECRET` | (dev placeholder) | JWT signing key (≥32 chars in prod) |
| `CORS_ALLOWED_ORIGINS` | localhost dev | Comma-separated allowed origins |

Optional: `SMTP_*`, `ASSETS_S3_*`, `DEVELOPER_API_KEY_PEPPER`.

## Project structure

See [`STRUCTURE.md`](STRUCTURE.md) for the full directory tree and architecture overview.

## License

Private / internal use.
