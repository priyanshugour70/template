package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/modules/billing/payment"
	"github.com/your-org/your-service/internal/modules/billing/pdf"
	"github.com/your-org/your-service/internal/modules/billing/pricing"
	"github.com/your-org/your-service/internal/modules/tenant"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/mail"
	"github.com/your-org/your-service/internal/pkg/storage"
	"github.com/your-org/your-service/internal/queue"
)

const (
	subCachePrefix = "sub:"
	subCacheTTL    = 10 * time.Minute
)

type Service struct {
	repo       *Repository
	log        *zap.Logger
	cache      cache.Cache
	producer   queue.Producer
	s3         *storage.S3     // nil when not configured — PDF endpoints check before use
	tenantSvc  *tenant.Service // used to render bill-to blocks on invoice PDFs
	mailer     mail.Sender     // payment receipt + invoice issued emails; falls back to NoopSender
	processors map[payment.Method]payment.Processor
}

func NewService(
	repo *Repository,
	log *zap.Logger,
	c cache.Cache,
	p queue.Producer,
	s3 *storage.S3,
	tenantSvc *tenant.Service,
	mailer mail.Sender,
) *Service {
	if mailer == nil {
		mailer = mail.NoopSender{}
	}
	// Wire the payment processors. Cash/bank/cheque all flow through Manual.
	// Gateway points at the stub until a real integration lands in Phase 6+.
	manual := payment.NewManual()
	procs := map[payment.Method]payment.Processor{
		payment.MethodCash:         manual,
		payment.MethodBankTransfer: manual,
		payment.MethodCheque:       manual,
		payment.MethodGateway:      payment.NewGatewayStub(),
	}
	return &Service{
		repo: repo, log: log, cache: c, producer: p,
		s3: s3, tenantSvc: tenantSvc, mailer: mailer,
		processors: procs,
	}
}

// ── feature catalog ───────────────────────────────────────────────────────

// ListFeatures returns the live feature catalog. Cached for 10 min in Redis.
// Plan builder UI pulls this on every visit; the cache keeps the route hot.
func (s *Service) ListFeatures(ctx context.Context) ([]Feature, error) {
	rows, err := s.repo.ListFeatures(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list features failed", err)
	}
	return rows, nil
}

