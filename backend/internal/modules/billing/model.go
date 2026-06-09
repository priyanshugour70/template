package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

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
	IsCustom       bool        `gorm:"not null;default:false"        json:"isCustom"`
}

func (Plan) TableName() string { return "billing_plans" }

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
	BillingState          string      `                                        json:"billingState,omitempty"`
	Features              model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb"  json:"features"`
	Limits                model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"  json:"limits"`
	Metadata              model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"  json:"metadata"`
	Notes                 string      `                                        json:"notes,omitempty"`
}

func (Subscription) TableName() string { return "billing_subscriptions" }

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

func (UsageCounter) TableName() string { return "billing_usage_counters" }

// ── DTOs ───────────────────────────────────────────────────────────────────

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
	HSNSAC           string      `gorm:"column:hsn_sac;not null;default:'998314'" json:"hsnSac"`
	PlaceOfSupply    string      `                                       json:"placeOfSupply,omitempty"`
	CGSTCents        int64       `gorm:"column:cgst_cents;not null;default:0" json:"cgstCents"`
	SGSTCents        int64       `gorm:"column:sgst_cents;not null;default:0" json:"sgstCents"`
	IGSTCents        int64       `gorm:"column:igst_cents;not null;default:0" json:"igstCents"`
	PDFStorageKey    string      `gorm:"column:pdf_storage_key"          json:"pdfStorageKey,omitempty"`
	Metadata         model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (Invoice) TableName() string { return "billing_invoices" }

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

func (Coupon) TableName() string { return "billing_coupons" }

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

func (CouponRedemption) TableName() string { return "billing_coupon_redemptions" }

// ── Transactions (payments) ───────────────────────────────────────────────

// Transaction maps to billing_transactions — one row per payment captured
// against an invoice. Many transactions can attach to one invoice (partial
// payments, retries); Phase 5 only emits one per invoice.
type Transaction struct {
	model.Base
	TenantID             uuid.UUID   `gorm:"type:uuid;not null;index"           json:"tenantId"`
	OrganizationID       uuid.UUID   `gorm:"type:uuid;not null;index"           json:"organizationId"`
	InvoiceID            uuid.UUID   `gorm:"type:uuid;not null;index"           json:"invoiceId"`
	ReceiptNumber        string      `gorm:"not null;uniqueIndex"               json:"receiptNumber"`
	Method               string      `gorm:"not null"                            json:"method"`
	Status               string      `gorm:"not null;default:recorded"          json:"status"`
	AmountCents          int64       `gorm:"not null"                            json:"amountCents"`
	Currency             string      `gorm:"not null;default:INR"                json:"currency"`
	Reference            string      `                                            json:"reference,omitempty"`
	Gateway              string      `                                            json:"gateway,omitempty"`
	GatewayTransactionID string      `gorm:"column:gateway_transaction_id"      json:"gatewayTransactionId,omitempty"`
	PaidAt               time.Time   `gorm:"not null;default:now()"             json:"paidAt"`
	RefundedAt           *time.Time  `                                            json:"refundedAt,omitempty"`
	RefundAmountCents    int64       `gorm:"not null;default:0"                 json:"refundAmountCents"`
	PDFStorageKey        string      `gorm:"column:pdf_storage_key"             json:"pdfStorageKey,omitempty"`
	Notes                string      `                                            json:"notes,omitempty"`
	Metadata             model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"     json:"metadata,omitempty"`
}

func (Transaction) TableName() string { return "billing_transactions" }

// RecordPaymentRequest is the body of POST /billing/invoices/:id/pay.
type RecordPaymentRequest struct {
	Method      string `json:"method"      binding:"required,oneof=cash bank_transfer cheque gateway"`
	AmountCents int64  `json:"amountCents" binding:"required,min=1"`
	Reference   string `json:"reference,omitempty" binding:"omitempty,max=200"`
	Notes       string `json:"notes,omitempty"     binding:"omitempty,max=2000"`
}

