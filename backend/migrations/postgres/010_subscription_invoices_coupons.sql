-- 010_subscription_invoices_coupons.sql — Invoices + coupons for the
-- subscription module. Adds endpoints' missing storage so the UI can ship
-- a real plan-switch / billing experience without a gateway integration.

-- ── subscription_invoices ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscription_invoices (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id    UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  subscription_id    UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
  number             TEXT NOT NULL UNIQUE,
  status             TEXT NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open', 'paid', 'void', 'uncollectible', 'refunded')),
  currency           TEXT NOT NULL DEFAULT 'INR',
  subtotal_cents     BIGINT NOT NULL DEFAULT 0,
  discount_cents     BIGINT NOT NULL DEFAULT 0,
  tax_cents          BIGINT NOT NULL DEFAULT 0,
  total_cents        BIGINT NOT NULL DEFAULT 0,
  amount_due_cents   BIGINT NOT NULL DEFAULT 0,
  amount_paid_cents  BIGINT NOT NULL DEFAULT 0,
  coupon_code        TEXT,
  description        TEXT,
  line_items         JSONB NOT NULL DEFAULT '[]'::jsonb,
  period_start       TIMESTAMPTZ,
  period_end         TIMESTAMPTZ,
  issued_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  due_at             TIMESTAMPTZ,
  paid_at            TIMESTAMPTZ,
  voided_at          TIMESTAMPTZ,
  gateway            TEXT,
  gateway_invoice_id TEXT,
  metadata           JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at         TIMESTAMPTZ,
  created_by         UUID,
  updated_by         UUID,
  deleted_by         UUID
);

CREATE INDEX IF NOT EXISTS idx_invoices_org_issued ON subscription_invoices
  (organization_id, issued_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_invoices_status ON subscription_invoices
  (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_invoices_subscription ON subscription_invoices
  (subscription_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_invoices_updated_at ON subscription_invoices;
CREATE TRIGGER trg_invoices_updated_at BEFORE UPDATE ON subscription_invoices
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── subscription_coupons ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscription_coupons (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code              CITEXT NOT NULL UNIQUE,
  name              TEXT NOT NULL,
  description       TEXT,
  -- exactly one of percent_off / amount_off_cents is set
  percent_off       INT CHECK (percent_off IS NULL OR (percent_off > 0 AND percent_off <= 100)),
  amount_off_cents  BIGINT CHECK (amount_off_cents IS NULL OR amount_off_cents > 0),
  currency          TEXT,                                  -- required when amount_off_cents is set
  duration          TEXT NOT NULL DEFAULT 'once'
                    CHECK (duration IN ('once', 'forever', 'repeating')),
  duration_months   INT,                                    -- when duration='repeating'
  max_redemptions   INT,                                    -- null = unlimited
  redemptions       INT NOT NULL DEFAULT 0,
  valid_from        TIMESTAMPTZ,
  valid_until       TIMESTAMPTZ,
  applies_to_plans  JSONB NOT NULL DEFAULT '[]'::jsonb,    -- empty = all plans
  is_active         BOOLEAN NOT NULL DEFAULT true,
  metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ,
  created_by        UUID,
  updated_by        UUID,
  deleted_by        UUID,
  CONSTRAINT coupon_discount_xor CHECK (
    (percent_off IS NOT NULL AND amount_off_cents IS NULL) OR
    (percent_off IS NULL AND amount_off_cents IS NOT NULL)
  )
);

CREATE INDEX IF NOT EXISTS idx_coupons_active ON subscription_coupons
  (is_active) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_coupons_updated_at ON subscription_coupons;
CREATE TRIGGER trg_coupons_updated_at BEFORE UPDATE ON subscription_coupons
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── coupon_redemptions ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS coupon_redemptions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  coupon_id       UUID NOT NULL REFERENCES subscription_coupons(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
  invoice_id      UUID REFERENCES subscription_invoices(id) ON DELETE SET NULL,
  amount_off_cents BIGINT NOT NULL DEFAULT 0,
  redeemed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon ON coupon_redemptions (coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_org ON coupon_redemptions (organization_id);


-- ── permissions ────────────────────────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  ('invoice.read',   'invoice', 'read',   'View subscription invoices', 'subscription', false),
  ('invoice.export', 'invoice', 'export', 'Export / download invoices', 'subscription', false),
  ('coupon.read',    'coupon',  'read',   'View coupons',               'subscription', false),
  ('coupon.create',  'coupon',  'create', 'Issue coupons',              'subscription', true),
  ('coupon.update',  'coupon',  'update', 'Modify coupons',             'subscription', true),
  ('subscription.pause', 'subscription', 'pause', 'Pause subscription', 'subscription', true)
ON CONFLICT (key) DO NOTHING;


-- ── seed test coupons ─────────────────────────────────────────────────────
INSERT INTO subscription_coupons (code, name, description, percent_off, duration, max_redemptions, is_active) VALUES
  ('WELCOME20', 'Welcome — 20% off', 'New customers — first cycle', 20, 'once', 1000, true),
  ('LOYALTY50', 'Loyalty — 50% off',  'Annual rebate, applies for 3 months', 50, 'repeating', NULL, true)
ON CONFLICT (code) DO NOTHING;

UPDATE subscription_coupons SET duration_months = 3 WHERE code = 'LOYALTY50' AND duration_months IS NULL;

INSERT INTO subscription_coupons (code, name, description, amount_off_cents, currency, duration, max_redemptions, is_active) VALUES
  ('FLAT500', 'Flat ₹500 off', 'Flat amount off any plan, one-time', 50000, 'INR', 'once', 500, true)
ON CONFLICT (code) DO NOTHING;