// PreviewQuote computes a quotation without persisting it. Drives the live
// pricing panel as the user toggles features in the plan-builder UI.
//
// Resolution order:
//  1. Load all active features into a catalog map.
//  2. Load tax config (home state + rates).
//  3. Pick the customer's billing state — req override → current subscription's
//     stored billing_state → home state (defaults to intra-state CGST+SGST).
//  4. Hand off to the pricing package (pure functions, fully unit-tested).
func (s *Service) PreviewQuote(ctx context.Context, req PreviewQuoteRequest) (*pricing.Quote, error) {
	rows, err := s.repo.ListFeatures(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load features failed", err)
	}
	catalog := make(map[string]pricing.Feature, len(rows))
	for _, f := range rows {
		catalog[f.Key] = pricing.Feature{
			Key:               f.Key,
			Name:              f.Name,
			Description:       f.Description,
			Category:          f.Category,
			BasePriceCents:    f.BasePriceCents,
			PerUserPriceCents: f.PerUserPriceCents,
			IncludedUsers:     f.IncludedUsers,
			IsCore:            f.IsCore,
			Requires:          []string(f.Requires),
			SortOrder:         f.SortOrder,
		}
	}

	tax, err := s.repo.GetTaxConfig(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load tax config failed", err)
	}

	customerState := req.CustomerState
	if customerState == "" {
		// Future: fall back to the org's active subscription billing_state when
		// one exists. For Phase 2 the preview-only endpoint defaults to home
		// state (intra-state CGST+SGST).
		customerState = tax.HomeState
	}

	q, err := pricing.BuildQuote(catalog, pricing.QuoteInput{
		SelectedKeys:  req.FeatureKeys,
		UserCount:     req.UserCount,
		CustomerState: customerState,
		HSNSAC:        tax.DefaultHSNSAC,
		Currency:      tax.Currency,
		Rates: pricing.Rates{
			HomeState: tax.HomeState,
			CGSTPct:   tax.DefaultCGSTPct,
			SGSTPct:   tax.DefaultSGSTPct,
			IGSTPct:   tax.DefaultIGSTPct,
		},
	})
	if err != nil {
		return nil, apperr.New(apperr.CodeValidation, err.Error(), err)
	}
	return &q, nil
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
// on the default plan. Idempotent: if an org already has a live subscription
// (trial/active/past_due/paused), returns it unchanged so the onboarding
// "Start trial" button can be hit repeatedly without exploding on the
// uidx_subscriptions_one_active_per_org partial unique index.
func (s *Service) ProvisionTrial(ctx context.Context, tenantID, orgID uuid.UUID) (*Subscription, error) {
	if existing, err := s.repo.GetActiveByOrg(ctx, orgID); err == nil && existing != nil {
		return existing, nil
	}
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

// ── billing info ──────────────────────────────────────────────────────────

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

// GetInvoicePDF returns the rendered PDF for an invoice. Lazy: first call
// renders, uploads to S3, persists pdf_storage_key on the invoice row; later
// calls stream the cached object straight from S3. Returns ([]byte, error).
//
// S3 is required — if not configured, we return a 503-style error so the API
// surfaces a clear "not configured" message instead of crashing.
func (s *Service) GetInvoicePDF(ctx context.Context, orgID, id uuid.UUID) ([]byte, error) {
	if s.s3 == nil {
		return nil, apperr.New(apperr.CodeInternal, "pdf storage not configured", nil)
	}
	inv, err := s.repo.GetInvoice(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "invoice not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load invoice failed", err)
	}

	// Cache hit — stream from S3.
	if inv.PDFStorageKey != "" {
		body, _, err := s.s3.GetBytes(ctx, inv.PDFStorageKey)
		if err == nil {
			return body, nil
		}
		s.log.Warn("invoice pdf cache miss, will re-render",
			zap.String("invoice", inv.Number), zap.String("key", inv.PDFStorageKey), zap.Error(err))
	}

	// Cache miss — render fresh, upload, persist the key.
	body, err := s.renderInvoicePDF(ctx, inv)
	if err != nil {
		return nil, err
	}
	key := s.s3.Key("invoices", inv.TenantID.String(), inv.Number+".pdf")
	if err := s.s3.PutBytes(ctx, key, "application/pdf", body); err != nil {
		// We can still hand the bytes back even if upload failed — the user
		// gets a PDF; we just lose caching for this invoice. Log and continue.
		s.log.Warn("invoice pdf upload failed", zap.String("invoice", inv.Number), zap.Error(err))
		return body, nil
	}
	// Best-effort cache key persist; if this fails we just re-render next time.
	if updErr := s.repo.UpdateInvoice(ctx, inv.ID, map[string]interface{}{"pdf_storage_key": key}); updErr != nil {
		s.log.Warn("invoice pdf key persist failed", zap.String("invoice", inv.Number), zap.Error(updErr))
	}
	return body, nil
}

// renderInvoicePDF collects everything the renderer needs (tax config, tenant
// + org info for "bill to", invoice + line items) and hands it off to the
// pdf package.
func (s *Service) renderInvoicePDF(ctx context.Context, inv *Invoice) ([]byte, error) {
	tax, err := s.repo.GetTaxConfig(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load tax config failed", err)
	}

	// Pull customer info from tenant/org so an empty billing_name on the
	// subscription still renders something on the PDF. Then overlay the
	// subscription's billing_email + billing_name (set at quotation time)
	// if available.
	customerName := inv.OrganizationID.String()
	customerAddress := ""
	customerEmail := ""
	if s.tenantSvc != nil {
		if t, _ := s.tenantSvc.GetTenant(ctx, inv.TenantID); t != nil {
			customerName = t.Name
		}
		if o, _ := s.tenantSvc.GetOrganization(ctx, inv.TenantID, inv.OrganizationID); o != nil && o.Name != "" {
			customerName = o.Name
		}
	}
	if inv.SubscriptionID != nil {
		var sub Subscription
		if err := s.repo.DB().WithContext(ctx).
			Where("id = ?", *inv.SubscriptionID).First(&sub).Error; err == nil {
			if sub.BillingName != "" {
				customerName = sub.BillingName
			}
			if sub.BillingEmail != "" {
				customerEmail = sub.BillingEmail
			}
		}
	}

	lines, err := s.repo.ListInvoiceLines(ctx, inv.ID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load invoice lines failed", err)
	}
	pdfLines := make([]pdf.InvoiceLineRow, 0, len(lines))
	for _, l := range lines {
		pdfLines = append(pdfLines, pdf.InvoiceLineRow{
			Description:        l.Description,
			HSNSAC:             l.HSNSAC,
			Quantity:           l.Quantity,
			UnitPriceCents:     l.UnitPriceCents,
			TaxableAmountCents: l.TaxableAmountCents,
			CGSTCents:          l.CGSTCents,
			SGSTCents:          l.SGSTCents,
			IGSTCents:          l.IGSTCents,
			TotalCents:         l.TotalCents,
		})
	}

	// IntraState flag: the renderer decides which tax columns to show. We
	// reuse the same comparison the tax calculator uses.
	intra := inv.IGSTCents == 0

	dueAt := time.Time{}
	if inv.DueAt != nil {
		dueAt = *inv.DueAt
	}
	periodStart, periodEnd := time.Time{}, time.Time{}
	if inv.PeriodStart != nil {
		periodStart = *inv.PeriodStart
	}
	if inv.PeriodEnd != nil {
		periodEnd = *inv.PeriodEnd
	}

	in := pdf.InvoiceInput{
		CompanyName:     tax.CompanyName,
		CompanyAddress:  tax.CompanyAddress,
		CompanyGSTIN:    tax.GSTIN,
		BankName:        tax.BankName,
		BankAccount:     tax.BankAccountNumber,
		BankIFSC:        tax.BankIFSC,
		BankAccountName: tax.BankAccountName,
		InvoiceTerms:    tax.InvoiceTerms,

		CustomerName:    customerName,
		CustomerEmail:   customerEmail,
		CustomerAddress: customerAddress,
		PlaceOfSupply:   inv.PlaceOfSupply,

		InvoiceNumber: inv.Number,
		IssuedAt:      inv.IssuedAt,
		DueAt:         dueAt,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Status:        inv.Status,

		Currency:        inv.Currency,
		Lines:           pdfLines,
		SubtotalCents:   inv.SubtotalCents,
		DiscountCents:   inv.DiscountCents,
		CGSTCents:       inv.CGSTCents,
		SGSTCents:       inv.SGSTCents,
		IGSTCents:       inv.IGSTCents,
		TotalCents:      inv.TotalCents,
		AmountDueCents:  inv.AmountDueCents,
		AmountPaidCents: inv.AmountPaidCents,
		IntraState:      intra,
	}

	renderer := pdf.NewRenderer()
	body, err := renderer.RenderInvoice(in)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "render invoice pdf failed", err)
	}
	return body, nil
}

