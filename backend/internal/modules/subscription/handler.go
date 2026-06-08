package subscription

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
	plans := g.Group("/subscription-plans", auth)
	{
		plans.GET("", h.listPlans)
	}
	sub := g.Group("/subscriptions", auth)
	{
		sub.GET("/active", perm("subscription.read"), h.getActive)
		sub.POST("/change", perm("subscription.update"), h.changePlan)
		sub.POST("/cancel", perm("subscription.cancel"), h.cancel)
		sub.GET("/features", h.featureSet)
		sub.GET("/usage", perm("subscription.read"), h.listUsage)
	}
}

func (h *Handler) listPlans(c *gin.Context) {
	rows, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) getActive(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	sub, err := h.svc.GetActive(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) changePlan(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req ChangePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	sub, err := h.svc.ChangePlan(c.Request.Context(), oid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) cancel(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req CancelRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.Cancel(c.Request.Context(), oid, req); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) featureSet(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.OK(c, &FeatureSet{Features: map[string]bool{}, Limits: map[string]int64{}})
		return
	}
	fs, err := h.svc.ResolveFeatureSet(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, fs)
}

func (h *Handler) listUsage(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	rows, err := h.svc.ListUsage(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}
