package audit

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// Insert is called from the worker consumer to persist a captured event.
func (r *Repository) Insert(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// BatchInsert flushes a slice of events in one statement.
func (r *Repository) BatchInsert(ctx context.Context, logs []Log) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 200).Error
}

// List returns audit rows filtered by the caller's tenant. The caller must
// pass a non-nil tenantID; cross-tenant audit access requires super-admin.
func (r *Repository) List(
	ctx context.Context,
	tenantID *uuid.UUID,
	orgID *uuid.UUID,
	filter ListFilter,
	p pagination.Params,
) ([]Log, int64, error) {
	q := r.db.WithContext(ctx).Model(&Log{})
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.UserEmail != "" {
		q = q.Where("user_email = ?", strings.ToLower(filter.UserEmail))
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != nil {
		q = q.Where("target_id = ?", *filter.TargetID)
	}
	if filter.Method != "" {
		q = q.Where("method = ?", strings.ToUpper(filter.Method))
	}
	if filter.Path != "" {
		q = q.Where("path ILIKE ?", "%"+filter.Path+"%")
	}
	if filter.StatusFrom > 0 {
		q = q.Where("status_code >= ?", filter.StatusFrom)
	}
	if filter.StatusTo > 0 {
		q = q.Where("status_code <= ?", filter.StatusTo)
	}
	if filter.OccurredFrom != nil {
		q = q.Where("occurred_at >= ?", *filter.OccurredFrom)
	}
	if filter.OccurredTo != nil {
		q = q.Where("occurred_at <= ?", *filter.OccurredTo)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("user_email ILIKE ? OR path ILIKE ? OR action ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []Log{}
	if err := q.
		Order("occurred_at DESC").
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ── aggregation queries (for the dashboard) ───────────────────────────────

// StatsFilter is the shared shape consumed by every aggregation query.
// It mirrors the read-side filter but drops pagination + free-text search.
type StatsFilter struct {
	UserID       *uuid.UUID
	UserEmail    string
	Action       string
	Method       string
	Path         string
	OccurredFrom *time.Time
	OccurredTo   *time.Time
}

func (r *Repository) baseStatsQuery(
	ctx context.Context,
	tenantID *uuid.UUID,
	orgID *uuid.UUID,
	f StatsFilter,
) *gorm.DB {
	q := r.db.WithContext(ctx).Table("audit_log")
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.UserEmail != "" {
		q = q.Where("user_email = ?", strings.ToLower(f.UserEmail))
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.Method != "" {
		q = q.Where("method = ?", strings.ToUpper(f.Method))
	}
	if f.Path != "" {
		q = q.Where("path ILIKE ?", "%"+f.Path+"%")
	}
	if f.OccurredFrom != nil {
		q = q.Where("occurred_at >= ?", *f.OccurredFrom)
	}
	if f.OccurredTo != nil {
		q = q.Where("occurred_at <= ?", *f.OccurredTo)
	}
	return q
}

// StatsSummary is the response of the overview endpoint.
type StatsSummary struct {
	TotalRequests   int64   `json:"totalRequests"`
	Success2xx      int64   `json:"success2xx"`
	Redirect3xx     int64   `json:"redirect3xx"`
	ClientError4xx  int64   `json:"clientError4xx"`
	ServerError5xx  int64   `json:"serverError5xx"`
	UniqueUsers     int64   `json:"uniqueUsers"`
	UniquePaths     int64   `json:"uniquePaths"`
	AvgLatencyMs    float64 `json:"avgLatencyMs"`
	P95LatencyMs    float64 `json:"p95LatencyMs"`
	ErrorRatePct    float64 `json:"errorRatePct"`
}

// Stats — single overview query, computed in one SELECT.
func (r *Repository) Stats(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, f StatsFilter) (*StatsSummary, error) {
	var s StatsSummary
	q := r.baseStatsQuery(ctx, tenantID, orgID, f)
	err := q.Select(`
		COUNT(*) AS total_requests,
		COUNT(*) FILTER (WHERE status_code BETWEEN 200 AND 299) AS success2xx,
		COUNT(*) FILTER (WHERE status_code BETWEEN 300 AND 399) AS redirect3xx,
		COUNT(*) FILTER (WHERE status_code BETWEEN 400 AND 499) AS client_error4xx,
		COUNT(*) FILTER (WHERE status_code >= 500) AS server_error5xx,
		COUNT(DISTINCT user_id) FILTER (WHERE user_id IS NOT NULL) AS unique_users,
		COUNT(DISTINCT path) FILTER (WHERE path IS NOT NULL) AS unique_paths,
		COALESCE(AVG(latency_ms), 0)::float8 AS avg_latency_ms,
		COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms), 0)::float8 AS p95_latency_ms
	`).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	if s.TotalRequests > 0 {
		errors := s.ClientError4xx + s.ServerError5xx
		s.ErrorRatePct = (float64(errors) / float64(s.TotalRequests)) * 100.0
	}
	return &s, nil
}

// TimeseriesBucket is one row of the timeseries response.
type TimeseriesBucket struct {
	Bucket         time.Time `json:"bucket"`
	Total          int64     `json:"total"`
	Success2xx     int64     `json:"success2xx"`
	Redirect3xx    int64     `json:"redirect3xx"`
	ClientError4xx int64     `json:"clientError4xx"`
	ServerError5xx int64     `json:"serverError5xx"`
}

// Timeseries returns request counts bucketed by interval. Interval is one of
// "minute", "hour", "day", "week" — anything else falls back to "hour".
func (r *Repository) Timeseries(
	ctx context.Context,
	tenantID *uuid.UUID,
	orgID *uuid.UUID,
	f StatsFilter,
	interval string,
) ([]TimeseriesBucket, error) {
	allowed := map[string]bool{"minute": true, "hour": true, "day": true, "week": true}
	if !allowed[interval] {
		interval = "hour"
	}
	rows := []TimeseriesBucket{}
	q := r.baseStatsQuery(ctx, tenantID, orgID, f)
	// date_trunc is parameter-safe via injection of a whitelisted literal.
	err := q.Select(`
		date_trunc('` + interval + `', occurred_at) AS bucket,
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE status_code BETWEEN 200 AND 299) AS success2xx,
		COUNT(*) FILTER (WHERE status_code BETWEEN 300 AND 399) AS redirect3xx,
		COUNT(*) FILTER (WHERE status_code BETWEEN 400 AND 499) AS client_error4xx,
		COUNT(*) FILTER (WHERE status_code >= 500) AS server_error5xx
	`).Group("bucket").Order("bucket ASC").Scan(&rows).Error
	return rows, err
}

// TopRow is a generic (key, count) shape — used for top-users, top-paths etc.
type TopRow struct {
	Key   string `json:"key"`
	Label string `json:"label,omitempty"`
	Count int64  `json:"count"`
}

// TopUsers — most active by user_email (cheap; user_id requires a join).
func (r *Repository) TopUsers(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, f StatsFilter, limit int) ([]TopRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows := []TopRow{}
	err := r.baseStatsQuery(ctx, tenantID, orgID, f).
		Select("COALESCE(user_email, 'anonymous') AS key, COUNT(*) AS count").
		Where("user_email IS NOT NULL").
		Group("user_email").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// TopFailingPaths — paths with the most 4xx/5xx responses.
func (r *Repository) TopFailingPaths(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, f StatsFilter, limit int) ([]TopRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows := []TopRow{}
	err := r.baseStatsQuery(ctx, tenantID, orgID, f).
		Select("COALESCE(path, '-') AS key, COUNT(*) AS count").
		Where("status_code >= 400").
		Group("path").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// TopActions — most common action labels (e.g. "user.create").
func (r *Repository) TopActions(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, f StatsFilter, limit int) ([]TopRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows := []TopRow{}
	err := r.baseStatsQuery(ctx, tenantID, orgID, f).
		Select("COALESCE(action, 'unknown') AS key, COUNT(*) AS count").
		Where("action IS NOT NULL AND action <> ''").
		Group("action").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// StatusBreakdown — count per HTTP status code (sorted desc).
func (r *Repository) StatusBreakdown(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, f StatsFilter) ([]TopRow, error) {
	rows := []TopRow{}
	err := r.baseStatsQuery(ctx, tenantID, orgID, f).
		Select("status_code::text AS key, COUNT(*) AS count").
		Where("status_code IS NOT NULL AND status_code > 0").
		Group("status_code").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

// Get returns a single audit row, scoped to tenant when provided.
func (r *Repository) Get(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) (*Log, error) {
	var l Log
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if err := q.First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}
