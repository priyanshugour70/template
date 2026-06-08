package audit

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// Insert is called from the worker consumer to persist a captured event.
func (r *Repository) Insert(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// BatchInsert flushes a slice of events in one statement.
func (r *Repository) BatchInsert(ctx context.Context, logs []Log) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 200).Error
}

// List returns audit rows filtered by the caller's tenant. The caller must
// pass a non-nil tenantID; cross-tenant audit access requires super-admin.
func (r *Repository) List(
	ctx context.Context,
	tenantID *uuid.UUID,
	orgID *uuid.UUID,
	filter ListFilter,
	p pagination.Params,
) ([]Log, int64, error) {
	q := r.db.WithContext(ctx).Model(&Log{})
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.UserEmail != "" {
		q = q.Where("user_email = ?", strings.ToLower(filter.UserEmail))
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != nil {
		q = q.Where("target_id = ?", *filter.TargetID)
	}
	if filter.Method != "" {
		q = q.Where("method = ?", strings.ToUpper(filter.Method))
	}
	if filter.Path != "" {
		q = q.Where("path ILIKE ?", "%"+filter.Path+"%")
	}
	if filter.StatusFrom > 0 {
		q = q.Where("status_code >= ?", filter.StatusFrom)
	}
	if filter.StatusTo > 0 {
		q = q.Where("status_code <= ?", filter.StatusTo)
	}
	if filter.OccurredFrom != nil {
		q = q.Where("occurred_at >= ?", *filter.OccurredFrom)
	}
	if filter.OccurredTo != nil {
		q = q.Where("occurred_at <= ?", *filter.OccurredTo)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("user_email ILIKE ? OR path ILIKE ? OR action ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Log{}
	if err := q.
		Order("occurred_at DESC").
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// Get returns a single audit row, scoped to tenant when provided.
func (r *Repository) Get(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) (*Log, error) {
	var l Log
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if err := q.First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}