// ── quotations (Phase 3) ──────────────────────────────────────────────────

const (
	quotationValidityDays = 30
	subscriptionCycleDays = 30
	invoiceDueDays        = 7
)

// CreateQuotation persists a new draft. Pricing is computed at draft time and
// re-computed at activation time; in between, the snapshot here keeps the UI
// responsive even when the catalog changes underneath.
func (s *Service) CreateQuotation(ctx context.Context, req CreateQuotationRequest) (*Quotation, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	if tid == uuid.Nil || oid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no tenant/org context", nil)
	}

	tax, catalog, err := s.loadCatalogAndTax(ctx)
	if err != nil {
		return nil, err
	}

	customerState := req.CustomerState
	if customerState == "" {
		customerState = tax.HomeState
	}
	q, err := s.runPricing(ctx, catalog, tax, req.FeatureKeys, req.UserCount, customerState)
	if err != nil {
		return nil, err
	}

	num, err := s.repo.NextQuotationNumber(ctx, tid, time.Now().Year())
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "next quotation number failed", err)
	}

	lineItemsJSON, _ := json.Marshal(q.Lines)
	addressJSON := []byte("{}")
	if req.BillingAddress != nil {
		addressJSON, _ = json.Marshal(req.BillingAddress)
	}

	row := &Quotation{
		TenantID:       tid,
		OrganizationID: oid,
		Number:         num,
		Status:         "draft",
		FeatureKeys:    pqStringArray(req.FeatureKeys),
		UserCount:      req.UserCount,
		SubtotalCents:  q.SubtotalCents,
		CGSTCents:      q.CGSTCents,
		SGSTCents:      q.SGSTCents,
		IGSTCents:      q.IGSTCents,
		TotalCents:     q.TotalCents,
		Currency:       q.Currency,
		PlaceOfSupply:  q.PlaceOfSupply,
		LineItems:      lineItemsJSON,
		BillingEmail:   req.BillingEmail,
		BillingName:    req.BillingName,
		BillingAddress: addressJSON,
		BillingState:   customerState,
		Notes:          req.Notes,
		ExpiresAt:      time.Now().Add(quotationValidityDays * 24 * time.Hour),
		Metadata:       []byte("{}"),
	}
	if err := s.repo.CreateQuotation(ctx, row); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "persist quotation failed", err)
	}
	return row, nil
}

// GetQuotation returns a quotation scoped to the caller's org. Super-admin
// bypass is intentional: cross-tenant billing access is restricted by the
// route's permission, not the lookup.
func (s *Service) GetQuotation(ctx context.Context, id uuid.UUID) (*Quotation, error) {
	oid := appctx.OrganizationID(ctx)
	row, err := s.repo.GetQuotation(ctx, oid, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "quotation not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch quotation failed", err)
	}
	return row, nil
}

func (s *Service) ListQuotations(ctx context.Context, status string, limit, offset int) ([]Quotation, int64, error) {
	oid := appctx.OrganizationID(ctx)
	if oid == uuid.Nil {
		return nil, 0, apperr.New(apperr.CodeForbidden, "no org context", nil)
	}
	rows, total, err := s.repo.ListQuotations(ctx, oid, status, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list quotations failed", err)
	}
	return rows, total, nil
}

// UpdateQuotation patches a draft and re-computes pricing if the selection
// changed. Activated/rejected/expired drafts are immutable.
func (s *Service) UpdateQuotation(ctx context.Context, id uuid.UUID, req UpdateQuotationRequest) (*Quotation, error) {
	row, err := s.GetQuotation(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status != "draft" {
		return nil, apperr.New(apperr.CodeValidation, "only draft quotations can be edited", nil)
	}

	// Re-price if any of the inputs that drive pricing actually changed.
	repriced := false
	keys := []string(row.FeatureKeys)
	if req.FeatureKeys != nil {
		keys = *req.FeatureKeys
		repriced = true
	}
	userCount := row.UserCount
	if req.UserCount != nil {
		userCount = *req.UserCount
		repriced = true
	}
	customerState := row.BillingState
	if req.CustomerState != nil {
		customerState = *req.CustomerState
		repriced = true
	}

	patch := map[string]interface{}{}
	if repriced {
		tax, catalog, err := s.loadCatalogAndTax(ctx)
		if err != nil {
			return nil, err
		}
		if customerState == "" {
			customerState = tax.HomeState
		}
		q, err := s.runPricing(ctx, catalog, tax, keys, userCount, customerState)
		if err != nil {
			return nil, err
		}
		lineItemsJSON, _ := json.Marshal(q.Lines)
		patch["feature_keys"] = pqStringArray(keys)
		patch["user_count"] = userCount
		patch["billing_state"] = customerState
		patch["subtotal_cents"] = q.SubtotalCents
		patch["cgst_cents"] = q.CGSTCents
		patch["sgst_cents"] = q.SGSTCents
		patch["igst_cents"] = q.IGSTCents
		patch["total_cents"] = q.TotalCents
		patch["place_of_supply"] = q.PlaceOfSupply
		patch["line_items"] = lineItemsJSON
	}
	if req.BillingEmail != nil {
		patch["billing_email"] = *req.BillingEmail
	}
	if req.BillingName != nil {
		patch["billing_name"] = *req.BillingName
	}
	if req.BillingAddress != nil {
		b, _ := json.Marshal(*req.BillingAddress)
		patch["billing_address"] = b
	}
	if req.Notes != nil {
		patch["notes"] = *req.Notes
	}

	if len(patch) == 0 {
		return row, nil
	}
	if err := s.repo.UpdateQuotation(ctx, id, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "update quotation failed", err)
	}
	return s.GetQuotation(ctx, id)
}

