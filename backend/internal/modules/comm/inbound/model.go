// Package inbound implements the public POST /comm/inbound/:token endpoint
// that lets external systems (GitHub, Sentry, custom services) drop messages
// into channels. Each channel can have N inbound hooks; each hook is a
// secret token bound to one (tenant, org, conversation).
//
// Auth model:
//   - Hook CRUD endpoints under /comm/conversations/:id/hooks live in the
//     comm.Handler — gated by the `comm.inbound_hook.manage` permission.
//   - The public POST endpoint /comm/inbound/:token has NO auth middleware.
//     The token in the URL is the credential. Look it up by SHA256 hash; the
//     plaintext token is never persisted.
//
// Phase 2 is intentionally minimal: text body + username/icon override + raw
// attachments JSON. Phase 4 can layer Slack-style block kit / mrkdwn parsing.
package inbound

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ChannelHook is a single inbound webhook on one conversation. The Token is
// only known at creation time — we store its SHA256 hash and return the
// plaintext exactly once in the create response.
type ChannelHook struct {
	model.Base
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index"     json:"tenantId"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index"     json:"organizationId"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index"     json:"conversationId"`

	Name        string `gorm:"not null"      json:"name"`
	IconURL     string `gorm:"column:icon_url" json:"iconUrl,omitempty"`
	DisplayName string `                       json:"displayName,omitempty"`

	TokenHash []byte `gorm:"type:bytea;not null;uniqueIndex" json:"-"`

	IsActive   bool       `gorm:"not null;default:true" json:"isActive"`
	LastUsedAt *time.Time `                                json:"lastUsedAt,omitempty"`
	UseCount   int64      `gorm:"not null;default:0"    json:"useCount"`
}

func (ChannelHook) TableName() string { return "comm_channel_hooks" }

// ── DTOs ──────────────────────────────────────────────────────────────────

type CreateHookRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=100"`
	IconURL     string `json:"iconUrl,omitempty"     binding:"omitempty,url,max=2048"`
	DisplayName string `json:"displayName,omitempty" binding:"omitempty,max=100"`
}

// CreateHookResponse exposes the plaintext token EXACTLY ONCE. Subsequent
// reads of the hook (via list/get) return only the hashed-only HookView.
type CreateHookResponse struct {
	Hook  ChannelHook `json:"hook"`
	Token string      `json:"token"`
	URL   string      `json:"url"` // fully-qualified POST endpoint
}

// InboundMessageRequest is the JSON shape external senders POST to the public
// endpoint. Loosely Slack-compatible so existing wiring works without
// rewrites. `text` is the only required field.
type InboundMessageRequest struct {
	Text        string                   `json:"text"        binding:"required,min=1,max=10000"`
	Username    string                   `json:"username,omitempty"    binding:"omitempty,max=100"`
	IconURL     string                   `json:"icon_url,omitempty"   binding:"omitempty,url,max=2048"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
}
