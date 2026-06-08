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

func (r *Repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	out := []Permission{}
	if err := r.db.WithContext(ctx).Order("category, key").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
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

func (r *Repository) ListRoles(ctx context.Context, orgID uuid.UUID) ([]Role, error) {
	out := []Role{}
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("priority DESC, name").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
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

func (r *Repository) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	out := []Permission{}
	if err := r.db.WithContext(ctx).
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID).
		Order("permissions.category, permissions.key").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
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
// the user in the given org through any of their roles. Used to populate the
// permission cache.
func (r *Repository) ResolvePermissionsForUserOrg(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	rows := []string{}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT p.key
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			JOIN membership_roles mr ON mr.role_id = rp.role_id
			JOIN memberships m       ON m.id = mr.membership_id
			JOIN roles r             ON r.id = rp.role_id
			WHERE m.user_id = ?
			  AND m.organization_id = ?
			  AND m.deleted_at IS NULL
			  AND r.deleted_at IS NULL
			  AND p.deleted_at IS NULL
			  AND m.status = 'active'
		`, userID, orgID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
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
