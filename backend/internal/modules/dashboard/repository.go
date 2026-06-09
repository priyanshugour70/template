package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository is a thin read-only wrapper around the cross-module tables we
// aggregate. Every method is scoped by tenantID + (optional) orgID — super-
// admins without an org context pass uuid.Nil for orgID to opt into
// tenant-wide rollups.
type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// ── KPIs ───────────────────────────────────────────────────────────────────

// SumActiveMRR returns total monthly recurring revenue across active subs
// scoped to the org (or tenant when orgID is nil). 0 when nothing matches —
// not an error.
func (r *Repository) SumActiveMRR(ctx context.Context, tenantID, orgID uuid.UUID) (int64, error) {
	var v int64
	q := r.db.WithContext(ctx).Table("billing_subscriptions").
		Where("status = 'active' AND deleted_at IS NULL").
		Where("tenant_id = ?", tenantID)
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Select("COALESCE(SUM(total_cents), 0)").Row().Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

// SumInvoicedBetween returns the total_cents on invoices issued in [from, to).
// Used both for "invoiced this month" and the previous-month comparator.
func (r *Repository) SumInvoicedBetween(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time) (int64, error) {
	var v int64
	q := r.db.WithContext(ctx).Table("billing_invoices").
		Where("issued_at >= ? AND issued_at < ?", from, to).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL")
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Select("COALESCE(SUM(total_cents), 0)").Row().Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

// CountActiveUsersBetween counts distinct user_ids seen in the audit log in
// [from, to). Approximation of MAU/WAU; trustworthy because the audit
// middleware fires on every authenticated request.
func (r *Repository) CountActiveUsersBetween(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time) (int64, error) {
	var v int64
	q := r.db.WithContext(ctx).Table("audit_log").
		Where("occurred_at >= ? AND occurred_at < ?", from, to).
		Where("user_id IS NOT NULL").
		Where("tenant_id = ?", tenantID)
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Select("COUNT(DISTINCT user_id)").Row().Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

// OutstandingInvoiceTotals returns the count + sum of open invoices.
func (r *Repository) OutstandingInvoiceTotals(ctx context.Context, tenantID, orgID uuid.UUID) (count int64, due int64, err error) {
	row := struct {
		Count int64 `gorm:"column:cnt"`
		Due   int64 `gorm:"column:due"`
	}{}
	q := r.db.WithContext(ctx).Table("billing_invoices").
		Select("COUNT(*) AS cnt, COALESCE(SUM(amount_due_cents), 0) AS due").
		Where("status = 'open'").
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL")
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err = q.Scan(&row).Error; err != nil {
		return 0, 0, err
	}
	return row.Count, row.Due, nil
}

// ── revenue chart ─────────────────────────────────────────────────────────

// RevenueByMonth returns one row per UTC month in [from, to). issued_cents is
// the sum on billing_invoices.issued_at; paid_cents is the sum on
// billing_invoices.paid_at (so the same invoice contributes to one issued
// bucket and a possibly-later paid bucket — keeps the trends honest).
//
// We union two date_trunc queries instead of FULL OUTER JOIN because gorm's
// Row() scanner doesn't love OUTER JOIN against an empty side.
func (r *Repository) RevenueByMonth(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time) ([]RevenueBucket, error) {
	rows := []struct {
		Bucket time.Time `gorm:"column:bucket"`
		Issued int64     `gorm:"column:issued"`
		Paid   int64     `gorm:"column:paid"`
	}{}

	orgFilter := ""
	if orgID != uuid.Nil {
		orgFilter = "AND organization_id = '" + orgID.String() + "'"
	}

	// One pass that computes both sums in a single CTE per month. The
	// generate_series gives us zero-filled months so the chart doesn't show
	// gaps for quiet periods.
	// Pin every bucket calc to UTC. The default session timezone is
	// Asia/Kolkata (per the DSN), and `date_trunc('month', timestamptz)` uses
	// the session zone — so without the AT TIME ZONE 'UTC' coercion the agg
	// CTEs produce IST midnights while generate_series produces UTC midnights
	// and the LEFT JOIN silently fails.
	q := `
WITH months AS (
  SELECT generate_series(?::timestamptz, ?::timestamptz - interval '1 day', interval '1 month') AS bucket
), issued AS (
  SELECT date_trunc('month', issued_at AT TIME ZONE 'UTC') AS bucket, COALESCE(SUM(total_cents),0) AS amt
  FROM billing_invoices
  WHERE issued_at >= ? AND issued_at < ?
    AND tenant_id = ? AND deleted_at IS NULL ` + orgFilter + `
  GROUP BY 1
), paid AS (
  SELECT date_trunc('month', paid_at AT TIME ZONE 'UTC') AS bucket, COALESCE(SUM(total_cents),0) AS amt
  FROM billing_invoices
  WHERE paid_at IS NOT NULL AND paid_at >= ? AND paid_at < ?
    AND tenant_id = ? AND deleted_at IS NULL ` + orgFilter + `
  GROUP BY 1
)
SELECT m.bucket AS bucket,
       COALESCE(issued.amt, 0) AS issued,
       COALESCE(paid.amt, 0)   AS paid
FROM months m
LEFT JOIN issued ON issued.bucket = m.bucket AT TIME ZONE 'UTC'
LEFT JOIN paid   ON paid.bucket   = m.bucket AT TIME ZONE 'UTC'
ORDER BY m.bucket ASC`

	if err := r.db.WithContext(ctx).Raw(q,
		from, to,
		from, to, tenantID,
		from, to, tenantID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]RevenueBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, RevenueBucket{
			Month:       row.Bucket,
			IssuedCents: row.Issued,
			PaidCents:   row.Paid,
		})
	}
	return out, nil
}

// ── request chart + status donut + top endpoints ──────────────────────────

// RequestsByDay returns one row per UTC day in [from, to) with total +
// 5xx counts. Zero-filled across the window via generate_series.
func (r *Repository) RequestsByDay(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time) ([]RequestBucket, error) {
	rows := []struct {
		Bucket   time.Time `gorm:"column:bucket"`
		Requests int64     `gorm:"column:requests"`
		Errors   int64     `gorm:"column:errors"`
	}{}

	orgFilter := ""
	if orgID != uuid.Nil {
		orgFilter = "AND organization_id = '" + orgID.String() + "'"
	}

	// Same UTC pinning rationale as RevenueByMonth — see the comment there.
	q := `
WITH days AS (
  SELECT generate_series(?::timestamptz, ?::timestamptz - interval '1 day', interval '1 day') AS bucket
), agg AS (
  SELECT date_trunc('day', occurred_at AT TIME ZONE 'UTC') AS bucket,
         COUNT(*) AS requests,
         COUNT(*) FILTER (WHERE status_code >= 500) AS errors
  FROM audit_log
  WHERE occurred_at >= ? AND occurred_at < ?
    AND tenant_id = ? ` + orgFilter + `
  GROUP BY 1
)
SELECT d.bucket AS bucket,
       COALESCE(agg.requests, 0) AS requests,
       COALESCE(agg.errors, 0)   AS errors
FROM days d
LEFT JOIN agg ON agg.bucket = d.bucket AT TIME ZONE 'UTC'
ORDER BY d.bucket ASC`

	if err := r.db.WithContext(ctx).Raw(q,
		from, to,
		from, to, tenantID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]RequestBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, RequestBucket{
			Day:      row.Bucket,
			Requests: row.Requests,
			Errors:   row.Errors,
		})
	}
	return out, nil
}

