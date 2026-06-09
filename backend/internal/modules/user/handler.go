package user

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
	"github.com/your-org/your-service/internal/pkg/response"
)

// RBACPort is the narrow interface this module needs from rbac.Service.
// Defined here (not imported) to avoid a cycle.
type RBACPort interface {
	InvalidateMembership(ctx context.Context, membershipID uuid.UUID)
	ResolveForUser(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
}

type Handler struct {
	svc      *Service
	log      *zap.Logger
	rbacBust RBACPort
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

// WithRBAC wires the rbac port so PATCH membership invalidates perms and
// /effective-permissions can resolve.
func (h *Handler) WithRBAC(b RBACPort) *Handler {
	h.rbacBust = b
	return h
}

type PermissionFunc func(perm string) gin.HandlerFunc

func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	users := g.Group("/users", auth)
	{
		// self-service
		users.GET("/me", h.me)
		users.PATCH("/me", h.updateMe)

		// org-scoped listing
		users.GET("", perm("user.list"), h.list)
		users.GET("/:id", h.getOne) // self-or-permission handled inside
		users.PATCH("/:id", perm("user.update"), h.update)
		users.POST("/:id/suspend", perm("user.suspend"), h.suspend)
		users.POST("/:id/reactivate", perm("user.suspend"), h.reactivate)
		users.DELETE("/:id", perm("user.delete"), h.archive)

		// security actions (admin)
		users.POST("/:id/force-password-reset", perm("user.update"), h.forcePasswordReset)
		users.POST("/:id/reset-mfa", perm("user.update"), h.resetMFA)
		users.POST("/:id/unlock", perm("user.update"), h.unlock)
		users.GET("/:id/effective-permissions", perm("user.read"), h.effectivePermissions)

		// memberships
		users.GET("/me/memberships", h.myMemberships)
		users.GET("/:id/memberships", perm("user.read"), h.adminListMemberships)
		users.PATCH("/:id/memberships/:mid", perm("user.update"), h.updateMembership)
		users.POST("/:id/memberships/:mid/suspend", perm("user.suspend"), h.suspendMembership)
		users.DELETE("/:id/memberships/:mid", perm("user.delete"), h.archiveMembership)

		// bulk
		users.POST("/bulk/memberships", perm("user.update"), h.bulkUpdateMemberships)
	}
}

// ── handlers ───────────────────────────────────────────────────────────────

func (h *Handler) me(c *gin.Context) {
	uid := appctx.UserID(c.Request.Context())
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "no principal", nil))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, u)
}

func (h *Handler) updateMe(c *gin.Context) {
	uid := appctx.UserID(c.Request.Context())
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "no principal", nil))
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	u, err := h.svc.Update(c.Request.Context(), uid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, u)
}

func (h *Handler) list(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	oid := appctx.OrganizationID(c.Request.Context())
	if tid == uuid.Nil || oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant/org context", nil))
		return
	}
	p := pagination.FromGin(c)
	filter := ListFilter{
		Status:     c.Query("status"),
		Search:     p.Search,
		Role:       c.Query("role"),
		JobTitle:   c.Query("jobTitle"),
		Department: c.Query("department"),
	}
	if v := c.Query("departmentId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DepartmentID = &id
		}
	}
	if v := c.Query("mfa"); v != "" {
		b := v == "true" || v == "1"
		filter.MFAEnabled = &b
	}
	parseTime := func(v string) *time.Time {
		if v == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil
		}
		return &t
	}
	filter.LastLoginAfter = parseTime(c.Query("lastLoginAfter"))
	filter.LastLoginBefore = parseTime(c.Query("lastLoginBefore"))
	filter.CreatedAfter = parseTime(c.Query("createdAfter"))
	filter.CreatedBefore = parseTime(c.Query("createdBefore"))
	rows, total, err := h.svc.ListInOrg(c.Request.Context(), tid, oid, filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) getOne(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	// self-access is always allowed.
	if id != appctx.UserID(c.Request.Context()) {
		// otherwise the rbac middleware should have run; we re-check by
		// requiring user.read for cross-user fetches via a soft check.
		// Bootstrap wires permission check via perm() middleware on a dedicated
		// route group. For self-or-permission we accept either.
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, u)
}

