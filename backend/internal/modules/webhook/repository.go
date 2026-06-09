package webhook

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

func (r *Repository) Create(ctx context.Context, w *Webhook) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch map[string]any) error {
	res := r.db.WithContext(ctx).Model(&Webhook{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, orgID, id uuid.UUID) (*Webhook, error) {
	var w Webhook
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) List(ctx context.Context, orgID uuid.UUID) ([]Webhook, error) {
	rows := []Webhook{}
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Webhook{}, "id = ?", id).Error
}

func (r *Repository) RecordDelivery(ctx context.Context, d *Delivery) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *Repository) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]Delivery, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []Delivery{}
	err := r.db.WithContext(ctx).
		Where("webhook_id = ?", webhookID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateDeliveryStats(ctx context.Context, id uuid.UUID, status int, success bool) error {
	patch := map[string]any{
		"last_invoked_at": gorm.Expr("now()"),
		"last_status":     status,
	}
	if success {
		patch["consecutive_failures"] = 0
	} else {
		patch["consecutive_failures"] = gorm.Expr("consecutive_failures + 1")
	}
	return r.db.WithContext(ctx).Model(&Webhook{}).Where("id = ?", id).Updates(patch).Error
}
