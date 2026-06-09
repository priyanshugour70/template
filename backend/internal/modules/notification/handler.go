package notification

import (
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

// Routes mounts the notification endpoints. Reads are gated on
// `notification.read` so unscoped tokens can't fetch the list.
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, perm PermissionFunc) {
	n := g.Group("/notifications", auth)
	{
		n.GET("", h.list)
		n.GET("/unread-count", h.unreadCount)
		n.POST("/:id/read", h.markRead)
		n.POST("/mark-all-read", h.markAllRead)
	}
}

func (h *Handler) list(c *gin.Context) {
	ctx := c.Request.Context()
	uid := appctx.UserID(ctx)
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "authentication required", nil))
		return
	}
	p := pagination.FromGin(c)
	filter := ListFilter{
		UnreadOnly: c.Query("unread") == "true" || c.Query("unread") == "1",
	}
	rows, total, err := h.svc.List(ctx, uid, filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedOK(c, rows, p.Page, p.Limit, int(total))
}

func (h *Handler) unreadCount(c *gin.Context) {
	ctx := c.Request.Context()
	uid := appctx.UserID(ctx)
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "authentication required", nil))
		return
	}
	n, err := h.svc.UnreadCount(ctx, uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"unreadCount": n})
}

func (h *Handler) markRead(c *gin.Context) {
	ctx := c.Request.Context()
	uid := appctx.UserID(ctx)
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "authentication required", nil))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeValidation, "invalid id", err))
		return
	}
	if err := h.svc.MarkRead(ctx, uid, id); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) markAllRead(c *gin.Context) {
	ctx := c.Request.Context()
	uid := appctx.UserID(ctx)
	if uid == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "authentication required", nil))
		return
	}
	n, err := h.svc.MarkAllRead(ctx, uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"updated": n})
}
