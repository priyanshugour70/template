package tenant

import (
	"net/http"

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

// PermissionFunc returns a Gin middleware that requires the named permission.
// Provided by the rbac module via dependency injection (bootstrap wires it).
type PermissionFunc func(perm string) gin.HandlerFunc

// Routes attaches every tenant + organization endpoint. auth is the JWT
// middleware; perm is the rbac permission-check factory.
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	tenants := g.Group("/tenants", auth)
	{
		// super-admin scope
		tenants.POST("", perm("tenant.update"), h.createTenant) // gated by tenant.update — adjust to super-admin gate when rolled out
		tenants.GET("", perm("tenant.read"), h.listTenants)

		// current-tenant scope (derived from JWT)
		me := tenants.Group("/me")
		me.GET("", perm("tenant.read"), h.getMyTenant)
		me.PATCH("", perm("tenant.update"), h.updateMyTenant)
		me.DELETE("", perm("tenant.delete"), h.archiveMyTenant)
		me.GET("/organizations", perm("org.list"), h.listMyOrganizations)
		me.POST("/organizations", perm("org.create"), h.createMyOrganization)
	}

	orgs := g.Group("/organizations", auth)
	{
		orgs.GET("/:id", perm("org.read"), h.getOrganization)
		orgs.PATCH("/:id", perm("org.update"), h.updateOrganization)
		orgs.DELETE("/:id", perm("org.delete"), h.archiveOrganization)
	}
}

// ── tenants ────────────────────────────────────────────────────────────────

// createTenant onboards a new tenant + its default organization.
// @Summary  Create a new tenant
// @Tags     tenant
// @Accept   json
// @Produce  json
// @Param    body body CreateTenantRequest true "tenant"
// @Success  201 {object} response.Body
// @Router   /tenants [post]
func (h *Handler) createTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	t, org, err := h.svc.CreateTenant(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, gin.H{"tenant": t, "defaultOrganization": org})
}

// listTenants returns a paginated list. Super-admin scope.
// @Summary  List tenants
// @Tags     tenant
// @Produce  json
// @Param    page   query int    false "page"
// @Param    limit  query int    false "limit"
// @Param    status query string false "status filter"
// @Param    q      query string false "search"
// @Router   /tenants [get]
func (h *Handler) listTenants(c *gin.Context) {
	if !appctx.IsSuperAdmin(c.Request.Context()) {
		response.Error(c, apperr.New(apperr.CodeForbidden, "super-admin required", nil))
		return
	}
	p := pagination.FromGin(c)
	filter := ListFilter{Status: c.Query("status"), Search: p.Search}
	rows, total, err := h.svc.ListTenants(c.Request.Context(), filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) getMyTenant(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	if tid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant context", nil))
		return
	}
	t, err := h.svc.GetTenant(c.Request.Context(), tid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func (h *Handler) updateMyTenant(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	if tid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant context", nil))
		return
	}
	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	t, err := h.svc.UpdateTenant(c.Request.Context(), tid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func (h *Handler) archiveMyTenant(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	if tid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant context", nil))
		return
	}
	if err := h.svc.ArchiveTenant(c.Request.Context(), tid); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── organizations ──────────────────────────────────────────────────────────

func (h *Handler) listMyOrganizations(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	if tid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant context", nil))
		return
	}
	p := pagination.FromGin(c)
	filter := ListFilter{Status: c.Query("status"), Search: p.Search}
	rows, total, err := h.svc.ListOrganizations(c.Request.Context(), tid, filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) createMyOrganization(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	if tid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no tenant context", nil))
		return
	}
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	o, err := h.svc.CreateOrganization(c.Request.Context(), tid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, o)
}

func (h *Handler) getOrganization(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	o, err := h.svc.GetOrganization(c.Request.Context(), tid, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, o)
}

func (h *Handler) updateOrganization(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	o, err := h.svc.UpdateOrganization(c.Request.Context(), tid, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, o)
}

func (h *Handler) archiveOrganization(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ArchiveOrganization(c.Request.Context(), tid, id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
