-- 016_comm_phase2.sql — Communication module Phase 2.
--
-- Adds inbound channel webhooks. The realtime stack (WebSockets, presence,
-- typing) is Redis-backed and stateless — no schema needed.

BEGIN;

CREATE TABLE IF NOT EXISTS comm_channel_hooks (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  conversation_id UUID NOT NULL REFERENCES comm_conversations(id) ON DELETE CASCADE,

  -- Display fields.
  name            TEXT NOT NULL,
  icon_url        TEXT,
  display_name    TEXT,

  -- Token hashed (SHA256). The raw token is returned ONCE at creation and
  -- never persisted. Token length matches tokens.New(32) → base64 ~43 chars.
  token_hash      BYTEA NOT NULL,

  is_active       BOOLEAN NOT NULL DEFAULT true,
  last_used_at    TIMESTAMPTZ,
  use_count       BIGINT NOT NULL DEFAULT 0,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_comm_channel_hooks_token
  ON comm_channel_hooks (token_hash) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_channel_hooks_conversation
  ON comm_channel_hooks (conversation_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_comm_channel_hooks_updated_at ON comm_channel_hooks;
CREATE TRIGGER trg_comm_channel_hooks_updated_at
  BEFORE UPDATE ON comm_channel_hooks
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