// DeleteQuotation soft-deletes a draft. Activated drafts can't be removed —
// they're the historical record behind a live subscription.
func (s *Service) DeleteQuotation(ctx context.Context, id uuid.UUID) error {
	row, err := s.GetQuotation(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == "accepted" {
		return apperr.New(apperr.CodeValidation, "cannot delete an activated quotation", nil)
	}
	if err := s.repo.DeleteQuotation(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "delete quotation failed", err)
	}
	return nil
}

// ActivateQuotation turns a draft into a live subscription + open invoice.
// Atomic: either every write succeeds or none of them do.
//
// Re-prices using the current catalog so a stale draft can't lock in obsolete
// pricing. The activated plan + plan_features rows preserve the snapshot —
// future catalog changes won't retroactively re-bill the customer.
func (s *Service) ActivateQuotation(ctx context.Context, id uuid.UUID) (*ActivateQuotationResponse, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if tid == uuid.Nil || oid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no tenant/org context", nil)
	}

	row, err := s.repo.GetQuotation(ctx, oid, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "quotation not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch quotation failed", err)
	}
	if row.Status != "draft" {
		return nil, apperr.New(apperr.CodeValidation, "only draft quotations can be activated", nil)
	}
	if row.ExpiresAt.Before(time.Now()) {
		return nil, apperr.New(apperr.CodeValidation, "quotation has expired", nil)
	}

	tax, catalog, err := s.loadCatalogAndTax(ctx)
	if err != nil {
		return nil, err
	}
	customerState := row.BillingState
	if customerState == "" {
		customerState = tax.HomeState
	}
	q, err := s.runPricing(ctx, catalog, tax, []string(row.FeatureKeys), row.UserCount, customerState)
	if err != nil {
		return nil, err
	}

	// Load the underlying feature rows so we can FK every PlanFeature to a
	// real billing_features.id (the PlanFeature snapshot doesn't soft-couple
	// via key alone — RESTRICT delete on the FK enforces referential safety).
	allFeats, err := s.repo.ListFeatures(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load features failed", err)
	}
	featByKey := make(map[string]Feature, len(allFeats))
	for _, f := range allFeats {
		featByKey[f.Key] = f
	}

	// Distinct feature keys present in the quote (auto-included core +
	// requires-chain + extra_users may not match the user's input).
	keysInQuote := uniqueLineKeys(q.Lines)

	now := time.Now()
	periodStart := now
	periodEnd := now.Add(subscriptionCycleDays * 24 * time.Hour)

	invoiceNum, err := s.repo.NextInvoiceNumber(ctx, tid, now.Year())
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "next invoice number failed", err)
	}

	var (
		respPlan        Plan
		respSubscription Subscription
		respInvoice     Invoice
		respLines       []InvoiceLine
	)

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create the custom plan (snapshot of feature selection).
		featuresJSON, _ := json.Marshal(keysInQuote)
		limits := map[string]int64{
			"users.max": int64(row.UserCount),
		}
		limitsJSON, _ := json.Marshal(limits)
		plan := Plan{
			Code:         "custom-" + strings.ToLower(row.Number),
			Name:         "Custom plan " + row.Number,
			Tier:         99,
			BillingCycle: "monthly",
			PriceCents:   q.TotalCents,
			Currency:     q.Currency,
			IsActive:     true,
			IsDefault:    false,
			IsPublic:     false,
			IsAddon:      false,
			IsCustom:     true,
			Features:     featuresJSON,
			Limits:       limitsJSON,
			Metadata:     []byte(`{"source":"quotation","quotation":"` + row.Number + `"}`),
		}
		if err := tx.Create(&plan).Error; err != nil {
			return fmt.Errorf("create plan: %w", err)
		}

		// 2. Snapshot price for each feature in the plan.
		pfRows := make([]PlanFeature, 0, len(keysInQuote))
		for _, key := range keysInQuote {
			f, ok := featByKey[key]
			if !ok {
				continue
			}
			qty := 1
			// extra_users line carries the actual additional headcount as qty.
			if key == "extra_users" {
				qty = q.ExtraUsers
				if qty <= 0 {
					continue
				}
			}
			pfRows = append(pfRows, PlanFeature{
				PlanID:            plan.ID,
				FeatureID:         f.ID,
				FeatureKey:        f.Key,
				BasePriceCents:    f.BasePriceCents,
				PerUserPriceCents: f.PerUserPriceCents,
				IncludedUsers:     f.IncludedUsers,
				Quantity:          qty,
			})
		}
		if len(pfRows) > 0 {
			if err := tx.Create(&pfRows).Error; err != nil {
				return fmt.Errorf("create plan features: %w", err)
			}
		}

		// 3. Subscription — status='pending' (payment due). 'pending' is the
		//    existing CHECK-constraint value closest to "awaiting first payment".
		taxCents := q.CGSTCents + q.SGSTCents + q.IGSTCents
		emptyBilling := []byte("{}")
		sub := Subscription{
			TenantID:           tid,
			OrganizationID:     oid,
			PlanID:             plan.ID,
			PlanCode:           plan.Code,
			Status:             "pending",
			BillingCycle:       plan.BillingCycle,
			Quantity:           1,
			UnitPriceCents:     q.SubtotalCents,
			DiscountCents:      0,
			TaxCents:           taxCents,
			TotalCents:         q.TotalCents,
			Currency:           q.Currency,
			StartedAt:          now,
			CurrentPeriodStart: &periodStart,
			CurrentPeriodEnd:   &periodEnd,
			NextBillingAt:      &periodEnd,
			BillingEmail:       row.BillingEmail,
			BillingName:        row.BillingName,
			BillingAddress:     coalesceJSONB(row.BillingAddress, emptyBilling),
			BillingState:       customerState,
			Features:           plan.Features,
			Limits:             plan.Limits,
			Metadata:           []byte(`{"source":"quotation","quotation":"` + row.Number + `"}`),
		}
		if err := tx.Create(&sub).Error; err != nil {
			return fmt.Errorf("create subscription: %w", err)
		}

		// 4. Invoice — open, due in 7 days.
		due := now.Add(invoiceDueDays * 24 * time.Hour)
		lineItemsJSON, _ := json.Marshal(q.Lines)
		inv := Invoice{
			TenantID:        tid,
			OrganizationID:  oid,
			SubscriptionID:  &sub.ID,
			Number:          invoiceNum,
			Status:          "open",
			Currency:        q.Currency,
			SubtotalCents:   q.SubtotalCents,
			DiscountCents:   0,
			TaxCents:        taxCents,
			TotalCents:      q.TotalCents,
			AmountDueCents:  q.TotalCents,
			AmountPaidCents: 0,
			LineItems:       lineItemsJSON,
			PeriodStart:     &periodStart,
			PeriodEnd:       &periodEnd,
			IssuedAt:        now,
			DueAt:           &due,
			HSNSAC:          tax.DefaultHSNSAC,
			PlaceOfSupply:   customerState,
			CGSTCents:       q.CGSTCents,
			SGSTCents:       q.SGSTCents,
			IGSTCents:       q.IGSTCents,
			Metadata:        []byte(`{"source":"quotation","quotation":"` + row.Number + `"}`),
		}
		if err := tx.Create(&inv).Error; err != nil {
			return fmt.Errorf("create invoice: %w", err)
		}

		// 5. Relational invoice lines (alongside the JSONB copy on the invoice).
		lineRows := make([]InvoiceLine, 0, len(q.Lines))
		for _, l := range q.Lines {
			lineRows = append(lineRows, InvoiceLine{
				InvoiceID:          inv.ID,
				FeatureKey:         l.FeatureKey,
				Description:        l.Description,
				HSNSAC:             l.HSNSAC,
				Quantity:           l.Quantity,
				UnitPriceCents:     l.UnitPriceCents,
				TaxableAmountCents: l.TaxableAmountCents,
				CGSTCents:          l.Tax.CGSTCents,
				SGSTCents:          l.Tax.SGSTCents,
				IGSTCents:          l.Tax.IGSTCents,
				TotalCents:         l.TotalCents,
				SortOrder:          l.SortOrder,
				Metadata:           []byte("{}"),
			})
		}
		if len(lineRows) > 0 {
			if err := tx.Create(&lineRows).Error; err != nil {
				return fmt.Errorf("create invoice lines: %w", err)
			}
		}

		// 6. Mark quotation accepted, link the new plan + subscription.
		acceptedAt := time.Now()
		planID, subID := plan.ID, sub.ID
		if err := tx.Model(&Quotation{}).
			Where("id = ?", row.ID).
			Updates(map[string]interface{}{
				"status":                    "accepted",
				"accepted_at":               acceptedAt,
				"activated_plan_id":         planID,
				"activated_subscription_id": subID,
				"updated_by":                nullableUUID(uid),
			}).Error; err != nil {
			return fmt.Errorf("update quotation: %w", err)
		}

		respPlan = plan
		respSubscription = sub
		respInvoice = inv
		respLines = lineRows
		return nil
	})
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "activate quotation failed", err)
	}

	// Refresh quotation for the response.
	updated, _ := s.repo.GetQuotation(ctx, oid, id)
	if updated == nil {
		updated = row
	}

	// Bust the org's cached feature set so middleware sees the new plan.
	s.invalidateCache(ctx, oid)

	return &ActivateQuotationResponse{
		Quotation:    *updated,
		Plan:         respPlan,
		Subscription: respSubscription,
		Invoice:      respInvoice,
		InvoiceLines: respLines,
	}, nil
}

