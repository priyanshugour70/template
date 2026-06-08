# backend

Go backend template: **modular monolith** with Gin, **GORM** (MariaDB ORM), and Redis. Separate **API** and **Worker** processes. Designed for **high-throughput** at scale.

> Replace `github.com/your-org/your-service` everywhere with the real module path before committing.

## Tech stack

- Go 1.25+
- Gin (HTTP framework)
- GORM (ORM for MariaDB)
- MariaDB 11 (primary database)
- Redis 7 (cache + pub/sub queue)
- Swagger (API documentation, generated from annotations)
- Zap (structured logging with daily file rotation)
- OpenTelemetry (optional traces over OTLP HTTP)
- Meilisearch (optional, for product/document discovery)
- AWS S3 (optional, for asset uploads)

## Quick start

```bash
# 1. Copy env
cp .env.example .env

# 2. Start Redis (and MariaDB on your own / Docker Desktop / RDS)
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
make docker-up    # starts API + Worker + Redis (MariaDB external)
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
| `make migrate` | Apply SQL migrations |
| `make seed` | Apply seed SQL files |
| `make apply-sql FILE=…` | Apply a single SQL file |
| `make docker-deps` | Start Redis only |
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

See [`.env.example`](.env.example) for the full list. Required for any environment:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `SERVER_PORT` | `8080` | HTTP port |
| `SERVER_MODE` | `debug` | Gin mode (`debug` / `release`) |
| `MARIADB_HOST` | `localhost` | MariaDB host |
| `MARIADB_PORT` | `3306` | MariaDB port |
| `MARIADB_USER` | `root` | MariaDB user |
| `MARIADB_PASSWORD` | `root` | MariaDB password |
| `MARIADB_DATABASE` | `app_db` | Database name |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `AUTH_JWT_SECRET` | (dev placeholder) | JWT signing key (≥32 chars in prod) |
| `CORS_ALLOWED_ORIGINS` | localhost dev origins | Comma-separated allowed origins |

Optional integrations: `SMTP_*`, `S3_*`, `MEILISEARCH_*`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OAUTH_GOOGLE_*`.

## Project structure

See [`STRUCTURE.md`](STRUCTURE.md) for the full directory tree and architecture overview.

## License

Private / internal use.
