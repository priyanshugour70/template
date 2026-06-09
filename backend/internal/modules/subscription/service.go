package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/queue"
)

const (
	subCachePrefix = "sub:"
	subCacheTTL    = 10 * time.Minute
)

type Service struct {
	repo     *Repository
	log      *zap.Logger
	cache    cache.Cache
	producer queue.Producer
}

func NewService(repo *Repository, log *zap.Logger, c cache.Cache, p queue.Producer) *Service {
	return &Service{repo: repo, log: log, cache: c, producer: p}
}

// ── plans ──────────────────────────────────────────────────────────────────

func (s *Service) ListPlans(ctx context.Context, limit, offset int) ([]Plan, int64, error) {
	rows, total, err := s.repo.ListActivePlans(ctx, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list plans failed", err)
	}
	return rows, total, nil
}

// ── org subscription ───────────────────────────────────────────────────────

func (s *Service) GetActive(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "no active subscription", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch subscription failed", err)
	}
	return sub, nil
}

// provisionOnPlan creates a brand-new subscription for an org on the given
// plan. Used by both first-time onboarding (called from ChangePlan when no
// active sub exists) and the explicit ProvisionTrial helper.
func (s *Service) provisionOnPlan(
	ctx context.Context,
	tenantID, orgID uuid.UUID,
	plan *Plan,
	req ChangePlanRequest,
) (*Subscription, error) {
	now := time.Now()
	trialEnd := now
	status := "active"
	if plan.TrialDays > 0 {
		trialEnd = now.Add(time.Duration(plan.TrialDays) * 24 * time.Hour)
		status = "trial"
	}
	cycle := plan.BillingCycle
	if req.BillingCycle != "" {
		cycle = req.BillingCycle
	}
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	sub := &Subscription{
		TenantID:           tenantID,
		OrganizationID:     orgID,
		PlanID:             plan.ID,
		PlanCode:           plan.Code,
		Status:             status,
		BillingCycle:       cycle,
		Quantity:           qty,
		UnitPriceCents:     plan.PriceCents,
		TotalCents:         int64(qty) * plan.PriceCents,
		Currency:           plan.Currency,
		StartedAt:          now,
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   &trialEnd,
		Features:           plan.Features,
		Limits:             plan.Limits,
		Metadata:           []byte(`{}`),
		CouponCode:         req.CouponCode,
	}
	if plan.TrialDays > 0 {
		sub.TrialStartedAt = &now
		sub.TrialEndsAt = &trialEnd
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create subscription failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return sub, nil
}

// ProvisionTrial creates a trial subscription for a newly-onboarded org based
// on the default plan. Called by the tenant onboarding flow.
func (s *Service) ProvisionTrial(ctx context.Context, tenantID, orgID uuid.UUID) (*Subscription, error) {
	plan, err := s.repo.GetDefaultPlan(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "no default plan configured", err)
	}
	now := time.Now()
	trialEnd := now
	if plan.TrialDays > 0 {
		trialEnd = now.Add(time.Duration(plan.TrialDays) * 24 * time.Hour)
	}
	status := "trial"
	if plan.TrialDays == 0 {
		status = "active"
	}
	sub := &Subscription{
		TenantID:           tenantID,
		OrganizationID:     orgID,
		PlanID:             plan.ID,
		PlanCode:           plan.Code,
		Status:             status,
		BillingCycle:       plan.BillingCycle,
		Quantity:           1,
		UnitPriceCents:     plan.PriceCents,
		TotalCents:         plan.PriceCents,
		Currency:           plan.Currency,
		StartedAt:          now,
		TrialStartedAt:     &now,
		TrialEndsAt:        &trialEnd,
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   &trialEnd,
		Features:           plan.Features,
		Limits:             plan.Limits,
		Metadata:           []byte("{}"),
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create subscription failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return sub, nil
}

func (s *Service) ChangePlan(ctx context.Context, orgID uuid.UUID, req ChangePlanRequest) (*Subscription, error) {
	plan, err := s.repo.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "plan not found", nil)
	}
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		// No active subscription — first-time onboarding. Provision a fresh
		// subscription on the chosen plan. We need a tenant ID; pull it from
		// the request context since the caller is authenticated.
		tenantID := appctx.TenantID(ctx)
		if tenantID == uuid.Nil {
			return nil, apperr.New(apperr.CodeForbidden, "no tenant context", nil)
		}
		return s.provisionOnPlan(ctx, tenantID, orgID, plan, req)
	}
	patch := map[string]interface{}{
		"plan_id":          plan.ID,
		"plan_code":        plan.Code,
		"unit_price_cents": plan.PriceCents,
		"currency":         plan.Currency,
		"features":         []byte(plan.Features),
		"limits":           []byte(plan.Limits),
	}
	if req.BillingCycle != "" {
		patch["billing_cycle"] = req.BillingCycle
	}
	if req.Quantity > 0 {
		patch["quantity"] = req.Quantity
		patch["total_cents"] = int64(req.Quantity) * plan.PriceCents
	}
	if req.CouponCode != "" {
		patch["coupon_code"] = req.CouponCode
	}
	if req.StartImmediately {
		now := time.Now()
		patch["current_period_start"] = now
		// next-period-end computation lives in a separate billing engine.
	}
	if err := s.repo.Update(ctx, sub.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "change plan failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return s.repo.GetActiveByOrg(ctx, orgID)
}

