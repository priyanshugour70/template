-- 012_billing_overhaul.sql
-- Phase 1 of the billing overhaul: rename subscription_* tables to billing_*,
-- add columns for GST + PDF storage + per-feature plan composition, create the
-- new tables (features catalog, plan_features, quotations, invoice_lines,
-- transactions, tax_config), and rename the subscription.* permission keys.
--
-- Data preservation: existing rows (the seeded Acme tenant + its trial sub)
-- survive the rename. Subsequent phases will evolve the data model further
-- (e.g. retire the free / pro / enterprise preset rows and keep only Starter).

BEGIN;

-- ── rename existing tables ────────────────────────────────────────────────

ALTER TABLE IF EXISTS subscription_plans          RENAME TO billing_plans;
ALTER TABLE IF EXISTS subscriptions               RENAME TO billing_subscriptions;
ALTER TABLE IF EXISTS usage_counters              RENAME TO billing_usage_counters;
ALTER TABLE IF EXISTS subscription_invoices       RENAME TO billing_invoices;
ALTER TABLE IF EXISTS subscription_coupons        RENAME TO billing_coupons;
ALTER TABLE IF EXISTS coupon_redemptions          RENAME TO billing_coupon_redemptions;

-- ── new columns on renamed tables ─────────────────────────────────────────

-- billing_plans: mark preset-vs-custom and store an aggregated price.
ALTER TABLE billing_plans
  ADD COLUMN IF NOT EXISTS is_custom BOOLEAN NOT NULL DEFAULT false;

-- billing_subscriptions: customer's billing-state for GST place-of-supply.
ALTER TABLE billing_subscriptions
  ADD COLUMN IF NOT EXISTS billing_state TEXT;

-- billing_invoices: GST breakdown + PDF storage + HSN/SAC code.
ALTER TABLE billing_invoices
  ADD COLUMN IF NOT EXISTS hsn_sac           TEXT NOT NULL DEFAULT '998314',
  ADD COLUMN IF NOT EXISTS place_of_supply   TEXT,
  ADD COLUMN IF NOT EXISTS cgst_cents        BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS sgst_cents        BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS igst_cents        BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS pdf_storage_key   TEXT;

-- ── new table: feature catalog ────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS billing_features (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key                    TEXT NOT NULL UNIQUE,
  name                   TEXT NOT NULL,
  description            TEXT NOT NULL DEFAULT '',
  category               TEXT NOT NULL
                         CHECK (category IN ('core', 'admin', 'compliance', 'integrations', 'limits')),
  base_price_cents       BIGINT NOT NULL DEFAULT 0,
  per_user_price_cents   BIGINT NOT NULL DEFAULT 0,
  included_users         INT NOT NULL DEFAULT 0,
  is_core                BOOLEAN NOT NULL DEFAULT false,
  is_starter_default     BOOLEAN NOT NULL DEFAULT false,
  is_active              BOOLEAN NOT NULL DEFAULT true,
  requires               TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  metadata               JSONB NOT NULL DEFAULT '{}'::jsonb,
  sort_order             INT NOT NULL DEFAULT 0,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at             TIMESTAMPTZ,
  created_by             UUID,
  updated_by             UUID,
  deleted_by             UUID
);

CREATE INDEX IF NOT EXISTS idx_billing_features_active_cat ON billing_features (is_active, category, sort_order) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_billing_features_updated_at ON billing_features;
CREATE TRIGGER trg_billing_features_updated_at BEFORE UPDATE ON billing_features FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── new table: plan ↔ feature junction ────────────────────────────────────

CREATE TABLE IF NOT EXISTS billing_plan_features (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  plan_id                UUID NOT NULL REFERENCES billing_plans(id) ON DELETE CASCADE,
  feature_id             UUID NOT NULL REFERENCES billing_features(id) ON DELETE RESTRICT,
  feature_key            TEXT NOT NULL,
  -- Snapshot of pricing at the moment the plan was activated. Catalog price
  -- changes do NOT retroactively re-bill — we charge what the customer agreed.
  base_price_cents       BIGINT NOT NULL DEFAULT 0,
  per_user_price_cents   BIGINT NOT NULL DEFAULT 0,
  included_users         INT NOT NULL DEFAULT 0,
  -- For "extra_users"-style features: how many units the customer added.
  quantity               INT NOT NULL DEFAULT 1,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (plan_id, feature_id)
);

CREATE INDEX IF NOT EXISTS idx_billing_plan_features_plan ON billing_plan_features (plan_id);
CREATE INDEX IF NOT EXISTS idx_billing_plan_features_feature ON billing_plan_features (feature_id);