// ── quotation helpers ─────────────────────────────────────────────────────

func (s *Service) loadCatalogAndTax(ctx context.Context) (*TaxConfig, map[string]pricing.Feature, error) {
	rows, err := s.repo.ListFeatures(ctx)
	if err != nil {
		return nil, nil, apperr.New(apperr.CodeInternal, "load features failed", err)
	}
	tax, err := s.repo.GetTaxConfig(ctx)
	if err != nil {
		return nil, nil, apperr.New(apperr.CodeInternal, "load tax config failed", err)
	}
	catalog := make(map[string]pricing.Feature, len(rows))
	for _, f := range rows {
		catalog[f.Key] = pricing.Feature{
			Key:               f.Key,
			Name:              f.Name,
			Description:       f.Description,
			Category:          f.Category,
			BasePriceCents:    f.BasePriceCents,
			PerUserPriceCents: f.PerUserPriceCents,
			IncludedUsers:     f.IncludedUsers,
			IsCore:            f.IsCore,
			Requires:          []string(f.Requires),
			SortOrder:         f.SortOrder,
		}
	}
	return tax, catalog, nil
}

func (s *Service) runPricing(_ context.Context, catalog map[string]pricing.Feature, tax *TaxConfig, keys []string, userCount int, customerState string) (*pricing.Quote, error) {
	q, err := pricing.BuildQuote(catalog, pricing.QuoteInput{
		SelectedKeys:  keys,
		UserCount:     userCount,
		CustomerState: customerState,
		HSNSAC:        tax.DefaultHSNSAC,
		Currency:      tax.Currency,
		Rates: pricing.Rates{
			HomeState: tax.HomeState,
			CGSTPct:   tax.DefaultCGSTPct,
			SGSTPct:   tax.DefaultSGSTPct,
			IGSTPct:   tax.DefaultIGSTPct,
		},
	})
	if err != nil {
		return nil, apperr.New(apperr.CodeValidation, err.Error(), err)
	}
	return &q, nil
}

