package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/jwt"
	"github.com/your-org/your-service/internal/pkg/response"
)

// AuthRequired verifies the bearer token, builds a Principal, and pushes it
// onto the request context for downstream middleware and handlers.
func AuthRequired(signer *jwt.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearerToken(c)
		if tok == "" {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "missing bearer token", nil))
			c.Abort()
			return
		}
		claims, err := signer.Verify(tok)
		if err != nil {
			code := apperr.CodeInvalidToken
			msg := "invalid token"
			if err == jwt.ErrExpired {
				code = apperr.CodeTokenExpired
				msg = "token expired"
			}
			response.Error(c, apperr.New(code, msg, err))
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(appctx.With(c.Request.Context(), principalFromClaims(c, claims)))
		c.Next()
	}
}

// AuthOptional populates context if a token is present but never aborts.
func AuthOptional(signer *jwt.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearerToken(c)
		if tok == "" {
			c.Next()
			return
		}
		if claims, err := signer.Verify(tok); err == nil {
			c.Request = c.Request.WithContext(appctx.With(c.Request.Context(), principalFromClaims(c, claims)))
		}
		c.Next()
	}
}

func principalFromClaims(c *gin.Context, claims *jwt.Claims) appctx.Principal {
	return appctx.Principal{
		UserID:         claims.UserID,
		TenantID:       claims.TenantID,
		OrganizationID: claims.OrganizationID,
		MembershipID:   claims.MembershipID,
		Email:          claims.Email,
		IP:             c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		JTI:            claims.ID,
		IsSuperAdmin:   claims.IsSuperAdmin,
	}
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
