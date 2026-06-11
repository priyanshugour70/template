-- 015_comm_module.sql — Communication module Phase 1.
--
-- Tables:
--   comm_conversations         — DMs and channels share this table (type discriminator)
--   comm_conversation_members  — who's in each conversation + per-user state (unread, prefs)
--   comm_messages              — every message; soft-deleted, threadable
--   comm_message_mentions      — extracted @mentions per message (user|here|channel|everyone)
--   comm_message_reactions     — emoji reactions
--
-- Permissions seeded under the "Communication" category and bound to the seed
-- system roles (owner, admin, member) across every existing organisation, so
-- the feature works for current tenants the moment the migration runs.

BEGIN;

-- ── conversations ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comm_conversations (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id    UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  type               TEXT NOT NULL CHECK (type IN ('dm', 'channel')),

  -- channels only
  slug               CITEXT,
  name               TEXT,
  topic              TEXT,
  description        TEXT,
  is_private         BOOLEAN NOT NULL DEFAULT false,

  -- DMs only — sorted "userA:userB" key for natural per-org uniqueness so
  -- creating a DM between the same two people twice idempotently returns the
  -- existing conversation instead of creating a duplicate.
  dm_key             TEXT,

  created_by_user    UUID,
  archived_at        TIMESTAMPTZ,

  -- Denormalised for fast sidebar sorting and unread badges.
  last_message_at    TIMESTAMPTZ,
  message_count      INTEGER NOT NULL DEFAULT 0,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at         TIMESTAMPTZ,
  created_by         UUID,
  updated_by         UUID,
  deleted_by         UUID
);

-- Channel slugs are unique per org. NULL slug (DMs) excluded.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_comm_conversations_channel_slug
  ON comm_conversations (organization_id, slug)
  WHERE type = 'channel' AND slug IS NOT NULL AND deleted_at IS NULL;

-- DMs unique per (org, pair-of-users).
CREATE UNIQUE INDEX IF NOT EXISTS uidx_comm_conversations_dm_key
  ON comm_conversations (organization_id, dm_key)
  WHERE type = 'dm' AND dm_key IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_conversations_org_activity
  ON comm_conversations (organization_id, last_message_at DESC NULLS LAST)
  WHERE deleted_at IS NULL AND archived_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_conversations_tenant
  ON comm_conversations (tenant_id);

DROP TRIGGER IF EXISTS trg_comm_conversations_updated_at ON comm_conversations;
CREATE TRIGGER trg_comm_conversations_updated_at
  BEFORE UPDATE ON comm_conversations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── conversation members ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comm_conversation_members (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id      UUID NOT NULL REFERENCES comm_conversations(id) ON DELETE CASCADE,
  user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- The tenant-membership the user occupies. Optional, but used by the
  -- notification trigger so a user who leaves the tenant stops receiving
  -- notifications even if they're still a comm member somehow.
  membership_id        UUID,

  role                 TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),

  joined_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  left_at              TIMESTAMPTZ,

  -- Read state, denormalised so the sidebar query stays single-table.
  last_read_message_id UUID,
  last_read_at         TIMESTAMPTZ,
  unread_count         INTEGER NOT NULL DEFAULT 0,

  notification_pref    TEXT NOT NULL DEFAULT 'all' CHECK (notification_pref IN ('all', 'mentions', 'none')),
  muted_until          TIMESTAMPTZ,

  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at           TIMESTAMPTZ,
  created_by           UUID,
  updated_by           UUID,
  deleted_by           UUID
);

-- Only one ACTIVE membership per (conversation, user). Re-joining a left
-- conversation reuses the row via a separate UPDATE.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_comm_members_conv_user_active
  ON comm_conversation_members (conversation_id, user_id)
  WHERE left_at IS NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_members_user_active
  ON comm_conversation_members (user_id) WHERE left_at IS NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_members_conversation
  ON comm_conversation_members (conversation_id);

