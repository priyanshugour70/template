package group

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
	gr := g.Group("/groups", auth)
	{
		gr.GET("", perm("group.list"), h.list)
		gr.POST("", perm("group.create"), h.create)
		gr.GET("/:id", perm("group.read"), h.get)
		gr.PATCH("/:id", perm("group.update"), h.update)
		gr.DELETE("/:id", perm("group.delete"), h.delete)
		gr.GET("/:id/members", perm("group.read"), h.listMembers)
		gr.POST("/:id/members", perm("group.assign"), h.addMember)
		gr.DELETE("/:id/members/:memberId", perm("group.assign"), h.removeMember)
		gr.GET("/:id/roles", perm("group.read"), h.listRoles)
		gr.PUT("/:id/roles", perm("group.assign"), h.assignRoles)
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

func (h *Handler) actor(c *gin.Context) *uuid.UUID {
	uid := appctx.UserID(c.Request.Context())
	if uid == uuid.Nil {
		return nil
	}
	u := uid
	return &u
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
	g, err := h.svc.Create(c.Request.Context(), tid, oid, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, g)
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
	g, err := h.svc.repo.Get(c.Request.Context(), oid, id)
	if err != nil {
		if IsNotFound(err) {
			response.Error(c, apperr.New(apperr.CodeNotFound, "group not found", nil))
			return
		}
		response.Error(c, apperr.New(apperr.CodeInternal, "load group failed", err))
		return
	}
	response.OK(c, g)
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
	g, err := h.svc.Update(c.Request.Context(), oid, id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, g)
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

func (h *Handler) listMembers(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListMembers(c.Request.Context(), oid, id, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) addMember(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in AddMemberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	if err := h.svc.AddMember(c.Request.Context(), oid, id, in, h.actor(c)); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) removeMember(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid memberId", err))
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), oid, id, memberID); err != nil {
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
	if err := h.svc.AssignRoles(c.Request.Context(), oid, id, in.RoleIDs, h.actor(c)); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