// RecordPaymentResponse is what /pay returns — enough for the UI to redirect
// the user to the new transaction / receipt.
type RecordPaymentResponse struct {
	Transaction  Transaction   `json:"transaction"`
	Invoice      Invoice       `json:"invoice"`
	Subscription *Subscription `json:"subscription,omitempty"`
	ReceiptURL   string        `json:"receiptUrl,omitempty"`
}

// ── Feature catalog ───────────────────────────────────────────────────────

// Feature is one row of billing_features — the catalog of capabilities a
// customer can pick from when building a custom plan.
type Feature struct {
	model.Base
	Key                string         `gorm:"uniqueIndex;not null"           json:"key"`
	Name               string         `gorm:"not null"                       json:"name"`
	Description        string         `                                       json:"description"`
	Category           string         `gorm:"not null"                       json:"category"` // core|admin|compliance|integrations|limits
	BasePriceCents     int64          `gorm:"not null;default:0"             json:"basePriceCents"`
	PerUserPriceCents  int64          `gorm:"not null;default:0"             json:"perUserPriceCents"`
	IncludedUsers      int            `gorm:"not null;default:0"             json:"includedUsers"`
	IsCore             bool           `gorm:"not null;default:false"         json:"isCore"`
	IsStarterDefault   bool           `gorm:"not null;default:false"         json:"isStarterDefault"`
	IsActive           bool           `gorm:"not null;default:true"          json:"isActive"`
	Requires           pq.StringArray `gorm:"type:text[]"                    json:"requires"`
	Metadata           model.JSONB    `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata,omitempty"`
	SortOrder          int            `gorm:"not null;default:0"             json:"sortOrder"`
}

func (Feature) TableName() string { return "billing_features" }

// TaxConfig is the singleton row in billing_tax_config that holds the
// company's GSTIN + home state + default GST rates + bank details for invoices.
type TaxConfig struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Singleton         bool      `gorm:"not null;default:true"                          json:"-"`
	CompanyName       string    `gorm:"not null;default:''"                            json:"companyName"`
	CompanyAddress    string    `gorm:"not null;default:''"                            json:"companyAddress"`
	GSTIN             string    `gorm:"not null;default:''"                            json:"gstin"`
	HomeState         string    `gorm:"not null;default:Karnataka"                     json:"homeState"`
	DefaultCGSTPct    float64   `gorm:"column:default_cgst_pct;type:numeric(5,2)"      json:"defaultCgstPct"`
	DefaultSGSTPct    float64   `gorm:"column:default_sgst_pct;type:numeric(5,2)"      json:"defaultSgstPct"`
	DefaultIGSTPct    float64   `gorm:"column:default_igst_pct;type:numeric(5,2)"      json:"defaultIgstPct"`
	DefaultHSNSAC     string    `gorm:"column:default_hsn_sac;not null;default:'998314'" json:"defaultHsnSac"`
	Currency          string    `gorm:"not null;default:INR"                           json:"currency"`
	BankName          string    `gorm:"not null;default:''"                            json:"bankName"`
	BankAccountNumber string    `gorm:"not null;default:''"                            json:"bankAccountNumber"`
	BankIFSC          string    `gorm:"column:bank_ifsc;not null;default:''"           json:"bankIfsc"`
	BankAccountName   string    `gorm:"not null;default:''"                            json:"bankAccountName"`
	InvoiceTerms      string    `gorm:"not null;default:''"                            json:"invoiceTerms"`
	Metadata          model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"               json:"metadata,omitempty"`
	CreatedAt         time.Time `gorm:"not null;default:now()"                         json:"createdAt"`
	UpdatedAt         time.Time `gorm:"not null;default:now()"                         json:"updatedAt"`
}

func (TaxConfig) TableName() string { return "billing_tax_config" }

// ── Quotation preview DTO ─────────────────────────────────────────────────

// PreviewQuoteRequest is the input to POST /billing/quotations/preview. Same
// shape gets re-used by POST /billing/quotations when we persist the draft in
// Phase 3.
type PreviewQuoteRequest struct {
	FeatureKeys   []string `json:"featureKeys"             binding:"required"`
	UserCount     int      `json:"userCount"               binding:"required,min=1"`
	// CustomerState drives the CGST/SGST vs IGST decision. Optional — if
	// blank, falls back to billing_tax_config.home_state (intra-state default).
	CustomerState string   `json:"customerState,omitempty" binding:"omitempty,max=64"`
}