DROP TRIGGER IF EXISTS trg_billing_plan_features_updated_at ON billing_plan_features;
CREATE TRIGGER trg_billing_plan_features_updated_at BEFORE UPDATE ON billing_plan_features FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── new table: quotations (draft plans) ───────────────────────────────────

CREATE TABLE IF NOT EXISTS billing_quotations (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id              UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  number                 TEXT NOT NULL UNIQUE,
  status                 TEXT NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft', 'accepted', 'rejected', 'expired')),
  -- The selected feature keys + per-user count + computed totals at draft time.
  feature_keys           TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  user_count             INT NOT NULL DEFAULT 1,
  subtotal_cents         BIGINT NOT NULL DEFAULT 0,
  discount_cents         BIGINT NOT NULL DEFAULT 0,
  cgst_cents             BIGINT NOT NULL DEFAULT 0,
  sgst_cents             BIGINT NOT NULL DEFAULT 0,
  igst_cents             BIGINT NOT NULL DEFAULT 0,
  total_cents            BIGINT NOT NULL DEFAULT 0,
  currency               TEXT NOT NULL DEFAULT 'INR',
  place_of_supply        TEXT,
  -- Line-item snapshot for the draft (rendered straight into the quotation PDF).
  line_items             JSONB NOT NULL DEFAULT '[]'::jsonb,
  billing_email          CITEXT,
  billing_name           TEXT,
  billing_address        JSONB,
  billing_state          TEXT,
  notes                  TEXT,
  expires_at             TIMESTAMPTZ NOT NULL,
  accepted_at            TIMESTAMPTZ,
  rejected_at            TIMESTAMPTZ,
  -- When activated, this points to the plan + subscription rows it produced.
  activated_plan_id      UUID REFERENCES billing_plans(id) ON DELETE SET NULL,
  activated_subscription_id UUID REFERENCES billing_subscriptions(id) ON DELETE SET NULL,
  metadata               JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at             TIMESTAMPTZ,
  created_by             UUID,
  updated_by             UUID,
  deleted_by             UUID
);

