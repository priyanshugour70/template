// Package pdf renders billing artefacts (invoices, receipts) as PDFs.
//
// The package exposes a `Renderer` interface so the implementation is easy to
// swap. Today we use signintech/gopdf because it's pure-Go and needs no
// external runtime (no Chromium). When the design demands richer layouts,
// swap in an HTML→PDF implementation behind the same interface.
package pdf

import "time"

// InvoiceInput is the structured data needed to render a single tax invoice.
// All fields are denormalised here so the renderer never touches the database
// — the caller (billing service) collects everything it needs first.
type InvoiceInput struct {
	// Issuer / company info from billing_tax_config.
	CompanyName    string
	CompanyAddress string
	CompanyGSTIN   string
	BankName       string
	BankAccount    string
	BankIFSC       string
	BankAccountName string
	InvoiceTerms   string

	// Customer / "bill to" — pulled from the tenant + invoice billing fields.
	CustomerName    string
	CustomerEmail   string
	CustomerAddress string
	CustomerGSTIN   string
	PlaceOfSupply   string

	// Invoice header.
	InvoiceNumber string
	IssuedAt      time.Time
	DueAt         time.Time
	PeriodStart   time.Time
	PeriodEnd     time.Time
	Status        string // "open" / "paid" / "void" — drives the stamp at the top

	// Money.
	Currency      string
	Lines         []InvoiceLineRow
	SubtotalCents int64
	DiscountCents int64
	CGSTCents     int64
	SGSTCents     int64
	IGSTCents     int64
	TotalCents    int64
	AmountDueCents int64
	AmountPaidCents int64
	// IntraState toggles which tax columns we show (CGST+SGST vs IGST) — set
	// to true when CustomerState matches the company's home state.
	IntraState bool
}

// InvoiceLineRow is one row of the line-items table.
type InvoiceLineRow struct {
	Description        string
	HSNSAC             string
	Quantity           int
	UnitPriceCents     int64
	TaxableAmountCents int64
	CGSTCents          int64
	SGSTCents          int64
	IGSTCents          int64
	TotalCents         int64
}

// ReceiptInput is the structured data needed to render a payment receipt
// (Phase 5 will use this; declared here so the package shape is stable).
type ReceiptInput struct {
	CompanyName     string
	CompanyAddress  string
	CompanyGSTIN    string
	ReceiptNumber   string
	PaidAt          time.Time
	CustomerName    string
	CustomerEmail   string
	InvoiceNumber   string
	Method          string // "cash" | "bank_transfer" | "cheque" | "gateway"
	Reference       string
	AmountCents     int64
	Currency        string
}

// Renderer is the interface every PDF backend implements.
type Renderer interface {
	RenderInvoice(in InvoiceInput) ([]byte, error)
	RenderReceipt(in ReceiptInput) ([]byte, error)
}
