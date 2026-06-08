# migrations/

Database migrations and idempotent seeds.

| Folder | Purpose |
|--------|---------|
| `mariadb/` | Forward-only schema migrations. Files are applied in lexicographic order and recorded in `schema_migrations`. Never edit a merged file — add a new one. |
| `seeds/` | Idempotent seed SQL applied by `make seed`. Use `INSERT IGNORE` or `ON DUPLICATE KEY UPDATE` so seeds can run repeatedly. |

## Naming

- Migrations: `NNN_short_name.sql` (3-digit numeric prefix). Examples: `001_init.sql`, `017_add_audit_log_index.sql`.
- Seeds: `NNN_short_name.sql`. Examples: `001_auth_roles.sql`, `002_default_categories.sql`.

## Workflow

1. Create a new file with the next number.
2. Run `make migrate` (or `make seed`) locally to apply.
3. Commit the file; CI will apply it on next deploy.

`cmd/migrate` runs **both** schema migrations and seeds; `cmd/seed` only runs seeds (assuming schema is current).
