// Package payment is the pluggable payment-processor abstraction for billing.
//
// Today: cash + bank_transfer + cheque all flow through the Manual processor,
// which simply records what the admin says happened. Tomorrow: Razorpay /
// Stripe / etc. will satisfy the same Processor interface so the calling
// service code doesn't change.
package payment

import (
	"context"
	"fmt"
	"strings"
)

// Method is the payment channel — what we actually charge or accept.
type Method string

const (
	MethodCash         Method = "cash"
	MethodBankTransfer Method = "bank_transfer"
	MethodCheque       Method = "cheque"
	MethodGateway      Method = "gateway"
)

// IsValid is true for the methods we support today.
func (m Method) IsValid() bool {
	switch m {
	case MethodCash, MethodBankTransfer, MethodCheque, MethodGateway:
		return true
	}
	return false
}

// MethodLabel renders a human-friendly name for emails / receipts.
// Kept here so the payment package owns the canonical label set.
func MethodLabel(m Method) string {
	switch m {
	case MethodCash:
		return "Cash"
	case MethodBankTransfer:
		return "Bank transfer"
	case MethodCheque:
		return "Cheque"
	case MethodGateway:
		return "Online payment"
	default:
		return string(m)
	}
}

// RecordRequest is what the billing service hands to the processor. The
// processor returns a Receipt describing what was actually captured —
// importantly, gateway processors may return a different amount than
// requested (e.g. tax + currency conversion) so the service trusts the
// receipt over its own request.
type RecordRequest struct {
	InvoiceID      string
	Method         Method
	AmountCents    int64
	Currency       string
	Reference      string // bank txn id / cheque number / gateway txn id
	Notes          string
	// Identity of the recording admin; passed through to audit/receipts.
	RecordedByID    string
	RecordedByEmail string
}

// Receipt is the processor's record of a captured payment.
type Receipt struct {
	Method                Method
	AmountCents           int64
	Currency              string
	Reference             string
	Gateway               string // e.g. "razorpay" / "stripe" — empty for manual
	GatewayTransactionID  string
	Status                string // "recorded" for manual, "pending"/"recorded" for gateway
}

// Processor is the single seam between billing.Service and any payment backend.
type Processor interface {
	Name() string
	Record(ctx context.Context, req RecordRequest) (*Receipt, error)
}

// Manual handles the three offline methods we collect today: cash, bank
// transfer, and cheque. There's no actual capture — the admin tells us money
// arrived; we just persist what they say.
type Manual struct{}

func NewManual() *Manual { return &Manual{} }

func (Manual) Name() string { return "manual" }

func (Manual) Record(_ context.Context, req RecordRequest) (*Receipt, error) {
	if !req.Method.IsValid() {
		return nil, fmt.Errorf("unknown payment method %q", req.Method)
	}
	if req.Method == MethodGateway {
		return nil, fmt.Errorf("gateway method requires a gateway processor, not Manual")
	}
	if req.AmountCents <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	// Bank transfer + cheque normally have a reference; cash often does not.
	if req.Method == MethodBankTransfer && strings.TrimSpace(req.Reference) == "" {
		return nil, fmt.Errorf("bank_transfer requires a reference (UTR / transaction id)")
	}
	if req.Method == MethodCheque && strings.TrimSpace(req.Reference) == "" {
		return nil, fmt.Errorf("cheque requires the cheque number as reference")
	}
	return &Receipt{
		Method:      req.Method,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Reference:   strings.TrimSpace(req.Reference),
		Status:      "recorded",
	}, nil
}

// GatewayStub is a placeholder for the future Razorpay/Stripe processor. It
// satisfies the interface so wiring is in place but every call fails fast —
// nobody should hit this path until the gateway is wired in Phase 6+.
type GatewayStub struct{}

func NewGatewayStub() *GatewayStub { return &GatewayStub{} }
func (GatewayStub) Name() string   { return "gateway-stub" }
func (GatewayStub) Record(_ context.Context, _ RecordRequest) (*Receipt, error) {
	return nil, fmt.Errorf("payment gateway not configured")
}
