package apikey

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

func (r *Repository) Create(ctx context.Context, k *APIKey) error {
	return r.db.WithContext(ctx).Create(k).Error
}

// ListForOrg returns non-deleted keys; revoked keys remain visible so the UI
// can show "revoked" badges. Filter client-side if you only want active.
func (r *Repository) ListForOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]APIKey, int64, error) {
	q := r.db.WithContext(ctx).Model(&APIKey{}).Where("organization_id = ?", orgID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []APIKey{}
	tx := q.Order("created_at DESC")
	if limit > 0 {
		tx = tx.Limit(limit).Offset(offset)
	}
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Get(ctx context.Context, orgID, id uuid.UUID) (*APIKey, error) {
	var k APIKey
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&k).Error
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// Revoke marks the key as revoked without deleting it. The token_hash stays
// indexed so future attempts to authenticate with the plaintext token can
// short-circuit to a 401 with a clear "revoked" reason.
func (r *Repository) Revoke(ctx context.Context, id uuid.UUID, by *uuid.UUID) error {
	now := time.Now()
	patch := map[string]interface{}{"revoked_at": now}
	if by != nil {
		patch["revoked_by"] = *by
	}
	res := r.db.WithContext(ctx).Model(&APIKey{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindByHash is the auth path lookup. Returns nil if revoked or expired.
func (r *Repository) FindByHash(ctx context.Context, hash string) (*APIKey, error) {
	var k APIKey
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Where("expires_at IS NULL OR expires_at > now()").
		First(&k).Error
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// TouchLastUsed updates last_used_at/ip without bumping updated_by. Best-effort.
func (r *Repository) TouchLastUsed(ctx context.Context, id uuid.UUID, ip string) error {
	now := time.Now()
	patch := map[string]interface{}{"last_used_at": now}
	if ip != "" {
		patch["last_used_ip"] = ip
	}
	return r.db.WithContext(ctx).
		Model(&APIKey{}).
		Where("id = ?", id).
		UpdateColumns(patch).Error
}
