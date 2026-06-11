package tenant

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Service struct {
	repo  *Repository
	log   *zap.Logger
	cache cache.Cache
}

func NewService(repo *Repository, log *zap.Logger, c cache.Cache) *Service {
	return &Service{repo: repo, log: log, cache: c}
}

// ── tenants ────────────────────────────────────────────────────────────────

// CreateTenant provisions a new tenant + its default organization in one TX.
// Designed to be called by a super-admin / out-of-band onboarding flow.
// The first admin user creation/invite is handled by the auth module after the
// returned tenant ID is known.
func (s *Service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, *Organization, error) {
	slug := normaliseSlug(req.Slug)
	if existing, err := s.repo.GetTenantBySlug(ctx, slug); err == nil && existing != nil {
		return nil, nil, apperr.New(apperr.CodeAlreadyExists, "tenant slug already taken", nil)
	}

	t := &Tenant{
		Slug:         slug,
		Name:         strings.TrimSpace(req.Name),
		LegalName:    req.LegalName,
		DisplayName:  firstNonEmpty(req.DisplayName, req.Name),
		Description:  req.Description,
		LogoURL:      req.LogoURL,
		PrimaryColor: req.PrimaryColor,
		SupportEmail: req.SupportEmail,
		SupportPhone: req.SupportPhone,
		WebsiteURL:   req.WebsiteURL,
		Status:       "trial",
		PlanCode:     firstNonEmpty(req.PlanCode, "free"),
		Country:      req.Country,
		Timezone:     firstNonEmpty(req.Timezone, "Asia/Kolkata"),
		Locale:       firstNonEmpty(req.Locale, "en-IN"),
		Currency:     firstNonEmpty(req.Currency, "INR"),
		BillingEmail: req.BillingEmail,
		TaxID:        req.TaxID,
		Settings:     []byte("{}"),
		Features:     []byte("{}"),
		Metadata:     []byte("{}"),
	}
	now := time.Now()
	t.ActivatedAt = &now

	if err := s.repo.CreateTenant(ctx, t); err != nil {
		return nil, nil, apperr.New(apperr.CodeInternal, "create tenant failed", err)
	}
	defaultOrg := &Organization{
		TenantID:    t.ID,
		Slug:        slug,
		Name:        t.Name,
		DisplayName: t.DisplayName,
		Status:      "active",
		IsDefault:   true,
		Timezone:    t.Timezone,
		Locale:      t.Locale,
		Currency:    t.Currency,
		Settings:    []byte("{}"),
		Features:    []byte("{}"),
		Metadata:    []byte("{}"),
	}
	defaultOrg.ActivatedAt = &now
	if err := s.repo.CreateOrganization(ctx, defaultOrg); err != nil {
		return nil, nil, apperr.New(apperr.CodeInternal, "create default org failed", err)
	}
	return t, defaultOrg, nil
}

func (s *Service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	t, err := s.repo.GetTenantByID(ctx, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "tenant not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch tenant failed", err)
	}
	return t, nil
}

func (s *Service) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	t, err := s.repo.GetTenantBySlug(ctx, normaliseSlug(slug))
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "tenant not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch tenant failed", err)
	}
	return t, nil
}

func (s *Service) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error) {
	patch := map[string]interface{}{}
	setIfNotNil(patch, "name", req.Name)
	setIfNotNil(patch, "legal_name", req.LegalName)
	setIfNotNil(patch, "display_name", req.DisplayName)
	setIfNotNil(patch, "description", req.Description)
	setIfNotNil(patch, "logo_url", req.LogoURL)
	setIfNotNil(patch, "favicon_url", req.FaviconURL)
	setIfNotNil(patch, "primary_color", req.PrimaryColor)
	setIfNotNil(patch, "secondary_color", req.SecondaryColor)
	setIfNotNil(patch, "support_email", req.SupportEmail)
	setIfNotNil(patch, "support_phone", req.SupportPhone)
	setIfNotNil(patch, "website_url", req.WebsiteURL)
	setIfNotNil(patch, "country", req.Country)
	setIfNotNil(patch, "timezone", req.Timezone)
	setIfNotNil(patch, "locale", req.Locale)
	setIfNotNil(patch, "currency", req.Currency)
	setIfNotNil(patch, "billing_email", req.BillingEmail)
	setIfNotNil(patch, "billing_name", req.BillingName)
	setIfNotNil(patch, "tax_id", req.TaxID)
	if req.BillingAddress != nil {
		patch["billing_address"] = []byte(req.BillingAddress)
	}
	if req.Settings != nil {
		patch["settings"] = []byte(req.Settings)
	}
	if req.Features != nil {
		patch["features"] = []byte(req.Features)
	}
	if req.Metadata != nil {
		patch["metadata"] = []byte(req.Metadata)
	}
	if len(patch) == 0 {
		return s.GetTenant(ctx, id)
	}
	if err := s.repo.UpdateTenant(ctx, id, patch); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "tenant not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "update tenant failed", err)
	}
	return s.GetTenant(ctx, id)
}

