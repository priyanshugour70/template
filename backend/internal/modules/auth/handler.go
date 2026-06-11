package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/middleware"
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
	a := g.Group("/auth")
	{
		// Per-IP rate limits on public auth endpoints. Tuned for the threat,
		// not for legit users — the login UI may call /discover on every
		// keystroke so it gets a higher cap. Burst (one-minute window) limits.
		a.POST("/discover", middleware.RateLimit(20), h.discover)
		a.POST("/login", middleware.RateLimit(10), h.login)
		a.POST("/refresh", middleware.RateLimit(30), h.refresh)
		a.POST("/forgot-password", middleware.RateLimit(5), h.forgotPassword)
		a.POST("/reset-password", middleware.RateLimit(10), h.resetPassword)
		a.POST("/accept-invite", middleware.RateLimit(5), h.acceptInvite)
		a.POST("/register", middleware.RateLimit(5), h.register)
		// Multi-tenant subdomain support — both public:
		//   /tenant-by-slug : tenant subdomain pages call this to paint brand
		//                     and resolve tenant_id before login.
		//   /check-slug     : signup form calls this for inline availability.
		//   /handoff/consume: tenant subdomain swaps a one-time apex token for
		//                     a real access+refresh pair (sets cookies scoped
		//                     to the subdomain).
		a.GET("/tenant-by-slug", middleware.RateLimit(60), h.tenantBySlug)
		a.GET("/check-slug", middleware.RateLimit(30), h.checkSlug)
		a.POST("/handoff/consume", middleware.RateLimit(20), h.handoffConsume)
	}
	authed := g.Group("/auth", auth)
	{
		authed.POST("/logout", h.logout)
		authed.POST("/switch-org", h.switchOrg)
		authed.POST("/change-password", h.changePassword)
		authed.GET("/sessions", h.listSessions)
		authed.DELETE("/sessions/:jti", h.revokeSession)
		// /handoff/issue is authed — only an already-logged-in apex session can
		// mint a handoff token to redirect to a tenant subdomain.
		authed.POST("/handoff/issue", h.handoffIssue)
	}
	invites := g.Group("/invites", auth)
	{
		invites.POST("", perm("user.invite"), h.invite)
	}
}

// ── handlers ───────────────────────────────────────────────────────────────

func (h *Handler) discover(c *gin.Context) {
	var req DiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	tenants, _ := h.svc.Discover(c.Request.Context(), req.Email)
	response.OK(c, gin.H{"tenants": tenants})
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) logout(c *gin.Context) {
	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)
	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	c.Status(http.StatusNoContent)
}

func (h *Handler) switchOrg(c *gin.Context) {
	var req SwitchOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.SwitchOrg(c.Request.Context(), req.OrganizationID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) forgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	_ = h.svc.ForgotPassword(c.Request.Context(), req.Email, c.ClientIP(), c.Request.UserAgent())
	c.Status(http.StatusAccepted)
}

func (h *Handler) resetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) changePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listSessions(c *gin.Context) {
	p := pagination.FromGin(c)
	rows, total, err := h.svc.ListSessions(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) revokeSession(c *gin.Context) {
	jti, err := uuid.Parse(c.Param("jti"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid jti", err))
		return
	}
	if err := h.svc.RevokeSession(c.Request.Context(), jti); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) invite(c *gin.Context) {
	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	inv, _, err := h.svc.Invite(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, inv)
}

func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.Register(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, out)
}

func (h *Handler) acceptInvite(c *gin.Context) {
	var req AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.AcceptInvite(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ── multi-tenant subdomain handlers ────────────────────────────────────────

func (h *Handler) tenantBySlug(c *gin.Context) {
	slug := c.Query("slug")
	out, err := h.svc.TenantBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) checkSlug(c *gin.Context) {
	out := h.svc.CheckSlug(c.Request.Context(), c.Query("slug"))
	response.OK(c, out)
}

func (h *Handler) handoffIssue(c *gin.Context) {
	out, err := h.svc.IssueHandoff(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) handoffConsume(c *gin.Context) {
	var req HandoffConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.ConsumeHandoff(c.Request.Context(), req.Token, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
