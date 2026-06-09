package user

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) DB() *gorm.DB { return r.db }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// ── users ──────────────────────────────────────────────────────────────────

func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", strings.TrimSpace(email)).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) UpdateUserLogin(ctx context.Context, id uuid.UUID, ip, userAgent string) error {
	patch := map[string]interface{}{
		"last_login_at":         time.Now(),
		"last_login_user_agent": userAgent,
		"failed_login_count":    0,
	}
	// Only write last_login_ip when we have a parsed IP — Postgres inet
	// rejects empty strings.
	if ip != "" {
		if parsed := net.ParseIP(ip); parsed != nil {
			patch["last_login_ip"] = parsed.String()
		}
	}
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(patch).Error
}

func (r *Repository) IncrementFailedLogin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		UpdateColumn("failed_login_count", gorm.Expr("failed_login_count + 1")).Error
}

func (r *Repository) LockUser(ctx context.Context, id uuid.UUID, until time.Time) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("locked_until", until).Error
}

func (r *Repository) SuspendUser(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("status", "suspended")
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}

func (r *Repository) ActivateUser(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("status", "active")
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}

func (r *Repository) ArchiveUser(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("status", "archived")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error
}

// ListUsersInOrg returns users belonging to an org (via memberships), filtered + paginated.
func (r *Repository) ListUsersInOrg(
	ctx context.Context,
	tenantID, orgID uuid.UUID,
	filter ListFilter,
	p pagination.Params,
) ([]User, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&User{}).
		Joins("INNER JOIN memberships m ON m.user_id = users.id AND m.deleted_at IS NULL").
		Where("m.tenant_id = ? AND m.organization_id = ?", tenantID, orgID)

	if filter.Status != "" {
		q = q.Where("users.status = ?", filter.Status)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where(
			"users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.display_name ILIKE ?",
			like, like, like, like,
		)
	}
	if filter.JobTitle != "" {
		q = q.Where("users.job_title ILIKE ?", "%"+filter.JobTitle+"%")
	}
	if filter.Department != "" {
		q = q.Where("users.department ILIKE ?", "%"+filter.Department+"%")
	}
	if filter.DepartmentID != nil {
		q = q.Where("m.department_id = ?", *filter.DepartmentID)
	}
	if filter.MFAEnabled != nil {
		q = q.Where("users.mfa_enabled = ?", *filter.MFAEnabled)
	}
	if filter.LastLoginAfter != nil {
		q = q.Where("users.last_login_at >= ?", *filter.LastLoginAfter)
	}
	if filter.LastLoginBefore != nil {
		q = q.Where("users.last_login_at <= ?", *filter.LastLoginBefore)
	}
	if filter.Role != "" {
		q = q.Where(`EXISTS (
			SELECT 1
			FROM membership_roles mr
			JOIN roles r ON r.id = mr.role_id
			WHERE mr.membership_id = m.id
			  AND r.key = ?
			  AND r.deleted_at IS NULL
		)`, filter.Role)
	}
	if filter.CreatedAfter != nil {
		q = q.Where("users.created_at >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		q = q.Where("users.created_at <= ?", *filter.CreatedBefore)
	}

	var total int64
	if err := q.Distinct("users.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]bool{
		"created_at": true, "updated_at": true, "email": true, "last_login_at": true,
		"first_name": true, "last_name": true, "status": true,
	}
	sortCol := p.SortClause(allowed, "created_at")
	out := []User{}
	if err := q.
		Distinct("users.*").
		Order("users." + sortCol).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── memberships ────────────────────────────────────────────────────────────

func (r *Repository) CreateMembership(ctx context.Context, m *Membership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetMembership(ctx context.Context, id uuid.UUID) (*Membership, error) {
	var m Membership
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*Membership, error) {
	var m Membership
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListMembershipsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Membership, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&Membership{}).
		Where("user_id = ? AND status IN ?", userID, []string{"active", "invited"})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Membership{}
	tx := q.Order("created_at DESC")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// UserHasMembershipInTenant returns true iff the user has any non-deleted
// membership in the given tenant. Used as the tenant-scope gate for user-level
// admin operations (suspend, archive, force-reset, etc.) to prevent cross-tenant
// IDOR.
func (r *Repository) UserHasMembershipInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&Membership{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ListTenantIDsByUserEmail(ctx context.Context, email string) ([]uuid.UUID, error) {
	out := []uuid.UUID{}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT m.tenant_id
			FROM memberships m
			JOIN users u ON u.id = m.user_id
			WHERE u.email = ?
			  AND u.deleted_at IS NULL
			  AND m.deleted_at IS NULL
			  AND m.status IN ('active','invited')
		`, strings.TrimSpace(email)).
		Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateMembership(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Membership{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) SuspendMembership(ctx context.Context, id uuid.UUID) error {
	return r.UpdateMembership(ctx, id, map[string]interface{}{"status": "suspended"})
}

func (r *Repository) ArchiveMembership(ctx context.Context, id uuid.UUID) error {
	if err := r.UpdateMembership(ctx, id, map[string]interface{}{"status": "archived"}); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&Membership{}, "id = ?", id).Error
}
