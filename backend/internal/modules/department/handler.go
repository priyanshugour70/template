package department

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
	d := g.Group("/departments", auth)
	{
		d.GET("", perm("department.list"), h.list)
		d.GET("/tree", perm("department.list"), h.tree)
		d.POST("", perm("department.create"), h.create)
		d.GET("/:id", perm("department.read"), h.get)
		d.PATCH("/:id", perm("department.update"), h.update)
		d.POST("/:id/move", perm("department.update"), h.move)
		d.DELETE("/:id", perm("department.delete"), h.delete)
		d.GET("/:id/roles", perm("department.read"), h.listRoles)
		d.PUT("/:id/roles", perm("department.assign"), h.assignRoles)
	}
}

func (h *Handler) orgID(c *gin.Context) (uuid.UUID, bool) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization is required", nil))
		return uuid.Nil, false
	}
	return oid, true
}

func (h *Handler) list(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	p := pagination.FromGin(c)
	rows, total, err := h.svc.List(c.Request.Context(), oid, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) tree(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	tree, err := h.svc.Tree(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tree)
}

func (h *Handler) create(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	tid := appctx.TenantID(c.Request.Context())
	d, err := h.svc.Create(c.Request.Context(), tid, oid, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, d)
}

func (h *Handler) get(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	d, err := h.svc.repo.Get(c.Request.Context(), oid, id)
	if err != nil {
		if IsNotFound(err) {
			response.Error(c, apperr.New(apperr.CodeNotFound, "department not found", nil))
			return
		}
		response.Error(c, apperr.New(apperr.CodeInternal, "load department failed", err))
		return
	}
	response.OK(c, d)
}

func (h *Handler) update(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	d, err := h.svc.Update(c.Request.Context(), oid, id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, d)
}

func (h *Handler) move(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in MoveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	if err := h.svc.Move(c.Request.Context(), oid, id, in.ParentID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) delete(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), oid, id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listRoles(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	ids, err := h.svc.ListRoles(c.Request.Context(), oid, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, ids)
}

func (h *Handler) assignRoles(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in AssignRolesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	uid := appctx.UserID(c.Request.Context())
	var by *uuid.UUID
	if uid != uuid.Nil {
		u := uid
		by = &u
	}
	if err := h.svc.AssignRoles(c.Request.Context(), oid, id, in.RoleIDs, by); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
