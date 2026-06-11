package ws

import (
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Hub is the in-process registry of live WebSocket connections. There is one
// hub per backend instance; cross-instance delivery flows through Redis (see
// pubsub.go).
//
// Two indexes are maintained:
//
//   - byUser[userID] → set of conns belonging to that user (multiple tabs,
//     mobile + desktop, etc). Used for presence broadcasts and the rare
//     "force-close a user's sessions" admin action.
//   - byConv[convID] → set of conns currently subscribed to that conversation.
//     Used as the primary delivery path for message/typing/read events.
//
// Subscription is per-connection, not per-user — a user open on two tabs
// with different conversations active gets the right events on the right
// tabs without cross-talk.
type Hub struct {
	mu     sync.RWMutex
	byUser map[uuid.UUID]map[*Conn]struct{}
	byConv map[uuid.UUID]map[*Conn]struct{}
	log    *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		byUser: make(map[uuid.UUID]map[*Conn]struct{}),
		byConv: make(map[uuid.UUID]map[*Conn]struct{}),
		log:    log,
	}
}

// Register adds a fresh connection to the byUser index. Subscriptions follow
// later via the client's `subscribe` events. The caller (handler.go) holds
// the conn's lifetime — Hub never closes connections, only forgets them.
func (h *Hub) Register(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.byUser[c.userID]; !ok {
		h.byUser[c.userID] = make(map[*Conn]struct{})
	}
	h.byUser[c.userID][c] = struct{}{}
}

// Unregister removes the connection from every index it appears in. Safe to
// call on an already-unregistered conn — concurrent disconnect + admin force-
// close shouldn't double-free anything.
func (h *Hub) Unregister(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.byUser[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.byUser, c.userID)
		}
	}
	for convID := range c.subs {
		if conns, ok := h.byConv[convID]; ok {
			delete(conns, c)
			if len(conns) == 0 {
				delete(h.byConv, convID)
			}
		}
	}
}

// Subscribe adds the conn to the byConv index for fast fan-out.
func (h *Hub) Subscribe(c *Conn, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.byConv[convID]; !ok {
		h.byConv[convID] = make(map[*Conn]struct{})
	}
	h.byConv[convID][c] = struct{}{}
	c.subs[convID] = struct{}{}
}

// Unsubscribe removes the conv from this conn's subscription set.
func (h *Hub) Unsubscribe(c *Conn, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.byConv[convID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.byConv, convID)
		}
	}
	delete(c.subs, convID)
}

// BroadcastToConversation pushes the frame to every conn subscribed to convID.
// `excludeUser` (if not Nil) skips connections owned by that user — used so a
// typer doesn't see their own typing indicator echoed back.
//
// Delivery is non-blocking: if a conn's send buffer is full, the frame is
// dropped for that conn and its writer goroutine will close the socket. We
// prefer dropping the slow client over stalling the rest of the fan-out.
func (h *Hub) BroadcastToConversation(convID uuid.UUID, frame ServerFrame, excludeUser *uuid.UUID) {
	h.mu.RLock()
	targets := h.byConv[convID]
	conns := make([]*Conn, 0, len(targets))
	for c := range targets {
		if excludeUser != nil && c.userID == *excludeUser {
			continue
		}
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	for _, c := range conns {
		c.tryEnqueue(frame)
	}
}

// BroadcastToUsers pushes the frame to every conn of every userID in the
// list. Used by presence events ("Alice came online — tell everyone in her
// org"). Order across users is unspecified.
func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, frame ServerFrame) {
	h.mu.RLock()
	conns := make([]*Conn, 0, len(userIDs))
	for _, uid := range userIDs {
		for c := range h.byUser[uid] {
			conns = append(conns, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range conns {
		c.tryEnqueue(frame)
	}
}

// LocalUserCount returns the number of distinct users currently connected
// to THIS instance — useful for health endpoints and rough metrics.
func (h *Hub) LocalUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byUser)
}

// LocalConnCount returns total connections (multiple tabs per user counted
// separately).
func (h *Hub) LocalConnCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, set := range h.byUser {
		n += len(set)
	}
	return n
}

// IsUserConnected reports whether at least one connection from this user is
// active on this instance. Used by the notifier path to decide whether to
// suppress an in-app notification (the user is actively watching, no need
// to ring the bell on top of the WS message they just saw).
func (h *Hub) IsUserConnected(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.byUser[userID]
	return ok
}