func (s *Service) Cancel(ctx context.Context, orgID uuid.UUID, req CancelRequest) error {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "no active subscription", nil)
		}
		return apperr.New(apperr.CodeInternal, "fetch subscription failed", err)
	}
	if err := s.repo.Cancel(ctx, sub.ID, req.Reason, req.Immediate); err != nil {
		return apperr.New(apperr.CodeInternal, "cancel failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return nil
}

// ── feature/quota resolution ───────────────────────────────────────────────

func (s *Service) ResolveFeatureSet(ctx context.Context, orgID uuid.UUID) (*FeatureSet, error) {
	if orgID == uuid.Nil {
		return &FeatureSet{Features: map[string]bool{}, Limits: map[string]int64{}}, nil
	}
	key := subCachePrefix + orgID.String()
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil && raw != "" {
			var fs FeatureSet
			if err := json.Unmarshal([]byte(raw), &fs); err == nil {
				return &fs, nil
			}
		}
	}
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		if IsNotFound(err) {
			return &FeatureSet{Features: map[string]bool{}, Limits: map[string]int64{}}, nil
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch subscription failed", err)
	}

	fs := FeatureSet{
		PlanCode: sub.PlanCode,
		Status:   sub.Status,
		Features: map[string]bool{},
		Limits:   map[string]int64{},
	}
	if len(sub.Features) > 0 {
		var arr []string
		if err := json.Unmarshal(sub.Features, &arr); err == nil {
			for _, f := range arr {
				fs.Features[f] = true
			}
		}
	}
	if len(sub.Limits) > 0 {
		var raw map[string]json.Number
		if err := json.Unmarshal(sub.Limits, &raw); err == nil {
			for k, v := range raw {
				n, _ := v.Int64()
				fs.Limits[k] = n
			}
		}
	}
	if s.cache != nil {
		if b, err := json.Marshal(fs); err == nil {
			_ = s.cache.Set(ctx, key, string(b), subCacheTTL)
		}
	}
	return &fs, nil
}

func (s *Service) HasFeature(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
	fs, err := s.ResolveFeatureSet(ctx, orgID)
	if err != nil {
		return false, err
	}
	return fs.Features[feature], nil
}

// IsWithinQuota reports whether incrementing `key` by `by` would still be
// within the org's configured limit. -1 limit means unlimited.
func (s *Service) IsWithinQuota(ctx context.Context, tenantID, orgID uuid.UUID, key string, by int64) (bool, error) {
	fs, err := s.ResolveFeatureSet(ctx, orgID)
	if err != nil {
		return false, err
	}
	limit, ok := fs.Limits[key]
	if !ok || limit < 0 {
		return true, nil
	}
	// fetch current usage for the active period
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	u, err := s.repo.GetUsage(ctx, orgID, key, periodStart)
	current := int64(0)
	if err == nil && u != nil {
		current = u.Count
	}
	return current+by <= limit, nil
}

func (s *Service) IncrementUsage(ctx context.Context, tenantID, orgID uuid.UUID, key string, by int64) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)
	return s.repo.IncrementUsage(ctx, tenantID, orgID, key, by, periodStart, periodEnd)
}

func (s *Service) ListUsage(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]UsageCounter, int64, error) {
	rows, total, err := s.repo.ListUsage(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list usage failed", err)
	}
	return rows, total, nil
}

// ── cache ──────────────────────────────────────────────────────────────────

func (s *Service) invalidateCache(ctx context.Context, orgID uuid.UUID) {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, subCachePrefix+orgID.String())
	}
	if s.producer != nil {
		_ = s.producer.Publish(ctx, queue.ChannelSubscriptionInvalidate, map[string]interface{}{
			"organizationId": orgID.String(),
		})
	}
}

