package group

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

func (r *Repository) Create(ctx context.Context, g *Group) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch map[string]any) error {
	return r.db.WithContext(ctx).Model(&Group{}).Where("id = ?", id).Updates(patch).Error
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Group{}, "id = ?", id).Error
}

func (r *Repository) Get(ctx context.Context, orgID, id uuid.UUID) (*Group, error) {
	var g Group
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Group, int64, error) {
	q := r.db.WithContext(ctx).Model(&Group{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Group{}
	tx := q.Order("name ASC")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── members ────────────────────────────────────────────────────────────────

func (r *Repository) AddUser(ctx context.Context, groupID, userID uuid.UUID, by *uuid.UUID) error {
	m := Member{GroupID: groupID, MemberUserID: &userID, AddedBy: by}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *Repository) AddSubGroup(ctx context.Context, parentID, childID uuid.UUID, by *uuid.UUID) error {
	if parentID == childID {
		return errors.New("cannot add a group to itself")
	}
	m := Member{GroupID: parentID, MemberGroupID: &childID, AddedBy: by}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *Repository) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Member{}, "id = ?", memberID).Error
}

func (r *Repository) ListMembers(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]Member, int64, error) {
	q := r.db.WithContext(ctx).Model(&Member{}).Where("group_id = ?", groupID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Member{}
	tx := q.Order("added_at ASC")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── role bindings ──────────────────────────────────────────────────────────

func (r *Repository) ReplaceRoles(ctx context.Context, groupID uuid.UUID, roleIDs []uuid.UUID, _ *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&GroupRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]GroupRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			rows = append(rows, GroupRole{GroupID: groupID, RoleID: rid})
		}
		return tx.CreateInBatches(rows, 50).Error
	})
}

func (r *Repository) ListRoles(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := r.db.WithContext(ctx).
		Table("group_roles").
		Where("group_id = ?", groupID).
		Pluck("role_id", &out).Error
	return out, err
}

// MembershipIDsAffectedByGroup walks user members of the group (transitive via
// nested sub-groups) and returns membership IDs in the same org for cache
// invalidation.
func (r *Repository) MembershipIDsAffectedByGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE expand(group_id) AS (
		  SELECT ? AS group_id
		  UNION
		  SELECT gm.member_group_id
		  FROM group_members gm
		  JOIN expand e ON e.group_id = gm.group_id
		  WHERE gm.member_group_id IS NOT NULL
		)
		SELECT DISTINCT m.id
		FROM expand e
		JOIN group_members gm ON gm.group_id = e.group_id AND gm.member_user_id IS NOT NULL
		JOIN memberships m ON m.user_id = gm.member_user_id
		JOIN groups g ON g.id = ?
		WHERE m.organization_id = g.organization_id
		  AND m.deleted_at IS NULL
	`, groupID, groupID).Scan(&out).Error
	return out, err
}
