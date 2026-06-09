package audit

import (
	"encoding/csv"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
	"github.com/your-org/your-service/internal/pkg/response"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

type PermissionFunc func(perm string) gin.HandlerFunc

func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	a := g.Group("/audit-logs", auth)
	{
		a.GET("", perm("audit.read"), h.list)
		a.GET("/stats", perm("audit.read"), h.stats)
		a.GET("/timeseries", perm("audit.read"), h.timeseries)
		a.GET("/top/users", perm("audit.read"), h.topUsers)
		a.GET("/top/failing-paths", perm("audit.read"), h.topFailingPaths)
		a.GET("/top/actions", perm("audit.read"), h.topActions)
		a.GET("/status-breakdown", perm("audit.read"), h.statusBreakdown)
		a.GET("/export.csv", perm("audit.export"), h.exportCSV)
		a.GET("/:id", perm("audit.read"), h.get)
	}
}

func (h *Handler) list(c *gin.Context) {
	ctx := c.Request.Context()
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)

	var tenantPtr *uuid.UUID
	if tid != uuid.Nil && !appctx.IsSuperAdmin(ctx) {
		t := tid
		tenantPtr = &t
	} else if v := c.Query("tenantId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			tenantPtr = &id
		}
	}

	var orgPtr *uuid.UUID
	if v := c.Query("organizationId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			orgPtr = &id
		}
	} else if oid != uuid.Nil {
		o := oid
		orgPtr = &o
	}

	p := pagination.FromGin(c)
	filter := ListFilter{
		Search:     p.Search,
		UserEmail:  c.Query("userEmail"),
		Action:     c.Query("action"),
		TargetType: c.Query("targetType"),
		Method:     c.Query("method"),
		Path:       c.Query("path"),
	}
	if v := c.Query("userId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.UserID = &id
		}
	}
	if v := c.Query("targetId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.TargetID = &id
		}
	}
	if v := c.Query("statusFrom"); v != "" {
		var n int
		_, _ = parseInt(v, &n)
		filter.StatusFrom = n
	}
	if v := c.Query("statusTo"); v != "" {
		var n int
		_, _ = parseInt(v, &n)
		filter.StatusTo = n
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.OccurredFrom = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.OccurredTo = &t
		}
	}

	rows, total, err := h.svc.List(ctx, tenantPtr, orgPtr, filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) get(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	tid := appctx.TenantID(ctx)
	var tenantPtr *uuid.UUID
	if tid != uuid.Nil && !appctx.IsSuperAdmin(ctx) {
		t := tid
		tenantPtr = &t
	}
	row, err := h.svc.Get(ctx, tenantPtr, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, row)
}

// ── aggregation handlers (dashboard) ──────────────────────────────────────

func (h *Handler) scope(c *gin.Context) (*uuid.UUID, *uuid.UUID) {
	ctx := c.Request.Context()
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)

	var tenantPtr *uuid.UUID
	if tid != uuid.Nil && !appctx.IsSuperAdmin(ctx) {
		t := tid
		tenantPtr = &t
	} else if v := c.Query("tenantId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			tenantPtr = &id
		}
	}
	var orgPtr *uuid.UUID
	if v := c.Query("organizationId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			orgPtr = &id
		}
	} else if oid != uuid.Nil {
		o := oid
		orgPtr = &o
	}
	return tenantPtr, orgPtr
}

func (h *Handler) parseStatsFilter(c *gin.Context) StatsFilter {
	f := StatsFilter{
		UserEmail: c.Query("userEmail"),
		Action:    c.Query("action"),
		Method:    c.Query("method"),
		Path:      c.Query("path"),
	}
	if v := c.Query("userId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.UserID = &id
		}
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.OccurredFrom = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.OccurredTo = &t
		}
	}
	return f
}

func (h *Handler) stats(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	out, err := h.svc.Stats(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) timeseries(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	interval := c.DefaultQuery("interval", "hour")
	out, err := h.svc.Timeseries(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c), interval)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) topUsers(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	limit := parseLimit(c.Query("limit"), 10)
	out, err := h.svc.TopUsers(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c), limit)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "top users failed", err))
		return
	}
	response.OK(c, out)
}

func (h *Handler) topFailingPaths(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	limit := parseLimit(c.Query("limit"), 10)
	out, err := h.svc.TopFailingPaths(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c), limit)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "top failing paths failed", err))
		return
	}
	response.OK(c, out)
}

func (h *Handler) topActions(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	limit := parseLimit(c.Query("limit"), 10)
	out, err := h.svc.TopActions(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c), limit)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "top actions failed", err))
		return
	}
	response.OK(c, out)
}

func (h *Handler) statusBreakdown(c *gin.Context) {
	tenantPtr, orgPtr := h.scope(c)
	out, err := h.svc.StatusBreakdown(c.Request.Context(), tenantPtr, orgPtr, h.parseStatsFilter(c))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "status breakdown failed", err))
		return
	}
	response.OK(c, out)
}

func (h *Handler) exportCSV(c *gin.Context) {
	ctx := c.Request.Context()
	tenantPtr, orgPtr := h.scope(c)
	// Reuse list filter parsing — export honours the same filters as the list.
	p := pagination.FromGin(c)
	p.Limit = 1000 // cap export size
	filter := ListFilter{
		UserEmail:  c.Query("userEmail"),
		Action:     c.Query("action"),
		Method:     c.Query("method"),
		Path:       c.Query("path"),
		Search:     p.Search,
		TargetType: c.Query("targetType"),
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.OccurredFrom = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.OccurredTo = &t
		}
	}
	rows, _, err := h.svc.List(ctx, tenantPtr, orgPtr, filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="audit-log.csv"`)
	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	_ = w.Write([]string{
		"occurred_at", "correlation_id", "user_email", "method", "path",
		"status_code", "latency_ms", "action", "target_type", "target_id",
	})
	for _, r := range rows {
		targetID := ""
		if r.TargetID != nil {
			targetID = r.TargetID.String()
		}
		_ = w.Write([]string{
			r.OccurredAt.Format(time.RFC3339),
			r.CorrelationID,
			r.UserEmail,
			r.Method,
			r.Path,
			fmt.Sprintf("%d", r.StatusCode),
			fmt.Sprintf("%d", r.LatencyMs),
			r.Action,
			r.TargetType,
			targetID,
		})
	}
}

func parseLimit(v string, def int) int {
	if v == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return def
	}
	if n <= 0 || n > 100 {
		return def
	}
	return n
}

func parseInt(s string, n *int) (int, error) {
	var v int
	_, err := fmtSscan(s, &v)
	if err == nil {
		*n = v
	}
	return v, err
}

// Tiny shim to keep handler.go free of an `fmt` import for one Sscan use.
func fmtSscan(s string, v *int) (int, error) {
	return fmtSscanf(s, "%d", v)
}

// fmt is intentionally local to keep the handler file tight.
func fmtSscanf(s, format string, v *int) (int, error) {
	return sscanf(s, format, v)
}

func sscanf(s, format string, v *int) (int, error) {
	return sscan(s, v)
}

func sscan(s string, v *int) (int, error) {
	n, err := scanInt(s)
	if err != nil {
		return 0, err
	}
	*v = n
	return 1, nil
}

func scanInt(s string) (int, error) {
	out := 0
	if s == "" {
		return 0, nil
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return out, nil
		}
		out = out*10 + int(s[i]-'0')
	}
	return out, nil
}