// InvalidateCacheForOrg is called by the worker consumer.
func (s *Service) InvalidateCacheForOrg(ctx context.Context, orgID uuid.UUID) {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, subCachePrefix+orgID.String())
	}
}

// ── lifecycle (pause / resume / reactivate / billing) ─────────────────────

// Pause stops billing without losing the subscription. ResumeAt is optional —
// when set, the worker will flip status back to active on that date.
func (s *Service) Pause(ctx context.Context, orgID uuid.UUID, req PauseRequest) (*Subscription, error) {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "no active subscription", nil)
	}
	if sub.Status == "paused" {
		return sub, nil
	}
	now := time.Now()
	patch := map[string]interface{}{
		"status":    "paused",
		"paused_at": now,
	}
	if req.ResumeAt != nil {
		patch["resume_at"] = *req.ResumeAt
	}
	if req.Reason != "" {
		patch["notes"] = req.Reason
	}
	if err := s.repo.Update(ctx, sub.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "pause failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return s.repo.GetActiveByOrg(ctx, orgID)
}

func (s *Service) Resume(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "no subscription", nil)
	}
	if sub.Status != "paused" {
		return sub, nil
	}
	patch := map[string]interface{}{
		"status":     "active",
		"paused_at":  nil,
		"resume_at":  nil,
	}
	if err := s.repo.Update(ctx, sub.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "resume failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return s.repo.GetActiveByOrg(ctx, orgID)
}

// Reactivate un-cancels a subscription whose cancel_at is in the future.
func (s *Service) Reactivate(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "no subscription", nil)
	}
	if sub.CancelAt == nil && sub.Status != "cancelled" {
		return sub, nil
	}
	patch := map[string]interface{}{
		"status":           "active",
		"cancel_at":        nil,
		"cancelled_at":     nil,
		"cancel_reason":    "",
		"cancel_immediate": false,
		"ended_at":         nil,
	}
	if err := s.repo.Update(ctx, sub.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "reactivate failed", err)
	}
	s.invalidateCache(ctx, orgID)
	return s.repo.GetActiveByOrg(ctx, orgID)
}

func (s *Service) UpdateBilling(ctx context.Context, orgID uuid.UUID, req UpdateBillingRequest) (*Subscription, error) {
	sub, err := s.repo.GetActiveByOrg(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "no subscription", nil)
	}
	patch := map[string]interface{}{}
	if req.BillingEmail != nil {
		patch["billing_email"] = *req.BillingEmail
	}
	if req.BillingName != nil {
		patch["billing_name"] = *req.BillingName
	}
	if req.BillingAddress != nil {
		b, _ := json.Marshal(req.BillingAddress)
		patch["billing_address"] = b
	}
	if len(patch) == 0 {
		return sub, nil
	}
	if err := s.repo.Update(ctx, sub.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "update billing failed", err)
	}
	return s.repo.GetActiveByOrg(ctx, orgID)
}

// ── coupons ───────────────────────────────────────────────────────────────

// ValidateCoupon checks code applicability without redeeming. UI-facing —
// returns a structured response with the reason on failure so the form can
// render a helpful inline message.
func (s *Service) ValidateCoupon(ctx context.Context, req ValidateCouponRequest) *ValidateCouponResponse {
	c, err := s.repo.GetCouponByCode(ctx, req.Code)
	if err != nil {
		return &ValidateCouponResponse{Valid: false, Reason: "coupon not found"}
	}
	if !c.IsActive || c.DeletedAt.Valid {
		return &ValidateCouponResponse{Valid: false, Reason: "coupon is inactive"}
	}
	now := time.Now()
	if c.ValidFrom != nil && now.Before(*c.ValidFrom) {
		return &ValidateCouponResponse{Valid: false, Reason: "coupon is not yet active"}
	}
	if c.ValidUntil != nil && now.After(*c.ValidUntil) {
		return &ValidateCouponResponse{Valid: false, Reason: "coupon has expired"}
	}
	if c.MaxRedemptions != nil && c.Redemptions >= *c.MaxRedemptions {
		return &ValidateCouponResponse{Valid: false, Reason: "coupon redemption limit reached"}
	}
	if req.PlanCode != "" && len(c.AppliesToPlans) > 2 {
		// AppliesToPlans is a JSON array; empty = [] = "any". If non-empty,
		// check the requested plan is in the list.
		var plans []string
		if err := json.Unmarshal([]byte(c.AppliesToPlans), &plans); err == nil && len(plans) > 0 {
			found := false
			for _, p := range plans {
				if p == req.PlanCode {
					found = true
					break
				}
			}
			if !found {
				return &ValidateCouponResponse{Valid: false, Reason: "coupon does not apply to this plan"}
			}
		}
	}
	return &ValidateCouponResponse{
		Valid:          true,
		Code:           c.Code,
		Name:           c.Name,
		PercentOff:     c.PercentOff,
		AmountOffCents: c.AmountOffCents,
		Currency:       c.Currency,
		Duration:       c.Duration,
	}
}

