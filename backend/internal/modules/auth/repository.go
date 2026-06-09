package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// ── invites ────────────────────────────────────────────────────────────────

func (r *Repository) CreateInvite(ctx context.Context, inv *Invite) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *Repository) GetInviteByTokenHash(ctx context.Context, hash []byte) (*Invite, error) {
	var inv Invite
	if err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *Repository) MarkInviteAccepted(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Invite{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "accepted",
			"accepted_at": time.Now(),
			"accepted_by": userID,
		}).Error
}

func (r *Repository) ListInvitesByEmail(ctx context.Context, email string) ([]Invite, error) {
	out := []Invite{}
	if err := r.db.WithContext(ctx).
		Where("email = ? AND status = ?", strings.ToLower(strings.TrimSpace(email)), "pending").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ── password reset ─────────────────────────────────────────────────────────

func (r *Repository) CreatePasswordReset(ctx context.Context, t *PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repository) GetPasswordResetByHash(ctx context.Context, hash []byte) (*PasswordResetToken, error) {
	var t PasswordResetToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL", hash).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("id = ?", id).
		Update("used_at", time.Now()).Error
}

// ── refresh tokens ─────────────────────────────────────────────────────────

func (r *Repository) CreateRefresh(ctx context.Context, t *RefreshToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repository) GetRefreshByHash(ctx context.Context, hash []byte) (*RefreshToken, error) {
	var t RefreshToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetRefreshByJTI(ctx context.Context, jti uuid.UUID) (*RefreshToken, error) {
	var t RefreshToken
	if err := r.db.WithContext(ctx).First(&t, "id = ?", jti).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) ListActiveRefreshByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]RefreshToken, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > now()", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []RefreshToken{}
	tx := q.Order("issued_at DESC")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, jti uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", jti).
		Updates(map[string]interface{}{
			"revoked_at":     time.Now(),
			"revoked_reason": reason,
		}).Error
}

func (r *Repository) RevokeRefreshFamily(ctx context.Context, familyID uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Updates(map[string]interface{}{
			"revoked_at":     time.Now(),
			"revoked_reason": reason,
		}).Error
}

func (r *Repository) TouchRefresh(ctx context.Context, jti uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("id = ?", jti).
		Update("last_used_at", time.Now()).Error
}
