# internal/repository/

Database connection and GORM setup for PostgreSQL.

- `OpenPostgres(ctx, cfg, log)` returns a `*gorm.DB` (pgx driver) with a tuned pool (defaults: 50 open, 25 idle, 5-min lifetime).
- Prepared statements enabled; foreign-key constraints relaxed during `AutoMigrate` so the SQL migration files in `migrations/postgres/` remain the source of truth.
- `AutoMigrate(db, models...)` is a thin pass-through — only use in dev or for module-local helper tables.

Modules each define their own `Repository` struct that embeds or wraps `*gorm.DB`. Keep raw GORM calls inside repositories so the service layer stays portable.
