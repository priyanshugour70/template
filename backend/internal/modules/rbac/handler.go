package rbac

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

type PermissionFunc func(perm string) gin.HandlerFunc

func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	perms := g.Group("/permissions", auth)
	{
		perms.GET("", perm("role.read"), h.listPermissions)
	}
	roles := g.Group("/roles", auth)
	{
		roles.GET("", perm("role.list"), h.listRoles)
		roles.POST("", perm("role.create"), h.createRole)
		roles.GET("/:id", perm("role.read"), h.getRole)
		roles.PATCH("/:id", perm("role.update"), h.updateRole)
		roles.DELETE("/:id", perm("role.delete"), h.archiveRole)
		roles.GET("/:id/permissions", perm("role.read"), h.listRolePerms)
	}
	memberships := g.Group("/memberships", auth)
	{
		memberships.POST("/:id/roles", perm("user.assign_role"), h.assignRoles)
		memberships.DELETE("/:id/roles/:roleId", perm("user.assign_role"), h.removeRole)
		memberships.GET("/:id/roles", perm("user.read"), h.listMembershipRoles)
	}
}

func (h *Handler) listPermissions(c *gin.Context) {
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListPermissions(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) listRoles(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListRoles(c.Request.Context(), oid, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) createRole(c *gin.Context) {
	tid := appctx.TenantID(c.Request.Context())
	oid := appctx.OrganizationID(c.Request.Context())
	uid := appctx.UserID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	actor := ptrUUID(uid)
	r, err := h.svc.CreateRole(c.Request.Context(), tid, oid, req, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, r)
}

func (h *Handler) getRole(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	r, err := h.svc.GetRole(c.Request.Context(), oid, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, r)
}

func (h *Handler) updateRole(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	uid := appctx.UserID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	r, err := h.svc.UpdateRole(c.Request.Context(), oid, id, req, ptrUUID(uid))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, r)
}

func (h *Handler) archiveRole(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ArchiveRole(c.Request.Context(), oid, id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listRolePerms(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListRolePermissions(c.Request.Context(), id, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) assignRoles(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	uid := appctx.UserID(c.Request.Context())
	mid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	if err := h.svc.AssignRolesToMembership(c.Request.Context(), oid, mid, req.RoleKeys, ptrUUID(uid)); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) removeRole(c *gin.Context) {
	mid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	rid, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid roleId", err))
		return
	}
	if err := h.svc.RemoveRoleFromMembership(c.Request.Context(), mid, rid); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listMembershipRoles(c *gin.Context) {
	mid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	rows, err := h.svc.ListMembershipRoles(c.Request.Context(), mid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}
