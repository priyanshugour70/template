package tenant

import (
	"context"
	"errors"
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

// ── tenants ────────────────────────────────────────────────────────────────

func (r *Repository) CreateTenant(ctx context.Context, t *Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repository) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	var t Tenant
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var t Tenant
	if err := r.db.WithContext(ctx).First(&t, "slug = ?", strings.TrimSpace(slug)).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) UpdateTenant(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Tenant{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// HardDeleteTenant physically removes the row, bypassing GORM soft-delete.
// Used by the signup rollback path: when user/membership creation fails after
// the tenant insert, we must release the unique slug so the same payload can
// be retried. CASCADE on tenant_id cleans up organizations + memberships.
func (r *Repository) HardDeleteTenant(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&Tenant{}).Error
}

func (r *Repository) ArchiveTenant(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Model(&Tenant{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "archived",
			"archived_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Delete(&Tenant{}, "id = ?", id).Error
}

func (r *Repository) ListTenants(ctx context.Context, filter ListFilter, p pagination.Params) ([]Tenant, int64, error) {
	q := r.db.WithContext(ctx).Model(&Tenant{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		q = q.Where("name ILIKE ? OR slug ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowedSort := map[string]bool{"created_at": true, "updated_at": true, "name": true, "slug": true, "status": true}
	out := []Tenant{}
	if err := q.Order(p.SortClause(allowedSort, "created_at")).
		Limit(p.Limit).Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── organizations ──────────────────────────────────────────────────────────

func (r *Repository) CreateOrganization(ctx context.Context, o *Organization) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *Repository) GetOrganization(ctx context.Context, tenantID, id uuid.UUID) (*Organization, error) {
	var o Organization
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) GetOrganizationBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*Organization, error) {
	var o Organization
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ?", tenantID, strings.TrimSpace(slug)).
		First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) UpdateOrganization(ctx context.Context, tenantID, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&Organization{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ArchiveOrganization(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Model(&Organization{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{"status": "archived", "archived_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&Organization{}).Error
}

func (r *Repository) ListOrganizations(ctx context.Context, tenantID uuid.UUID, filter ListFilter, p pagination.Params) ([]Organization, int64, error) {
	q := r.db.WithContext(ctx).Model(&Organization{}).Where("tenant_id = ?", tenantID)
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		q = q.Where("name ILIKE ? OR slug ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowedSort := map[string]bool{"created_at": true, "updated_at": true, "name": true, "slug": true, "status": true}
	out := []Organization{}
	if err := q.Order(p.SortClause(allowedSort, "created_at")).
		Limit(p.Limit).Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── transactional helpers ──────────────────────────────────────────────────

// WithTx runs fn inside a single DB transaction. Repository methods used
// inside fn must accept a *gorm.DB; this template helper exposes the tx db.
func (r *Repository) WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := fn(tx); err != nil {
			return err
		}
		return nil
	})
}

// IsNotFound is a convenience for services that want to map gorm.ErrRecordNotFound.
func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
