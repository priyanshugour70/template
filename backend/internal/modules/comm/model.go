// Package comm is the Communication module: direct messages, channels, and
// the messages that flow through them.
//
// Tables in this module are prefixed `comm_*`. DMs and channels share one
// `conversations` table with a `type` discriminator — the alternative (two
// tables) duplicates message storage logic for no real upside. Inbound
// webhooks, presence, typing, and the WebSocket hub arrive in Phase 2; this
// file is the REST-only Phase 1 surface.
package comm

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

// Conversation is the umbrella for both DMs and channels. Distinguishing
// fields:
//   - type='dm':       Slug/Name/Topic empty; DMKey holds "userA:userB" with
//                      user IDs sorted lex so a DM between the same pair is
//                      idempotent under the partial unique index on (org_id,
//                      dm_key) WHERE type='dm'.
//   - type='channel':  Slug + Name required; DMKey empty. Slug is unique per
//                      org via the matching partial index.
type Conversation struct {
	model.Base
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenantId"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organizationId"`
	Type           string    `gorm:"not null"                  json:"type"` // dm | channel

	Slug        string `gorm:"type:citext"          json:"slug,omitempty"`
	Name        string `                              json:"name,omitempty"`
	Topic       string `                              json:"topic,omitempty"`
	Description string `                              json:"description,omitempty"`
	IsPrivate   bool   `gorm:"not null;default:false" json:"isPrivate"`

	DMKey string `gorm:"column:dm_key" json:"-"`

	ArchivedAt *time.Time `json:"archivedAt,omitempty"`

	LastMessageAt *time.Time `                                 json:"lastMessageAt,omitempty"`
	MessageCount  int        `gorm:"not null;default:0"        json:"messageCount"`
}

func (Conversation) TableName() string { return "comm_conversations" }

// ConversationMember tracks per-user state inside a conversation: their role,
// when they joined/left, read state (denormalised so the sidebar count stays
// cheap), and per-conversation notification preferences.
type ConversationMember struct {
	model.Base
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversationId"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	MembershipID   *uuid.UUID `gorm:"type:uuid"                 json:"membershipId,omitempty"`
	Role           string     `gorm:"not null;default:member"   json:"role"`

	JoinedAt time.Time  `gorm:"not null;default:now()" json:"joinedAt"`
	LeftAt   *time.Time `                                json:"leftAt,omitempty"`

	LastReadMessageID *uuid.UUID `gorm:"type:uuid"             json:"lastReadMessageId,omitempty"`
	LastReadAt        *time.Time `                              json:"lastReadAt,omitempty"`
	UnreadCount       int        `gorm:"not null;default:0"    json:"unreadCount"`

	NotificationPref string     `gorm:"not null;default:all" json:"notificationPref"`
	MutedUntil       *time.Time `                              json:"mutedUntil,omitempty"`
}

func (ConversationMember) TableName() string { return "comm_conversation_members" }

// Message is a single post in a conversation. SenderType discriminates:
//   - 'user':    SenderUserID is set, real-person message
//   - 'system':  both sender_user_id and sender_webhook_id null, sender_user_id
//                may carry the actor; rendered with system-message styling
//   - 'webhook': SenderWebhookID is set (Phase 2 inbound feature)
//
// Body is the original markdown source. Soft delete: body becomes "[deleted]"
// at the API layer when DeletedAt is non-null but the row is kept so threads
// don't break.
type Message struct {
	model.Base
	ConversationID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversationId"`
	TenantID        uuid.UUID  `gorm:"type:uuid;not null"        json:"tenantId"`
	OrganizationID  uuid.UUID  `gorm:"type:uuid;not null"        json:"organizationId"`
	ParentMessageID *uuid.UUID `gorm:"type:uuid"                  json:"parentMessageId,omitempty"`

	SenderType      string     `gorm:"not null"   json:"senderType"`
	SenderUserID    *uuid.UUID `gorm:"type:uuid"  json:"senderUserId,omitempty"`
	SenderWebhookID *uuid.UUID `gorm:"type:uuid"  json:"senderWebhookId,omitempty"`

	Body        string      `gorm:"not null;default:''"             json:"body"`
	BodyFormat  string      `gorm:"not null;default:markdown"       json:"bodyFormat"`
	Attachments model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb" json:"attachments"`
	Metadata    model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`

	EditedAt *time.Time `json:"editedAt,omitempty"`
}

func (Message) TableName() string { return "comm_messages" }

// MessageMention records one resolved mention. MentionType:
//   - 'user':      TargetUserID is the resolved user
//   - 'here':      notify online members only (Phase 2 needs presence)
//   - 'channel':   notify every channel member
//   - 'everyone':  notify every org member (admin-gated in service)
type MessageMention struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID    uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"messageId"`
	MentionType  string     `gorm:"not null"                                       json:"mentionType"`
	TargetUserID *uuid.UUID `gorm:"type:uuid"                                       json:"targetUserId,omitempty"`
	IndexInBody  int        `gorm:"not null;default:0"                             json:"indexInBody"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"                         json:"createdAt"`
}

