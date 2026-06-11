-- 014_billing_number_per_tenant.sql — scope numbering uniqueness to tenants.
--
-- Bug: billing_quotations, subscription_invoices, and billing_transactions
-- all declared UNIQUE on the document number alone. NextQuotationNumber /
-- NextInvoiceNumber / NextReceiptNumber compute per-tenant (count rows for
-- THIS tenant + 1), so tenant A and tenant B both produce QUO-2026-000001 →
-- the second INSERT fails with "duplicate key value violates unique
-- constraint billing_quotations_number_key".
--
-- Fix: drop the global UNIQUE; add a composite UNIQUE(tenant_id, number).
-- Existing rows are already per-tenant-unique, so this is a pure constraint
-- swap with no data backfill needed.
--
-- Idempotent: safe to re-run.

BEGIN;

-- ── billing_quotations ────────────────────────────────────────────────────
ALTER TABLE billing_quotations
  DROP CONSTRAINT IF EXISTS billing_quotations_number_key;
CREATE UNIQUE INDEX IF NOT EXISTS uidx_billing_quotations_tenant_number
  ON billing_quotations (tenant_id, number);

-- ── billing_invoices (table was renamed from subscription_invoices in 012;
--    the constraint kept the old name) ──────────────────────────────────────
ALTER TABLE billing_invoices
  DROP CONSTRAINT IF EXISTS subscription_invoices_number_key;
ALTER TABLE billing_invoices
  DROP CONSTRAINT IF EXISTS billing_invoices_number_key;
CREATE UNIQUE INDEX IF NOT EXISTS uidx_billing_invoices_tenant_number
  ON billing_invoices (tenant_id, number);

-- ── billing_transactions (receipt_number) ─────────────────────────────────
ALTER TABLE billing_transactions
  DROP CONSTRAINT IF EXISTS billing_transactions_receipt_number_key;
-- Receipt numbers can be NULL while a transaction is still pending; only
-- enforce uniqueness on rows that have one assigned.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_billing_transactions_tenant_receipt
  ON billing_transactions (tenant_id, receipt_number)
  WHERE receipt_number IS NOT NULL;

COMMIT;
