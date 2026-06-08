# cmd/

Application entrypoints. Each subdirectory contains a `main.go` that builds an independent binary. All binaries share the same `internal/` packages.

| Binary | Purpose |
|--------|---------|
| **api/** | HTTP API server. Bootstraps config, DB, Redis, router, registers modules, starts Gin with graceful shutdown. Swagger annotations live in `cmd/api/main.go`. |
| **worker/** | Background worker. Connects to Redis and processes Pub/Sub channels asynchronously. Hosts cron-like jobs (`internal/jobs/*`). |
| **migrate/** | Applies forward-only SQL migrations from `migrations/mariadb/` and idempotent seeds from `migrations/seeds/`. |
| **seed/** | Re-applies the idempotent seed files (never drops data). |
| **apply-sql/** | Apply a single SQL file. Use for one-off admin tasks: `go run ./cmd/apply-sql path/to.sql`. |
| **hashpass/** | Hash a password using the same algorithm as the auth module. Useful for creating seed admin accounts. |
| **reindex-search/** | Rebuilds Meilisearch indices from the database (run after schema changes or to recover from index loss). |
| **repair-data/** | One-off data repair commands (used to recover from production incidents). Delete unused subcommands in your project. |

Add a new entrypoint by creating `cmd/<name>/main.go`. The Dockerfile builds every `cmd/*` so the same image is reusable across binaries.
