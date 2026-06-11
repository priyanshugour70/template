package comm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// DB exposes the underlying gorm handle for the few service methods that
// need transactions across multiple repos (create-channel-with-members).
func (r *Repository) DB() *gorm.DB { return r.db }

// ── conversations ──────────────────────────────────────────────────────────

func (r *Repository) CreateConversation(ctx context.Context, c *Conversation) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) GetConversation(ctx context.Context, orgID, id uuid.UUID) (*Conversation, error) {
	var c Conversation
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if orgID != uuid.Nil {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GetChannelBySlug looks up a non-DM conversation by its slug, scoped to the
// caller's org. Used by both the API (GET /channels/:slug) and the reserved-
// slug check on create.
func (r *Repository) GetChannelBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*Conversation, error) {
	var c Conversation
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND type = 'channel' AND slug = ?", orgID, slug).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GetDMByKey returns an existing 1:1 DM between the two users in the org or
// gorm.ErrRecordNotFound. The dmKey must be the canonical sorted "userA:userB"
// form produced by buildDMKey in the service.
func (r *Repository) GetDMByKey(ctx context.Context, orgID uuid.UUID, dmKey string) (*Conversation, error) {
	var c Conversation
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND type = 'dm' AND dm_key = ?", orgID, dmKey).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateConversation patches by id. Used for rename/topic/archive.
func (r *Repository) UpdateConversation(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	res := r.db.WithContext(ctx).
		Model(&Conversation{}).
		Where("id = ?", id).
		Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ArchiveConversation soft-archives (sets archived_at) so it disappears from
// sidebars but stays queryable via the all-conversations view.
func (r *Repository) ArchiveConversation(ctx context.Context, id uuid.UUID) error {
	return r.UpdateConversation(ctx, id, map[string]interface{}{"archived_at": time.Now()})
}

// ListConversationsForUser returns conversations the user is a member of,
// sorted by last activity. Filter knobs power the sidebar dropdowns:
//
//	type=""        → both DMs and channels (default for the "all" view)
//	type="dm"      → just DMs
//	type="channel" → just channels (sidebar's "Channels ▾ recent 5")
//	limit          → caller decides; 0 means no limit
func (r *Repository) ListConversationsForUser(
	ctx context.Context,
	orgID, userID uuid.UUID,
	f ListConversationsFilter,
) ([]ConversationListItem, error) {
	q := r.db.WithContext(ctx).
		Table("comm_conversations c").
		Joins("JOIN comm_conversation_members m ON m.conversation_id = c.id "+
			"AND m.user_id = ? AND m.left_at IS NULL AND m.deleted_at IS NULL", userID).
		Where("c.organization_id = ?", orgID).
		Where("c.deleted_at IS NULL")
	if !f.IncludeArchived {
		q = q.Where("c.archived_at IS NULL")
	}
	if f.Type != "" {
		q = q.Where("c.type = ?", f.Type)
	}
	q = q.Order("c.last_message_at DESC NULLS LAST, c.created_at DESC")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	out := []ConversationListItem{}
	if err := q.
		Select("c.*, m.unread_count AS unread_count, m.last_read_message_id AS last_read_message_id").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// BumpLastMessage is called after every successful insert so the sidebar
// activity sort stays accurate. Done in the same transaction as the message
// insert; a partial bump (message inserted, conversation un-bumped) leaves
// the sidebar slightly stale but doesn't corrupt anything.
func (r *Repository) BumpLastMessage(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Conversation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_message_at": at,
			"message_count":   gorm.Expr("message_count + 1"),
		}).Error
}

// ── members ────────────────────────────────────────────────────────────────

func (r *Repository) CreateMember(ctx context.Context, m *ConversationMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// AddMembersIfMissing inserts (conv, user) pairs that don't already have an
// active membership. Idempotent. Returns the new member rows it created.
func (r *Repository) AddMembersIfMissing(
	ctx context.Context,
	convID uuid.UUID,
	userIDs []uuid.UUID,
	role string,
	invitedBy *uuid.UUID,
) ([]ConversationMember, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	// Find which userIDs already have an active membership so we don't
	// trip the partial unique index.
	var existing []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id IN ? AND left_at IS NULL AND deleted_at IS NULL", convID, userIDs).
		Pluck("user_id", &existing).Error; err != nil {
		return nil, err
	}
	skip := make(map[uuid.UUID]struct{}, len(existing))
	for _, id := range existing {
		skip[id] = struct{}{}
	}
	toCreate := make([]ConversationMember, 0, len(userIDs)-len(existing))
	for _, uid := range userIDs {
		if _, ok := skip[uid]; ok {
			continue
		}
		toCreate = append(toCreate, ConversationMember{
			ConversationID:   convID,
			UserID:           uid,
			Role:             role,
			JoinedAt:         time.Now(),
			NotificationPref: "all",
		})
	}
	if len(toCreate) == 0 {
		return nil, nil
	}
	if err := r.db.WithContext(ctx).Create(&toCreate).Error; err != nil {
		return nil, err
	}
	return toCreate, nil
}

func (r *Repository) GetMember(ctx context.Context, convID, userID uuid.UUID) (*ConversationMember, error) {
	var m ConversationMember
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND user_id = ? AND left_at IS NULL AND deleted_at IS NULL", convID, userID).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ListMembers(ctx context.Context, convID uuid.UUID) ([]ConversationMember, error) {
	out := []ConversationMember{}
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND left_at IS NULL AND deleted_at IS NULL", convID).
		Order("joined_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// MarkMemberLeft soft-leaves (set left_at). The unique index excludes
// left_at IS NOT NULL so the same user can re-join later.
func (r *Repository) MarkMemberLeft(ctx context.Context, convID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id = ? AND left_at IS NULL", convID, userID).
		Update("left_at", time.Now()).Error
}

func (r *Repository) UpdateMember(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&ConversationMember{}).
		Where("id = ?", id).
		Updates(patch).Error
}

// MarkRead advances the read state for a single member and zeroes their
// unread count. Caller passes a known message id; the conversation should
// have been validated upstream.
func (r *Repository) MarkRead(ctx context.Context, convID, userID, lastReadMessageID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id = ? AND left_at IS NULL", convID, userID).
		Updates(map[string]interface{}{
			"last_read_message_id": lastReadMessageID,
			"last_read_at":         time.Now(),
			"unread_count":         0,
		}).Error
}

// IncrementUnreadForOthers bumps the cached unread count for every active
// member EXCEPT the sender, called inside the message-send transaction. Keeps
// the sidebar badge accurate without a per-render COUNT(*).
func (r *Repository) IncrementUnreadForOthers(ctx context.Context, convID, senderUserID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id <> ? AND left_at IS NULL", convID, senderUserID).
		Update("unread_count", gorm.Expr("unread_count + 1")).Error
}

// ── messages ───────────────────────────────────────────────────────────────

func (r *Repository) CreateMessage(ctx context.Context, m *Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) GetMessage(ctx context.Context, id uuid.UUID) (*Message, error) {
	var m Message
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMessages returns paginated history. Cursor:
//
//	beforeMessageID = uuid.Nil → newest page first
//	beforeMessageID != nil     → strictly older than the cursor message
//
// The query uses the (conversation_id, created_at DESC, id DESC) index so it
// stays index-only even on hot conversations.
func (r *Repository) ListMessages(
	ctx context.Context,
	convID uuid.UUID,
	beforeMessageID uuid.UUID,
	limit int,
) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at DESC, id DESC").
		Limit(limit)
	if beforeMessageID != uuid.Nil {
		// (created_at, id) tuple compare for stable pagination across rows
		// posted in the same millisecond.
		q = q.Where(`(created_at, id) < (
			SELECT created_at, id FROM comm_messages WHERE id = ?
		)`, beforeMessageID)
	}
	out := []Message{}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateMessage(ctx context.Context, id uuid.UUID, patch map[string]interface{}) error {
	res := r.db.WithContext(ctx).
		Model(&Message{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(patch)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SoftDeleteMessage flips deleted_at. The body is preserved in DB for audit;
// the API rewrites it to "[deleted]" on read.
func (r *Repository) SoftDeleteMessage(ctx context.Context, id uuid.UUID, by *uuid.UUID) error {
	patch := map[string]interface{}{"deleted_at": time.Now()}
	if by != nil {
		patch["deleted_by"] = *by
	}
	return r.UpdateMessage(ctx, id, patch)
}

// ── mentions ──────────────────────────────────────────────────────────────

func (r *Repository) CreateMentions(ctx context.Context, rows []MessageMention) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *Repository) ListMentionsForMessages(ctx context.Context, ids []uuid.UUID) ([]MessageMention, error) {
	out := []MessageMention{}
	if len(ids) == 0 {
		return out, nil
	}
	if err := r.db.WithContext(ctx).
		Where("message_id IN ?", ids).
		Order("created_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) DeleteMentionsForMessage(ctx context.Context, messageID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Delete(&MessageMention{}).Error
}

// ── reactions ─────────────────────────────────────────────────────────────

// AddReaction is idempotent — the unique index makes the second insert a
// no-op (returned err is checked + swallowed when it's a unique violation).
func (r *Repository) AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	row := MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}
	err := r.db.WithContext(ctx).Create(&row).Error
	if err == nil {
		return nil
	}
	// Idempotent: ignore unique violation.
	if isUniqueViolation(err) {
		return nil
	}
	return err
}

func (r *Repository) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	return r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&MessageReaction{}).Error
}

func (r *Repository) ListReactionsForMessages(ctx context.Context, ids []uuid.UUID) ([]MessageReaction, error) {
	out := []MessageReaction{}
	if len(ids) == 0 {
		return out, nil
	}
	if err := r.db.WithContext(ctx).
		Where("message_id IN ?", ids).
		Order("created_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func isUniqueViolation(err error) bool {
	// Postgres SQLSTATE 23505. We match by string to avoid pulling in pgconn.
	return err != nil && (containsAny(err.Error(),
		"duplicate key value violates unique constraint",
		"SQLSTATE 23505",
	))
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && stringContains(s, sub) {
			return true
		}
	}
	return false
}

// stringContains is a thin wrapper so this file doesn't need to import
// strings just for one call.
func stringContains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	n, m := len(s), len(sub)
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