func (MessageMention) TableName() string { return "comm_message_mentions" }

// MessageReaction is one emoji from one user on one message. Unique
// constraint (message_id, user_id, emoji) makes "add reaction" idempotent.
type MessageReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"messageId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"                              json:"userId"`
	Emoji     string    `gorm:"not null"                                        json:"emoji"`
	CreatedAt time.Time `gorm:"not null;default:now()"                         json:"createdAt"`
}

func (MessageReaction) TableName() string { return "comm_message_reactions" }

// ── Request DTOs ───────────────────────────────────────────────────────────

type CreateChannelRequest struct {
	Slug        string      `json:"slug"        binding:"required,min=2,max=64"`
	Name        string      `json:"name"        binding:"required,min=1,max=200"`
	Topic       string      `json:"topic,omitempty"       binding:"omitempty,max=500"`
	Description string      `json:"description,omitempty" binding:"omitempty,max=2000"`
	IsPrivate   bool        `json:"isPrivate"`
	MemberIDs   []uuid.UUID `json:"memberIds,omitempty"`
}

type CreateDMRequest struct {
	RecipientUserID uuid.UUID `json:"recipientUserId" binding:"required"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty"        binding:"omitempty,min=1,max=200"`
	Topic       *string `json:"topic,omitempty"       binding:"omitempty,max=500"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=2000"`
	IsPrivate   *bool   `json:"isPrivate,omitempty"`
}

type AddMembersRequest struct {
	UserIDs []uuid.UUID `json:"userIds" binding:"required,min=1"`
}

type SendMessageRequest struct {
	Body            string                   `json:"body"            binding:"required,min=1,max=10000"`
	BodyFormat      string                   `json:"bodyFormat,omitempty" binding:"omitempty,oneof=markdown plain"`
	ParentMessageID *uuid.UUID               `json:"parentMessageId,omitempty"`
	Attachments     []map[string]interface{} `json:"attachments,omitempty"`
}

type EditMessageRequest struct {
	Body string `json:"body" binding:"required,min=1,max=10000"`
}

type MarkReadRequest struct {
	LastReadMessageID uuid.UUID `json:"lastReadMessageId" binding:"required"`
}

type UpdateMemberPrefsRequest struct {
	NotificationPref *string    `json:"notificationPref,omitempty" binding:"omitempty,oneof=all mentions none"`
	MutedUntil       *time.Time `json:"mutedUntil,omitempty"`
}

type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=64"`
}

// ── List filters ──────────────────────────────────────────────────────────

type ListConversationsFilter struct {
	// Type filters the result set: empty = both DMs and channels; 'dm' or
	// 'channel' to scope to one type. Powers the sidebar's "recent 5 chats"
	// and "recent 5 channels" sections.
	Type string
	// IncludeArchived defaults to false. The dashboard "all conversations"
	// page can flip it to true for an archive view.
	IncludeArchived bool
	// Limit applied after sorting by last_message_at DESC.
	Limit int
}

// ── Response shapes ───────────────────────────────────────────────────────

// ConversationView wraps a Conversation with hydrated members and the
// caller's own membership state. Used by GET /conversations/:id.
type ConversationView struct {
	Conversation
	Members      []ConversationMemberView `json:"members,omitempty"`
	MyMembership *ConversationMemberView  `json:"myMembership,omitempty"`
}

// ConversationMemberView denormalises the user record alongside the
// membership so the client doesn't have to do an N+1 fetch for avatars/names.
type ConversationMemberView struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"userId"`
	Role              string     `json:"role"`
	JoinedAt          time.Time  `json:"joinedAt"`
	LastReadMessageID *uuid.UUID `json:"lastReadMessageId,omitempty"`
	UnreadCount       int        `json:"unreadCount"`
	NotificationPref  string     `json:"notificationPref"`

	UserEmail       string `json:"userEmail,omitempty"`
	UserDisplayName string `json:"userDisplayName,omitempty"`
	UserAvatarURL   string `json:"userAvatarUrl,omitempty"`
}

// MessageView attaches mentions, reactions, and sender hydration to a
// Message. The API ALWAYS returns the view shape — the bare Message is for
// internal use only.
type MessageView struct {
	Message
	Mentions  []MessageMention  `json:"mentions,omitempty"`
	Reactions []MessageReaction `json:"reactions,omitempty"`

	SenderDisplayName string `json:"senderDisplayName,omitempty"`
	SenderAvatarURL   string `json:"senderAvatarUrl,omitempty"`
}
