-- 011_settings_apikeys_webhooks.sql — Developer settings infrastructure.
--
-- Adds API keys (with hashed secrets) and webhooks (with HMAC secret).
-- Both tables are tenant + organization scoped. Permissions seeded so the
-- Settings page can gate the Developer tab.

-- ── api_keys ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS api_keys (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
  name            TEXT NOT NULL,
  prefix          TEXT NOT NULL,                       -- first 8 chars of plaintext key, shown in UI
  token_hash      TEXT NOT NULL UNIQUE,                -- sha256 of plaintext; plaintext never stored
  scopes          JSONB NOT NULL DEFAULT '[]'::jsonb,  -- list of permission keys this token can use
  rate_limit_rpm  INT,                                  -- optional per-key throttle
  last_used_at    TIMESTAMPTZ,
  last_used_ip    INET,
  expires_at      TIMESTAMPTZ,
  revoked_at      TIMESTAMPTZ,
  revoked_by      UUID REFERENCES users(id) ON DELETE SET NULL,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_api_keys_org ON api_keys (organization_id)
  WHERE deleted_at IS NULL AND revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys (user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires ON api_keys (expires_at) WHERE expires_at IS NOT NULL AND deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_api_keys_updated_at ON api_keys;
CREATE TRIGGER trg_api_keys_updated_at BEFORE UPDATE ON api_keys
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── webhooks ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS webhooks (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  url             TEXT NOT NULL,
  events          JSONB NOT NULL DEFAULT '[]'::jsonb,   -- e.g. ["user.created","subscription.cancelled"]
  secret_hash     TEXT,                                  -- HMAC secret (sha256-of-plaintext); plaintext returned once
  is_active       BOOLEAN NOT NULL DEFAULT true,
  description     TEXT,
  headers         JSONB NOT NULL DEFAULT '{}'::jsonb,   -- extra static headers to send
  last_invoked_at TIMESTAMPTZ,
  last_status     INT,                                   -- HTTP status of last delivery
  consecutive_failures INT NOT NULL DEFAULT 0,
  disabled_at     TIMESTAMPTZ,
  disabled_reason TEXT,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE INDEX IF NOT EXISTS idx_webhooks_org_active ON webhooks (organization_id, is_active)
  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON webhooks USING gin (events) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_webhooks_updated_at ON webhooks;
CREATE TRIGGER trg_webhooks_updated_at BEFORE UPDATE ON webhooks
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── webhook_deliveries ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS webhook_deliveries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  webhook_id      UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
  event           TEXT NOT NULL,
  payload         JSONB NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'success', 'failed', 'dropped')),
  attempt         INT NOT NULL DEFAULT 1,
  response_status INT,
  response_body   TEXT,
  error_message   TEXT,
  duration_ms     INT,
  delivered_at    TIMESTAMPTZ,
  next_retry_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_hook ON webhook_deliveries (webhook_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries (status);


-- ── permissions ────────────────────────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  ('api_key.read',    'api_key', 'read',    'View API keys',                 'developer', false),
  ('api_key.create',  'api_key', 'create',  'Issue API keys',                'developer', true),
  ('api_key.delete',  'api_key', 'delete',  'Revoke API keys',               'developer', true),
  ('webhook.read',    'webhook', 'read',    'View webhooks',                 'developer', false),
  ('webhook.create',  'webhook', 'create',  'Register webhooks',             'developer', true),
  ('webhook.update',  'webhook', 'update',  'Modify webhooks',               'developer', true),
  ('webhook.delete',  'webhook', 'delete',  'Remove webhooks',               'developer', true),
  ('webhook.test',    'webhook', 'test',    'Fire test webhook deliveries',  'developer', false)
ON CONFLICT (key) DO NOTHING;
