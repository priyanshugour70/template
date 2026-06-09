package rbac

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// ── permissions ────────────────────────────────────────────────────────────

func (r *Repository) ListPermissions(ctx context.Context, limit, offset int) ([]Permission, int64, error) {
	q := r.db.WithContext(ctx).Model(&Permission{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Permission{}
	tx := q.Order("category, key")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) GetPermissionsByKeys(ctx context.Context, keys []string) ([]Permission, error) {
	if len(keys) == 0 {
		return []Permission{}, nil
	}
	out := []Permission{}
	if err := r.db.WithContext(ctx).Where("key IN ?", keys).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ── roles ──────────────────────────────────────────────────────────────────

func (r *Repository) CreateRole(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *Repository) GetRoleByID(ctx context.Context, orgID, id uuid.UUID) (*Role, error) {
	var rl Role
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&rl).Error; err != nil {
		return nil, err
	}
	return &rl, nil
}

func (r *Repository) GetRoleByKey(ctx context.Context, orgID uuid.UUID, key string) (*Role, error) {
	var rl Role
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND key = ?", orgID, key).
		First(&rl).Error; err != nil {
		return nil, err
	}
	return &rl, nil
}

func (r *Repository) ListRoles(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Role, int64, error) {
	q := r.db.WithContext(ctx).Model(&Role{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Role{}
	tx := q.Order("priority DESC, name")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) UpdateRole(ctx context.Context, orgID, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&Role{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ArchiveRole(ctx context.Context, orgID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		Delete(&Role{}).Error
}

// ── role_permissions ───────────────────────────────────────────────────────

func (r *Repository) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		if len(permissionIDs) == 0 {
			return nil
		}
		rows := make([]RolePermission, 0, len(permissionIDs))
		for _, pid := range permissionIDs {
			rows = append(rows, RolePermission{RoleID: roleID, PermissionID: pid, GrantedBy: grantedBy})
		}
		return tx.Create(&rows).Error
	})
}

func (r *Repository) ListRolePermissions(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]Permission, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&Permission{}).
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Permission{}
	tx := q.Order("permissions.category, permissions.key")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── membership_roles ───────────────────────────────────────────────────────

func (r *Repository) AssignRolesToMembership(ctx context.Context, membershipID uuid.UUID, roleIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	if len(roleIDs) == 0 {
		return nil
	}
	rows := make([]MembershipRole, 0, len(roleIDs))
	for _, rid := range roleIDs {
		rows = append(rows, MembershipRole{MembershipID: membershipID, RoleID: rid, GrantedBy: grantedBy})
	}
	return r.db.WithContext(ctx).
		Clauses(/* on conflict do nothing — composite PK suffices */).
		Create(&rows).Error
}

func (r *Repository) RemoveRoleFromMembership(ctx context.Context, membershipID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("membership_id = ? AND role_id = ?", membershipID, roleID).
		Delete(&MembershipRole{}).Error
}

func (r *Repository) ListMembershipRoles(ctx context.Context, membershipID uuid.UUID) ([]Role, error) {
	out := []Role{}
	if err := r.db.WithContext(ctx).
		Joins("INNER JOIN membership_roles mr ON mr.role_id = roles.id").
		Where("mr.membership_id = ? AND roles.deleted_at IS NULL", membershipID).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ResolvePermissionsForUserOrg returns the union of permission keys held by
// the user in the given org. Sources:
//   (1) direct membership_roles
//   (2) department_roles on ancestors of the membership's department (via closure)
//   (3) group_roles on every group containing the user (transitive via nested groups)
func (r *Repository) ResolvePermissionsForUserOrg(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	rows := []string{}
	err := r.db.WithContext(ctx).
		Raw(`
			WITH active_membership AS (
			  SELECT id, department_id
			  FROM memberships
			  WHERE user_id = ?
			    AND organization_id = ?
			    AND deleted_at IS NULL
			    AND status = 'active'
			  LIMIT 1
			),
			user_groups AS (
			  -- transitive: every group containing the user directly OR via nested groups
			  WITH RECURSIVE direct_groups AS (
			    SELECT gm.group_id
			    FROM group_members gm
			    JOIN groups g ON g.id = gm.group_id
			    WHERE gm.member_user_id = ?
			      AND g.organization_id = ?
			      AND g.deleted_at IS NULL
			    UNION
			    SELECT gm.group_id
			    FROM group_members gm
			    JOIN direct_groups dg ON dg.group_id = gm.member_group_id
			  )
			  SELECT group_id FROM direct_groups
			),
			role_ids AS (
			  SELECT mr.role_id
			  FROM membership_roles mr
			  JOIN active_membership am ON am.id = mr.membership_id
			  UNION
			  SELECT dr.role_id
			  FROM department_roles dr
			  JOIN department_closure dc ON dc.ancestor_id = dr.department_id
			  JOIN active_membership am ON am.department_id = dc.descendant_id
			  UNION
			  SELECT gr.role_id
			  FROM group_roles gr
			  JOIN user_groups ug ON ug.group_id = gr.group_id
			)
			SELECT DISTINCT p.key
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			JOIN role_ids ri ON ri.role_id = rp.role_id
			JOIN roles r ON r.id = rp.role_id
			WHERE r.deleted_at IS NULL
			  AND p.deleted_at IS NULL
		`, userID, orgID, userID, orgID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetMembershipScope returns the (tenant_id, organization_id) of a membership
// — used by the service layer to verify a membership belongs to the caller's
// scope before role assignment/removal. Returns ErrRecordNotFound if the
// membership doesn't exist or has been deleted.
func (r *Repository) GetMembershipScope(ctx context.Context, mid uuid.UUID) (tenantID, orgID uuid.UUID, err error) {
	var row struct {
		TenantID       uuid.UUID `gorm:"column:tenant_id"`
		OrganizationID uuid.UUID `gorm:"column:organization_id"`
	}
	if err := r.db.WithContext(ctx).
		Table("memberships").
		Select("tenant_id, organization_id").
		Where("id = ? AND deleted_at IS NULL", mid).
		Take(&row).Error; err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return row.TenantID, row.OrganizationID, nil
}

// MembershipIDsForRole returns all memberships currently holding the given role
// — used to invalidate the permission cache when the role changes.
func (r *Repository) MembershipIDsForRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	ids := []uuid.UUID{}
	err := r.db.WithContext(ctx).
		Raw(`SELECT membership_id FROM membership_roles WHERE role_id = ?`, roleID).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
