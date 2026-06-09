-- seeds.sql — initial demo data for local development.
--
-- Creates a single tenant ("acme") with one organization, a super-admin user,
-- the three system roles (owner / admin / member) with their permission
-- bundles, and a Free-plan subscription. All inserts are idempotent.
--
-- Login credentials (dev only — change before any real deployment):
--   email:    admin@acme.example
--   password: Admin@123
--   tenantId: 11111111-1111-1111-1111-111111111111

BEGIN;

-- ── tenant ────────────────────────────────────────────────────────────────
INSERT INTO tenants (
  id, slug, name, display_name, description,
  logo_url, primary_color, support_email,
  status, plan_code,
  timezone, locale, currency,
  settings, features, metadata,
  activated_at, created_at, updated_at
) VALUES (
  '11111111-1111-1111-1111-111111111111',
  'acme',
  'Acme Corporation',
  'Acme',
  'Seeded demo tenant for local development.',
  'https://placehold.co/96x96/2E7D32/ffffff?text=A',
  '#2E7D32',
  'support@acme.example',
  'active',
  'free',
  'Asia/Kolkata',
  'en-IN',
  'INR',
  '{}', '{}', '{"seeded": true}',
  now(), now(), now()
) ON CONFLICT (id) DO NOTHING;

-- ── organization ──────────────────────────────────────────────────────────
INSERT INTO organizations (
  id, tenant_id, slug, name, display_name,
  status, is_default,
  timezone, locale, currency,
  settings, features, metadata,
  activated_at, created_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'acme',
  'Acme HQ',
  'Acme Headquarters',
  'active', true,
  'Asia/Kolkata', 'en-IN', 'INR',
  '{}', '{}', '{"seeded": true}',
  now(), now(), now()
) ON CONFLICT (id) DO NOTHING;

-- ── super-admin user (password: Admin@123) ────────────────────────────────
INSERT INTO users (
  id, email, password_hash, password_algo, password_changed_at,
  first_name, last_name, display_name,
  status, locale, timezone,
  is_super_admin,
  primary_tenant_id, primary_organization_id,
  email_verified_at,
  preferences, notification_preferences, metadata,
  created_at, updated_at
) VALUES (
  '33333333-3333-3333-3333-333333333333',
  'admin@acme.example',
  '$2a$12$KjjzLu3HnVsmN/f8t1k5w.5/pPyfA3FpkoGrIOvw2l1EipqMDtyca',
  'bcrypt', now(),
  'Acme', 'Admin', 'Acme Admin',
  'active', 'en-IN', 'Asia/Kolkata',
  true,
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  now(),
  '{}', '{}', '{"seeded": true}',
  now(), now()
) ON CONFLICT (id) DO NOTHING;

-- ── membership (admin user → acme org) ────────────────────────────────────
INSERT INTO memberships (
  id, user_id, tenant_id, organization_id,
  status, is_default, is_owner, is_billing_contact,
  joined_at, settings, metadata,
  created_at, updated_at
) VALUES (
  '44444444-4444-4444-4444-444444444444',
  '33333333-3333-3333-3333-333333333333',
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  'active', true, true, true,
  now(), '{}', '{"seeded": true}',
  now(), now()
) ON CONFLICT (id) DO NOTHING;

-- ── system roles ──────────────────────────────────────────────────────────
INSERT INTO roles (
  id, tenant_id, organization_id, key, name, description,
  is_system, is_default, is_assignable, priority, metadata,
  created_at, updated_at
) VALUES
  ('55555555-5555-5555-5555-555555555555',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'owner', 'Owner', 'Full access to everything in the organization.',
    true, false, true, 100, '{}',
    now(), now()),
  ('66666666-6666-6666-6666-666666666666',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'admin', 'Admin', 'Administrative access. Cannot delete tenant/org or impersonate.',
    true, false, true, 80, '{}',
    now(), now()),
  ('77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'member', 'Member', 'Read-only baseline access.',
    true, true, true, 10, '{}',
    now(), now())
ON CONFLICT (organization_id, key) DO NOTHING;

-- Owner: every permission in the catalog.
INSERT INTO role_permissions (role_id, permission_id, granted_at)
SELECT '55555555-5555-5555-5555-555555555555', p.id, now()
FROM permissions p
WHERE p.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- Admin: everything except dangerous deletes + impersonate.
INSERT INTO role_permissions (role_id, permission_id, granted_at)
SELECT '66666666-6666-6666-6666-666666666666', p.id, now()
FROM permissions p
WHERE p.deleted_at IS NULL
  AND p.key NOT IN ('tenant.delete', 'org.delete', 'user.impersonate')
ON CONFLICT DO NOTHING;

-- Member: read/list permissions only.
INSERT INTO role_permissions (role_id, permission_id, granted_at)
SELECT '77777777-7777-7777-7777-777777777777', p.id, now()
FROM permissions p
WHERE p.deleted_at IS NULL
  AND (p.action = 'read' OR p.action = 'list')
ON CONFLICT DO NOTHING;

-- ── assign owner role to seed admin's membership ─────────────────────────
INSERT INTO membership_roles (membership_id, role_id, granted_at)
VALUES (
  '44444444-4444-4444-4444-444444444444',
  '55555555-5555-5555-5555-555555555555',
  now()
) ON CONFLICT DO NOTHING;

-- ── seed subscription (Free plan, active) ────────────────────────────────
-- Table was renamed in migration 012 (subscription_plans → billing_plans,
-- subscriptions → billing_subscriptions). Same column shape, new prefixes.
INSERT INTO billing_subscriptions (
  id, tenant_id, organization_id, plan_id, plan_code,
  status, billing_cycle, quantity, unit_price_cents, total_cents, currency,
  started_at, current_period_start, current_period_end,
  features, limits, metadata,
  created_at, updated_at
)
SELECT
  '88888888-8888-8888-8888-888888888888',
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  p.id, p.code,
  'active', p.billing_cycle, 1, p.price_cents, p.price_cents, p.currency,
  now(), now(), now() + interval '30 days',
  p.features, p.limits, '{"seeded": true}',
  now(), now()
FROM billing_plans p
WHERE p.code = 'free'
ON CONFLICT (id) DO NOTHING;

COMMIT;