func uniqueLineKeys(lines []pricing.Line) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if _, ok := seen[l.FeatureKey]; ok {
			continue
		}
		seen[l.FeatureKey] = struct{}{}
		out = append(out, l.FeatureKey)
	}
	return out
}

func nullableUUID(u uuid.UUID) *uuid.UUID {
	if u == uuid.Nil {
		return nil
	}
	id := u
	return &id
}

func coalesceJSONB(primary, fallback []byte) []byte {
	if len(primary) > 0 && string(primary) != "null" {
		return primary
	}
	return fallback
}

// pqStringArray wraps a regular slice in pq.StringArray so it round-trips
// through the gorm `text[]` mapper.
func pqStringArray(s []string) pq.StringArray { return pq.StringArray(s) }

// ── payments / receipts (Phase 5) ─────────────────────────────────────────

// RecordPayment captures a payment against an open invoice. End-to-end:
//   1. Load + validate the invoice (status="open", amount matches due).
//   2. Hand off to the configured processor (manual today; gateway in Phase 6+).
//   3. Atomically:
//        a. mint a receipt number;
//        b. insert a billing_transactions row;
//        c. flip the invoice to "paid";
//        d. if a pending subscription is attached, flip it to "active".
//   4. Render the receipt PDF + upload to S3 (lazy — best-effort).
//   5. Email the receipt to the subscription's billing_email (with PDF attached).
//
// PDF upload + email send are best-effort: a failure there does NOT roll back
// the transaction. We've taken the customer's money — we'd rather log the
// failure and let an admin retry the email than reject the payment.
func (s *Service) RecordPayment(ctx context.Context, invoiceID uuid.UUID, req RecordPaymentRequest) (*RecordPaymentResponse, error) {
	oid := appctx.OrganizationID(ctx)
	if oid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no org context", nil)
	}

	inv, err := s.repo.GetInvoice(ctx, oid, invoiceID)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "invoice not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load invoice failed", err)
	}
	if inv.Status != "open" {
		return nil, apperr.New(apperr.CodeValidation,
			fmt.Sprintf("invoice is %q — only open invoices can be paid", inv.Status), nil)
	}
	// Phase 5 collects payment in one go. Partial payments would change the
	// validation to amount <= due + bookkeeping for amount_paid_cents.
	if req.AmountCents != inv.AmountDueCents {
		return nil, apperr.New(apperr.CodeValidation,
			fmt.Sprintf("amount %d does not match invoice due %d", req.AmountCents, inv.AmountDueCents), nil)
	}

	method := payment.Method(req.Method)
	processor, ok := s.processors[method]
	if !ok {
		return nil, apperr.New(apperr.CodeValidation, "unsupported payment method", nil)
	}

	// Audit-trail breadcrumb on the receipt (who recorded the payment).
	actorID := appctx.UserID(ctx)
	actorEmail := appctx.Email(ctx)

	receipt, err := processor.Record(ctx, payment.RecordRequest{
		InvoiceID:       inv.ID.String(),
		Method:          method,
		AmountCents:     req.AmountCents,
		Currency:        inv.Currency,
		Reference:       req.Reference,
		Notes:           req.Notes,
		RecordedByID:    actorID.String(),
		RecordedByEmail: actorEmail,
	})
	if err != nil {
		return nil, apperr.New(apperr.CodeValidation, err.Error(), err)
	}

	// Build the receipt number outside the txn so we can roll back the txn
	// without consuming a number. NextReceiptNumber is a count() so it isn't
	// strictly atomic — fine for v1, swap for a postgres sequence if collisions
	// ever materialize.
	receiptNum, err := s.repo.NextReceiptNumber(ctx, inv.TenantID, time.Now().Year())
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "next receipt number failed", err)
	}

	now := time.Now()
	txn := Transaction{
		TenantID:             inv.TenantID,
		OrganizationID:       inv.OrganizationID,
		InvoiceID:            inv.ID,
		ReceiptNumber:        receiptNum,
		Method:               string(receipt.Method),
		Status:               receipt.Status,
		AmountCents:          receipt.AmountCents,
		Currency:             receipt.Currency,
		Reference:            receipt.Reference,
		Gateway:              receipt.Gateway,
		GatewayTransactionID: receipt.GatewayTransactionID,
		PaidAt:               now,
		Notes:                req.Notes,
		Metadata:             []byte(`{}`),
	}

	var (
		respTxn          Transaction
		respInvoice      Invoice
		respSubscription *Subscription
	)

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Persist the transaction row.
		if err := tx.Create(&txn).Error; err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		// 2. Flip the invoice to paid. amount_paid_cents reflects the captured
		//    amount (== amount_due_cents after the validation above).
		invPatch := map[string]interface{}{
			"status":            "paid",
			"amount_paid_cents": gorm.Expr("amount_paid_cents + ?", receipt.AmountCents),
			"amount_due_cents":  gorm.Expr("amount_due_cents - ?", receipt.AmountCents),
			"paid_at":           now,
		}
		if err := tx.Model(&Invoice{}).Where("id = ?", inv.ID).Updates(invPatch).Error; err != nil {
			return fmt.Errorf("update invoice: %w", err)
		}
		// Reload to populate response with the new amount_paid / amount_due.
		var refreshedInv Invoice
		if err := tx.Where("id = ?", inv.ID).First(&refreshedInv).Error; err != nil {
			return fmt.Errorf("reload invoice: %w", err)
		}
		respInvoice = refreshedInv

		// 3. If a pending subscription is attached to this invoice, this
		//    payment is the activation trigger. Flip it to "active".
		//
		//    A partial unique index (uidx_subscriptions_one_active_per_org)
		//    only allows one row per org in {trial, active, past_due, paused}.
		//    So if the org has an existing trial/active subscription (e.g. the
		//    seed Pro plan), we expire it first — paying for a new plan via
		//    quotation activation is semantically a replacement.
		if inv.SubscriptionID != nil {
			var sub Subscription
			if err := tx.Where("id = ?", *inv.SubscriptionID).First(&sub).Error; err != nil {
				return fmt.Errorf("load subscription: %w", err)
			}
			if sub.Status == "pending" {
				// Retire any prior live subscription for this org so the
				// unique-active index doesn't trip on the flip below.
				if err := tx.Model(&Subscription{}).
					Where(`organization_id = ?
					       AND id <> ?
					       AND status IN ('trial', 'active', 'past_due', 'paused')`,
						sub.OrganizationID, sub.ID).
					Updates(map[string]interface{}{
						"status":        "expired",
						"ended_at":      now,
						"cancel_reason": "replaced by " + sub.PlanCode,
					}).Error; err != nil {
					return fmt.Errorf("retire prior active subscription: %w", err)
				}
				if err := tx.Model(&Subscription{}).
					Where("id = ?", sub.ID).
					Updates(map[string]interface{}{
						"status":               "active",
						"last_billed_at":       now,
						"current_period_start": now,
					}).Error; err != nil {
					return fmt.Errorf("activate subscription: %w", err)
				}
				// Re-fetch so the response carries the new status.
				if err := tx.Where("id = ?", sub.ID).First(&sub).Error; err != nil {
					return fmt.Errorf("reload subscription: %w", err)
				}
			}
			respSubscription = &sub
		}

		respTxn = txn
		return nil
	})
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "record payment failed", err)
	}

	// Bust subscription cache so feature/quota middleware sees the activated state.
	s.invalidateCache(ctx, oid)

	// Render receipt PDF + upload + send email. All best-effort — payment is
	// already committed. We capture any failure on the log and continue so the
	// caller doesn't see a 500 just because email was misconfigured.
	pdfBytes, pdfErr := s.renderReceiptPDF(ctx, &respTxn, &respInvoice)
	if pdfErr != nil {
		s.log.Warn("receipt pdf render failed",
			zap.String("receipt", respTxn.ReceiptNumber), zap.Error(pdfErr))
	}

	if pdfErr == nil && s.s3 != nil {
		key := s.s3.Key("receipts", respTxn.TenantID.String(), respTxn.ReceiptNumber+".pdf")
		if upErr := s.s3.PutBytes(ctx, key, "application/pdf", pdfBytes); upErr != nil {
			s.log.Warn("receipt pdf upload failed",
				zap.String("receipt", respTxn.ReceiptNumber), zap.Error(upErr))
		} else if updErr := s.repo.UpdateTransaction(ctx, respTxn.ID, map[string]interface{}{
			"pdf_storage_key": key,
		}); updErr != nil {
			s.log.Warn("receipt pdf key persist failed",
				zap.String("receipt", respTxn.ReceiptNumber), zap.Error(updErr))
		} else {
			respTxn.PDFStorageKey = key
		}
	}

	// Email — only if we know who to email AND we successfully rendered.
	recipient := s.resolveBillingEmail(ctx, &respInvoice, respSubscription)
	if recipient != "" {
		paidAtFmt := respTxn.PaidAt.Format("02 Jan 2006, 15:04 MST")
		amtFmt := formatMoney(respTxn.Currency, respTxn.AmountCents)
		body := mail.EmailPaymentReceipt(
			respTxn.ReceiptNumber, respInvoice.Number, amtFmt,
			payment.MethodLabel(method), paidAtFmt)
		var mailErr error
		if pdfErr == nil && len(pdfBytes) > 0 {
			mailErr = s.mailer.SendWithAttachments(recipient,
				"Payment received — receipt "+respTxn.ReceiptNumber, body,
				mail.Attachment{
					Filename: "receipt-" + respTxn.ReceiptNumber + ".pdf",
					MIMEType: "application/pdf",
					Data:     pdfBytes,
				})
		} else {
			mailErr = s.mailer.Send(recipient,
				"Payment received — receipt "+respTxn.ReceiptNumber, body)
		}
		if mailErr != nil {
			s.log.Warn("receipt email send failed",
				zap.String("to", recipient), zap.Error(mailErr))
		}
	}

	resp := &RecordPaymentResponse{
		Transaction:  respTxn,
		Invoice:      respInvoice,
		Subscription: respSubscription,
	}
	if respTxn.PDFStorageKey != "" {
		resp.ReceiptURL = fmt.Sprintf("/api/v1/billing/receipts/%s/pdf", respTxn.ID.String())
	}
	return resp, nil
}

