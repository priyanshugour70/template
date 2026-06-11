package comm

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

// RBACPort is the narrow rbac dependency the handler needs — only for the
// author-or-moderator branch on delete. Defined here so the comm package
// stays free of an rbac import.
type RBACPort interface {
	HasPermission(ctx context.Context, userID, orgID uuid.UUID, perm string) (bool, error)
}

type Handler struct {
	svc  *Service
	rbac RBACPort
	log  *zap.Logger
}

func NewHandler(svc *Service, rbac RBACPort, log *zap.Logger) *Handler {
	return &Handler{svc: svc, rbac: rbac, log: log}
}

type PermissionFunc func(perm string) gin.HandlerFunc

// Routes mounts every REST endpoint for the comm module under /comm. The
// top-level perm gate is `comm.read` for any read; writes carry their own
// (`comm.message.send`, `comm.channel.create`, `comm.channel.manage`,
// `comm.message.moderate`).
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	c := g.Group("/comm", auth, perm("comm.read"))
	{
		c.GET("/conversations", h.listConversations)
		c.POST("/conversations/channels", perm("comm.channel.create"), h.createChannel)
		c.POST("/conversations/dms", h.createOrGetDM)
		c.GET("/conversations/:id", h.getConversation)
		c.PATCH("/conversations/:id", perm("comm.channel.manage"), h.updateChannel)
		c.DELETE("/conversations/:id", perm("comm.channel.manage"), h.archiveChannel)

		c.GET("/conversations/:id/members", h.listMembers)
		c.POST("/conversations/:id/members", perm("comm.channel.manage"), h.addMembers)
		c.DELETE("/conversations/:id/members/:userId", h.removeMember)
		c.PATCH("/conversations/:id/members/me", h.updateMyPrefs)

		c.GET("/conversations/:id/messages", h.listMessages)
		c.POST("/conversations/:id/messages", perm("comm.message.send"), h.sendMessage)
		c.POST("/conversations/:id/read", h.markRead)

		c.PATCH("/messages/:id", perm("comm.message.send"), h.editMessage)
		c.DELETE("/messages/:id", h.deleteMessage)

		c.POST("/messages/:id/reactions", perm("comm.message.send"), h.addReaction)
		c.DELETE("/messages/:id/reactions/:emoji", perm("comm.message.send"), h.removeReaction)
	}
}

// ── conversations ─────────────────────────────────────────────────────────

func (h *Handler) listConversations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	f := ListConversationsFilter{
		Type:            c.Query("type"),
		IncludeArchived: c.Query("includeArchived") == "true",
		Limit:           limit,
	}
	rows, err := h.svc.ListMyConversations(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) createChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	conv, err := h.svc.CreateChannel(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, conv)
}

func (h *Handler) createOrGetDM(c *gin.Context) {
	var req CreateDMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	conv, err := h.svc.CreateOrGetDM(c.Request.Context(), req.RecipientUserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	// 200 for idempotent get-or-create — could be an existing DM.
	response.OK(c, conv)
}

func (h *Handler) getConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	view, err := h.svc.GetConversationView(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, view)
}

func (h *Handler) updateChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	conv, err := h.svc.UpdateChannel(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, conv)
}

func (h *Handler) archiveChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.ArchiveChannel(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── members ───────────────────────────────────────────────────────────────

func (h *Handler) listMembers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	rows, err := h.svc.ListConversationMembers(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) addMembers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req AddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	added, err := h.svc.AddMembers(c.Request.Context(), id, req.UserIDs)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, added)
}

func (h *Handler) removeMember(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	uid, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid userId", err))
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), id, uid); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) updateMyPrefs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req UpdateMemberPrefsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.UpdateMyMemberPrefs(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ── messages ──────────────────────────────────────────────────────────────

func (h *Handler) listMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var beforeID uuid.UUID
	if s := c.Query("before"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			beforeID = id
		}
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	rows, err := h.svc.ListMessages(c.Request.Context(), id, beforeID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) sendMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	msg, err := h.svc.SendMessage(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, msg)
}

func (h *Handler) editMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	out, err := h.svc.EditMessage(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// deleteMessage runs author-vs-moderator gating in the service. The route
// has no perm gate (anyone with comm.read could try to delete THEIR OWN
// message) — the service decides. Moderation through this endpoint requires
// the caller's principal to carry comm.message.moderate, checked via
// hasPermission on the gin context (audit middleware logs the action).
func (h *Handler) deleteMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	ctx := c.Request.Context()
	isMod := appctx.IsSuperAdmin(ctx)
	if !isMod && h.rbac != nil {
		ok, _ := h.rbac.HasPermission(ctx, appctx.UserID(ctx), appctx.OrganizationID(ctx), "comm.message.moderate")
		isMod = ok
	}
	if err := h.svc.DeleteMessage(ctx, id, isMod); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) markRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), id, req.LastReadMessageID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── reactions ─────────────────────────────────────────────────────────────

func (h *Handler) addReaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	var req ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid request body", err))
		return
	}
	if err := h.svc.AddReaction(c.Request.Context(), id, req.Emoji); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) removeReaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	emoji := c.Param("emoji")
	if err := h.svc.RemoveReaction(c.Request.Context(), id, emoji); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