// StatusBreakdown groups audit_log status codes into 2xx/3xx/4xx/5xx for the
// donut chart. Window is the same as the request series (14 days by default).
func (r *Repository) StatusBreakdown(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time) ([]StatusSlice, error) {
	rows := []struct {
		Grp   string `gorm:"column:grp"`
		Count int64  `gorm:"column:count"`
	}{}

	q := r.db.WithContext(ctx).Table("audit_log").
		Select(`CASE
		         WHEN status_code BETWEEN 200 AND 299 THEN '2xx'
		         WHEN status_code BETWEEN 300 AND 399 THEN '3xx'
		         WHEN status_code BETWEEN 400 AND 499 THEN '4xx'
		         WHEN status_code BETWEEN 500 AND 599 THEN '5xx'
		         ELSE 'other'
		       END AS grp, COUNT(*) AS count`).
		Where("occurred_at >= ? AND occurred_at < ?", from, to).
		Where("tenant_id = ?", tenantID).
		Group("grp").Order("grp ASC")
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]StatusSlice, 0, len(rows))
	for _, row := range rows {
		out = append(out, StatusSlice{Group: row.Grp, Count: row.Count})
	}
	return out, nil
}

// TopEndpoints returns the most-hit (method, route) pairs over the window
// with their request count + average latency. Limit usually 8 for the UI.
// Route falls back to path when route is unset (older audit rows).
func (r *Repository) TopEndpoints(ctx context.Context, tenantID, orgID uuid.UUID, from, to time.Time, limit int) ([]EndpointBucket, error) {
	rows := []struct {
		Method string `gorm:"column:method"`
		Route  string `gorm:"column:route"`
		Count  int64  `gorm:"column:count"`
		Avg    int64  `gorm:"column:avg_latency_ms"`
	}{}

	q := r.db.WithContext(ctx).Table("audit_log").
		Select(`method, COALESCE(NULLIF(route, ''), path) AS route, COUNT(*) AS count, COALESCE(AVG(latency_ms),0)::bigint AS avg_latency_ms`).
		Where("occurred_at >= ? AND occurred_at < ?", from, to).
		Where("tenant_id = ?", tenantID).
		Where("method IS NOT NULL AND method <> ''").
		Group("method, COALESCE(NULLIF(route, ''), path)").
		Order("count DESC").Limit(limit)
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]EndpointBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, EndpointBucket{
			Method: row.Method, Route: row.Route, Count: row.Count, AvgLatencyMs: row.Avg,
		})
	}
	return out, nil
}

