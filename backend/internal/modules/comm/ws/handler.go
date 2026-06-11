package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/modules/comm/presence"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

// Handler mounts /comm/ws/ticket and /comm/ws.
type Handler struct {
	tickets        *TicketStore
	hub            *Hub
	pubsub         *PubSub
	tracker        *presence.Tracker
	log            *zap.Logger
	allowedOrigins []string
}

func NewHandler(
	tickets *TicketStore,
	hub *Hub,
	pubsub *PubSub,
	tracker *presence.Tracker,
	allowedOrigins []string,
	log *zap.Logger,
) *Handler {
	return &Handler{
		tickets:        tickets,
		hub:            hub,
		pubsub:         pubsub,
		tracker:        tracker,
		allowedOrigins: allowedOrigins,
		log:            log,
	}
}

// Routes — ticket issuance is authed via the standard JWT middleware. The WS
// upgrade itself has NO auth middleware: the browser cannot reliably attach
// Authorization headers to WS handshakes, so we authenticate by ticket
// (consumed from the query string within seconds of issue).
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc) {
	g.POST("/comm/ws/ticket", auth, h.issueTicket)
	g.GET("/comm/ws", h.upgrade)
}

// ── ticket endpoint ───────────────────────────────────────────────────────

type ticketResponse struct {
	Ticket    string    `json:"ticket"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (h *Handler) issueTicket(c *gin.Context) {
	ctx := c.Request.Context()
	t := Ticket{
		UserID:         appctx.UserID(ctx),
		TenantID:       appctx.TenantID(ctx),
		OrganizationID: appctx.OrganizationID(ctx),
	}
	if t.UserID == uuid.Nil {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil))
		return
	}
	tok, exp, err := h.tickets.Issue(ctx, t)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInternal, "ticket issue failed", err))
		return
	}
	response.OK(c, ticketResponse{Ticket: tok, ExpiresAt: exp})
}

// ── upgrade endpoint ──────────────────────────────────────────────────────

func (h *Handler) upgrade(c *gin.Context) {
	token := c.Query("ticket")
	if token == "" {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "missing ticket", nil))
		return
	}
	ticket, err := h.tickets.Consume(c.Request.Context(), token)
	if err != nil {
		// Don't leak the difference between "expired" and "never existed".
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "invalid or expired ticket", nil))
		return
	}

	// Origin allow-list — the same CORS rules that gate REST apply here too.
	// AcceptOptions hands us a callback to decide; pass through the configured
	// list. Empty list means: allow same-origin only (the default for the
	// websocket lib).
	acceptOpts := &websocket.AcceptOptions{
		OriginPatterns: h.allowedOrigins,
		// Compression is off by default in coder/websocket; explicit for clarity.
		CompressionMode: websocket.CompressionDisabled,
	}
	wsConn, err := websocket.Accept(c.Writer, c.Request, acceptOpts)
	if err != nil {
		h.log.Warn("comm/ws: upgrade failed", zap.Error(err))
		// Accept already wrote a response on failure.
		return
	}

	conn := NewConn(wsConn, h.hub, h.pubsub, *ticket, h.log)
	h.hub.Register(conn)

	// Presence: mark online. Broadcast only if this is the user's first
	// connection (other tab might already have them online).
	if h.tracker != nil {
		wasOffline, perr := h.tracker.SetOnline(c.Request.Context(), ticket.UserID, ticket.OrganizationID)
		if perr != nil {
			h.log.Warn("comm/ws: set online failed", zap.Error(perr))
		} else if wasOffline {
			h.broadcastPresence(c.Request.Context(), ticket, "online")
		}
	}

	// Hand off to a goroutine so the HTTP handler returns. Run() blocks until
	// the conn closes; cleanup happens after.
	go func() {
		// Use a fresh context — the gin context is cancelled when the HTTP
		// handler returns, which would kill our pumps immediately.
		runCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start a heartbeat goroutine to extend presence TTL.
		if h.tracker != nil {
			hbCtx, hbCancel := context.WithCancel(runCtx)
			defer hbCancel()
			go h.heartbeatLoop(hbCtx, ticket)
		}

		conn.Run(runCtx)

		// Conn closed — deregister and update presence.
		h.hub.Unregister(conn)
		if h.tracker != nil {
			// Only mark offline if NO other conn for this user remains.
			if !h.hub.IsUserConnected(ticket.UserID) {
				if _, err := h.tracker.SetOffline(runCtx, ticket.UserID, ticket.OrganizationID); err != nil {
					h.log.Debug("comm/ws: set offline failed", zap.Error(err))
				}
				h.broadcastPresence(runCtx, ticket, "offline")
			}
		}
	}()
}

func (h *Handler) heartbeatLoop(ctx context.Context, ticket *Ticket) {
	// Heartbeat every PresenceTTL/3 so even with bursty network we keep the
	// key alive comfortably.
	tick := time.NewTicker(presence.PresenceTTL / 3)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := h.tracker.Heartbeat(ctx, ticket.UserID); err != nil {
				h.log.Debug("comm/ws: heartbeat failed", zap.Error(err))
			}
		}
	}
}

// broadcastPresence resolves the org's online members (minus the changing
// user) and pushes a presence event. Skipped if the publisher is nil.
func (h *Handler) broadcastPresence(ctx context.Context, ticket *Ticket, status string) {
	if h.pubsub == nil || h.tracker == nil {
		return
	}
	online, err := h.tracker.OrgOnlineUsers(ctx, ticket.OrganizationID)
	if err != nil {
		return
	}
	// Trim the actor — they don't need their own presence event.
	recipients := make([]uuid.UUID, 0, len(online))
	for _, uid := range online {
		if uid != ticket.UserID {
			recipients = append(recipients, uid)
		}
	}
	frame := NewPresence(ticket.UserID, status)
	_ = h.pubsub.BroadcastToOrganization(ctx, ticket.OrganizationID, frame, recipients)
}

// alias to silence import of net/http via gin types
var _ = http.MethodGet
