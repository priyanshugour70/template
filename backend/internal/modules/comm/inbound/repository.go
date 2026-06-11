package inbound

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func (r *Repository) Create(ctx context.Context, h *ChannelHook) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*ChannelHook, error) {
	var h ChannelHook
	if err := r.db.WithContext(ctx).First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

// GetByTokenHash is the auth path for the public POST endpoint. We match on
// token_hash AND is_active AND not-deleted; the partial unique index ensures
// at most one row can match.
func (r *Repository) GetByTokenHash(ctx context.Context, hash []byte) (*ChannelHook, error) {
	var h ChannelHook
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_active = true", hash).
		First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Repository) ListByConversation(ctx context.Context, convID uuid.UUID) ([]ChannelHook, error) {
	out := []ChannelHook{}
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at DESC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ChannelHook{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// BumpUse records that the hook just delivered a message. Best-effort —
// failures here don't block delivery.
func (r *Repository) BumpUse(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ChannelHook{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": time.Now(),
			"use_count":    gorm.Expr("use_count + 1"),
		}).Error
}
