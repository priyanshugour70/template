# migrations/

Forward-only PostgreSQL migrations.

| Folder | Purpose |
|--------|---------|
| `postgres/` | Schema migrations applied by `scripts/migrate.sh`. Files run in lexicographic order; applied filenames are tracked in `schema_migrations`. |

## Naming

`NNN_short_name.sql` (3-digit numeric prefix). Examples: `001_init.sql`, `017_add_audit_log_index.sql`.

## Workflow

1. Create a new file with the next number.
2. Run `make migrate` (or `bash scripts/migrate.sh`) locally to apply.
3. Commit the file; CI will apply it on next deploy (the EC2 deploy script runs `migrate.sh` inside the api container before starting services).

## Rules

- **Never edit a merged migration.** Add a new file instead.
- All migrations are forward-only — no down migrations. If a change is destructive, write the safe equivalent (e.g. add column nullable, backfill, then drop old column in a later migration).
- Idempotency: use `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`, `ALTER TABLE … ADD COLUMN IF NOT EXISTS`, etc., so re-running a failed deploy is safe.
