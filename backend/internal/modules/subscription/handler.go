package subscription

import (
	"fmt"
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
		sub.POST("/preview-change", perm("subscription.read"), h.previewChange)
		sub.POST("/cancel", perm("subscription.cancel"), h.cancel)
		sub.POST("/reactivate", perm("subscription.update"), h.reactivate)
		sub.POST("/pause", perm("subscription.pause"), h.pause)
		sub.POST("/resume", perm("subscription.pause"), h.resume)
		sub.PATCH("/billing", perm("subscription.update"), h.updateBilling)
		sub.GET("/features", h.featureSet)
		sub.GET("/usage", perm("subscription.read"), h.listUsage)
		sub.GET("/invoices", perm("invoice.read"), h.listInvoices)
		sub.GET("/invoices/:id", perm("invoice.read"), h.getInvoice)
		sub.POST("/coupons/validate", perm("subscription.read"), h.validateCoupon)
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
	sub, inv, err := h.svc.ChangePlanWithInvoice(c.Request.Context(), oid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"subscription": sub, "invoice": inv})
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

// ── lifecycle ─────────────────────────────────────────────────────────────

func (h *Handler) previewChange(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req PreviewChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.PreviewChange(c.Request.Context(), oid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) reactivate(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	sub, err := h.svc.Reactivate(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) pause(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req PauseRequest
	_ = c.ShouldBindJSON(&req)
	sub, err := h.svc.Pause(c.Request.Context(), oid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) resume(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	sub, err := h.svc.Resume(c.Request.Context(), oid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) updateBilling(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	var req UpdateBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	sub, err := h.svc.UpdateBilling(c.Request.Context(), oid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sub)
}

// ── invoices ───────────────────────────────────────────────────────────────

func (h *Handler) listInvoices(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	limit := 50
	if v := c.Query("limit"); v != "" {
		// silent fallback to default on parse error
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	rows, err := h.svc.ListInvoices(c.Request.Context(), oid, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) getInvoice(c *gin.Context) {
	oid := appctx.OrganizationID(c.Request.Context())
	if oid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	inv, err := h.svc.GetInvoice(c.Request.Context(), oid, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inv)
}

// ── coupons ────────────────────────────────────────────────────────────────

func (h *Handler) validateCoupon(c *gin.Context) {
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	response.OK(c, h.svc.ValidateCoupon(c.Request.Context(), req))
}
