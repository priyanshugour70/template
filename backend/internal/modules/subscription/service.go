package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
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

func (s *Service) ListPlans(ctx context.Context) ([]Plan, error) {
	rows, err := s.repo.ListActivePlans(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list plans failed", err)
	}
	return rows, nil
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
		// No active subscription — create one immediately.
		return s.ProvisionTrial(ctx, sub.TenantID, orgID)
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

func (s *Service) ListUsage(ctx context.Context, orgID uuid.UUID) ([]UsageCounter, error) {
	rows, err := s.repo.ListUsage(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list usage failed", err)
	}
	return rows, nil
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