DROP TRIGGER IF EXISTS trg_comm_members_updated_at ON comm_conversation_members;
CREATE TRIGGER trg_comm_members_updated_at
  BEFORE UPDATE ON comm_conversation_members
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── messages ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comm_messages (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id      UUID NOT NULL REFERENCES comm_conversations(id) ON DELETE CASCADE,
  -- Redundant with conversations.tenant_id/organization_id, but indexed
  -- separately so cross-conversation queries (e.g. "all messages where I'm
  -- mentioned in tenant X") don't require joins.
  tenant_id            UUID NOT NULL,
  organization_id      UUID NOT NULL,

  -- Threading. NULL = top-level message.
  parent_message_id    UUID REFERENCES comm_messages(id) ON DELETE SET NULL,

  sender_type          TEXT NOT NULL CHECK (sender_type IN ('user', 'system', 'webhook')),
  sender_user_id       UUID,
  sender_webhook_id    UUID,

  body                 TEXT NOT NULL DEFAULT '',
  body_format          TEXT NOT NULL DEFAULT 'markdown' CHECK (body_format IN ('markdown', 'plain')),
  attachments          JSONB NOT NULL DEFAULT '[]'::jsonb,
  metadata             JSONB NOT NULL DEFAULT '{}'::jsonb,

  edited_at            TIMESTAMPTZ,

  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at           TIMESTAMPTZ,
  created_by           UUID,
  updated_by           UUID,
  deleted_by           UUID
);

-- Cursor pagination = (conversation_id, created_at DESC, id DESC). DESC NULLS
-- LAST is the natural index for "most recent first".
CREATE INDEX IF NOT EXISTS idx_comm_messages_conv_timeline
  ON comm_messages (conversation_id, created_at DESC, id DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_messages_sender_user
  ON comm_messages (sender_user_id, created_at DESC)
  WHERE sender_user_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_comm_messages_parent
  ON comm_messages (parent_message_id)
  WHERE parent_message_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_comm_messages_tenant
  ON comm_messages (tenant_id);

DROP TRIGGER IF EXISTS trg_comm_messages_updated_at ON comm_messages;
CREATE TRIGGER trg_comm_messages_updated_at
  BEFORE UPDATE ON comm_messages
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── mentions ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comm_message_mentions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  message_id      UUID NOT NULL REFERENCES comm_messages(id) ON DELETE CASCADE,
  mention_type    TEXT NOT NULL CHECK (mention_type IN ('user', 'here', 'channel', 'everyone')),
  target_user_id  UUID REFERENCES users(id) ON DELETE SET NULL,
  index_in_body   INTEGER NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_comm_mentions_message
  ON comm_message_mentions (message_id);

-- "all messages where I'm mentioned" — listed by recency.
CREATE INDEX IF NOT EXISTS idx_comm_mentions_target_user
  ON comm_message_mentions (target_user_id, created_at DESC)
  WHERE target_user_id IS NOT NULL;

-- ── reactions ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comm_message_reactions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  message_id  UUID NOT NULL REFERENCES comm_messages(id) ON DELETE CASCADE,
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  emoji       TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_comm_reactions_unique
  ON comm_message_reactions (message_id, user_id, emoji);

CREATE INDEX IF NOT EXISTS idx_comm_reactions_message
  ON comm_message_reactions (message_id);

-- ── RBAC permissions ───────────────────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_system, is_dangerous, metadata)
VALUES
  ('comm.read',                'comm', 'read',     'View conversations and messages',   'Communication', true, false, '{}'::jsonb),
  ('comm.channel.create',      'comm', 'create',   'Create new channels',                'Communication', true, false, '{}'::jsonb),
  ('comm.channel.manage',      'comm', 'manage',   'Rename channel, manage members',     'Communication', true, false, '{}'::jsonb),
  ('comm.message.send',        'comm', 'send',     'Post messages and reactions',        'Communication', true, false, '{}'::jsonb),
  ('comm.message.moderate',    'comm', 'moderate', 'Delete other users'' messages',     'Communication', true, true,  '{}'::jsonb),
  ('comm.inbound_hook.manage', 'comm', 'manage',   'Manage inbound channel webhooks',   'Communication', true, false, '{}'::jsonb)
ON CONFLICT (key) DO NOTHING;

-- Bind to existing system roles across every organisation. The seed roles
-- ("owner", "admin", "member") are recreated per-org by tenant onboarding,
-- so we grant comm perms to every row that exists today.

-- Members get the basics: read, create channels, send messages.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.key = 'member' AND r.deleted_at IS NULL
  AND p.key IN ('comm.read', 'comm.channel.create', 'comm.message.send')
ON CONFLICT DO NOTHING;

-- Owners + admins get everything in the module.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.key IN ('owner', 'admin') AND r.deleted_at IS NULL
  AND p.key IN (
    'comm.read', 'comm.channel.create', 'comm.channel.manage',
    'comm.message.send', 'comm.message.moderate', 'comm.inbound_hook.manage'
  )
ON CONFLICT DO NOTHING;

COMMIT;
