package tenant

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

type Tenant struct {
	model.Base
	Slug             string      `gorm:"type:citext;not null;uniqueIndex" json:"slug"`
	Name             string      `gorm:"not null"                          json:"name"`
	LegalName        string      `                                          json:"legalName,omitempty"`
	DisplayName      string      `                                          json:"displayName,omitempty"`
	Description      string      `                                          json:"description,omitempty"`
	LogoURL          string      `gorm:"column:logo_url"                   json:"logoUrl,omitempty"`
	FaviconURL       string      `gorm:"column:favicon_url"                json:"faviconUrl,omitempty"`
	PrimaryColor     string      `                                          json:"primaryColor,omitempty"`
	SecondaryColor   string      `                                          json:"secondaryColor,omitempty"`
	SupportEmail     string      `gorm:"type:citext"                       json:"supportEmail,omitempty"`
	SupportPhone     string      `                                          json:"supportPhone,omitempty"`
	WebsiteURL       string      `gorm:"column:website_url"                json:"websiteUrl,omitempty"`
	Status           string      `gorm:"not null;default:active"           json:"status"`
	PlanCode         string      `                                          json:"planCode,omitempty"`
	SeatLimit        *int        `                                          json:"seatLimit,omitempty"`
	Country          string      `                                          json:"country,omitempty"`
	Timezone         string      `gorm:"not null;default:Asia/Kolkata"     json:"timezone"`
	Locale           string      `gorm:"not null;default:en-IN"            json:"locale"`
	Currency         string      `gorm:"not null;default:INR"              json:"currency"`
	BillingEmail     string      `gorm:"type:citext"                       json:"billingEmail,omitempty"`
	BillingName      string      `                                          json:"billingName,omitempty"`
	BillingAddress   model.JSONB `gorm:"type:jsonb"                         json:"billingAddress,omitempty"`
	TaxID            string      `gorm:"column:tax_id"                     json:"taxId,omitempty"`
	Settings         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"    json:"settings"`
	Features         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"    json:"features"`
	Metadata         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"    json:"metadata"`
	TrialEndsAt      *time.Time  `                                          json:"trialEndsAt,omitempty"`
	ActivatedAt      *time.Time  `                                          json:"activatedAt,omitempty"`
	SuspendedAt      *time.Time  `                                          json:"suspendedAt,omitempty"`
	SuspensionReason string      `                                          json:"suspensionReason,omitempty"`
	ArchivedAt       *time.Time  `                                          json:"archivedAt,omitempty"`
}

func (Tenant) TableName() string { return "tenants" }