// ── Quotation (draft plan) ────────────────────────────────────────────────

// Quotation maps to billing_quotations. A draft persists the user's feature
// selection and the resolved pricing snapshot. Activating a draft mints a Plan
// + Subscription + first Invoice in one transaction.
type Quotation struct {
	model.Base
	TenantID                uuid.UUID      `gorm:"type:uuid;not null"        json:"tenantId"`
	OrganizationID          uuid.UUID      `gorm:"type:uuid;not null"        json:"organizationId"`
	Number                  string         `gorm:"not null;uniqueIndex"      json:"number"`
	Status                  string         `gorm:"not null;default:draft"    json:"status"` // draft|accepted|rejected|expired
	FeatureKeys             pq.StringArray `gorm:"type:text[]"               json:"featureKeys"`
	UserCount               int            `gorm:"not null;default:1"        json:"userCount"`
	SubtotalCents           int64          `gorm:"not null;default:0"        json:"subtotalCents"`
	DiscountCents           int64          `gorm:"not null;default:0"        json:"discountCents"`
	CGSTCents               int64          `gorm:"column:cgst_cents;not null;default:0" json:"cgstCents"`
	SGSTCents               int64          `gorm:"column:sgst_cents;not null;default:0" json:"sgstCents"`
	IGSTCents               int64          `gorm:"column:igst_cents;not null;default:0" json:"igstCents"`
	TotalCents              int64          `gorm:"not null;default:0"        json:"totalCents"`
	Currency                string         `gorm:"not null;default:INR"      json:"currency"`
	PlaceOfSupply           string         `                                  json:"placeOfSupply,omitempty"`
	LineItems               model.JSONB    `gorm:"type:jsonb;default:'[]'::jsonb" json:"lineItems"`
	BillingEmail            string         `gorm:"type:citext"                json:"billingEmail,omitempty"`
	BillingName             string         `                                  json:"billingName,omitempty"`
	BillingAddress          model.JSONB    `gorm:"type:jsonb"                 json:"billingAddress,omitempty"`
	BillingState            string         `                                  json:"billingState,omitempty"`
	Notes                   string         `                                  json:"notes,omitempty"`
	ExpiresAt               time.Time      `gorm:"not null"                  json:"expiresAt"`
	AcceptedAt              *time.Time     `                                  json:"acceptedAt,omitempty"`
	RejectedAt              *time.Time     `                                  json:"rejectedAt,omitempty"`
	ActivatedPlanID         *uuid.UUID     `gorm:"type:uuid"                  json:"activatedPlanId,omitempty"`
	ActivatedSubscriptionID *uuid.UUID     `gorm:"type:uuid"                  json:"activatedSubscriptionId,omitempty"`
	Metadata                model.JSONB    `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata,omitempty"`
}

func (Quotation) TableName() string { return "billing_quotations" }

// PlanFeature maps to billing_plan_features — the price snapshot for each
// feature attached to a plan. Snapshotting prevents catalog price changes
// from retroactively re-billing existing customers.
type PlanFeature struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlanID            uuid.UUID `gorm:"type:uuid;not null"                              json:"planId"`
	FeatureID         uuid.UUID `gorm:"type:uuid;not null"                              json:"featureId"`
	FeatureKey        string    `gorm:"not null"                                        json:"featureKey"`
	BasePriceCents    int64     `gorm:"not null;default:0"                              json:"basePriceCents"`
	PerUserPriceCents int64     `gorm:"not null;default:0"                              json:"perUserPriceCents"`
	IncludedUsers     int       `gorm:"not null;default:0"                              json:"includedUsers"`
	Quantity          int       `gorm:"not null;default:1"                              json:"quantity"`
	CreatedAt         time.Time `gorm:"not null;default:now()"                          json:"createdAt"`
	UpdatedAt         time.Time `gorm:"not null;default:now()"                          json:"updatedAt"`
}

