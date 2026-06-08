-- 001_init.sql — starter PostgreSQL schema for the template.
-- Replace with your real first migration. Never edit after deploy; add 002_*.sql, 003_*.sql, …
--
-- The schema_migrations table is created by scripts/migrate.sh — this file just
-- ensures common extensions and timestamp helpers exist.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- for gen_random_uuid()

-- updated_at auto-touch trigger function (call from per-table triggers below).
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
