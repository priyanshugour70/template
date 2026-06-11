// Package ws implements the WebSocket layer of the communication module:
// ticket-based auth, an in-process connection hub, per-connection read/write
// goroutines, and Redis pub/sub fan-out for cross-instance delivery.
//
// The protocol is JSON over WebSocket. A small number of inbound types
// (subscribe / unsubscribe / typing / ping) drive a larger outbound vocabulary
// of conversation/message/presence/typing events. ClientFrame and ServerFrame
// are deliberately flat structs with omitempty on every optional field so
// adding a new event type doesn't require a versioned schema bump.
package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ── Inbound (client → server) ─────────────────────────────────────────────

// ClientFrame is what every WebSocket message FROM the browser looks like.
// Unrecognised Type values are silently ignored — the protocol is
// forward-compatible: an older server can run alongside newer clients.
type ClientFrame struct {
	Type           string     `json:"type"`
	ConversationID *uuid.UUID `json:"conversationId,omitempty"`
}

// Inbound type constants. Kept here so the conn read-loop and any future
// docs reference the same canonical strings.
const (
	ClientTypeSubscribe   = "subscribe"
	ClientTypeUnsubscribe = "unsubscribe"
	ClientTypeTyping      = "typing"
	ClientTypePing        = "ping"
)

// ── Outbound (server → client) ────────────────────────────────────────────

// ServerFrame is the union shape for every message we push to clients. The
// `Type` field is mandatory; everything else is per-event. Routing fields
// (ConversationID, UserID) are also part of the payload so a single client
// can demux events without tracking subscription state.
type ServerFrame struct {
	Type string `json:"type"`

	ConversationID *uuid.UUID `json:"conversationId,omitempty"`
	UserID         *uuid.UUID `json:"userId,omitempty"`
	MessageID      *uuid.UUID `json:"messageId,omitempty"`

	// Body payloads marshalled by the publisher. We keep them as raw JSON so
	// the hub's broadcast path stays type-free — the publisher already knows
	// the right shape; the hub just forwards bytes.
	Message      json.RawMessage `json:"message,omitempty"`
	Conversation json.RawMessage `json:"conversation,omitempty"`
	Member       json.RawMessage `json:"member,omitempty"`

	Emoji   string `json:"emoji,omitempty"`
	UntilMS int64  `json:"untilMs,omitempty"`
	Status  string `json:"status,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Outbound type constants — exhaustive list, intentionally narrow so the
// frontend can switch() on every value without a "default: warn" branch.
const (
	ServerTypeMessageCreated     = "message.created"
	ServerTypeMessageUpdated     = "message.updated"
	ServerTypeMessageDeleted     = "message.deleted"
	ServerTypeReactionAdded      = "reaction.added"
	ServerTypeReactionRemoved    = "reaction.removed"
	ServerTypeTyping             = "typing"
	ServerTypePresence           = "presence"
	ServerTypeRead               = "read"
	ServerTypeMemberAdded        = "member.added"
	ServerTypeMemberRemoved      = "member.removed"
	ServerTypeConversationUpdate = "conversation.updated"
	ServerTypePong               = "pong"
	ServerTypeError              = "error"
	ServerTypeHello              = "hello"
)

// ── Constructors ──────────────────────────────────────────────────────────
//
// Each constructor encodes the routing intent (which conversation / user this
// targets) and the payload shape. Callers in the service or hub use these so
// the JSON shape is centralised here and never ad-hoc.

func NewMessageCreated(convID uuid.UUID, message json.RawMessage) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeMessageCreated,
		ConversationID: &convID,
		Message:        message,
	}
}

func NewMessageUpdated(convID uuid.UUID, message json.RawMessage) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeMessageUpdated,
		ConversationID: &convID,
		Message:        message,
	}
}

func NewMessageDeleted(convID, messageID uuid.UUID) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeMessageDeleted,
		ConversationID: &convID,
		MessageID:      &messageID,
	}
}

func NewReactionAdded(convID, messageID, userID uuid.UUID, emoji string) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeReactionAdded,
		ConversationID: &convID,
		MessageID:      &messageID,
		UserID:         &userID,
		Emoji:          emoji,
	}
}

func NewReactionRemoved(convID, messageID, userID uuid.UUID, emoji string) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeReactionRemoved,
		ConversationID: &convID,
		MessageID:      &messageID,
		UserID:         &userID,
		Emoji:          emoji,
	}
}

// NewTyping is sent to every subscriber of convID EXCEPT the typer. The
// until-millis tells the receiver when to stop showing the indicator
// regardless of follow-up messages (typer goes offline mid-keystroke, etc).
func NewTyping(convID, userID uuid.UUID, until time.Time) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeTyping,
		ConversationID: &convID,
		UserID:         &userID,
		UntilMS:        until.UnixMilli(),
	}
}

// NewPresence is broadcast to every user that shares an org with the user
// whose status changed. `status` is one of "online" | "away" | "offline".
func NewPresence(userID uuid.UUID, status string) ServerFrame {
	return ServerFrame{
		Type:   ServerTypePresence,
		UserID: &userID,
		Status: status,
	}
}

// NewRead announces that `userID` read up to `messageID` in `convID`. Used
// for "Seen by X" UI affordances.
func NewRead(convID, userID, messageID uuid.UUID) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeRead,
		ConversationID: &convID,
		UserID:         &userID,
		MessageID:      &messageID,
	}
}

func NewMemberAdded(convID, userID uuid.UUID, member json.RawMessage) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeMemberAdded,
		ConversationID: &convID,
		UserID:         &userID,
		Member:         member,
	}
}

func NewMemberRemoved(convID, userID uuid.UUID) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeMemberRemoved,
		ConversationID: &convID,
		UserID:         &userID,
	}
}

func NewConversationUpdated(convID uuid.UUID, conversation json.RawMessage) ServerFrame {
	return ServerFrame{
		Type:           ServerTypeConversationUpdate,
		ConversationID: &convID,
		Conversation:   conversation,
	}
}

func NewPong() ServerFrame { return ServerFrame{Type: ServerTypePong} }

// NewError is the catch-all server-side error frame. The conn read-loop sends
// these on bad input (unknown conversation, malformed JSON, etc) and then
// CONTINUES — the connection is only force-closed for hard failures.
func NewError(reason, message string) ServerFrame {
	return ServerFrame{Type: ServerTypeError, Reason: reason, Error: message}
}

// NewHello is sent once immediately after a successful upgrade. It echoes the
// authenticated user so clients can confirm the session went through.
func NewHello(userID uuid.UUID) ServerFrame {
	return ServerFrame{Type: ServerTypeHello, UserID: &userID}
}
