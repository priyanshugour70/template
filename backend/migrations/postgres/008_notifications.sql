-- 008_notifications.sql — User-facing notifications.
--
-- Per-user notifications surfaced in the dashboard bell. Multi-tenant; soft
-- deletes preserved. List endpoint orders by created_at DESC.

CREATE TABLE IF NOT EXISTS notifications (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind            TEXT NOT NULL DEFAULT 'info'
                  CHECK (kind IN ('info', 'success', 'warning', 'error')),
  title           TEXT NOT NULL,
  message         TEXT,
  link            TEXT,
  is_read         BOOLEAN NOT NULL DEFAULT false,
  read_at         TIMESTAMPTZ,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
  ON notifications (user_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
  ON notifications (user_id)
  WHERE deleted_at IS NULL AND is_read = false;

CREATE INDEX IF NOT EXISTS idx_notifications_tenant
  ON notifications (tenant_id)
  WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_notifications_updated_at ON notifications;
CREATE TRIGGER trg_notifications_updated_at BEFORE UPDATE ON notifications
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── permissions ────────────────────────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  ('notification.read',  'notification', 'read',  'Read own notifications', 'notification', false),
  ('notification.write', 'notification', 'write', 'Create notifications',   'notification', true)
ON CONFLICT (key) DO NOTHING;
