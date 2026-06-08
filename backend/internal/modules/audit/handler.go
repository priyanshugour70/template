package audit

import (
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