type Organization struct {
	model.Base
	TenantID       uuid.UUID   `gorm:"type:uuid;not null;index"        json:"tenantId"`
	Slug           string      `gorm:"type:citext;not null"            json:"slug"`
	Name           string      `gorm:"not null"                         json:"name"`
	DisplayName    string      `                                         json:"displayName,omitempty"`
	Description    string      `                                         json:"description,omitempty"`
	LogoURL        string      `gorm:"column:logo_url"                  json:"logoUrl,omitempty"`
	CoverURL       string      `gorm:"column:cover_url"                 json:"coverUrl,omitempty"`
	PrimaryColor   string      `                                         json:"primaryColor,omitempty"`
	SecondaryColor string      `                                         json:"secondaryColor,omitempty"`
	WebsiteURL     string      `gorm:"column:website_url"               json:"websiteUrl,omitempty"`
	ContactEmail   string      `gorm:"type:citext"                      json:"contactEmail,omitempty"`
	ContactPhone   string      `                                         json:"contactPhone,omitempty"`
	Industry       string      `                                         json:"industry,omitempty"`
	Size           string      `                                         json:"size,omitempty"`
	Country        string      `                                         json:"country,omitempty"`
	State          string      `                                         json:"state,omitempty"`
	City           string      `                                         json:"city,omitempty"`
	PostalCode     string      `                                         json:"postalCode,omitempty"`
	Timezone       string      `gorm:"not null;default:Asia/Kolkata"    json:"timezone"`
	Locale         string      `gorm:"not null;default:en-IN"           json:"locale"`
	Currency       string      `gorm:"not null;default:INR"             json:"currency"`
	Status         string      `gorm:"not null;default:active"          json:"status"`
	IsDefault      bool        `                                         json:"isDefault"`
	Settings       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"settings"`
	Features       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"features"`
	Metadata       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"metadata"`
	Address        model.JSONB `gorm:"type:jsonb"                        json:"address,omitempty"`
	ActivatedAt    *time.Time  `                                         json:"activatedAt,omitempty"`
	SuspendedAt    *time.Time  `                                         json:"suspendedAt,omitempty"`
	ArchivedAt     *time.Time  `                                         json:"archivedAt,omitempty"`
}

func (Organization) TableName() string { return "organizations" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type CreateTenantRequest struct {
	Slug           string `json:"slug" binding:"required,min=2,max=64"`
	Name           string `json:"name" binding:"required,min=1,max=200"`
	LegalName      string `json:"legalName,omitempty"`
	DisplayName    string `json:"displayName,omitempty"`
	Description    string `json:"description,omitempty"`
	LogoURL        string `json:"logoUrl,omitempty"`
	PrimaryColor   string `json:"primaryColor,omitempty"`
	SupportEmail   string `json:"supportEmail,omitempty" binding:"omitempty,email"`
	SupportPhone   string `json:"supportPhone,omitempty"`
	WebsiteURL     string `json:"websiteUrl,omitempty"`
	PlanCode       string `json:"planCode,omitempty"`
	Country        string `json:"country,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	Locale         string `json:"locale,omitempty"`
	Currency       string `json:"currency,omitempty"`
	BillingEmail   string `json:"billingEmail,omitempty" binding:"omitempty,email"`
	TaxID          string `json:"taxId,omitempty"`
	// Initial admin user (will receive an invite).
	AdminEmail     string `json:"adminEmail" binding:"required,email"`
	AdminFirstName string `json:"adminFirstName,omitempty"`
	AdminLastName  string `json:"adminLastName,omitempty"`
}

type UpdateTenantRequest struct {
	Name           *string     `json:"name,omitempty"           binding:"omitempty,min=1,max=200"`
	LegalName      *string     `json:"legalName,omitempty"      binding:"omitempty,max=200"`
	DisplayName    *string     `json:"displayName,omitempty"    binding:"omitempty,max=200"`
	Description    *string     `json:"description,omitempty"    binding:"omitempty,max=2000"`
	LogoURL        *string     `json:"logoUrl,omitempty"        binding:"omitempty,url,max=2048"`
	FaviconURL     *string     `json:"faviconUrl,omitempty"     binding:"omitempty,url,max=2048"`
	PrimaryColor   *string     `json:"primaryColor,omitempty"   binding:"omitempty,hexcolor"`
	SecondaryColor *string     `json:"secondaryColor,omitempty" binding:"omitempty,hexcolor"`
	SupportEmail   *string     `json:"supportEmail,omitempty"   binding:"omitempty,email,max=254"`
	SupportPhone   *string     `json:"supportPhone,omitempty"   binding:"omitempty,max=32"`
	WebsiteURL     *string     `json:"websiteUrl,omitempty"     binding:"omitempty,url,max=2048"`
	Country        *string     `json:"country,omitempty"        binding:"omitempty,max=64"`
	Timezone       *string     `json:"timezone,omitempty"       binding:"omitempty,max=64"`
	Locale         *string     `json:"locale,omitempty"         binding:"omitempty,max=16"`
	Currency       *string     `json:"currency,omitempty"       binding:"omitempty,max=8"`
	BillingEmail   *string     `json:"billingEmail,omitempty"   binding:"omitempty,email,max=254"`
	BillingName    *string     `json:"billingName,omitempty"    binding:"omitempty,max=200"`
	BillingAddress model.JSONB `json:"billingAddress,omitempty"`
	TaxID          *string     `json:"taxId,omitempty"          binding:"omitempty,max=64"`
	Settings       model.JSONB `json:"settings,omitempty"`
	Features       model.JSONB `json:"features,omitempty"`
	Metadata       model.JSONB `json:"metadata,omitempty"`
}

