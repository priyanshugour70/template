-- 013_billing_plan_features_sync.sql — align seeded plan features with the
-- billing_features catalog.
--
-- The original seed (006_subscription.sql) used dotted machine names like
-- "user.invite", "audit.read", "sso" — none of which exist in the
-- billing_features catalog populated by 012_billing_overhaul.sql, so the
-- quotation pricing engine rejected every plan with "unknown feature ...".
--
-- This migration overwrites the four canonical plan rows (free, starter, pro,
-- enterprise) with feature keys that DO exist in billing_features. Custom
-- plans (code starting with 'custom-') are left untouched.
--
-- Idempotent: re-running just re-sets the same JSONB arrays.

BEGIN;

UPDATE billing_plans
SET features = '["auth","basic_rbac","user_management","notifications","audit_log"]'::jsonb
WHERE code = 'free';

UPDATE billing_plans
SET features = '["auth","basic_rbac","user_management","notifications","audit_log","settings","multi_org","audit_export","starter_bundle"]'::jsonb
WHERE code = 'starter';

UPDATE billing_plans
SET features = '["auth","basic_rbac","user_management","notifications","audit_log","settings","multi_org","audit_export","starter_bundle","custom_roles","departments","groups","webhooks","api_keys"]'::jsonb
WHERE code = 'pro';

UPDATE billing_plans
SET features = '["auth","basic_rbac","user_management","notifications","audit_log","settings","multi_org","audit_export","starter_bundle","custom_roles","departments","groups","webhooks","api_keys","extra_users"]'::jsonb
WHERE code = 'enterprise';

COMMIT;
