// Package dashboard aggregates cross-module analytics for the home page —
// billing trends, request volume, top endpoints, activity feed. The package
// owns NO state of its own; it only reads from existing module tables.
package dashboard

import "time"

// Summary is the single response shape for GET /api/v1/dashboard/summary.
// Every field is org-scoped (or tenant-scoped for super-admins without an
// active org) — the service layer enforces that, not the model.
type Summary struct {
	// ── KPI tiles ──
	KPIs KPIs `json:"kpis"`

	// ── chart series ──
	RevenueByMonth   []RevenueBucket   `json:"revenueByMonth"`   // last 12 months
	RequestsByDay    []RequestBucket   `json:"requestsByDay"`    // last 14 days
	StatusBreakdown  []StatusSlice     `json:"statusBreakdown"`  // last 14 days
	TopEndpoints     []EndpointBucket  `json:"topEndpoints"`     // last 14 days, top 8
	InvoiceAging     []AgingBucket     `json:"invoiceAging"`     // open invoices bucketed
	RecentActivity   []ActivityEntry   `json:"recentActivity"`   // last 10 audit events

	// Metadata so the UI can stamp "as of …" + flag stale caches.
	GeneratedAt time.Time `json:"generatedAt"`
}

// KPIs are the four headline tiles. Every value is an absolute number, the
// `*Delta` fields express how that number compares to the previous period so
// the UI can render arrows + percentages.
type KPIs struct {
	MRRCents          int64   `json:"mrrCents"`           // sum of active subs' totalCents
	MRRDeltaPct       float64 `json:"mrrDeltaPct"`        // vs same metric a month ago
	InvoicedThisMonthCents int64 `json:"invoicedThisMonthCents"`
	InvoicedDeltaPct  float64 `json:"invoicedDeltaPct"`
	ActiveUsers7d     int64   `json:"activeUsers7d"`      // distinct user_ids in audit_log
	ActiveUsersDeltaPct float64 `json:"activeUsersDeltaPct"`
	OutstandingDueCents int64 `json:"outstandingDueCents"` // sum of open-invoice amount_due
	OpenInvoiceCount  int64   `json:"openInvoiceCount"`
}

// RevenueBucket — one month on the trend chart. Both values are in minor
// units; the UI formats. Month is the first-of-month timestamp in UTC.
type RevenueBucket struct {
	Month        time.Time `json:"month"`
	IssuedCents  int64     `json:"issuedCents"`
	PaidCents    int64     `json:"paidCents"`
}

// RequestBucket — one day on the activity chart.
type RequestBucket struct {
	Day      time.Time `json:"day"`
	Requests int64     `json:"requests"`
	Errors   int64     `json:"errors"` // status >= 500
}

// StatusSlice — one slice of the status-code donut. Group is "2xx" / "3xx" / etc.
type StatusSlice struct {
	Group string `json:"group"`
	Count int64  `json:"count"`
}

// EndpointBucket — one row of the "top endpoints" horizontal bar list.
type EndpointBucket struct {
	Route     string `json:"route"`     // e.g. "/api/v1/billing/quotations"
	Method    string `json:"method"`
	Count     int64  `json:"count"`
	AvgLatencyMs int64 `json:"avgLatencyMs"`
}

// AgingBucket — one bar on the invoice-aging chart. Bucket is "current" / "1-30"
// / "31-60" / "61-90" / "90+" — kept stringly so the UI can render them in order.
type AgingBucket struct {
	Bucket string `json:"bucket"`
	Count  int64  `json:"count"`
	TotalDueCents int64 `json:"totalDueCents"`
}

// ActivityEntry — one line in the recent-activity feed.
type ActivityEntry struct {
	OccurredAt time.Time `json:"occurredAt"`
	Action     string    `json:"action"`      // verb, e.g. "billing.invoice.pay"
	UserEmail  string    `json:"userEmail,omitempty"`
	Method     string    `json:"method,omitempty"`
	Path       string    `json:"path,omitempty"`
	StatusCode int       `json:"statusCode"`
	TargetType string    `json:"targetType,omitempty"`
}