type TenantPublicResponse struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName,omitempty"`
	LogoURL     string    `json:"logoUrl,omitempty"`
	PrimaryColor string   `json:"primaryColor,omitempty"`
}

func PublicResponse(t *Tenant) TenantPublicResponse {
	return TenantPublicResponse{
		ID:           t.ID,
		Slug:         t.Slug,
		Name:         t.Name,
		DisplayName:  t.DisplayName,
		LogoURL:      t.LogoURL,
		PrimaryColor: t.PrimaryColor,
	}
}

type CreateOrgRequest struct {
	Slug         string `json:"slug" binding:"required,min=2,max=64"`
	Name         string `json:"name" binding:"required,min=1,max=200"`
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
	LogoURL      string `json:"logoUrl,omitempty"`
	WebsiteURL   string `json:"websiteUrl,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty" binding:"omitempty,email"`
	ContactPhone string `json:"contactPhone,omitempty"`
	Industry     string `json:"industry,omitempty"`
	Size         string `json:"size,omitempty"`
	Country      string `json:"country,omitempty"`
	Timezone     string `json:"timezone,omitempty"`
	Locale       string `json:"locale,omitempty"`
	Currency     string `json:"currency,omitempty"`
	IsDefault    bool   `json:"isDefault,omitempty"`
}

type UpdateOrgRequest struct {
	Name           *string     `json:"name,omitempty"           binding:"omitempty,min=1,max=200"`
	DisplayName    *string     `json:"displayName,omitempty"    binding:"omitempty,max=200"`
	Description    *string     `json:"description,omitempty"    binding:"omitempty,max=2000"`
	LogoURL        *string     `json:"logoUrl,omitempty"        binding:"omitempty,url,max=2048"`
	CoverURL       *string     `json:"coverUrl,omitempty"       binding:"omitempty,url,max=2048"`
	PrimaryColor   *string     `json:"primaryColor,omitempty"   binding:"omitempty,hexcolor"`
	SecondaryColor *string     `json:"secondaryColor,omitempty" binding:"omitempty,hexcolor"`
	WebsiteURL     *string     `json:"websiteUrl,omitempty"     binding:"omitempty,url,max=2048"`
	ContactEmail   *string     `json:"contactEmail,omitempty"   binding:"omitempty,email,max=254"`
	ContactPhone   *string     `json:"contactPhone,omitempty"   binding:"omitempty,max=32"`
	Industry       *string     `json:"industry,omitempty"       binding:"omitempty,max=64"`
	Size           *string     `json:"size,omitempty"           binding:"omitempty,max=32"`
	Country        *string     `json:"country,omitempty"        binding:"omitempty,max=64"`
	State          *string     `json:"state,omitempty"          binding:"omitempty,max=64"`
	City           *string     `json:"city,omitempty"           binding:"omitempty,max=64"`
	PostalCode     *string     `json:"postalCode,omitempty"     binding:"omitempty,max=32"`
	Timezone       *string     `json:"timezone,omitempty"       binding:"omitempty,max=64"`
	Locale         *string     `json:"locale,omitempty"         binding:"omitempty,max=16"`
	Currency       *string     `json:"currency,omitempty"       binding:"omitempty,max=8"`
	Address        model.JSONB `json:"address,omitempty"`
	Settings       model.JSONB `json:"settings,omitempty"`
	Features       model.JSONB `json:"features,omitempty"`
	Metadata       model.JSONB `json:"metadata,omitempty"`
}

type ListFilter struct {
	Status string
	Search string
}
