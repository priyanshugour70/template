package apikey

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

type PermissionFunc func(perm string) gin.HandlerFunc

func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	r := g.Group("/api-keys", auth)
	{
		r.GET("", perm("api_key.read"), h.list)
		r.POST("", perm("api_key.create"), h.create)
		r.DELETE("/:id", perm("api_key.delete"), h.revoke)
	}
}

func (h *Handler) list(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization required", nil))
		return
	}
	rows, err := h.svc.List(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) create(c *gin.Context) {
	ctx := c.Request.Context()
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization required", nil))
		return
	}
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid payload", err))
		return
	}
	var userPtr *uuid.UUID
	if uid != uuid.Nil {
		u := uid
		userPtr = &u
	}
	out, err := h.svc.Create(ctx, tid, oid, userPtr, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, out)
}

func (h *Handler) revoke(c *gin.Context) {
	ctx := c.Request.Context()
	oid := appctx.OrganizationID(ctx)
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "active organization required", nil))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	uid := appctx.UserID(ctx)
	var by *uuid.UUID
	if uid != uuid.Nil {
		u := uid
		by = &u
	}
	if err := h.svc.Revoke(ctx, oid, id, by); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
