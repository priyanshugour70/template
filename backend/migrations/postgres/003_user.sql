-- 003_user.sql — Users and Memberships (user ↔ organization link).

-- ── users ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
  id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email                    CITEXT NOT NULL UNIQUE,
  email_verified_at        TIMESTAMPTZ,
  password_hash            TEXT,
  password_algo            TEXT NOT NULL DEFAULT 'bcrypt',
  password_changed_at      TIMESTAMPTZ,
  must_change_password     BOOLEAN NOT NULL DEFAULT false,
  first_name               TEXT,
  middle_name              TEXT,
  last_name                TEXT,
  display_name             TEXT,
  username                 CITEXT UNIQUE,
  avatar_url               TEXT,
  cover_url                TEXT,
  bio                      TEXT,
  phone                    TEXT,
  phone_verified_at        TIMESTAMPTZ,
  alt_email                CITEXT,
  date_of_birth            DATE,
  gender                   TEXT,
  job_title                TEXT,
  department               TEXT,
  employee_code            TEXT,
  status                   TEXT NOT NULL DEFAULT 'invited'
                           CHECK (status IN ('active', 'suspended', 'invited', 'pending', 'archived')),
  locale                   TEXT NOT NULL DEFAULT 'en-IN',
  timezone                 TEXT NOT NULL DEFAULT 'Asia/Kolkata',
  country                  TEXT,
  state                    TEXT,
  city                     TEXT,
  address                  JSONB,
  preferences              JSONB NOT NULL DEFAULT '{}'::jsonb,
  notification_preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata                 JSONB NOT NULL DEFAULT '{}'::jsonb,
  last_login_at            TIMESTAMPTZ,
  last_login_ip            INET,
  last_login_user_agent    TEXT,
  failed_login_count       INT NOT NULL DEFAULT 0,
  locked_until             TIMESTAMPTZ,
  mfa_enabled              BOOLEAN NOT NULL DEFAULT false,
  mfa_secret               TEXT,
  mfa_recovery_codes       JSONB,
  is_super_admin           BOOLEAN NOT NULL DEFAULT false,
  primary_tenant_id        UUID REFERENCES tenants(id) ON DELETE SET NULL,
  primary_organization_id  UUID REFERENCES organizations(id) ON DELETE SET NULL,
  signup_source            TEXT,
  referral_code            TEXT,
  marketing_opt_in         BOOLEAN NOT NULL DEFAULT false,
  terms_accepted_at        TIMESTAMPTZ,
  terms_version            TEXT,
  created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at               TIMESTAMPTZ,
  created_by               UUID,
  updated_by               UUID,
  deleted_by               UUID
);

CREATE INDEX IF NOT EXISTS idx_users_status_active ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_active ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_primary_tenant ON users (primary_tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_primary_org ON users (primary_organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users (last_login_at DESC) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── memberships (user ↔ organization) ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS memberships (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  status            TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended', 'invited', 'pending', 'archived')),
  is_default        BOOLEAN NOT NULL DEFAULT false,
  is_owner          BOOLEAN NOT NULL DEFAULT false,
  is_billing_contact BOOLEAN NOT NULL DEFAULT false,
  job_title         TEXT,
  department        TEXT,
  employee_code     TEXT,
  reports_to        UUID REFERENCES users(id) ON DELETE SET NULL,
  invited_by        UUID REFERENCES users(id) ON DELETE SET NULL,
  invited_at        TIMESTAMPTZ,
  joined_at         TIMESTAMPTZ,
  last_active_at    TIMESTAMPTZ,
  permissions_cache_key TEXT,
  settings          JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ,
  created_by        UUID,
  updated_by        UUID,
  deleted_by        UUID,
  UNIQUE (user_id, organization_id)
);

CREATE INDEX IF NOT EXISTS idx_memberships_user ON memberships (user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_memberships_org_status ON memberships (organization_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_memberships_tenant_user ON memberships (tenant_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_memberships_user_default ON memberships (user_id, is_default) WHERE deleted_at IS NULL AND is_default = true;

DROP TRIGGER IF EXISTS trg_memberships_updated_at ON memberships;
CREATE TRIGGER trg_memberships_updated_at
  BEFORE UPDATE ON memberships
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
