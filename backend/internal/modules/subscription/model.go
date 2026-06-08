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