// GetTransaction returns a single transaction scoped to the caller's org.
func (s *Service) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	oid := appctx.OrganizationID(ctx)
	t, err := s.repo.GetTransaction(ctx, oid, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "transaction not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load transaction failed", err)
	}
	return t, nil
}

func (s *Service) ListTransactions(ctx context.Context, limit, offset int) ([]Transaction, int64, error) {
	oid := appctx.OrganizationID(ctx)
	if oid == uuid.Nil {
		return nil, 0, apperr.New(apperr.CodeForbidden, "no org context", nil)
	}
	rows, total, err := s.repo.ListTransactions(ctx, oid, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list transactions failed", err)
	}
	return rows, total, nil
}

// GetReceiptPDF mirrors GetInvoicePDF: stream from S3 if cached, otherwise
// render fresh, upload, persist the key, and return the bytes.
func (s *Service) GetReceiptPDF(ctx context.Context, id uuid.UUID) ([]byte, error) {
	if s.s3 == nil {
		// Without S3 we can still render on demand — receipts are tiny.
		t, err := s.GetTransaction(ctx, id)
		if err != nil {
			return nil, err
		}
		inv, err := s.repo.GetInvoice(ctx, t.OrganizationID, t.InvoiceID)
		if err != nil {
			return nil, apperr.New(apperr.CodeInternal, "load invoice failed", err)
		}
		return s.renderReceiptPDF(ctx, t, inv)
	}

	t, err := s.GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}

	if t.PDFStorageKey != "" {
		body, _, err := s.s3.GetBytes(ctx, t.PDFStorageKey)
		if err == nil {
			return body, nil
		}
		s.log.Warn("receipt pdf cache miss, will re-render",
			zap.String("receipt", t.ReceiptNumber), zap.String("key", t.PDFStorageKey), zap.Error(err))
	}

	inv, err := s.repo.GetInvoice(ctx, t.OrganizationID, t.InvoiceID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load invoice failed", err)
	}
	body, err := s.renderReceiptPDF(ctx, t, inv)
	if err != nil {
		return nil, err
	}
	key := s.s3.Key("receipts", t.TenantID.String(), t.ReceiptNumber+".pdf")
	if err := s.s3.PutBytes(ctx, key, "application/pdf", body); err != nil {
		s.log.Warn("receipt pdf upload failed", zap.String("receipt", t.ReceiptNumber), zap.Error(err))
		return body, nil
	}
	if updErr := s.repo.UpdateTransaction(ctx, t.ID, map[string]interface{}{"pdf_storage_key": key}); updErr != nil {
		s.log.Warn("receipt pdf key persist failed", zap.String("receipt", t.ReceiptNumber), zap.Error(updErr))
	}
	return body, nil
}