func (PlanFeature) TableName() string { return "billing_plan_features" }

// InvoiceLine maps to billing_invoice_lines.
type InvoiceLine struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceID          uuid.UUID `gorm:"type:uuid;not null"                              json:"invoiceId"`
	FeatureKey         string    `                                                        json:"featureKey,omitempty"`
	Description        string    `gorm:"not null"                                        json:"description"`
	HSNSAC             string    `gorm:"column:hsn_sac;not null;default:'998314'"        json:"hsnSac"`
	Quantity           int       `gorm:"not null;default:1"                              json:"quantity"`
	UnitPriceCents     int64     `gorm:"not null;default:0"                              json:"unitPriceCents"`
	TaxableAmountCents int64     `gorm:"not null;default:0"                              json:"taxableAmountCents"`
	CGSTCents          int64     `gorm:"column:cgst_cents;not null;default:0"            json:"cgstCents"`
	SGSTCents          int64     `gorm:"column:sgst_cents;not null;default:0"            json:"sgstCents"`
	IGSTCents          int64     `gorm:"column:igst_cents;not null;default:0"            json:"igstCents"`
	TotalCents         int64     `gorm:"not null;default:0"                              json:"totalCents"`
	SortOrder          int       `gorm:"not null;default:0"                              json:"sortOrder"`
	Metadata           model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"                json:"metadata,omitempty"`
	CreatedAt          time.Time `gorm:"not null;default:now()"                          json:"createdAt"`
}

func (InvoiceLine) TableName() string { return "billing_invoice_lines" }

// ── Quotation request DTOs ────────────────────────────────────────────────

// CreateQuotationRequest persists a draft. Same selection inputs as preview
// plus the customer-facing details that get rendered into the quotation/invoice.
type CreateQuotationRequest struct {
	FeatureKeys    []string               `json:"featureKeys"               binding:"required"`
	UserCount      int                    `json:"userCount"                 binding:"required,min=1"`
	CustomerState  string                 `json:"customerState,omitempty"   binding:"omitempty,max=64"`
	BillingEmail   string                 `json:"billingEmail,omitempty"    binding:"omitempty,email,max=254"`
	BillingName    string                 `json:"billingName,omitempty"     binding:"omitempty,max=200"`
	BillingAddress map[string]interface{} `json:"billingAddress,omitempty"`
	Notes          string                 `json:"notes,omitempty"           binding:"omitempty,max=2000"`
}

// UpdateQuotationRequest patches a draft. All fields optional. Allowed only
// when status='draft'.
type UpdateQuotationRequest struct {
	FeatureKeys    *[]string               `json:"featureKeys,omitempty"`
	UserCount      *int                    `json:"userCount,omitempty"     binding:"omitempty,min=1"`
	CustomerState  *string                 `json:"customerState,omitempty" binding:"omitempty,max=64"`
	BillingEmail   *string                 `json:"billingEmail,omitempty"  binding:"omitempty,email,max=254"`
	BillingName    *string                 `json:"billingName,omitempty"   binding:"omitempty,max=200"`
	BillingAddress *map[string]interface{} `json:"billingAddress,omitempty"`
	Notes          *string                 `json:"notes,omitempty"         binding:"omitempty,max=2000"`
}

// ActivateQuotationResponse is what /:id/activate returns — the activation
// receipt the frontend uses to redirect the user to the new invoice.
type ActivateQuotationResponse struct {
	Quotation    Quotation     `json:"quotation"`
	Plan         Plan          `json:"plan"`
	Subscription Subscription  `json:"subscription"`
	Invoice      Invoice       `json:"invoice"`
	InvoiceLines []InvoiceLine `json:"invoiceLines"`
}

// ── DTOs (lifecycle) ──────────────────────────────────────────────────────

type UpdateBillingRequest struct {
	BillingEmail   *string                `json:"billingEmail,omitempty"`
	BillingName    *string                `json:"billingName,omitempty"`
	BillingAddress map[string]interface{} `json:"billingAddress,omitempty"`
}
