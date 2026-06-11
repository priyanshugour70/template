package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Conn wraps a single live WebSocket. Two goroutines run per conn:
//
//   - readPump:  parse incoming ClientFrames, react (subscribe / typing / ping)
//   - writePump: serialize outbound ServerFrames from the send channel
//
// Outbound events arrive on `send` (a buffered channel). The hub pushes
// frames there via tryEnqueue; writePump pulls and writes. If send fills up
// the conn is too slow — we drop further frames and close the socket so the
// client reconnects clean.
type Conn struct {
	id       uuid.UUID
	ws       *websocket.Conn
	userID   uuid.UUID
	tenantID uuid.UUID
	orgID    uuid.UUID

	hub *Hub
	out OutboundPort // typing/subscribe events to publish upstream

	mu   sync.RWMutex // guards subs
	subs map[uuid.UUID]struct{}

	send chan ServerFrame

	closeOnce sync.Once
	done      chan struct{}

	log *zap.Logger
}

// OutboundPort lets the conn publish events sourced from its read loop
// (typing throttle, etc) into the cross-instance pub/sub. The handler
// supplies an implementation backed by Redis Publish. Keeping it an
// interface lets unit tests stub the network.
type OutboundPort interface {
	PublishTyping(ctx context.Context, convID, userID uuid.UUID, untilMS int64) error
}

const (
	// sendBufferSize is per-connection. Bursts of 64 outbound frames buffered
	// while the write goroutine catches up. Anything past that drops events;
	// the client reconnects and replays via REST.
	sendBufferSize = 64

	// pingInterval is how often the server sends a WS-level Ping frame to
	// keep proxies happy. Client may also send ClientTypePing for the same
	// purpose; both refresh the read deadline.
	pingInterval = 25 * time.Second

	// readDeadline must be longer than pingInterval — a client that misses a
	// few keepalives gets force-disconnected.
	readDeadline = pingInterval*2 + 10*time.Second

	// typingThrottle silences typing emits to once per N seconds per
	// (conn, conversation). The client may spam events on every keystroke.
	typingThrottle = 2 * time.Second

	// typingDisplayWindow is how long the indicator stays visible to other
	// users without a refresh — tells them when to stop showing it if the
	// typer goes silent.
	typingDisplayWindow = 5 * time.Second
)

// NewConn registers a fresh upgraded socket. Caller is responsible for
// starting the pumps (Run) — the function returns immediately so the
// HTTP handler can release the request.
func NewConn(ws *websocket.Conn, hub *Hub, out OutboundPort, t Ticket, log *zap.Logger) *Conn {
	return &Conn{
		id:       uuid.New(),
		ws:       ws,
		userID:   t.UserID,
		tenantID: t.TenantID,
		orgID:    t.OrganizationID,
		hub:      hub,
		out:      out,
		subs:     make(map[uuid.UUID]struct{}),
		send:     make(chan ServerFrame, sendBufferSize),
		done:     make(chan struct{}),
		log:      log,
	}
}

// UserID exposes the principal for cross-package consumers (presence
// tracker needs to know which user just connected).
func (c *Conn) UserID() uuid.UUID { return c.userID }

// OrgID exposes the org for presence broadcasts targeted at org members.
func (c *Conn) OrgID() uuid.UUID { return c.orgID }

// Run starts the read and write pumps and blocks until either exits. The
// hub is responsible for calling Unregister AFTER Run returns — the conn
// itself just signals "I'm done".
func (c *Conn) Run(ctx context.Context) {
	// Send the hello frame immediately so clients can confirm auth.
	c.tryEnqueue(NewHello(c.userID))

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go c.writePump(rctx)
	c.readPump(rctx)
}

// Close terminates the underlying socket. Idempotent.
func (c *Conn) Close(reason string) {
	c.closeOnce.Do(func() {
		_ = c.ws.Close(websocket.StatusNormalClosure, reason)
		close(c.done)
	})
}

// tryEnqueue is the only entry point the hub uses to push frames. Non-
// blocking: drops the frame and closes the socket if the send buffer is
// full. Slow clients become disconnected clients — they reconnect and
// re-fetch history via REST.
func (c *Conn) tryEnqueue(frame ServerFrame) {
	select {
	case c.send <- frame:
	default:
		c.log.Warn("comm/ws: send buffer full, dropping conn",
			zap.String("connId", c.id.String()),
			zap.String("userId", c.userID.String()),
		)
		c.Close("send buffer overflow")
	}
}

// ── read pump ─────────────────────────────────────────────────────────────

// readPump consumes ClientFrames. It also maintains per-conv typing throttle
// state. Errors close the connection cleanly.
func (c *Conn) readPump(ctx context.Context) {
	defer c.Close("read loop exit")
	lastTyping := make(map[uuid.UUID]time.Time)

	for {
		rctx, cancel := context.WithTimeout(ctx, readDeadline)
		_, data, err := c.ws.Read(rctx)
		cancel()
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				c.log.Debug("comm/ws: read closed", zap.Error(err))
			}
			return
		}
		var frame ClientFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			c.tryEnqueue(NewError("invalid_json", "could not parse frame"))
			continue
		}
		switch frame.Type {
		case ClientTypePing:
			c.tryEnqueue(NewPong())
		case ClientTypeSubscribe:
			if frame.ConversationID == nil {
				c.tryEnqueue(NewError("missing_field", "conversationId required"))
				continue
			}
			c.hub.Subscribe(c, *frame.ConversationID)
		case ClientTypeUnsubscribe:
			if frame.ConversationID == nil {
				continue
			}
			c.hub.Unsubscribe(c, *frame.ConversationID)
		case ClientTypeTyping:
			if frame.ConversationID == nil {
				continue
			}
			now := time.Now()
			if last, ok := lastTyping[*frame.ConversationID]; ok && now.Sub(last) < typingThrottle {
				continue // throttled
			}
			lastTyping[*frame.ConversationID] = now
			until := now.Add(typingDisplayWindow)
			// Server-side membership check would be ideal but we deliberately
			// skip it here: typing is broadcast-fire-and-forget; the
			// downstream subscribers are by definition members of that conv,
			// and a non-member emitting typing would simply not be received.
			if c.out != nil {
				if err := c.out.PublishTyping(ctx, *frame.ConversationID, c.userID, until.UnixMilli()); err != nil {
					c.log.Debug("comm/ws: publish typing failed", zap.Error(err))
				}
			}
		default:
			// Unknown types are silently ignored for forward compat.
		}
	}
}

// ── write pump ────────────────────────────────────────────────────────────

// writePump serializes outbound frames and pings the socket periodically so
// load balancers + browsers don't idle-close. Exits when the conn is closed
// or the underlying socket errors.
func (c *Conn) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			return
		case frame := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			payload, err := json.Marshal(frame)
			if err != nil {
				cancel()
				continue
			}
			if err := c.ws.Write(wctx, websocket.MessageText, payload); err != nil {
				cancel()
				c.Close("write failed")
				return
			}
			cancel()
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := c.ws.Ping(pctx); err != nil {
				cancel()
				c.Close("ping failed")
				return
			}
			cancel()
		}
	}
}
