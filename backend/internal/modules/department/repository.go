package department

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

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) Create(ctx context.Context, d *Department) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch map[string]any) error {
	return r.db.WithContext(ctx).Model(&Department{}).Where("id = ?", id).Updates(patch).Error
}

func (r *Repository) Get(ctx context.Context, orgID, id uuid.UUID) (*Department, error) {
	var d Department
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]Department, error) {
	rows := []Department{}
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("sort_order ASC, name ASC").
		Find(&rows).Error
	return rows, err
}

// Reparent updates parent_id. The DB trigger rebuilds the closure rows.
func (r *Repository) Reparent(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Department{}).
		Where("id = ?", id).
		Update("parent_id", parentID).Error
}

// IsAncestor reports whether `ancestor` is an ancestor of (or equal to) `descendant`.
// Used to prevent cycles when reparenting.
func (r *Repository) IsAncestor(ctx context.Context, ancestor, descendant uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("department_closure").
		Where("ancestor_id = ? AND descendant_id = ?", ancestor, descendant).
		Count(&n).Error
	return n > 0, err
}

// Delete soft-deletes a department. Memberships pointing at it have their
// department_id set to NULL by the FK.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Department{}, "id = ?", id).Error
}

// ── role bindings ──────────────────────────────────────────────────────────

func (r *Repository) ReplaceRoles(ctx context.Context, deptID uuid.UUID, roleIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("department_id = ?", deptID).Delete(&DeptRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]DeptRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			rows = append(rows, DeptRole{DepartmentID: deptID, RoleID: rid})
		}
		return tx.CreateInBatches(rows, 50).Error
	})
}

func (r *Repository) ListRoles(ctx context.Context, deptID uuid.UUID) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := r.db.WithContext(ctx).
		Table("department_roles").
		Where("department_id = ?", deptID).
		Pluck("role_id", &out).Error
	return out, err
}

// MembershipIDsAffectedByDept returns memberships whose effective permissions
// might change when the given department (or any of its descendants) has its
// role grants modified.
func (r *Repository) MembershipIDsAffectedByDept(ctx context.Context, deptID uuid.UUID) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT m.id
		FROM memberships m
		JOIN department_closure c ON c.descendant_id = m.department_id
		WHERE c.ancestor_id = ?
		  AND m.deleted_at IS NULL
	`, deptID).Scan(&out).Error
	return out, err
}