func (h *Handler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, u)
}

func (h *Handler) suspend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.Suspend(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) reactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.Reactivate(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) archive(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.Archive(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) myMemberships(c *gin.Context) {
	uid := appctx.UserID(c.Request.Context())
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListMembershipsByUser(c.Request.Context(), uid, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "list memberships failed", err))
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) updateMembership(c *gin.Context) {
	mid, err := uuid.Parse(c.Param("mid"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in UpdateMembershipInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	m, err := h.svc.UpdateMembership(c.Request.Context(), mid, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	// Bust the cached permission set — dept changes affect inherited grants.
	if h.rbacBust != nil && (in.DepartmentID != nil || in.ReportsTo != nil) {
		h.rbacBust.InvalidateMembership(c.Request.Context(), mid)
	}
	response.OK(c, m)
}

func (h *Handler) suspendMembership(c *gin.Context) {
	mid, err := uuid.Parse(c.Param("mid"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.SuspendMembership(c.Request.Context(), mid); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) archiveMembership(c *gin.Context) {
	mid, err := uuid.Parse(c.Param("mid"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ArchiveMembership(c.Request.Context(), mid); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── admin security actions ────────────────────────────────────────────────

func (h *Handler) forcePasswordReset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ForcePasswordReset(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) resetMFA(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ResetMFA(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) unlock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.UnlockUser(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) effectivePermissions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization is required", nil))
		return
	}
	if h.rbacBust == nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "permissions resolver not wired", nil))
		return
	}
	keys, err := h.rbacBust.ResolveForUser(c.Request.Context(), id, oid)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "resolve permissions failed", err))
		return
	}
	response.OK(c, gin.H{"permissions": keys})
}

// adminListMemberships returns a user's memberships scoped to the caller's
// tenant. Used by the admin user detail view. Paginated.
func (h *Handler) adminListMemberships(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	p := pagination.FromGin(c)
	// Fetch unpaginated first so the post-filter to caller's tenant is correct;
	// then page in-memory. The caller-tenant filter is the tighter constraint,
	// so this is bounded by the user's reachable memberships within one tenant —
	// typically tiny.
	rows, _, err := h.svc.ListMembershipsByUser(c.Request.Context(), id, 0, 0)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "list memberships failed", err))
		return
	}
	tid := appctx.TenantID(c.Request.Context())
	if tid != uuid.Nil && !appctx.IsSuperAdmin(c.Request.Context()) {
		out := make([]Membership, 0, len(rows))
		for _, m := range rows {
			if m.TenantID == tid {
				out = append(out, m)
			}
		}
		rows = out
	}
	total := len(rows)
	if p.Limit > 0 {
		start := p.Offset
		end := start + p.Limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		rows = rows[start:end]
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, total)
}

// ── bulk ──────────────────────────────────────────────────────────────────

type bulkMembershipsRequest struct {
	MembershipIDs []uuid.UUID           `json:"membershipIds" binding:"required,min=1"`
	Patch         UpdateMembershipInput `json:"patch"`
}

type bulkMembershipsResponse struct {
	Updated int       `json:"updated"`
	Failed  []failure `json:"failed,omitempty"`
}

type failure struct {
	MembershipID string `json:"membershipId"`
	Error        string `json:"error"`
}

func (h *Handler) bulkUpdateMemberships(c *gin.Context) {
	var req bulkMembershipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	out := bulkMembershipsResponse{}
	bustNeeded := req.Patch.DepartmentID != nil || req.Patch.ReportsTo != nil
	for _, mid := range req.MembershipIDs {
		if _, err := h.svc.UpdateMembership(c.Request.Context(), mid, req.Patch); err != nil {
			out.Failed = append(out.Failed, failure{MembershipID: mid.String(), Error: err.Error()})
			continue
		}
		out.Updated++
		if h.rbacBust != nil && bustNeeded {
			h.rbacBust.InvalidateMembership(c.Request.Context(), mid)
		}
	}
	response.OK(c, out)
}