// HardDeleteTenant releases the unique slug by physically deleting the row.
// Intended for failed-signup rollback only; downstream rows (organizations,
// memberships) are removed by ON DELETE CASCADE.
func (s *Service) HardDeleteTenant(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.HardDeleteTenant(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "hard delete tenant failed", err)
	}
	return nil
}

func (s *Service) ArchiveTenant(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.ArchiveTenant(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "tenant not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "archive tenant failed", err)
	}
	return nil
}

func (s *Service) ListTenants(ctx context.Context, filter ListFilter, p pagination.Params) ([]Tenant, int64, error) {
	rows, total, err := s.repo.ListTenants(ctx, filter, p)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list tenants failed", err)
	}
	return rows, total, nil
}

// ── organizations ──────────────────────────────────────────────────────────

func (s *Service) CreateOrganization(ctx context.Context, tenantID uuid.UUID, req CreateOrgRequest) (*Organization, error) {
	slug := normaliseSlug(req.Slug)
	if existing, err := s.repo.GetOrganizationBySlug(ctx, tenantID, slug); err == nil && existing != nil {
		return nil, apperr.New(apperr.CodeAlreadyExists, "organization slug already taken", nil)
	}
	o := &Organization{
		TenantID:     tenantID,
		Slug:         slug,
		Name:         strings.TrimSpace(req.Name),
		DisplayName:  firstNonEmpty(req.DisplayName, req.Name),
		Description:  req.Description,
		LogoURL:      req.LogoURL,
		WebsiteURL:   req.WebsiteURL,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Industry:     req.Industry,
		Size:         req.Size,
		Country:      req.Country,
		Timezone:     firstNonEmpty(req.Timezone, "Asia/Kolkata"),
		Locale:       firstNonEmpty(req.Locale, "en-IN"),
		Currency:     firstNonEmpty(req.Currency, "INR"),
		Status:       "active",
		IsDefault:    req.IsDefault,
		Settings:     []byte("{}"),
		Features:     []byte("{}"),
		Metadata:     []byte("{}"),
	}
	now := time.Now()
	o.ActivatedAt = &now
	if err := s.repo.CreateOrganization(ctx, o); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create organization failed", err)
	}
	return o, nil
}

func (s *Service) GetOrganization(ctx context.Context, tenantID, id uuid.UUID) (*Organization, error) {
	o, err := s.repo.GetOrganization(ctx, tenantID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "organization not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch organization failed", err)
	}
	return o, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, tenantID, id uuid.UUID, req UpdateOrgRequest) (*Organization, error) {
	patch := map[string]interface{}{}
	setIfNotNil(patch, "name", req.Name)
	setIfNotNil(patch, "display_name", req.DisplayName)
	setIfNotNil(patch, "description", req.Description)
	setIfNotNil(patch, "logo_url", req.LogoURL)
	setIfNotNil(patch, "cover_url", req.CoverURL)
	setIfNotNil(patch, "primary_color", req.PrimaryColor)
	setIfNotNil(patch, "secondary_color", req.SecondaryColor)
	setIfNotNil(patch, "website_url", req.WebsiteURL)
	setIfNotNil(patch, "contact_email", req.ContactEmail)
	setIfNotNil(patch, "contact_phone", req.ContactPhone)
	setIfNotNil(patch, "industry", req.Industry)
	setIfNotNil(patch, "size", req.Size)
	setIfNotNil(patch, "country", req.Country)
	setIfNotNil(patch, "state", req.State)
	setIfNotNil(patch, "city", req.City)
	setIfNotNil(patch, "postal_code", req.PostalCode)
	setIfNotNil(patch, "timezone", req.Timezone)
	setIfNotNil(patch, "locale", req.Locale)
	setIfNotNil(patch, "currency", req.Currency)
	if req.Address != nil {
		patch["address"] = []byte(req.Address)
	}
	if req.Settings != nil {
		patch["settings"] = []byte(req.Settings)
	}
	if req.Features != nil {
		patch["features"] = []byte(req.Features)
	}
	if req.Metadata != nil {
		patch["metadata"] = []byte(req.Metadata)
	}
	if len(patch) == 0 {
		return s.GetOrganization(ctx, tenantID, id)
	}
	if err := s.repo.UpdateOrganization(ctx, tenantID, id, patch); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "organization not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "update organization failed", err)
	}
	return s.GetOrganization(ctx, tenantID, id)
}

func (s *Service) ArchiveOrganization(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := s.repo.ArchiveOrganization(ctx, tenantID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "organization not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "archive organization failed", err)
	}
	return nil
}

func (s *Service) ListOrganizations(ctx context.Context, tenantID uuid.UUID, filter ListFilter, p pagination.Params) ([]Organization, int64, error) {
	rows, total, err := s.repo.ListOrganizations(ctx, tenantID, filter, p)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list organizations failed", err)
	}
	return rows, total, nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func normaliseSlug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return s
}

func firstNonEmpty(parts ...string) string {
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			return p
		}
	}
	return ""
}

func setIfNotNil(m map[string]interface{}, col string, v *string) {
	if v != nil {
		m[col] = *v
	}
}