// applyCoupon takes a base amount + coupon and returns the discount cents.
func applyCoupon(amountCents int64, c *Coupon) int64 {
	if c == nil {
		return 0
	}
	if c.PercentOff != nil {
		return (amountCents * int64(*c.PercentOff)) / 100
	}
	if c.AmountOffCents != nil {
		off := *c.AmountOffCents
		if off > amountCents {
			off = amountCents
		}
		return off
	}
	return 0
}

// ── proration preview ──────────────────────────────────────────────────────

// PreviewChange computes the dollars/cents the user will be charged when
// switching plans, based on a simple day-pro-rata model. No payment is
// recorded; the response is informational only.
func (s *Service) PreviewChange(ctx context.Context, orgID uuid.UUID, req PreviewChangeRequest) (*PreviewChangeResponse, error) {
	target, err := s.repo.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "plan not found", nil)
	}
	current, _ := s.repo.GetActiveByOrg(ctx, orgID)

	now := time.Now()
	cycle := req.BillingCycle
	if cycle == "" {
		cycle = target.BillingCycle
	}
	cycleDays := cycleLengthDays(cycle)
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	base := target.PriceCents * int64(quantity)

	// Proration: credit unused days on the current plan; charge full base for
	// the new plan starting now. Daily-rate model.
	var proration int64
	unusedDays := 0
	fromCode := ""
	isUpgrade := true
	if current != nil {
		fromCode = current.PlanCode
		if current.CurrentPeriodEnd != nil && current.CurrentPeriodEnd.After(now) {
			unusedDays = int(current.CurrentPeriodEnd.Sub(now).Hours() / 24)
			if cycleDays > 0 && current.UnitPriceCents > 0 {
				dailyRate := current.UnitPriceCents / int64(cycleDays)
				// negative = credit
				proration = -(dailyRate * int64(unusedDays))
			}
		}
		if target.Tier < currentTier(s, ctx, current.PlanCode) {
			isUpgrade = false
		}
	}

	// Coupon
	var discount int64
	couponCode := ""
	if req.CouponCode != "" {
		c, err := s.repo.GetCouponByCode(ctx, req.CouponCode)
		if err == nil && c.IsActive {
			discount = applyCoupon(base, c)
			couponCode = c.Code
		}
	}

	// 18% GST on net of base + proration (proration can be negative)
	taxable := base + proration - discount
	if taxable < 0 {
		taxable = 0
	}
	tax := (taxable * 18) / 100
	total := taxable + tax
	if total < 0 {
		total = 0
	}

	return &PreviewChangeResponse{
		FromPlanCode:        fromCode,
		ToPlanCode:          target.Code,
		BillingCycle:        cycle,
		Currency:            target.Currency,
		BaseAmountCents:     base,
		ProrationCents:      proration,
		CouponCode:          couponCode,
		DiscountCents:       discount,
		TaxCents:            tax,
		TotalDueCents:       total,
		EffectiveAt:         now.UTC().Format(time.RFC3339),
		IsUpgrade:           isUpgrade,
		UnusedDaysRemaining: unusedDays,
	}, nil
}

func cycleLengthDays(cycle string) int {
	switch cycle {
	case "monthly":
		return 30
	case "quarterly":
		return 90
	case "yearly":
		return 365
	default:
		return 30
	}
}

// currentTier loads the current plan's tier; falls back to 0 on lookup error.
func currentTier(s *Service, ctx context.Context, code string) int {
	if code == "" {
		return 0
	}
	p, err := s.repo.GetPlanByCode(ctx, code)
	if err != nil {
		return 0
	}
	return p.Tier
}

// ── invoices ──────────────────────────────────────────────────────────────

func (s *Service) ListInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]Invoice, error) {
	rows, err := s.repo.ListInvoices(ctx, orgID, limit)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list invoices failed", err)
	}
	return rows, nil
}

func (s *Service) GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*Invoice, error) {
	inv, err := s.repo.GetInvoice(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "invoice not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load invoice failed", err)
	}
	return inv, nil
}

