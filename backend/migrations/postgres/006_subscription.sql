-- 006_subscription.sql — Plans, org subscriptions, usage counters.

-- ── subscription_plans (catalog) ───────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscription_plans (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code            TEXT NOT NULL UNIQUE,
  name            TEXT NOT NULL,
  description     TEXT,
  tagline         TEXT,
  tier            INT NOT NULL DEFAULT 0,
  billing_cycle   TEXT NOT NULL DEFAULT 'monthly'
                  CHECK (billing_cycle IN ('monthly', 'quarterly', 'yearly', 'custom', 'one_time')),
  price_cents     BIGINT NOT NULL DEFAULT 0,
  currency        TEXT NOT NULL DEFAULT 'INR',
  trial_days      INT NOT NULL DEFAULT 0,
  is_active       BOOLEAN NOT NULL DEFAULT true,
  is_default      BOOLEAN NOT NULL DEFAULT false,
  is_public       BOOLEAN NOT NULL DEFAULT true,
  is_addon        BOOLEAN NOT NULL DEFAULT false,
  features        JSONB NOT NULL DEFAULT '[]'::jsonb,
  limits          JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  effective_from  TIMESTAMPTZ,
  effective_until TIMESTAMPTZ,
  gateway         TEXT,
  external_ref    TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_active ON subscription_plans (is_active, tier) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_subscription_plans_updated_at ON subscription_plans;
CREATE TRIGGER trg_subscription_plans_updated_at BEFORE UPDATE ON subscription_plans FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── subscriptions (one active per org) ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscriptions (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id               UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  plan_id                 UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE RESTRICT,
  plan_code               TEXT NOT NULL,
  status                  TEXT NOT NULL DEFAULT 'trial'
                          CHECK (status IN ('trial', 'active', 'past_due', 'cancelled', 'expired', 'paused', 'pending')),
  billing_cycle           TEXT NOT NULL DEFAULT 'monthly',
  quantity                INT NOT NULL DEFAULT 1,
  unit_price_cents        BIGINT NOT NULL DEFAULT 0,
  discount_cents          BIGINT NOT NULL DEFAULT 0,
  tax_cents               BIGINT NOT NULL DEFAULT 0,
  total_cents             BIGINT NOT NULL DEFAULT 0,
  currency                TEXT NOT NULL DEFAULT 'INR',
  started_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  trial_started_at        TIMESTAMPTZ,
  trial_ends_at           TIMESTAMPTZ,
  current_period_start    TIMESTAMPTZ,
  current_period_end      TIMESTAMPTZ,
  next_billing_at         TIMESTAMPTZ,
  last_billed_at          TIMESTAMPTZ,
  cancel_at               TIMESTAMPTZ,
  cancelled_at            TIMESTAMPTZ,
  cancel_reason           TEXT,
  cancel_immediate        BOOLEAN NOT NULL DEFAULT false,
  ended_at                TIMESTAMPTZ,
  pause_at                TIMESTAMPTZ,
  paused_at               TIMESTAMPTZ,
  resume_at               TIMESTAMPTZ,
  gateway                 TEXT,
  gateway_customer_id     TEXT,
  gateway_subscription_id TEXT,
  external_ref            TEXT,
  coupon_code             TEXT,
  billing_email           CITEXT,
  billing_name            TEXT,
  billing_address         JSONB,
  features                JSONB NOT NULL DEFAULT '[]'::jsonb,
  limits                  JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata                JSONB NOT NULL DEFAULT '{}'::jsonb,
  notes                   TEXT,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at              TIMESTAMPTZ,
  created_by              UUID,
  updated_by              UUID,
  deleted_by              UUID
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_org ON subscriptions (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_status ON subscriptions (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_period_end ON subscriptions (current_period_end) WHERE deleted_at IS NULL AND status IN ('active', 'past_due');
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_billing ON subscriptions (next_billing_at) WHERE deleted_at IS NULL AND status = 'active';
CREATE UNIQUE INDEX IF NOT EXISTS uidx_subscriptions_one_active_per_org
  ON subscriptions (organization_id)
  WHERE deleted_at IS NULL AND status IN ('trial', 'active', 'past_due', 'paused');

DROP TRIGGER IF EXISTS trg_subscriptions_updated_at ON subscriptions;
CREATE TRIGGER trg_subscriptions_updated_at BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── usage_counters ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS usage_counters (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
  key             TEXT NOT NULL,
  count           BIGINT NOT NULL DEFAULT 0,
  limit_value     BIGINT,
  period_start    TIMESTAMPTZ NOT NULL,
  period_end      TIMESTAMPTZ NOT NULL,
  last_reset_at   TIMESTAMPTZ,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID,
  UNIQUE (organization_id, key, period_start)
);

CREATE INDEX IF NOT EXISTS idx_usage_counters_org_key ON usage_counters (organization_id, key) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_usage_counters_updated_at ON usage_counters;
CREATE TRIGGER trg_usage_counters_updated_at BEFORE UPDATE ON usage_counters FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── seed default plans ─────────────────────────────────────────────────────
INSERT INTO subscription_plans (code, name, description, tier, billing_cycle, price_cents, currency, trial_days, is_default, features, limits) VALUES
  ('free', 'Free', 'Free starter plan', 0, 'monthly', 0, 'INR', 0, true,
    '["user.invite","user.list","org.read","audit.read"]'::jsonb,
    '{"users.max":5,"storage.gb":1,"orgs.max":1,"api.calls.monthly":10000}'::jsonb),
  ('starter', 'Starter', 'Small teams', 1, 'monthly', 99900, 'INR', 14, false,
    '["user.invite","user.list","org.read","org.create","audit.read","audit.export","export.csv"]'::jsonb,
    '{"users.max":25,"storage.gb":10,"orgs.max":3,"api.calls.monthly":100000}'::jsonb),
  ('pro', 'Pro', 'Growing companies', 2, 'monthly', 299900, 'INR', 14, false,
    '["user.invite","user.list","org.read","org.create","audit.read","audit.export","export.csv","export.xlsx","webhook","sso"]'::jsonb,
    '{"users.max":250,"storage.gb":100,"orgs.max":10,"api.calls.monthly":1000000}'::jsonb),
  ('enterprise', 'Enterprise', 'Enterprise scale', 3, 'yearly', 0, 'INR', 30, false,
    '["user.invite","user.list","org.read","org.create","audit.read","audit.export","export.csv","export.xlsx","webhook","sso","saml","scim","priority_support","custom_branding"]'::jsonb,
    '{"users.max":-1,"storage.gb":-1,"orgs.max":-1,"api.calls.monthly":-1}'::jsonb)
ON CONFLICT (code) DO NOTHING;
