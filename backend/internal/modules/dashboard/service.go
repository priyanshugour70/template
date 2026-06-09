package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// GetSummary builds the home-page dashboard payload. Every panel queries a
// different slice of cross-module data, so we fan them out concurrently via
// errgroup — total wall-clock is the slowest single query, not the sum of all.
//
// Scoping rules:
//   - Authenticated non-super-admin: scoped to (tenant, org). org=nil is rejected.
//   - Super-admin with an active org context: same as above (org-scoped view).
//   - Super-admin without an active org: tenant-wide (org=Nil sentinel).
func (s *Service) GetSummary(ctx context.Context) (*Summary, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	if tid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no tenant context", nil)
	}
	// Non-super-admins must have an org context. Super-admins are allowed to
	// roll up across the tenant.
	if oid == uuid.Nil && !appctx.IsSuperAdmin(ctx) {
		return nil, apperr.New(apperr.CodeForbidden, "no org context", nil)
	}

	now := time.Now().UTC()
	// 12-month window starts at first-of-month, 11 months ago. The +1 month
	// upper bound includes the in-progress current month.
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	revFrom := monthStart.AddDate(0, -11, 0)
	revTo := monthStart.AddDate(0, 1, 0)

	// 14-day window for the activity chart + status donut + top endpoints.
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	reqFrom := dayStart.AddDate(0, 0, -13)
	reqTo := dayStart.AddDate(0, 0, 1)

	// Comparator windows for the KPI deltas.
	mtdFrom := monthStart                        // month-to-date
	mtdPrevFrom := monthStart.AddDate(0, -1, 0) // previous month full
	mtdPrevTo := monthStart
	wkFrom := now.AddDate(0, 0, -7)
	wkPrevFrom := now.AddDate(0, 0, -14)
	wkPrevTo := wkFrom

	var (
		mrr, invMTD, invPrev, due  int64
		users7d, usersPrev7d, openCount int64
		revenue                    []RevenueBucket
		reqs                       []RequestBucket
		statuses                   []StatusSlice
		endpoints                  []EndpointBucket
		aging                      []AgingBucket
		activity                   []ActivityEntry
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) { mrr, err = s.repo.SumActiveMRR(gctx, tid, oid); return })
	g.Go(func() (err error) {
		invMTD, err = s.repo.SumInvoicedBetween(gctx, tid, oid, mtdFrom, now)
		return
	})
	g.Go(func() (err error) {
		invPrev, err = s.repo.SumInvoicedBetween(gctx, tid, oid, mtdPrevFrom, mtdPrevTo)
		return
	})
	g.Go(func() (err error) {
		users7d, err = s.repo.CountActiveUsersBetween(gctx, tid, oid, wkFrom, now)
		return
	})
	g.Go(func() (err error) {
		usersPrev7d, err = s.repo.CountActiveUsersBetween(gctx, tid, oid, wkPrevFrom, wkPrevTo)
		return
	})
	g.Go(func() error {
		c, d, err := s.repo.OutstandingInvoiceTotals(gctx, tid, oid)
		openCount, due = c, d
		return err
	})
	g.Go(func() (err error) { revenue, err = s.repo.RevenueByMonth(gctx, tid, oid, revFrom, revTo); return })
	g.Go(func() (err error) { reqs, err = s.repo.RequestsByDay(gctx, tid, oid, reqFrom, reqTo); return })
	g.Go(func() (err error) { statuses, err = s.repo.StatusBreakdown(gctx, tid, oid, reqFrom, reqTo); return })
	g.Go(func() (err error) { endpoints, err = s.repo.TopEndpoints(gctx, tid, oid, reqFrom, reqTo, 8); return })
	g.Go(func() (err error) { aging, err = s.repo.InvoiceAging(gctx, tid, oid, now); return })
	g.Go(func() (err error) { activity, err = s.repo.RecentActivity(gctx, tid, oid, 10); return })

	if err := g.Wait(); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "build dashboard summary failed", err)
	}

	return &Summary{
		KPIs: KPIs{
			MRRCents:               mrr,
			MRRDeltaPct:            0, // MRR comparators need a snapshot table; deferred. Surface 0 = "—" on the UI.
			InvoicedThisMonthCents: invMTD,
			InvoicedDeltaPct:       pctDelta(invMTD, invPrev),
			ActiveUsers7d:          users7d,
			ActiveUsersDeltaPct:    pctDelta(users7d, usersPrev7d),
			OutstandingDueCents:    due,
			OpenInvoiceCount:       openCount,
		},
		RevenueByMonth:  revenue,
		RequestsByDay:   reqs,
		StatusBreakdown: statuses,
		TopEndpoints:    endpoints,
		InvoiceAging:    aging,
		RecentActivity:  activity,
		GeneratedAt:     now,
	}, nil
}

// pctDelta computes (current - previous) / previous × 100 with sane handling
// for zero baselines. Returns 0 when previous == 0 — UI renders that as "—".
func pctDelta(current, previous int64) float64 {
	if previous == 0 {
		return 0
	}
	return (float64(current-previous) / float64(previous)) * 100.0
}