// GenerateInvoiceForPlanChange creates an open invoice for the current
// plan-change preview total. Called automatically from ChangePlan so the
// history table populates as users switch.
func (s *Service) generateInvoiceForChange(
	ctx context.Context,
	tenantID, orgID uuid.UUID,
	sub *Subscription,
	preview *PreviewChangeResponse,
) (*Invoice, error) {
	if preview == nil {
		return nil, nil
	}
	year := time.Now().Year()
	number, err := s.repo.NextInvoiceNumber(ctx, tenantID, year)
	if err != nil {
		return nil, err
	}
	items := []LineItem{
		{
			Description: "Plan: " + preview.ToPlanCode + " (" + preview.BillingCycle + ")",
			Quantity:    1,
			UnitCents:   preview.BaseAmountCents,
			AmountCents: preview.BaseAmountCents,
		},
	}
	if preview.ProrationCents != 0 {
		items = append(items, LineItem{
			Description: "Proration credit for unused days on previous plan",
			Quantity:    1,
			UnitCents:   preview.ProrationCents,
			AmountCents: preview.ProrationCents,
		})
	}
	if preview.DiscountCents > 0 {
		items = append(items, LineItem{
			Description: "Coupon: " + preview.CouponCode,
			Quantity:    1,
			UnitCents:   -preview.DiscountCents,
			AmountCents: -preview.DiscountCents,
		})
	}
	if preview.TaxCents > 0 {
		items = append(items, LineItem{
			Description: "Tax (GST 18%)",
			Quantity:    1,
			UnitCents:   preview.TaxCents,
			AmountCents: preview.TaxCents,
		})
	}
	body, _ := json.Marshal(items)
	now := time.Now()
	inv := &Invoice{
		TenantID:        tenantID,
		OrganizationID:  orgID,
		Number:          number,
		Status:          "open",
		Currency:        preview.Currency,
		SubtotalCents:   preview.BaseAmountCents + preview.ProrationCents,
		DiscountCents:   preview.DiscountCents,
		TaxCents:        preview.TaxCents,
		TotalCents:      preview.TotalDueCents,
		AmountDueCents:  preview.TotalDueCents,
		AmountPaidCents: 0,
		CouponCode:      preview.CouponCode,
		Description:     "Plan change to " + preview.ToPlanCode,
		LineItems:       body,
		IssuedAt:        now,
	}
	if sub != nil {
		sid := sub.ID
		inv.SubscriptionID = &sid
		inv.PeriodStart = sub.CurrentPeriodStart
		inv.PeriodEnd = sub.CurrentPeriodEnd
	}
	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// ChangePlanWithInvoice runs the existing ChangePlan logic, then generates an
// invoice record from the proration preview. Exposed as a wrapper so the
// existing /change endpoint stays backwards-compatible.
func (s *Service) ChangePlanWithInvoice(ctx context.Context, orgID uuid.UUID, req ChangePlanRequest) (*Subscription, *Invoice, error) {
	preview, _ := s.PreviewChange(ctx, orgID, PreviewChangeRequest{
		PlanCode:     req.PlanCode,
		BillingCycle: req.BillingCycle,
		Quantity:     req.Quantity,
		CouponCode:   req.CouponCode,
	})
	sub, err := s.ChangePlan(ctx, orgID, req)
	if err != nil {
		return nil, nil, err
	}
	tenantID := appctx.TenantID(ctx)
	if tenantID == uuid.Nil {
		return sub, nil, nil
	}
	inv, _ := s.generateInvoiceForChange(ctx, tenantID, orgID, sub, preview)

	// If a coupon was used, record the redemption and bump the counter.
	if req.CouponCode != "" {
		if coupon, err := s.repo.GetCouponByCode(ctx, req.CouponCode); err == nil && coupon.IsActive {
			actor := appctx.UserID(ctx)
			var by *uuid.UUID
			if actor != uuid.Nil {
				a := actor
				by = &a
			}
			var invID *uuid.UUID
			if inv != nil {
				id := inv.ID
				invID = &id
			}
			_ = s.repo.RecordCouponRedemption(ctx, &CouponRedemption{
				CouponID:       coupon.ID,
				OrganizationID: orgID,
				SubscriptionID: &sub.ID,
				InvoiceID:      invID,
				AmountOffCents: preview.DiscountCents,
				CreatedBy:      by,
			})
			_ = s.repo.IncrementCouponRedemptions(ctx, coupon.ID)
		}
	}
	return sub, inv, nil
}
