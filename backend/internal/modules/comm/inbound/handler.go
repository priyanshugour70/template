package inbound

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

type PermissionFunc func(perm string) gin.HandlerFunc

// Routes mounts BOTH:
//   - /comm/conversations/:id/hooks (authed, perm-gated)
//   - /comm/inbound/:token            (PUBLIC — no auth, no perm)
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	managed := g.Group("/comm/conversations/:id/hooks", auth, perm("comm.inbound_hook.manage"))
	{
		managed.GET("", h.list)
		managed.POST("", h.create)
	}
	g.DELETE("/comm/hooks/:id", auth, perm("comm.inbound_hook.manage"), h.revoke)
	// Public — anyone with the token. No auth middleware.
	g.POST("/comm/inbound/:token", h.receive)
}

func (h *Handler) list(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	rows, err := h.svc.ListHooks(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) create(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req CreateHookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.CreateHook(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, out)
}

func (h *Handler) revoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.RevokeHook(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}

// receive is the PUBLIC endpoint. No auth — the token in the path IS the
// credential. Errors leak only generic shapes so probers can't tell apart
// "wrong token" from "revoked token".
func (h *Handler) receive(c *gin.Context) {
	token := c.Param("token")
	var req InboundMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	msgID, err := h.svc.PostFromExternal(c.Request.Context(), token, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, gin.H{"messageId": msgID, "ok": true})
}
