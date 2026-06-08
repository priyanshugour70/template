package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

// Middleware bundles permission-check factories. Construct one in bootstrap
// and pass into every other module's Routes().
type Middleware struct {
	svc *Service
}

func NewMiddleware(svc *Service) *Middleware { return &Middleware{svc: svc} }

// RequirePermission denies the request unless the caller has the named
// permission in their active org.
func (m *Middleware) RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if perm == "" {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		uid := appctx.UserID(ctx)
		oid := appctx.OrganizationID(ctx)
		if uid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil))
			c.Abort()
			return
		}
		ok, err := m.svc.HasPermission(ctx, uid, oid, perm)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "missing permission",
				map[string]interface{}{"required": perm}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *Middleware) RequireAnyPermission(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		uid := appctx.UserID(ctx)
		oid := appctx.OrganizationID(ctx)
		if uid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil))
			c.Abort()
			return
		}
		ok, err := m.svc.HasAny(ctx, uid, oid, perms...)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "missing one of required permissions",
				map[string]interface{}{"requiredAny": perms}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *Middleware) RequireAllPermissions(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		uid := appctx.UserID(ctx)
		oid := appctx.OrganizationID(ctx)
		if uid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil))
			c.Abort()
			return
		}
		ok, err := m.svc.HasAll(ctx, uid, oid, perms...)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "missing required permissions",
				map[string]interface{}{"requiredAll": perms}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole denies unless the caller's membership in the active org includes
// the named role.
func (m *Middleware) RequireRole(roleKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		mid := appctx.MembershipID(ctx)
		if mid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeForbidden, "no membership context", nil))
			c.Abort()
			return
		}
		roles, err := m.svc.ListMembershipRoles(ctx, mid)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		for _, r := range roles {
			if r.Key == roleKey {
				c.Next()
				return
			}
		}
		response.Error(c, apperr.NewWithDetails(
			apperr.CodeForbidden, "missing required role",
			map[string]interface{}{"requiredRole": roleKey}, nil,
		))
		c.Abort()
	}
}

// RequireSelfOrPermission allows the request if the path parameter idParam is
// equal to the caller's user ID, or if they hold the named permission.
func (m *Middleware) RequireSelfOrPermission(perm, idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		uid := appctx.UserID(ctx)
		if uid == uuid.Nil {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil))
			c.Abort()
			return
		}
		if appctx.IsSuperAdmin(ctx) {
			c.Next()
			return
		}
		if target, err := uuid.Parse(c.Param(idParam)); err == nil && target == uid {
			c.Next()
			return
		}
		ok, err := m.svc.HasPermission(ctx, uid, appctx.OrganizationID(ctx), perm)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.Error(c, apperr.NewWithDetails(
				apperr.CodeForbidden, "self-access denied and permission missing",
				map[string]interface{}{"required": perm}, nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
