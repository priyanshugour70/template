-- 004_rbac.sql — Permissions, Roles, Role↔Permission, Membership↔Role.

-- ── permissions (catalog, seeded) ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS permissions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key          TEXT NOT NULL UNIQUE,
  resource     TEXT NOT NULL,
  action       TEXT NOT NULL,
  description  TEXT,
  category     TEXT,
  is_system    BOOLEAN NOT NULL DEFAULT true,
  is_dangerous BOOLEAN NOT NULL DEFAULT false,
  metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at   TIMESTAMPTZ,
  created_by   UUID,
  updated_by   UUID,
  deleted_by   UUID
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions (resource) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions (category) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_permissions_updated_at ON permissions;
CREATE TRIGGER trg_permissions_updated_at
  BEFORE UPDATE ON permissions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── roles ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  key             TEXT NOT NULL,
  name            TEXT NOT NULL,
  description     TEXT,
  is_system       BOOLEAN NOT NULL DEFAULT false,
  is_default      BOOLEAN NOT NULL DEFAULT false,
  is_assignable   BOOLEAN NOT NULL DEFAULT true,
  priority        INT NOT NULL DEFAULT 0,
  color           TEXT,
  icon            TEXT,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID,
  UNIQUE (organization_id, key)
);

CREATE INDEX IF NOT EXISTS idx_roles_org ON roles (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_tenant ON roles (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_default ON roles (organization_id, is_default) WHERE deleted_at IS NULL AND is_default = true;

DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;
CREATE TRIGGER trg_roles_updated_at
  BEFORE UPDATE ON roles
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── role_permissions ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_permissions (
  role_id        UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id  UUID NOT NULL REFERENCES permissions(id) ON DELETE RESTRICT,
  granted_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by     UUID,
  PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_perm ON role_permissions (permission_id);


-- ── membership_roles ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS membership_roles (
  membership_id  UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
  role_id        UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
  granted_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by     UUID,
  expires_at     TIMESTAMPTZ,
  PRIMARY KEY (membership_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_membership_roles_role ON membership_roles (role_id);


-- ── seed system permissions catalog ────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  -- tenant
  ('tenant.read',          'tenant', 'read',          'View tenant profile',            'tenant', false),
  ('tenant.update',        'tenant', 'update',        'Edit tenant profile',            'tenant', false),
  ('tenant.delete',        'tenant', 'delete',        'Archive a tenant',               'tenant', true),
  -- organization
  ('org.read',             'org',    'read',          'View organization',              'org',    false),
  ('org.list',             'org',    'list',          'List organizations',             'org',    false),
  ('org.create',           'org',    'create',        'Create organizations',           'org',    false),
  ('org.update',           'org',    'update',        'Edit organization',              'org',    false),
  ('org.delete',           'org',    'delete',        'Archive organization',           'org',    true),
  -- user
  ('user.read',            'user',   'read',          'View user details',              'user',   false),
  ('user.list',            'user',   'list',          'List users',                     'user',   false),
  ('user.invite',          'user',   'invite',        'Invite users',                   'user',   false),
  ('user.update',          'user',   'update',        'Edit users',                     'user',   false),
  ('user.suspend',         'user',   'suspend',       'Suspend/reactivate users',       'user',   true),
  ('user.delete',          'user',   'delete',        'Archive users',                  'user',   true),
  ('user.assign_role',     'user',   'assign_role',   'Assign or remove roles',         'user',   true),
  ('user.impersonate',     'user',   'impersonate',   'Impersonate another user',       'user',   true),
  -- role
  ('role.read',            'role',   'read',          'View roles',                     'role',   false),
  ('role.list',            'role',   'list',          'List roles',                     'role',   false),
  ('role.create',          'role',   'create',        'Create custom roles',            'role',   false),
  ('role.update',          'role',   'update',        'Edit roles',                     'role',   false),
  ('role.delete',          'role',   'delete',        'Archive roles',                  'role',   true),
  ('role.assign',          'role',   'assign',        'Assign permissions to roles',    'role',   true),
  -- subscription
  ('subscription.read',    'subscription', 'read',    'View subscription',              'subscription', false),
  ('subscription.update',  'subscription', 'update',  'Change subscription plan',       'subscription', true),
  ('subscription.cancel',  'subscription', 'cancel',  'Cancel subscription',            'subscription', true),
  -- audit
  ('audit.read',           'audit',  'read',          'View audit log',                 'audit',  false),
  ('audit.export',         'audit',  'export',        'Export audit log',               'audit',  false),
  -- files
  ('file.upload',          'file',   'upload',        'Upload files',                   'file',   false),
  ('file.read',            'file',   'read',          'Download files',                 'file',   false),
  ('file.delete',          'file',   'delete',        'Delete files',                   'file',   true)
ON CONFLICT (key) DO NOTHING;
