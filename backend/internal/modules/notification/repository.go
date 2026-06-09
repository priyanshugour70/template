package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func (r *Repository) Insert(ctx context.Context, n *Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// List returns the user's notifications ordered by created_at DESC.
func (r *Repository) List(
	ctx context.Context,
	userID uuid.UUID,
	filter ListFilter,
	p pagination.Params,
) ([]Notification, int64, error) {
	q := r.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ?", userID)
	if filter.UnreadOnly {
		q = q.Where("is_read = false")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Notification{}
	if err := q.
		Order("created_at DESC").
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&n).Error
	return n, err
}

// MarkRead flips a single row to read; scoped to user_id so callers can't
// mark someone else's notification.
func (r *Repository) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ? AND user_id = ? AND is_read = false", id, userID).
		Updates(map[string]any{"is_read": true, "read_at": &now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// Either already read, or doesn't belong to the user — verify it exists.
		var n Notification
		if err := r.db.WithContext(ctx).
			Where("id = ? AND user_id = ?", id, userID).
			First(&n).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]any{"is_read": true, "read_at": &now})
	return res.RowsAffected, res.Error
}
