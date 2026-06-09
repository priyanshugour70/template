package webhook

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
	r := g.Group("/webhooks", auth)
	{
		r.GET("", perm("webhook.read"), h.list)
		r.POST("", perm("webhook.create"), h.create)
		r.GET("/:id", perm("webhook.read"), h.get)
		r.PATCH("/:id", perm("webhook.update"), h.update)
		r.DELETE("/:id", perm("webhook.delete"), h.delete)
		r.GET("/:id/deliveries", perm("webhook.read"), h.deliveries)
		r.POST("/:id/test", perm("webhook.test"), h.test)
	}
}

func (h *Handler) orgID(c *gin.Context) (uuid.UUID, bool) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization required", nil))
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

func (h *Handler) create(c *gin.Context) {
	ctx := c.Request.Context()
	tid := appctx.TenantID(ctx)
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	out, err := h.svc.Create(ctx, tid, oid, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, out)
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
	w, err := h.svc.repo.Get(c.Request.Context(), oid, id)
	if err != nil {
		if IsNotFound(err) {
			response.Error(c, apperr.New(apperr.CodeNotFound, "webhook not found", nil))
			return
		}
		response.Error(c, apperr.New(apperr.CodeInternal, "load webhook failed", err))
		return
	}
	response.OK(c, w)
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
	w, err := h.svc.Update(c.Request.Context(), oid, id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, w)
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

func (h *Handler) deliveries(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	rows, err := h.svc.ListDeliveries(c.Request.Context(), oid, id, 50)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) test(c *gin.Context) {
	oid, ok := h.orgID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var in TestFireInput
	_ = c.ShouldBindJSON(&in)
	out, err := h.svc.TestFire(c.Request.Context(), oid, id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
