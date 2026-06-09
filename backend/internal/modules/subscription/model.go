package subscription

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

type Plan struct {
	model.Base
	Code           string      `gorm:"not null;uniqueIndex"          json:"code"`
	Name           string      `gorm:"not null"                       json:"name"`
	Description    string      `                                       json:"description,omitempty"`
	Tagline        string      `                                       json:"tagline,omitempty"`
	Tier           int         `gorm:"not null;default:0"            json:"tier"`
	BillingCycle   string      `gorm:"not null;default:monthly"      json:"billingCycle"`
	PriceCents     int64       `gorm:"not null;default:0"            json:"priceCents"`
	Currency       string      `gorm:"not null;default:INR"          json:"currency"`
	TrialDays      int         `gorm:"not null;default:0"            json:"trialDays"`
	IsActive       bool        `gorm:"not null;default:true"         json:"isActive"`
	IsDefault      bool        `gorm:"not null;default:false"        json:"isDefault"`
	IsPublic       bool        `gorm:"not null;default:true"         json:"isPublic"`
	IsAddon        bool        `gorm:"not null;default:false"        json:"isAddon"`
	Features       model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb" json:"features"`
	Limits         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"limits"`
	Metadata       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
	EffectiveFrom  *time.Time  `                                       json:"effectiveFrom,omitempty"`
	EffectiveUntil *time.Time  `                                       json:"effectiveUntil,omitempty"`
	Gateway        string      `                                       json:"gateway,omitempty"`
	ExternalRef    string      `                                       json:"externalRef,omitempty"`
}

func (Plan) TableName() string { return "subscription_plans" }

type Subscription struct {
	model.Base
	TenantID              uuid.UUID   `gorm:"type:uuid;not null;index"      json:"tenantId"`
	OrganizationID        uuid.UUID   `gorm:"type:uuid;not null;index"      json:"organizationId"`
	PlanID                uuid.UUID   `gorm:"type:uuid;not null"             json:"planId"`
	PlanCode              string      `gorm:"not null"                        json:"planCode"`
	Status                string      `gorm:"not null;default:trial"         json:"status"`
	BillingCycle          string      `gorm:"not null;default:monthly"       json:"billingCycle"`
	Quantity              int         `gorm:"not null;default:1"             json:"quantity"`
	UnitPriceCents        int64       `gorm:"not null;default:0"             json:"unitPriceCents"`
	DiscountCents         int64       `gorm:"not null;default:0"             json:"discountCents"`
	TaxCents              int64       `gorm:"not null;default:0"             json:"taxCents"`
	TotalCents            int64       `gorm:"not null;default:0"             json:"totalCents"`
	Currency              string      `gorm:"not null;default:INR"           json:"currency"`
	StartedAt             time.Time   `gorm:"not null"                        json:"startedAt"`
	TrialStartedAt        *time.Time  `                                        json:"trialStartedAt,omitempty"`
	TrialEndsAt           *time.Time  `                                        json:"trialEndsAt,omitempty"`
	CurrentPeriodStart    *time.Time  `                                        json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd      *time.Time  `                                        json:"currentPeriodEnd,omitempty"`
	NextBillingAt         *time.Time  `                                        json:"nextBillingAt,omitempty"`
	LastBilledAt          *time.Time  `                                        json:"lastBilledAt,omitempty"`
	CancelAt              *time.Time  `                                        json:"cancelAt,omitempty"`
	CancelledAt           *time.Time  `                                        json:"cancelledAt,omitempty"`
	CancelReason          string      `                                        json:"cancelReason,omitempty"`
	CancelImmediate       bool        `gorm:"column:cancel_immediate"        json:"cancelImmediate"`
	EndedAt               *time.Time  `                                        json:"endedAt,omitempty"`
	PauseAt               *time.Time  `                                        json:"pauseAt,omitempty"`
	PausedAt              *time.Time  `                                        json:"pausedAt,omitempty"`
	ResumeAt              *time.Time  `                                        json:"resumeAt,omitempty"`
	Gateway               string      `                                        json:"gateway,omitempty"`
	GatewayCustomerID     string      `                                        json:"gatewayCustomerId,omitempty"`
	GatewaySubscriptionID string      `                                        json:"gatewaySubscriptionId,omitempty"`
	ExternalRef           string      `                                        json:"externalRef,omitempty"`
	CouponCode            string      `                                        json:"couponCode,omitempty"`
	BillingEmail          string      `gorm:"type:citext"                    json:"billingEmail,omitempty"`
	BillingName           string      `                                        json:"billingName,omitempty"`
	BillingAddress        model.JSONB `gorm:"type:jsonb"                      json:"billingAddress,omitempty"`
	Features              model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb"  json:"features"`
	Limits                model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"  json:"limits"`
	Metadata              model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"  json:"metadata"`
	Notes                 string      `                                        json:"notes,omitempty"`
}

