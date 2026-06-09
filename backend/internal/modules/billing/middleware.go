package billing

import (
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