// renderReceiptPDF gathers everything the renderer needs (tax config for
// issuer fields, customer info from tenant/subscription) and produces bytes.
func (s *Service) renderReceiptPDF(ctx context.Context, t *Transaction, inv *Invoice) ([]byte, error) {
	tax, err := s.repo.GetTaxConfig(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load tax config failed", err)
	}

	customerName := inv.OrganizationID.String()
	customerEmail := ""
	if s.tenantSvc != nil {
		if t2, _ := s.tenantSvc.GetTenant(ctx, inv.TenantID); t2 != nil {
			customerName = t2.Name
		}
		if o, _ := s.tenantSvc.GetOrganization(ctx, inv.TenantID, inv.OrganizationID); o != nil && o.Name != "" {
			customerName = o.Name
		}
	}
	if inv.SubscriptionID != nil {
		var sub Subscription
		if err := s.repo.DB().WithContext(ctx).
			Where("id = ?", *inv.SubscriptionID).First(&sub).Error; err == nil {
			if sub.BillingName != "" {
				customerName = sub.BillingName
			}
			if sub.BillingEmail != "" {
				customerEmail = sub.BillingEmail
			}
		}
	}

	in := pdf.ReceiptInput{
		CompanyName:    tax.CompanyName,
		CompanyAddress: tax.CompanyAddress,
		CompanyGSTIN:   tax.GSTIN,
		ReceiptNumber:  t.ReceiptNumber,
		PaidAt:         t.PaidAt,
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		InvoiceNumber:  inv.Number,
		Method:         t.Method,
		Reference:      t.Reference,
		AmountCents:    t.AmountCents,
		Currency:       t.Currency,
	}
	renderer := pdf.NewRenderer()
	body, err := renderer.RenderReceipt(in)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "render receipt pdf failed", err)
	}
	return body, nil
}

// resolveBillingEmail picks the address to send the receipt to. Priority:
//   1. subscription.billing_email (set at quotation time);
//   2. invoice/tenant fallback (handled inside the conditional).
// Returns "" when no usable address is known — caller logs and skips.
func (s *Service) resolveBillingEmail(ctx context.Context, inv *Invoice, sub *Subscription) string {
	if sub != nil && sub.BillingEmail != "" {
		return sub.BillingEmail
	}
	if inv != nil && inv.SubscriptionID != nil {
		var loaded Subscription
		if err := s.repo.DB().WithContext(ctx).
			Where("id = ?", *inv.SubscriptionID).First(&loaded).Error; err == nil && loaded.BillingEmail != "" {
			return loaded.BillingEmail
		}
	}
	return ""
}

// formatMoney renders cents as "₹1,234.56" / "$12.34". Only used in transactional
// emails — PDF renderer has its own formatter optimised for layout.
func formatMoney(currency string, cents int64) string {
	sym := currencySymbol(currency)
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%s%d.%02d", sym, whole, frac)
}

func currencySymbol(c string) string {
	switch strings.ToUpper(c) {
	case "INR":
		return "₹"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	default:
		return c + " "
	}
}