func (Subscription) TableName() string { return "subscriptions" }

type UsageCounter struct {
	model.Base
	TenantID       uuid.UUID   `gorm:"type:uuid;not null;index" json:"tenantId"`
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null;index" json:"organizationId"`
	SubscriptionID *uuid.UUID  `gorm:"type:uuid"                 json:"subscriptionId,omitempty"`
	Key            string      `gorm:"not null;index"             json:"key"`
	Count          int64       `gorm:"not null;default:0"        json:"count"`
	LimitValue     *int64      `gorm:"column:limit_value"        json:"limitValue,omitempty"`
	PeriodStart    time.Time   `gorm:"not null"                   json:"periodStart"`
	PeriodEnd      time.Time   `gorm:"not null"                   json:"periodEnd"`
	LastResetAt    *time.Time  `                                   json:"lastResetAt,omitempty"`
	Metadata       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (UsageCounter) TableName() string { return "usage_counters" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type ChangePlanRequest struct {
	PlanCode        string `json:"planCode" binding:"required"`
	BillingCycle    string `json:"billingCycle,omitempty"`
	Quantity        int    `json:"quantity,omitempty"`
	StartImmediately bool  `json:"startImmediately,omitempty"`
	CouponCode      string `json:"couponCode,omitempty"`
}

type CancelRequest struct {
	Reason    string `json:"reason,omitempty"`
	Immediate bool   `json:"immediate,omitempty"`
}

// FeatureSet is the cached, resolved snapshot for an org. Keys are feature
// strings (e.g. "export.csv"); limits maps a quota key to its numeric ceiling
// (-1 means unlimited).
type FeatureSet struct {
	PlanCode string           `json:"planCode"`
	Status   string           `json:"status"`
	Features map[string]bool  `json:"features"`
	Limits   map[string]int64 `json:"limits"`
}

// ── invoices ───────────────────────────────────────────────────────────────

type Invoice struct {
	model.Base
	TenantID         uuid.UUID   `gorm:"type:uuid;not null;index"      json:"tenantId"`
	OrganizationID   uuid.UUID   `gorm:"type:uuid;not null;index"      json:"organizationId"`
	SubscriptionID   *uuid.UUID  `gorm:"type:uuid"                      json:"subscriptionId,omitempty"`
	Number           string      `gorm:"not null;uniqueIndex"          json:"number"`
	Status           string      `gorm:"not null;default:'open'"       json:"status"`
	Currency         string      `gorm:"not null;default:'INR'"        json:"currency"`
	SubtotalCents    int64       `gorm:"not null;default:0"            json:"subtotalCents"`
	DiscountCents    int64       `gorm:"not null;default:0"            json:"discountCents"`
	TaxCents         int64       `gorm:"not null;default:0"            json:"taxCents"`
	TotalCents       int64       `gorm:"not null;default:0"            json:"totalCents"`
	AmountDueCents   int64       `gorm:"not null;default:0"            json:"amountDueCents"`
	AmountPaidCents  int64       `gorm:"not null;default:0"            json:"amountPaidCents"`
	CouponCode       string      `                                       json:"couponCode,omitempty"`
	Description      string      `                                       json:"description,omitempty"`
	LineItems        model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb" json:"lineItems"`
	PeriodStart      *time.Time  `                                       json:"periodStart,omitempty"`
	PeriodEnd        *time.Time  `                                       json:"periodEnd,omitempty"`
	IssuedAt         time.Time   `gorm:"not null;default:now()"        json:"issuedAt"`
	DueAt            *time.Time  `                                       json:"dueAt,omitempty"`
	PaidAt           *time.Time  `                                       json:"paidAt,omitempty"`
	VoidedAt         *time.Time  `                                       json:"voidedAt,omitempty"`
	Gateway          string      `                                       json:"gateway,omitempty"`
	GatewayInvoiceID string      `                                       json:"gatewayInvoiceId,omitempty"`
	Metadata         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (Invoice) TableName() string { return "subscription_invoices" }

// LineItem is the shape we serialize into Invoice.LineItems.
type LineItem struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitCents   int64  `json:"unitCents"`
	AmountCents int64  `json:"amountCents"`
}

// ── coupons ────────────────────────────────────────────────────────────────

type Coupon struct {
	model.Base
	Code            string      `gorm:"type:citext;not null;uniqueIndex" json:"code"`
	Name            string      `gorm:"not null"                          json:"name"`
	Description     string      `                                          json:"description,omitempty"`
	PercentOff      *int        `gorm:"column:percent_off"               json:"percentOff,omitempty"`
	AmountOffCents  *int64      `gorm:"column:amount_off_cents"          json:"amountOffCents,omitempty"`
	Currency        string      `                                          json:"currency,omitempty"`
	Duration        string      `gorm:"not null;default:'once'"          json:"duration"`
	DurationMonths  *int        `gorm:"column:duration_months"           json:"durationMonths,omitempty"`
	MaxRedemptions  *int        `gorm:"column:max_redemptions"           json:"maxRedemptions,omitempty"`
	Redemptions     int         `gorm:"not null;default:0"               json:"redemptions"`
	ValidFrom       *time.Time  `                                          json:"validFrom,omitempty"`
	ValidUntil      *time.Time  `                                          json:"validUntil,omitempty"`
	AppliesToPlans  model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb"   json:"appliesToPlans"`
	IsActive        bool        `gorm:"not null;default:true"            json:"isActive"`
	Metadata        model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"metadata"`
}

func (Coupon) TableName() string { return "subscription_coupons" }

type CouponRedemption struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CouponID       uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"couponId"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"organizationId"`
	SubscriptionID *uuid.UUID `gorm:"type:uuid"                                       json:"subscriptionId,omitempty"`
	InvoiceID      *uuid.UUID `gorm:"type:uuid"                                       json:"invoiceId,omitempty"`
	AmountOffCents int64      `gorm:"column:amount_off_cents;not null;default:0"     json:"amountOffCents"`
	RedeemedAt     time.Time  `gorm:"not null;default:now()"                         json:"redeemedAt"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid"                                       json:"createdBy,omitempty"`
}

func (CouponRedemption) TableName() string { return "coupon_redemptions" }

// ── DTOs (lifecycle) ──────────────────────────────────────────────────────

type PauseRequest struct {
	ResumeAt *time.Time `json:"resumeAt,omitempty"`
	Reason   string     `json:"reason,omitempty"`
}

type UpdateBillingRequest struct {
	BillingEmail   *string                `json:"billingEmail,omitempty"`
	BillingName    *string                `json:"billingName,omitempty"`
	BillingAddress map[string]interface{} `json:"billingAddress,omitempty"`
}

type PreviewChangeRequest struct {
	PlanCode     string `json:"planCode" binding:"required"`
	BillingCycle string `json:"billingCycle,omitempty"`
	Quantity     int    `json:"quantity,omitempty"`
	CouponCode   string `json:"couponCode,omitempty"`
}

// PreviewChangeResponse is what the UI renders before the user confirms a
// plan switch. Amounts are minor units (cents/paise); UI formats.
type PreviewChangeResponse struct {
	FromPlanCode        string `json:"fromPlanCode"`
	ToPlanCode          string `json:"toPlanCode"`
	BillingCycle        string `json:"billingCycle"`
	Currency            string `json:"currency"`
	BaseAmountCents     int64  `json:"baseAmountCents"`
	ProrationCents      int64  `json:"prorationCents"` // credit (-) or charge (+)
	CouponCode          string `json:"couponCode,omitempty"`
	DiscountCents       int64  `json:"discountCents"`
	TaxCents            int64  `json:"taxCents"`
	TotalDueCents       int64  `json:"totalDueCents"`
	EffectiveAt         string `json:"effectiveAt"`
	IsUpgrade           bool   `json:"isUpgrade"`
	UnusedDaysRemaining int    `json:"unusedDaysRemaining"`
}

type ValidateCouponRequest struct {
	Code     string `json:"code" binding:"required"`
	PlanCode string `json:"planCode,omitempty"`
}

type ValidateCouponResponse struct {
	Valid          bool   `json:"valid"`
	Reason         string `json:"reason,omitempty"`
	Code           string `json:"code,omitempty"`
	Name           string `json:"name,omitempty"`
	PercentOff     *int   `json:"percentOff,omitempty"`
	AmountOffCents *int64 `json:"amountOffCents,omitempty"`
	Currency       string `json:"currency,omitempty"`
	Duration       string `json:"duration,omitempty"`
}