// ── invoice aging ─────────────────────────────────────────────────────────

// InvoiceAging buckets open invoices by days-overdue. Buckets:
//   current (due_at >= today)
//   1-30 / 31-60 / 61-90 / 90+ days overdue
// Voided / paid invoices excluded.
func (r *Repository) InvoiceAging(ctx context.Context, tenantID, orgID uuid.UUID, now time.Time) ([]AgingBucket, error) {
	rows := []struct {
		Bucket string `gorm:"column:bucket"`
		Count  int64  `gorm:"column:count"`
		Due    int64  `gorm:"column:total_due"`
	}{}

	orgFilter := ""
	if orgID != uuid.Nil {
		orgFilter = "AND organization_id = '" + orgID.String() + "'"
	}

	q := `
SELECT bucket,
       COUNT(*) AS count,
       COALESCE(SUM(amount_due_cents),0) AS total_due
FROM (
  SELECT amount_due_cents,
         CASE
           WHEN due_at IS NULL OR due_at >= ?::timestamptz THEN 'current'
           WHEN due_at >= ?::timestamptz - interval '30 days' THEN '1-30'
           WHEN due_at >= ?::timestamptz - interval '60 days' THEN '31-60'
           WHEN due_at >= ?::timestamptz - interval '90 days' THEN '61-90'
           ELSE '90+'
         END AS bucket
  FROM billing_invoices
  WHERE status = 'open'
    AND tenant_id = ? AND deleted_at IS NULL ` + orgFilter + `
) t
GROUP BY bucket`

	if err := r.db.WithContext(ctx).Raw(q, now, now, now, now, tenantID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Order buckets logically (current first, 90+ last) — query returns
	// them alphabetically so we re-order in code.
	want := []string{"current", "1-30", "31-60", "61-90", "90+"}
	by := map[string]AgingBucket{}
	for _, row := range rows {
		by[row.Bucket] = AgingBucket{Bucket: row.Bucket, Count: row.Count, TotalDueCents: row.Due}
	}
	out := make([]AgingBucket, 0, len(want))
	for _, k := range want {
		if v, ok := by[k]; ok {
			out = append(out, v)
		} else {
			out = append(out, AgingBucket{Bucket: k})
		}
	}
	return out, nil
}

// ── activity feed ─────────────────────────────────────────────────────────

// RecentActivity returns the last N audit_log entries with the columns the
// feed UI cares about. Ordered newest-first.
func (r *Repository) RecentActivity(ctx context.Context, tenantID, orgID uuid.UUID, limit int) ([]ActivityEntry, error) {
	rows := []struct {
		OccurredAt time.Time `gorm:"column:occurred_at"`
		Action     string    `gorm:"column:action"`
		UserEmail  string    `gorm:"column:user_email"`
		Method     string    `gorm:"column:method"`
		Path       string    `gorm:"column:path"`
		StatusCode int       `gorm:"column:status_code"`
		TargetType string    `gorm:"column:target_type"`
	}{}

	q := r.db.WithContext(ctx).Table("audit_log").
		Select("occurred_at, action, user_email, method, path, status_code, target_type").
		Where("tenant_id = ?", tenantID).
		Order("occurred_at DESC").Limit(limit)
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ActivityEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, ActivityEntry{
			OccurredAt: r.OccurredAt,
			Action:     r.Action,
			UserEmail:  r.UserEmail,
			Method:     r.Method,
			Path:       r.Path,
			StatusCode: r.StatusCode,
			TargetType: r.TargetType,
		})
	}
	return out, nil
}
