package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// ── plans ──────────────────────────────────────────────────────────────────

func (r *Repository) ListActivePlans(ctx context.Context) ([]Plan, error) {
	out := []Plan{}
	if err := r.db.WithContext(ctx).
		Where("is_active = true AND is_public = true").
		Order("tier ASC, price_cents ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetPlanByCode(ctx context.Context, code string) (*Plan, error) {
	var p Plan
	if err := r.db.WithContext(ctx).First(&p, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	var p Plan
	if err := r.db.WithContext(ctx).
		Where("is_active = true AND is_default = true").
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ── subscriptions ──────────────────────────────────────────────────────────

func (r *Repository) Create(ctx context.Context, sub *Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *Repository) GetActiveByOrg(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	var s Subscription
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status IN ?", orgID, []string{"trial", "active", "past_due", "paused"}).
		Order("created_at DESC").
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	if len(patch) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Subscription{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID, reason string, immediate bool) error {
	now := time.Now()
	patch := map[string]interface{}{
		"cancelled_at":     now,
		"cancel_reason":    reason,
		"cancel_immediate": immediate,
	}
	if immediate {
		patch["status"] = "cancelled"
		patch["ended_at"] = now
	} else {
		patch["cancel_at"] = now
	}
	return r.Update(ctx, id, patch)
}

// ── usage counters ─────────────────────────────────────────────────────────

func (r *Repository) IncrementUsage(ctx context.Context, tenantID, orgID uuid.UUID, key string, by int64, periodStart, periodEnd time.Time) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO usage_counters (id, tenant_id, organization_id, key, count, period_start, period_end, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, now(), now())
		ON CONFLICT (organization_id, key, period_start)
		DO UPDATE SET count = usage_counters.count + EXCLUDED.count, updated_at = now()
	`, tenantID, orgID, key, by, periodStart, periodEnd).Error
}

func (r *Repository) GetUsage(ctx context.Context, orgID uuid.UUID, key string, periodStart time.Time) (*UsageCounter, error) {
	var u UsageCounter
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND key = ? AND period_start = ?", orgID, key, periodStart).
		First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ListUsage(ctx context.Context, orgID uuid.UUID) ([]UsageCounter, error) {
	out := []UsageCounter{}
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("period_start DESC").
		Find(&out).Error
	return out, err
}

// ── invoices ───────────────────────────────────────────────────────────────

func (r *Repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *Repository) ListInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]Invoice, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []Invoice{}
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("issued_at DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

func (r *Repository) GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*Invoice, error) {
	var inv Invoice
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

// NextInvoiceNumber issues a monotonically-increasing per-tenant invoice
// number. Cheap implementation: count existing rows + 1, formatted with a
// year prefix. For higher-throughput tenants you'd swap this for a sequence.
func (r *Repository) NextInvoiceNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	var n int64
	if err := r.db.WithContext(ctx).
		Model(&Invoice{}).
		Where("tenant_id = ? AND issued_at >= make_timestamptz(?, 1, 1, 0, 0, 0)", tenantID, year).
		Count(&n).Error; err != nil {
		return "", err
	}
	return formatInvoiceNumber(year, int(n)+1), nil
}

func formatInvoiceNumber(year, seq int) string {
	// e.g. INV-2026-000001
	return fmt.Sprintf("INV-%04d-%06d", year, seq)
}

// ── coupons ────────────────────────────────────────────────────────────────

func (r *Repository) GetCouponByCode(ctx context.Context, code string) (*Coupon, error) {
	var c Coupon
	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) IncrementCouponRedemptions(ctx context.Context, couponID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Coupon{}).
		Where("id = ?", couponID).
		UpdateColumn("redemptions", gorm.Expr("redemptions + 1")).Error
}

func (r *Repository) RecordCouponRedemption(ctx context.Context, red *CouponRedemption) error {
	return r.db.WithContext(ctx).Create(red).Error
}
