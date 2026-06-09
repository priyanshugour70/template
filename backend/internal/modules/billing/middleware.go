package billing

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

// Middleware bundles feature/quota gating factories that the bootstrap injects
// into every other module's Routes.
type Middleware struct {
	svc *Service
}

func NewMiddleware(svc *Service) *Middleware { return &Middleware{svc: svc} }

// RequireFeature denies the request unless the active org's subscription
// includes the named feature.
func (m *Middleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if feature == "" {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		oid := appctx.OrganizationID(ctx)
		if oid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
			c.Abort()
			return
		}
		ok, err := m.svc.HasFeature(ctx, oid, feature)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "feature not available in current plan",
				map[string]interface{}{"feature": feature}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// BillingGate locks the API down when the org's subscription is no longer in
// a paying state (expired or cancelled). Read-only access stays open so the
// customer can still see invoices + transactions and download a receipt.
//
// Decision flow per request:
//   1. Anonymous calls / GETs pass through unchanged.
//   2. /auth/*, /billing/*, /health, /swagger/* always pass — customers must
//      be able to authenticate and pay to unlock themselves.
//   3. Super-admin bypass (we still need the back-office to work even while
//      a customer is locked out).
//   4. Lookup the active subscription. expired / cancelled → 402.
//
// Returning 402 (instead of 403) lets the frontend route handler distinguish
// "billing problem — show the unlock banner" from "permission problem — show
// the access-denied page".
func (m *Middleware) BillingGate() gin.HandlerFunc {
	allowedPrefixes := []string{
		"/api/v1/billing",
		"/api/v1/auth",
		"/api/v1/subscriptions",       // legacy aliases used by current frontend
		"/api/v1/subscription-plans",
		"/health",
		"/swagger",
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		for _, p := range allowedPrefixes {
			if strings.HasPrefix(path, p) {
				c.Next()
				return
			}
		}
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		oid := appctx.OrganizationID(ctx)
		if oid == uuid.Nil {
			// No org context = unauth'd or token-only request — let downstream
			// auth middleware decide whether to allow.
			c.Next()
			return
		}
		sub, err := m.svc.GetActive(ctx, oid)
		if err != nil {
			// No active subscription either — fail closed for non-GET so a
			// newly-onboarded tenant without a plan can't bypass billing.
			response.Fail(c, http.StatusPaymentRequired, "BILLING_INACTIVE",
				"Choose a plan to continue", nil)
			c.Abort()
			return
		}
		switch sub.Status {
		case "expired", "cancelled", "past_due":
			response.Fail(c, http.StatusPaymentRequired, "BILLING_INACTIVE",
				"Your billing is "+sub.Status+". Settle the open invoice to continue.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireWithinQuota denies if the named quota is exhausted. Defaults to
// checking with by=1; bigger requests should call svc.IsWithinQuota directly.
func (m *Middleware) RequireWithinQuota(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		tid := appctx.TenantID(ctx)
		oid := appctx.OrganizationID(ctx)
		if oid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeForbidden, "no org context", nil))
			c.Abort()
			return
		}
		ok, err := m.svc.IsWithinQuota(ctx, tid, oid, key, 1)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "quota exceeded",
				map[string]interface{}{"quota": key}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
