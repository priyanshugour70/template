package billing

import (
	"fmt"
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
	// Legacy route paths kept while Phase 8 hasn't shipped the frontend rewrite.
	// All `subscription.*` permission keys were renamed to `billing.*` in
	// migration 012, so we point the guards at the new keys.
	plans := g.Group("/subscription-plans", auth)
	{
		plans.GET("", h.listPlans)
	}
	sub := g.Group("/subscriptions", auth)
	{
		sub.GET("/active", perm("billing.read"), h.getActive)
		sub.POST("/change", perm("billing.manage"), h.changePlan)
		sub.POST("/preview-change", perm("billing.read"), h.previewChange)
		sub.POST("/cancel", perm("billing.cancel"), h.cancel)
		sub.POST("/reactivate", perm("billing.manage"), h.reactivate)
		// Pause/resume are deprecated and dropped from the new model. The
		// permission no longer exists, so wiring them up would 500 the request.
		sub.PATCH("/billing", perm("billing.manage"), h.updateBilling)
		sub.GET("/features", h.featureSet)
		sub.GET("/usage", perm("billing.read"), h.listUsage)
		sub.GET("/invoices", perm("billing.invoice.read"), h.listInvoices)
		sub.GET("/invoices/:id", perm("billing.invoice.read"), h.getInvoice)
		sub.POST("/coupons/validate", perm("billing.read"), h.validateCoupon)
	}

	// New billing routes (Phase 2 + Phase 3). The frontend will move to
	// /api/v1/billing/* in Phase 8; until then the legacy /subscription-plans
	// + /subscriptions routes above stay as aliases.
	bill := g.Group("/billing", auth)
	{
		bill.GET("/features", perm("billing.read"), h.listFeatures)
		bill.POST("/quotations/preview", perm("billing.read"), h.previewQuote)

		// Quotation lifecycle (Phase 3).
		bill.GET("/quotations", perm("billing.quotation.read"), h.listQuotations)
		bill.POST("/quotations", perm("billing.quotation.manage"), h.createQuotation)
		bill.GET("/quotations/:id", perm("billing.quotation.read"), h.getQuotation)
		bill.PATCH("/quotations/:id", perm("billing.quotation.manage"), h.updateQuotation)
		bill.DELETE("/quotations/:id", perm("billing.quotation.manage"), h.deleteQuotation)
		bill.POST("/quotations/:id/activate", perm("billing.quotation.manage"), h.activateQuotation)

		// Invoice PDF (Phase 4). Full /invoices/* CRUD migrates from
		// /subscriptions/invoices in Phase 8.
		bill.GET("/invoices/:id/pdf", perm("billing.invoice.read"), h.getInvoicePDF)

		// Payments + receipts (Phase 5).
		bill.POST("/invoices/:id/pay", perm("billing.invoice.pay"), h.recordPayment)
		bill.GET("/transactions", perm("billing.transaction.read"), h.listTransactions)
		bill.GET("/transactions/:id", perm("billing.transaction.read"), h.getTransaction)
		bill.GET("/receipts/:id/pdf", perm("billing.transaction.read"), h.getReceiptPDF)
	}
}

func (h *Handler) listPlans(c *gin.Context) {
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListPlans(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
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
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListUsage(c.Request.Context(), oid, p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
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

// ── catalog + quote preview (Phase 2) ─────────────────────────────────────

func (h *Handler) listFeatures(c *gin.Context) {
	rows, err := h.svc.ListFeatures(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) previewQuote(c *gin.Context) {
	var req PreviewQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	q, err := h.svc.PreviewQuote(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, q)
}

// ── quotations (Phase 3) ──────────────────────────────────────────────────

func (h *Handler) listQuotations(c *gin.Context) {
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListQuotations(c.Request.Context(), c.Query("status"), p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) createQuotation(c *gin.Context) {
	var req CreateQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	q, err := h.svc.CreateQuotation(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, q)
}

func (h *Handler) getQuotation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	q, err := h.svc.GetQuotation(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, q)
}

func (h *Handler) updateQuotation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	q, err := h.svc.UpdateQuotation(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, q)
}

func (h *Handler) deleteQuotation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.DeleteQuotation(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) activateQuotation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	out, err := h.svc.ActivateQuotation(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// getInvoicePDF streams the invoice PDF. Lazy: renders on first call, caches
// in S3 keyed by tenant + invoice number, returns the cached bytes thereafter.
// ?download=1 forces an attachment Content-Disposition; default is inline so
// the browser opens it in a tab.
func (h *Handler) getInvoicePDF(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	ctx := c.Request.Context()
	oid := appctx.OrganizationID(ctx)
	if oid == uuid.Nil && !appctx.IsSuperAdmin(ctx) {
		response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
		return
	}
	body, err := h.svc.GetInvoicePDF(ctx, oid, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Filename based on invoice number so saved PDFs are self-describing.
	inv, _ := h.svc.GetInvoice(ctx, oid, id)
	filename := "invoice.pdf"
	if inv != nil {
		filename = inv.Number + ".pdf"
	}
	disp := "inline"
	if c.Query("download") == "1" {
		disp = "attachment"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disp, filename))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, "application/pdf", body)
}

// ── payments + receipts (Phase 5) ─────────────────────────────────────────

func (h *Handler) recordPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid invoice id", err))
		return
	}
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.RecordPayment(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, out)
}

func (h *Handler) listTransactions(c *gin.Context) {
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListTransactions(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) getTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	t, err := h.svc.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

// getReceiptPDF streams the receipt PDF. Same lazy-render + S3 cache pattern
// as the invoice endpoint. ?download=1 forces an attachment disposition.
func (h *Handler) getReceiptPDF(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	ctx := c.Request.Context()
	body, err := h.svc.GetReceiptPDF(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	t, _ := h.svc.GetTransaction(ctx, id)
	filename := "receipt.pdf"
	if t != nil {
		filename = t.ReceiptNumber + ".pdf"
	}
	disp := "inline"
	if c.Query("download") == "1" {
		disp = "attachment"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disp, filename))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, "application/pdf", body)
}
