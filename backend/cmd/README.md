# cmd/

Application entrypoints. Each subdirectory contains a `main.go` that builds an independent binary. All binaries share the same `internal/` packages.

| Binary | Purpose |
|--------|---------|
| **api/** | HTTP API server. Bootstraps config, DB, Redis, router, registers modules, starts Gin with graceful shutdown. Swagger annotations live in `cmd/api/main.go`. |
| **worker/** | Background worker. Connects to Redis and processes Pub/Sub channels asynchronously. Hosts cron-like jobs (`internal/jobs/*`). |

Migrations are applied via `scripts/migrate.sh` (which shells out to `psql`) — there is no separate `migrate` binary in this simplified template. Add a `cmd/<name>/main.go` if you need a new entrypoint; both Dockerfile variants will pick it up automatically once you update them.
