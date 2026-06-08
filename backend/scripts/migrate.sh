#!/usr/bin/env bash
# Applies SQL migrations in migrations/postgres/ via psql.
#
# Reads connection details from environment (POSTGRES_HOST, POSTGRES_PORT,
# POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DATABASE, POSTGRES_SSLMODE).
# Pulls .env from the project root unless ENV_FILE is set.
#
# A `schema_migrations` table records applied files; existing files are skipped.

set -euo pipefail
cd "$(dirname "$0")/.."

MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations/postgres}"

if [ -f "${ENV_FILE:-.env}" ]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE:-.env}"
  set +a
fi

: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${POSTGRES_DATABASE:?POSTGRES_DATABASE is required}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_SSLMODE="${POSTGRES_SSLMODE:-disable}"

export PGPASSWORD="$POSTGRES_PASSWORD"
PSQL=(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DATABASE" -v ON_ERROR_STOP=1 --set=sslmode="$POSTGRES_SSLMODE")

if ! command -v psql >/dev/null 2>&1; then
  echo "psql not found. Install postgresql-client and retry." >&2
  exit 1
fi

echo "==> ensuring schema_migrations table"
"${PSQL[@]}" -q <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  name        VARCHAR(255) PRIMARY KEY,
  applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

applied=$( "${PSQL[@]}" -At -c "SELECT name FROM schema_migrations" )

shopt -s nullglob
files=("$MIGRATIONS_DIR"/*.sql)
if [ ${#files[@]} -eq 0 ]; then
  echo "No .sql files in $MIGRATIONS_DIR"
  exit 0
fi

for path in "${files[@]}"; do
  name="$(basename "$path")"
  if printf '%s\n' "$applied" | grep -Fxq "$name"; then
    echo "  -> $name (skipped)"
    continue
  fi
  echo "  -> $name"
  "${PSQL[@]}" -q -f "$path"
  "${PSQL[@]}" -q -c "INSERT INTO schema_migrations (name) VALUES ('$name') ON CONFLICT DO NOTHING"
done

echo "Done."
