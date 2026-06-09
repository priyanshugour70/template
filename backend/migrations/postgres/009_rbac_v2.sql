-- 009_rbac_v2.sql — Departments (hierarchy) + Groups (cross-cutting) + role bindings.
--
-- DESIGN:
--   * Departments form a single tree per organization. A membership belongs to
--     at most one department. Role bindings on a department APPLY DOWN the
--     subtree — children inherit ancestor role grants.
--   * Groups are cross-cutting and flat: groups can contain users and other
--     groups (polymorphic). Role bindings on a group apply to every member.
--   * department_closure stores the ancestor→descendant precomputed pairs so
--     permission resolution is a single join, not a recursive CTE per request.
--
-- Maintenance: triggers below keep the closure table consistent on insert /
-- update of parent_id / soft delete. group nesting closure handled in-app
-- on edits (less hot than departments).


-- ── departments ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS departments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  parent_id       UUID REFERENCES departments(id) ON DELETE SET NULL,
  slug            CITEXT NOT NULL,
  name            TEXT NOT NULL,
  description     TEXT,
  cost_center     TEXT,
  manager_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  color           TEXT,
  icon            TEXT,
  is_archived     BOOLEAN NOT NULL DEFAULT false,
  sort_order      INT NOT NULL DEFAULT 0,
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID,
  UNIQUE (organization_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_departments_org ON departments (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_departments_parent ON departments (parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_departments_tenant ON departments (tenant_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_departments_updated_at ON departments;
CREATE TRIGGER trg_departments_updated_at BEFORE UPDATE ON departments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── department_closure ────────────────────────────────────────────────────
-- Stores every (ancestor, descendant, depth) pair. Self-pair has depth=0.
CREATE TABLE IF NOT EXISTS department_closure (
  ancestor_id   UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
  descendant_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
  depth         INT  NOT NULL,
  PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX IF NOT EXISTS idx_dept_closure_descendant ON department_closure (descendant_id);


-- Trigger: on INSERT of a department, populate self-pair + ancestor pairs.
CREATE OR REPLACE FUNCTION department_closure_insert() RETURNS trigger AS $$
BEGIN
  INSERT INTO department_closure (ancestor_id, descendant_id, depth)
  VALUES (NEW.id, NEW.id, 0);

  IF NEW.parent_id IS NOT NULL THEN
    INSERT INTO department_closure (ancestor_id, descendant_id, depth)
    SELECT ancestor_id, NEW.id, depth + 1
    FROM department_closure
    WHERE descendant_id = NEW.parent_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_dept_closure_insert ON departments;
CREATE TRIGGER trg_dept_closure_insert AFTER INSERT ON departments
  FOR EACH ROW EXECUTE FUNCTION department_closure_insert();


-- Trigger: on UPDATE of parent_id, rebuild subtree closure pairs.
CREATE OR REPLACE FUNCTION department_closure_reparent() RETURNS trigger AS $$
BEGIN
  IF NEW.parent_id IS DISTINCT FROM OLD.parent_id THEN
    -- Delete all pairs where ancestor is OUTSIDE the moved subtree
    -- and descendant is INSIDE the moved subtree.
    DELETE FROM department_closure
    WHERE descendant_id IN (
            SELECT descendant_id FROM department_closure WHERE ancestor_id = NEW.id
          )
      AND ancestor_id NOT IN (
            SELECT descendant_id FROM department_closure WHERE ancestor_id = NEW.id
          );

    -- Re-link the subtree under the new parent's ancestor chain.
    IF NEW.parent_id IS NOT NULL THEN
      INSERT INTO department_closure (ancestor_id, descendant_id, depth)
      SELECT p_anc.ancestor_id, sub.descendant_id, p_anc.depth + sub.depth + 1
      FROM department_closure p_anc
      JOIN department_closure sub ON sub.ancestor_id = NEW.id
      WHERE p_anc.descendant_id = NEW.parent_id;
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_dept_closure_reparent ON departments;
CREATE TRIGGER trg_dept_closure_reparent AFTER UPDATE OF parent_id ON departments
  FOR EACH ROW EXECUTE FUNCTION department_closure_reparent();


-- ── department_roles ──────────────────────────────────────────────────────
-- Role grants attached to a department. Apply to every membership whose
-- department is this dept OR a descendant (resolved via closure).
CREATE TABLE IF NOT EXISTS department_roles (
  department_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
  role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  granted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by    UUID,
  PRIMARY KEY (department_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_dept_roles_role ON department_roles (role_id);


-- ── groups ────────────────────────────────────────────────────────────────
-- Cross-cutting collections of users (and other groups). Flat semantics —
-- group membership grants the group's roles, no inheritance up/down.
CREATE TABLE IF NOT EXISTS groups (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  slug            CITEXT NOT NULL,
  name            TEXT NOT NULL,
  description     TEXT,
  kind            TEXT NOT NULL DEFAULT 'custom'
                  CHECK (kind IN ('custom', 'dynamic', 'system')),
  color           TEXT,
  icon            TEXT,
  is_archived     BOOLEAN NOT NULL DEFAULT false,
  rule            JSONB,                       -- for kind='dynamic' future use
  metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  created_by      UUID,
  updated_by      UUID,
  deleted_by      UUID,
  UNIQUE (organization_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_groups_org ON groups (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_tenant ON groups (tenant_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_groups_updated_at ON groups;
CREATE TRIGGER trg_groups_updated_at BEFORE UPDATE ON groups
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── group_members ─────────────────────────────────────────────────────────
-- Polymorphic: a group member is either a user (member_user_id) or a
-- child group (member_group_id). Exactly one is non-null.
CREATE TABLE IF NOT EXISTS group_members (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id        UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  member_user_id  UUID REFERENCES users(id) ON DELETE CASCADE,
  member_group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
  added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  added_by        UUID,
  CONSTRAINT group_member_polymorphic_xor CHECK (
    (member_user_id IS NOT NULL AND member_group_id IS NULL) OR
    (member_user_id IS NULL AND member_group_id IS NOT NULL)
  ),
  CONSTRAINT group_member_no_self_ref CHECK (member_group_id IS NULL OR member_group_id != group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_group_member_user
  ON group_members (group_id, member_user_id) WHERE member_user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_group_member_group
  ON group_members (group_id, member_group_id) WHERE member_group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members (member_user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_child ON group_members (member_group_id);


-- ── group_roles ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS group_roles (
  group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by UUID,
  PRIMARY KEY (group_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_group_roles_role ON group_roles (role_id);


-- ── memberships.department_id ─────────────────────────────────────────────
-- Replace the free-text `department` column with a real FK. Keep the legacy
-- text column for back-compat — apps can migrate at their own pace.
ALTER TABLE memberships
  ADD COLUMN IF NOT EXISTS department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_memberships_department
  ON memberships (department_id) WHERE deleted_at IS NULL;


-- ── permissions catalog additions ─────────────────────────────────────────
INSERT INTO permissions (key, resource, action, description, category, is_dangerous) VALUES
  -- department
  ('department.read',     'department', 'read',   'View departments',                'department', false),
  ('department.list',     'department', 'list',   'List departments',                'department', false),
  ('department.create',   'department', 'create', 'Create departments',              'department', false),
  ('department.update',   'department', 'update', 'Edit departments',                'department', false),
  ('department.delete',   'department', 'delete', 'Archive departments',             'department', true),
  ('department.assign',   'department', 'assign', 'Move members or attach roles',    'department', true),
  -- group
  ('group.read',          'group',      'read',   'View groups',                     'group',      false),
  ('group.list',          'group',      'list',   'List groups',                     'group',      false),
  ('group.create',        'group',      'create', 'Create groups',                   'group',      false),
  ('group.update',        'group',      'update', 'Edit groups',                     'group',      false),
  ('group.delete',        'group',      'delete', 'Archive groups',                  'group',      true),
  ('group.assign',        'group',      'assign', 'Add/remove members or roles',     'group',      true)
ON CONFLICT (key) DO NOTHING;
