-- 011_username_partial_unique.sql — relax users.username UNIQUE so empty
-- strings don't collide.
--
-- Bug: GORM inserts `username=''` for users with no username (Register flow,
-- imports, etc). A plain UNIQUE constraint treats '' as a value, so the
-- SECOND such user fails with "duplicate key value violates unique constraint
-- users_username_key". This blocked every new tenant signup once the seed
-- user (admin@acme.example, username='') existed.
--
-- Fix:
--   1. Normalise existing empty usernames to NULL (multiple NULLs allowed).
--   2. Drop the plain UNIQUE constraint.
--   3. Add a partial unique index that only enforces uniqueness when
--      username is a non-empty value.
-- Idempotent: safe to re-run.

BEGIN;

UPDATE users SET username = NULL WHERE username = '';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;

CREATE UNIQUE INDEX IF NOT EXISTS users_username_unique
  ON users (username)
  WHERE username IS NOT NULL AND username <> '';

COMMIT;
