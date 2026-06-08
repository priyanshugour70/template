-- 002_tenant.sql — Tenants and Organizations.
-- Tenant = SaaS customer (top-level isolation boundary).
-- Organization = workspace inside a tenant. All business data hangs off org_id.

CREATE EXTENSION IF NOT EXISTS citext;

-- ── tenants ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenants (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug              CITEXT NOT NULL UNIQUE,
  name              TEXT NOT NULL,
  legal_name        TEXT,
  display_name      TEXT,
  description       TEXT,
  logo_url          TEXT,
  favicon_url       TEXT,
  primary_color     TEXT,
  secondary_color   TEXT,
  support_email     CITEXT,
  support_phone     TEXT,
  website_url       TEXT,
  status            TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended', 'trial', 'pending', 'archived')),
  plan_code         TEXT,
  seat_limit        INT,
  country           TEXT,
  timezone          TEXT NOT NULL DEFAULT 'Asia/Kolkata',
  locale            TEXT NOT NULL DEFAULT 'en-IN',
  currency          TEXT NOT NULL DEFAULT 'INR',
  billing_email     CITEXT,
  billing_name      TEXT,
  billing_address   JSONB,
  tax_id            TEXT,
  settings          JSONB NOT NULL DEFAULT '{}'::jsonb,
  features          JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
  trial_ends_at     TIMESTAMPTZ,
  activated_at      TIMESTAMPTZ,
  suspended_at      TIMESTAMPTZ,
  suspension_reason TEXT,
  archived_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ,
  created_by        UUID,
  updated_by        UUID,
  deleted_by        UUID
);

CREATE INDEX IF NOT EXISTS idx_tenants_status_active ON tenants (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants (deleted_at);
CREATE INDEX IF NOT EXISTS idx_tenants_plan_code ON tenants (plan_code) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_tenants_updated_at ON tenants;
CREATE TRIGGER trg_tenants_updated_at
  BEFORE UPDATE ON tenants
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── organizations ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS organizations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  slug            CITEXT NOT NULL,
  name            TEXT NOT NULL,
  display_name    TEXT,
  description     TEXT,
  logo_url        TEXT,
  cover_url       TEXT,
  primary_color   TEXT,
  secondary_color TEXT,
  website_url     TEXT,
  contact_email   CITEXT,
  contact_phone   TEXT,
  industry        TEXT,
  size            TEXT,
  country         TEXT,
  state           TEXT,
  city            TEXT,
  postal_code     TEXT,
  timezone        TEXT NOT NULL DEFAULT 'Asia/Kolkata',
  locale          TEXT NOT NULL DEFAULT 'en-IN',
  currency        TEXT NOT NULL DEFAULT 'INR',
  status          TEXT NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active', 'suspended', 'pending', 'archived')),
  is_default      BOOLEAN NOT NULL DEFAULT false,
  settings        JSONB NOT NULL DEFAULT '{}'::jsonb,
  features        JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  address         JSONB,
  activated_at    TIMESTAMPTZ,
  suspended_at    TIMESTAMPTZ,
  archived_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID,
  UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_organizations_tenant_status ON organizations (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations (deleted_at);
CREATE INDEX IF NOT EXISTS idx_organizations_is_default ON organizations (tenant_id, is_default) WHERE deleted_at IS NULL AND is_default = true;

DROP TRIGGER IF EXISTS trg_organizations_updated_at ON organizations;
CREATE TRIGGER trg_organizations_updated_at
  BEFORE UPDATE ON organizations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