CREATE INDEX IF NOT EXISTS idx_billing_quotations_org ON billing_quotations (organization_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_billing_quotations_status ON billing_quotations (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_billing_quotations_expires ON billing_quotations (expires_at) WHERE deleted_at IS NULL AND status = 'draft';

DROP TRIGGER IF EXISTS trg_billing_quotations_updated_at ON billing_quotations;
CREATE TRIGGER trg_billing_quotations_updated_at BEFORE UPDATE ON billing_quotations FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── new table: relational invoice lines ───────────────────────────────────
-- (existing JSONB line_items on billing_invoices stays for now; both will be
-- populated until Phase 4 cuts over the read path. Phase 10 drops the JSONB
-- copy.)

CREATE TABLE IF NOT EXISTS billing_invoice_lines (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id             UUID NOT NULL REFERENCES billing_invoices(id) ON DELETE CASCADE,
  feature_key            TEXT,
  description            TEXT NOT NULL,
  hsn_sac                TEXT NOT NULL DEFAULT '998314',
  quantity               INT NOT NULL DEFAULT 1,
  unit_price_cents       BIGINT NOT NULL DEFAULT 0,
  taxable_amount_cents   BIGINT NOT NULL DEFAULT 0,
  cgst_cents             BIGINT NOT NULL DEFAULT 0,
  sgst_cents             BIGINT NOT NULL DEFAULT 0,
  igst_cents             BIGINT NOT NULL DEFAULT 0,
  total_cents            BIGINT NOT NULL DEFAULT 0,
  sort_order             INT NOT NULL DEFAULT 0,
  metadata               JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_billing_invoice_lines_invoice ON billing_invoice_lines (invoice_id, sort_order);

-- ── new table: payment transactions ───────────────────────────────────────

CREATE TABLE IF NOT EXISTS billing_transactions (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id              UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  invoice_id             UUID NOT NULL REFERENCES billing_invoices(id) ON DELETE RESTRICT,
  receipt_number         TEXT NOT NULL UNIQUE,
  method                 TEXT NOT NULL
                         CHECK (method IN ('cash', 'bank_transfer', 'cheque', 'gateway')),
  status                 TEXT NOT NULL DEFAULT 'recorded'
                         CHECK (status IN ('recorded', 'pending', 'failed', 'refunded')),
  amount_cents           BIGINT NOT NULL,
  currency               TEXT NOT NULL DEFAULT 'INR',
  reference              TEXT,                -- bank txn ref / cheque number / gateway txn id
  gateway                TEXT,                -- 'razorpay', 'stripe', null for cash
  gateway_transaction_id TEXT,
  paid_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  refunded_at            TIMESTAMPTZ,
  refund_amount_cents    BIGINT NOT NULL DEFAULT 0,
  pdf_storage_key        TEXT,
  notes                  TEXT,
  metadata               JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at             TIMESTAMPTZ,
  created_by             UUID,
  updated_by             UUID,
  deleted_by             UUID
);

CREATE INDEX IF NOT EXISTS idx_billing_transactions_org_paid ON billing_transactions (organization_id, paid_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_billing_transactions_invoice ON billing_transactions (invoice_id);

DROP TRIGGER IF EXISTS trg_billing_transactions_updated_at ON billing_transactions;
CREATE TRIGGER trg_billing_transactions_updated_at BEFORE UPDATE ON billing_transactions FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── new table: tax config (single row) ────────────────────────────────────

CREATE TABLE IF NOT EXISTS billing_tax_config (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- Only one row is meant to exist; we enforce that with a partial unique index.
  singleton              BOOLEAN NOT NULL DEFAULT true,
  company_name           TEXT NOT NULL DEFAULT '',
  company_address        TEXT NOT NULL DEFAULT '',
  gstin                  TEXT NOT NULL DEFAULT '',
  home_state             TEXT NOT NULL DEFAULT 'Karnataka',
  default_cgst_pct       NUMERIC(5, 2) NOT NULL DEFAULT 9.00,
  default_sgst_pct       NUMERIC(5, 2) NOT NULL DEFAULT 9.00,
  default_igst_pct       NUMERIC(5, 2) NOT NULL DEFAULT 18.00,
  default_hsn_sac        TEXT NOT NULL DEFAULT '998314',
  currency               TEXT NOT NULL DEFAULT 'INR',
  bank_name              TEXT NOT NULL DEFAULT '',
  bank_account_number    TEXT NOT NULL DEFAULT '',
  bank_ifsc              TEXT NOT NULL DEFAULT '',
  bank_account_name      TEXT NOT NULL DEFAULT '',
  invoice_terms          TEXT NOT NULL DEFAULT 'Payment due within 7 days. Computer-generated invoice — no signature required.',
  metadata               JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_billing_tax_config_singleton ON billing_tax_config (singleton);

DROP TRIGGER IF EXISTS trg_billing_tax_config_updated_at ON billing_tax_config;
CREATE TRIGGER trg_billing_tax_config_updated_at BEFORE UPDATE ON billing_tax_config FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── seed feature catalog ──────────────────────────────────────────────────

INSERT INTO billing_features
  (key, name, description, category, base_price_cents, per_user_price_cents, included_users, is_core, is_starter_default, requires, sort_order)
VALUES
  -- core (always included, free)
  ('auth',              'Authentication & sessions',        'Sign-in, MFA, refresh tokens, session management.',
   'core',          0,    0,  0, true,  true,  ARRAY[]::TEXT[], 10),
  ('settings',          'Settings hub',                     'Profile, security, notifications and preferences pages.',
   'core',          0,    0,  0, true,  true,  ARRAY[]::TEXT[], 20),
  ('user_management',   'User management',                  'Invite users, manage profiles, suspend/reactivate.',
   'core',          0,    0,  0, true,  true,  ARRAY[]::TEXT[], 30),
  ('basic_rbac',        'System roles',                     'Owner / Admin / Member built-in roles.',
   'core',          0,    0,  0, true,  true,  ARRAY[]::TEXT[], 40),
  ('notifications',     'In-app notifications',             'Notification bell, real-time alerts, mark-as-read.',
   'core',          0,    0,  0, true,  true,  ARRAY[]::TEXT[], 50),

  -- limits / pricing knobs
  ('starter_bundle',    'Starter package',                  'Bundles the MVP features for one organization with up to 10 users.',
   'limits',  1000000,    0, 10, false, true,  ARRAY[]::TEXT[], 100),
  ('extra_users',       'Additional users',                 'Adds capacity beyond the included user quota.',
   'limits',        0,  50000, 0, false, false, ARRAY[]::TEXT[], 110),

  -- admin / governance add-ons
  ('custom_roles',      'Custom roles & permissions',       'Define your own roles beyond Owner/Admin/Member.',
   'admin',    200000,    0,  0, false, false, ARRAY['basic_rbac']::TEXT[], 200),
  ('multi_org',         'Multiple organizations',           'Run several organizations under one tenant.',
   'admin',    300000,    0,  0, false, false, ARRAY[]::TEXT[], 210),
  ('departments',       'Department hierarchy',             'Nested departments with inherited role grants.',
   'admin',    250000,    0,  0, false, false, ARRAY[]::TEXT[], 220),
  ('groups',            'Groups',                           'Flat cross-cutting user groups for ad-hoc role bundles.',
   'admin',    200000,    0,  0, false, false, ARRAY[]::TEXT[], 230),

  -- compliance
  ('audit_log',         'Audit log',                        'Search and inspect every API request the workspace made.',
   'compliance', 150000,  0,  0, false, true,  ARRAY[]::TEXT[], 300),
  ('audit_export',      'Audit CSV export',                 'Export filtered audit logs as CSV for compliance reviews.',
   'compliance', 100000,  0,  0, false, false, ARRAY['audit_log']::TEXT[], 310),

  -- integrations
  ('webhooks',          'Outgoing webhooks',                'HMAC-signed webhooks for system events.',
   'integrations', 250000, 0,  0, false, false, ARRAY[]::TEXT[], 400),
  ('api_keys',          'API keys',                         'Issue scoped API tokens for headless integrations and CI.',
   'integrations', 250000, 0,  0, false, false, ARRAY[]::TEXT[], 410)
ON CONFLICT (key) DO NOTHING;

-- ── seed tax config row ───────────────────────────────────────────────────

INSERT INTO billing_tax_config (singleton, company_name, home_state, default_cgst_pct, default_sgst_pct, default_igst_pct, default_hsn_sac)
VALUES (true, 'Your Company Pvt Ltd', 'Karnataka', 9.00, 9.00, 18.00, '998314')
ON CONFLICT (singleton) DO NOTHING;

-- ── permissions: rename subscription.* → billing.* + add new ones ─────────

-- Rename existing keys in place. permissions.key is unique so each UPDATE is safe.
UPDATE permissions SET key = 'billing.read',   category = 'billing' WHERE key = 'subscription.read';
UPDATE permissions SET key = 'billing.manage', category = 'billing' WHERE key = 'subscription.update';
UPDATE permissions SET key = 'billing.cancel', category = 'billing' WHERE key = 'subscription.cancel';

-- Pause/resume goes away in the new model; drop both the perm and any role
-- assignments referencing it.
DELETE FROM role_permissions
  WHERE permission_id IN (SELECT id FROM permissions WHERE key = 'subscription.pause');
DELETE FROM permissions WHERE key = 'subscription.pause';

-- Coupon perms move to billing.coupon.* (kept around even though the new UX
-- doesn't surface coupons heavily — backend code still supports them in
-- Phase 1 for backwards compat).
UPDATE permissions SET key = 'billing.coupon.read',   category = 'billing' WHERE key = 'coupon.read';
UPDATE permissions SET key = 'billing.coupon.create', category = 'billing' WHERE key = 'coupon.create';
UPDATE permissions SET key = 'billing.coupon.update', category = 'billing' WHERE key = 'coupon.update';
UPDATE permissions SET key = 'billing.coupon.delete', category = 'billing' WHERE key = 'coupon.delete';

-- Existing invoice.* perms get rehomed under billing.invoice.*
UPDATE permissions SET key = 'billing.invoice.read',   category = 'billing' WHERE key = 'invoice.read';
UPDATE permissions SET key = 'billing.invoice.export', category = 'billing' WHERE key = 'invoice.export';

-- New permissions for the new flows.
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  ('billing.invoice.pay',      'invoice',     'pay',     'Mark invoices as paid by cash, bank transfer, or cheque.',  'billing', false),
  ('billing.transaction.read', 'transaction', 'read',    'View the payments ledger and download receipts.',           'billing', false),
  ('billing.admin',            'billing',     'admin',   'Configure tax settings, feature pricing, and bank details.','billing', true),
  ('billing.quotation.read',   'quotation',   'read',    'View draft and historical quotations.',                     'billing', false),
  ('billing.quotation.manage', 'quotation',   'manage',  'Create, edit, and activate quotations into subscriptions.', 'billing', false)
ON CONFLICT (key) DO NOTHING;

-- Owner / Admin already received subscription.* perms in 010; the UPDATEs above
-- propagated the rename via the role_permissions FK. Grant the brand-new perms
-- to every role that previously held subscription.update (now billing.manage).
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM role_permissions rp
JOIN permissions src ON src.id = rp.permission_id AND src.key = 'billing.manage'
CROSS JOIN permissions p
WHERE p.key IN (
  'billing.invoice.pay',
  'billing.transaction.read',
  'billing.quotation.read',
  'billing.quotation.manage'
)
ON CONFLICT DO NOTHING;

COMMIT;
