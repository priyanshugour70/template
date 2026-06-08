-- 005_auth.sql — Invites, password resets, refresh tokens.

-- ── invites ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS invites (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  email           CITEXT NOT NULL,
  first_name      TEXT,
  last_name       TEXT,
  job_title       TEXT,
  department      TEXT,
  token_hash      BYTEA NOT NULL UNIQUE,
  role_ids        UUID[] NOT NULL DEFAULT '{}'::uuid[],
  invited_by      UUID REFERENCES users(id) ON DELETE SET NULL,
  message         TEXT,
  status          TEXT NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
  expires_at      TIMESTAMPTZ NOT NULL,
  accepted_at     TIMESTAMPTZ,
  accepted_by     UUID REFERENCES users(id) ON DELETE SET NULL,
  revoked_at      TIMESTAMPTZ,
  revoked_by      UUID REFERENCES users(id) ON DELETE SET NULL,
  resend_count    INT NOT NULL DEFAULT 0,
  last_resent_at  TIMESTAMPTZ,
  ip              INET,
  user_agent      TEXT,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_invites_org_status ON invites (organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_invites_email_status ON invites (email, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_invites_expires_at ON invites (expires_at) WHERE deleted_at IS NULL AND status = 'pending';

DROP TRIGGER IF EXISTS trg_invites_updated_at ON invites;
CREATE TRIGGER trg_invites_updated_at BEFORE UPDATE ON invites FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── password reset tokens ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  BYTEA NOT NULL UNIQUE,
  ip          INET,
  user_agent  TEXT,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ,
  created_by  UUID,
  updated_by  UUID,
  deleted_by  UUID
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_reset_tokens (user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_password_resets_expires ON password_reset_tokens (expires_at) WHERE deleted_at IS NULL AND used_at IS NULL;

DROP TRIGGER IF EXISTS trg_password_resets_updated_at ON password_reset_tokens;
CREATE TRIGGER trg_password_resets_updated_at BEFORE UPDATE ON password_reset_tokens FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── refresh tokens ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- = jti
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id        UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id  UUID REFERENCES organizations(id) ON DELETE SET NULL,
  membership_id    UUID REFERENCES memberships(id) ON DELETE SET NULL,
  token_hash       BYTEA NOT NULL,
  family_id        UUID NOT NULL,
  parent_id        UUID,
  device_id        TEXT,
  device_name      TEXT,
  client           TEXT,
  ip               INET,
  user_agent       TEXT,
  issued_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at       TIMESTAMPTZ NOT NULL,
  last_used_at     TIMESTAMPTZ,
  revoked_at       TIMESTAMPTZ,
  revoked_reason   TEXT,
  metadata         JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at       TIMESTAMPTZ,
  created_by       UUID,
  updated_by       UUID,
  deleted_by       UUID
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_active ON refresh_tokens (user_id, revoked_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens (family_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens (expires_at) WHERE deleted_at IS NULL AND revoked_at IS NULL;

DROP TRIGGER IF EXISTS trg_refresh_tokens_updated_at ON refresh_tokens;
CREATE TRIGGER trg_refresh_tokens_updated_at BEFORE UPDATE ON refresh_tokens FOR EACH ROW EXECUTE FUNCTION set_updated_at();
