package comm

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/modules/user"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
)

// editWindowMinutes is how long after sending a user can still edit their
// own message. Phase 1 is permissive; we can clamp later.
const editWindowMinutes = 24 * 60

// BroadcasterPort is what comm.Service uses to push real-time events. Per-
// event methods keep the interface explicit (no generic "any frame") so the
// implementation in ws/ can use typed constructors and the contract is easy
// to mock for tests. nil-port is silently no-op for Phase 1 / single-process
// use where the WS layer isn't bootstrapped.
type BroadcasterPort interface {
	MessageCreated(ctx context.Context, convID uuid.UUID, view MessageView)
	MessageUpdated(ctx context.Context, convID uuid.UUID, view MessageView)
	MessageDeleted(ctx context.Context, convID, msgID uuid.UUID)
	ReactionAdded(ctx context.Context, convID, msgID, userID uuid.UUID, emoji string)
	ReactionRemoved(ctx context.Context, convID, msgID, userID uuid.UUID, emoji string)
	Read(ctx context.Context, convID, userID, lastReadMessageID uuid.UUID)
	MemberAdded(ctx context.Context, convID uuid.UUID, view ConversationMemberView)
	MemberRemoved(ctx context.Context, convID, userID uuid.UUID)
	ConversationUpdated(ctx context.Context, conv Conversation)
}

type Service struct {
	repo     *Repository
	userSvc  *user.Service
	notifier *notifier
	bcast    BroadcasterPort
	log      *zap.Logger
}

func NewService(repo *Repository, userSvc *user.Service, notif NotificationPort, log *zap.Logger) *Service {
	return &Service{
		repo:     repo,
		userSvc:  userSvc,
		notifier: newNotifier(notif, log),
		log:      log,
	}
}

// SetBroadcaster installs the broadcast publisher AFTER the ws layer has
// been constructed in bootstrap. Order: service first, then ws (which sees
// the service), then this method to close the loop.
func (s *Service) SetBroadcaster(b BroadcasterPort) { s.bcast = b }

// b returns a non-nil broadcaster (real or no-op) so call sites don't have
// to nil-check. The no-op is a zero-cost call.
func (s *Service) b() BroadcasterPort {
	if s.bcast == nil {
		return noopBroadcaster{}
	}
	return s.bcast
}

type noopBroadcaster struct{}

func (noopBroadcaster) MessageCreated(context.Context, uuid.UUID, MessageView)              {}
func (noopBroadcaster) MessageUpdated(context.Context, uuid.UUID, MessageView)              {}
func (noopBroadcaster) MessageDeleted(context.Context, uuid.UUID, uuid.UUID)                {}
func (noopBroadcaster) ReactionAdded(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string)   {}
func (noopBroadcaster) ReactionRemoved(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) {}
func (noopBroadcaster) Read(context.Context, uuid.UUID, uuid.UUID, uuid.UUID)               {}
func (noopBroadcaster) MemberAdded(context.Context, uuid.UUID, ConversationMemberView)      {}
func (noopBroadcaster) MemberRemoved(context.Context, uuid.UUID, uuid.UUID)                 {}
func (noopBroadcaster) ConversationUpdated(context.Context, Conversation)                   {}

// ── DMs ────────────────────────────────────────────────────────────────────

// CreateOrGetDM is idempotent — calling it twice for the same (caller,
// recipient) returns the existing conversation. The DM key encodes the pair
// of user IDs sorted lex so direction doesn't matter.
func (s *Service) CreateOrGetDM(ctx context.Context, recipientUserID uuid.UUID) (*Conversation, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if tid == uuid.Nil || oid == uuid.Nil || uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	if recipientUserID == uid {
		return nil, apperr.New(apperr.CodeValidation, "cannot DM yourself", nil)
	}

	// The recipient must be an active member of the same org. Otherwise a
	// malicious caller could probe the existence of users in other orgs by
	// trying random UUIDs.
	if _, err := s.userSvc.GetByID(ctx, recipientUserID); err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "recipient not found", nil)
	}

	key := buildDMKey(uid, recipientUserID)
	if existing, err := s.repo.GetDMByKey(ctx, oid, key); err == nil && existing != nil {
		return existing, nil
	}

	now := time.Now()
	conv := &Conversation{
		TenantID:       tid,
		OrganizationID: oid,
		Type:           "dm",
		DMKey:          key,
		LastMessageAt:  &now,
	}
	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			// A concurrent caller may have just inserted the same key — the
			// partial unique index will reject ours. Re-fetch and return.
			if existing, err2 := s.repo.GetDMByKey(ctx, oid, key); err2 == nil && existing != nil {
				*conv = *existing
				return nil
			}
			return err
		}
		members := []ConversationMember{
			{ConversationID: conv.ID, UserID: uid, Role: "member", JoinedAt: now, NotificationPref: "all"},
			{ConversationID: conv.ID, UserID: recipientUserID, Role: "member", JoinedAt: now, NotificationPref: "all"},
		}
		return tx.Create(&members).Error
	}); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create dm failed", err)
	}
	if peer := s.dmPeerName(ctx, conv.ID, uid); peer != "" {
		conv.Name = peer
	}
	return conv, nil
}

// ── Channels ───────────────────────────────────────────────────────────────

func (s *Service) CreateChannel(ctx context.Context, req CreateChannelRequest) (*Conversation, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if tid == uuid.Nil || oid == uuid.Nil || uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	slug := normaliseSlug(req.Slug)
	if !isValidSlug(slug) {
		return nil, apperr.New(apperr.CodeValidation,
			"slug must be 2-64 chars, lowercase letters, digits or hyphens", nil)
	}
	if existing, _ := s.repo.GetChannelBySlug(ctx, oid, slug); existing != nil {
		return nil, apperr.New(apperr.CodeAlreadyExists, "channel slug taken", nil)
	}

	now := time.Now()
	conv := &Conversation{
		TenantID:       tid,
		OrganizationID: oid,
		Type:           "channel",
		Slug:           slug,
		Name:           strings.TrimSpace(req.Name),
		Topic:          req.Topic,
		Description:    req.Description,
		IsPrivate:      req.IsPrivate,
		LastMessageAt:  &now,
	}
	// Seed members: creator as owner + any explicit invites as members.
	initialMembers := uniqueUUIDs(append([]uuid.UUID{uid}, req.MemberIDs...))
	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			return err
		}
		rows := make([]ConversationMember, 0, len(initialMembers))
		for _, mid := range initialMembers {
			role := "member"
			if mid == uid {
				role = "owner"
			}
			rows = append(rows, ConversationMember{
				ConversationID:   conv.ID,
				UserID:           mid,
				Role:             role,
				JoinedAt:         now,
				NotificationPref: "all",
			})
		}
		return tx.Create(&rows).Error
	}); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create channel failed", err)
	}
	return conv, nil
}

func (s *Service) UpdateChannel(ctx context.Context, id uuid.UUID, req UpdateChannelRequest) (*Conversation, error) {
	conv, _, err := s.requireChannelManager(ctx, id)
	if err != nil {
		return nil, err
	}
	patch := map[string]interface{}{}
	if req.Name != nil {
		patch["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Topic != nil {
		patch["topic"] = *req.Topic
	}
	if req.Description != nil {
		patch["description"] = *req.Description
	}
	if req.IsPrivate != nil {
		patch["is_private"] = *req.IsPrivate
	}
	if len(patch) == 0 {
		return conv, nil
	}
	if err := s.repo.UpdateConversation(ctx, id, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "update channel failed", err)
	}
	return s.repo.GetConversation(ctx, appctx.OrganizationID(ctx), id)
}

func (s *Service) ArchiveChannel(ctx context.Context, id uuid.UUID) error {
	if _, _, err := s.requireChannelManager(ctx, id); err != nil {
		return err
	}
	if err := s.repo.ArchiveConversation(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "archive failed", err)
	}
	return nil
}

// ── Conversation reads ────────────────────────────────────────────────────

func (s *Service) ListMyConversations(ctx context.Context, f ListConversationsFilter) ([]ConversationListItem, error) {
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if oid == uuid.Nil || uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	rows, err := s.repo.ListConversationsForUser(ctx, oid, uid, f)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list conversations failed", err)
	}
	// Hydrate DM titles with the OTHER member's display name so the sidebar
	// shows "Bob Builder" instead of a literal "DM" string.
	for i := range rows {
		if rows[i].Type == "dm" {
			if peer := s.dmPeerName(ctx, rows[i].ID, uid); peer != "" {
				rows[i].Name = peer
			}
		}
	}
	return rows, nil
}

// dmPeerName resolves the other member of a DM to a display string. Empty
// when the lookup fails (best-effort — the UI falls back to a placeholder).
func (s *Service) dmPeerName(ctx context.Context, convID, callerID uuid.UUID) string {
	members, err := s.repo.ListMembers(ctx, convID)
	if err != nil {
		return ""
	}
	for _, m := range members {
		if m.UserID == callerID {
			continue
		}
		u, err := s.userSvc.GetByID(ctx, m.UserID)
		if err != nil || u == nil {
			continue
		}
		if u.DisplayName != "" {
			return u.DisplayName
		}
		first, last := strings.TrimSpace(u.FirstName), strings.TrimSpace(u.LastName)
		full := strings.TrimSpace(first + " " + last)
		if full != "" {
			return full
		}
		return u.Email
	}
	return ""
}

// GetConversationView returns the conversation + hydrated members + caller's
// membership. Public channels are visible to non-members (with empty
// membership); DMs and private channels require active membership.
func (s *Service) GetConversationView(ctx context.Context, id uuid.UUID) (*ConversationView, error) {
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if oid == uuid.Nil || uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	conv, err := s.repo.GetConversation(ctx, oid, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "conversation not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch conversation failed", err)
	}

	myMember, _ := s.repo.GetMember(ctx, id, uid)
	if myMember == nil {
		// Non-member access: only public channels are visible.
		if conv.Type != "channel" || conv.IsPrivate {
			return nil, apperr.New(apperr.CodeForbidden, "not a member", nil)
		}
	}

	members, err := s.repo.ListMembers(ctx, id)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load members failed", err)
	}
	hydrated := s.hydrateMembers(ctx, members)
	view := &ConversationView{Conversation: *conv, Members: hydrated}
	if myMember != nil {
		me := s.hydrateMembers(ctx, []ConversationMember{*myMember})[0]
		view.MyMembership = &me
	}
	// DM titles: same hydration as the list endpoint — the conv pane header
	// reads view.Conversation.Name.
	if view.Type == "dm" {
		if peer := s.dmPeerName(ctx, id, uid); peer != "" {
			view.Name = peer
		}
	}
	return view, nil
}

func (s *Service) ListConversationMembers(ctx context.Context, convID uuid.UUID) ([]ConversationMemberView, error) {
	if _, err := s.requireMember(ctx, convID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListMembers(ctx, convID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "load members failed", err)
	}
	return s.hydrateMembers(ctx, rows), nil
}

// ── Membership ────────────────────────────────────────────────────────────

func (s *Service) AddMembers(ctx context.Context, convID uuid.UUID, userIDs []uuid.UUID) ([]ConversationMemberView, error) {
	conv, _, err := s.requireChannelManager(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv.Type != "channel" {
		return nil, apperr.New(apperr.CodeValidation, "cannot add members to a DM", nil)
	}
	rows, err := s.repo.AddMembersIfMissing(ctx, convID, uniqueUUIDs(userIDs), "member", ptrUUID(appctx.UserID(ctx)))
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "add members failed", err)
	}
	return s.hydrateMembers(ctx, rows), nil
}

func (s *Service) RemoveMember(ctx context.Context, convID, targetUserID uuid.UUID) error {
	uid := appctx.UserID(ctx)
	conv, err := s.requireMember(ctx, convID)
	if err != nil {
		return err
	}
	// Self-leave any time. Removing someone else requires manage perms.
	if targetUserID != uid {
		if _, _, err := s.requireChannelManager(ctx, convID); err != nil {
			return err
		}
		if conv.Type == "dm" {
			return apperr.New(apperr.CodeValidation, "cannot remove members from a DM", nil)
		}
	}
	if err := s.repo.MarkMemberLeft(ctx, convID, targetUserID); err != nil {
		return apperr.New(apperr.CodeInternal, "remove member failed", err)
	}
	return nil
}

func (s *Service) UpdateMyMemberPrefs(ctx context.Context, convID uuid.UUID, req UpdateMemberPrefsRequest) (*ConversationMember, error) {
	_, member, err := s.requireMemberWithRow(ctx, convID)
	if err != nil {
		return nil, err
	}
	patch := map[string]interface{}{}
	if req.NotificationPref != nil {
		patch["notification_pref"] = *req.NotificationPref
	}
	if req.MutedUntil != nil {
		patch["muted_until"] = *req.MutedUntil
	}
	if len(patch) == 0 {
		return member, nil
	}
	if err := s.repo.UpdateMember(ctx, member.ID, patch); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "update prefs failed", err)
	}
	return s.repo.GetMember(ctx, convID, member.UserID)
}

// ── Messages ──────────────────────────────────────────────────────────────

func (s *Service) SendMessage(ctx context.Context, convID uuid.UUID, req SendMessageRequest) (*MessageView, error) {
	conv, member, err := s.requireMemberWithRow(ctx, convID)
	if err != nil {
		return nil, err
	}
	uid := appctx.UserID(ctx)
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, apperr.New(apperr.CodeValidation, "body cannot be empty", nil)
	}
	bodyFormat := req.BodyFormat
	if bodyFormat == "" {
		bodyFormat = "markdown"
	}

	// Parse + resolve mentions BEFORE the transaction so a slow user-list
	// query doesn't hold a write lock.
	raw := ParseMentions(body)
	resolved, _ := s.resolveMentions(ctx, raw)

	msg := &Message{
		ConversationID: convID,
		TenantID:       conv.TenantID,
		OrganizationID: conv.OrganizationID,
		SenderType:     "user",
		SenderUserID:   ptrUUID(uid),
		Body:           body,
		BodyFormat:     bodyFormat,
	}
	if req.ParentMessageID != nil {
		if parent, err := s.repo.GetMessage(ctx, *req.ParentMessageID); err == nil && parent.ConversationID == convID {
			msg.ParentMessageID = req.ParentMessageID
		}
	}

	var mentions []MessageMention
	var members []ConversationMember
	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		mentions = BuildMentionsForMessage(msg.ID, raw, resolved)
		if len(mentions) > 0 {
			if err := tx.Create(&mentions).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&Conversation{}).Where("id = ?", convID).
			Updates(map[string]interface{}{
				"last_message_at": msg.CreatedAt,
				"message_count":   gorm.Expr("message_count + 1"),
			}).Error; err != nil {
			return err
		}
		// Bump unread for every other active member in the same query.
		if err := tx.Model(&ConversationMember{}).
			Where("conversation_id = ? AND user_id <> ? AND left_at IS NULL", convID, uid).
			Update("unread_count", gorm.Expr("unread_count + 1")).Error; err != nil {
			return err
		}
		// Zero the sender's own unread (sending implies reading).
		if err := tx.Model(&ConversationMember{}).
			Where("id = ?", member.ID).
			Updates(map[string]interface{}{
				"last_read_message_id": msg.ID,
				"last_read_at":         msg.CreatedAt,
				"unread_count":         0,
			}).Error; err != nil {
			return err
		}
		var inner error
		members, inner = s.repo.ListMembers(ctx, convID)
		return inner
	}); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "send message failed", err)
	}

	// Notifications happen AFTER commit so a slow notifier doesn't block the
	// send. Errors are logged and swallowed.
	s.notifier.PublishForMessage(ctx, conv, msg, mentions, members)

	view := s.hydrateMessages(ctx, []Message{*msg})
	var out MessageView
	if len(view) == 0 {
		out = MessageView{Message: *msg, Mentions: mentions}
	} else {
		out = view[0]
		out.Mentions = mentions
	}
	// Real-time broadcast — every subscriber sees the message land instantly.
	s.b().MessageCreated(ctx, convID, out)
	return &out, nil
}

// SendMessageAsWebhook is the inbound-hook entry point. The caller has
// already verified the hook token; we build a webhook-typed message with
// the hook's display name baked into metadata, persist, and broadcast.
// No mention parsing for webhook bodies (they're machine-generated; humans
// wouldn't be typing @mentions there).
func (s *Service) SendMessageAsWebhook(
	ctx context.Context,
	conversationID uuid.UUID,
	hookID uuid.UUID,
	displayName, iconURL, body string,
	attachments []map[string]interface{},
) (uuid.UUID, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return uuid.Nil, apperr.New(apperr.CodeValidation, "body cannot be empty", nil)
	}
	tid := appctx.TenantID(ctx)
	if tid == uuid.Nil {
		return uuid.Nil, apperr.New(apperr.CodeForbidden, "no tenant context", nil)
	}
	conv, err := s.repo.GetConversation(ctx, uuid.Nil, conversationID)
	if err != nil {
		return uuid.Nil, apperr.New(apperr.CodeNotFound, "conversation not found", nil)
	}
	if conv.TenantID != tid {
		// Hook crossed tenants somehow — defence in depth.
		return uuid.Nil, apperr.New(apperr.CodeForbidden, "hook does not belong to this tenant", nil)
	}
	meta := map[string]any{}
	if displayName != "" {
		meta["webhookDisplayName"] = displayName
	}
	if iconURL != "" {
		meta["webhookIconUrl"] = iconURL
	}
	metaJSON, _ := json.Marshal(meta)
	attachmentsJSON, _ := json.Marshal(attachments)
	hookIDPtr := hookID
	msg := &Message{
		ConversationID:  conversationID,
		TenantID:        conv.TenantID,
		OrganizationID:  conv.OrganizationID,
		SenderType:      "webhook",
		SenderWebhookID: &hookIDPtr,
		Body:            body,
		BodyFormat:      "markdown",
		Attachments:     attachmentsJSON,
		Metadata:        metaJSON,
	}
	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		if err := tx.Model(&Conversation{}).Where("id = ?", conversationID).
			Updates(map[string]interface{}{
				"last_message_at": msg.CreatedAt,
				"message_count":   gorm.Expr("message_count + 1"),
			}).Error; err != nil {
			return err
		}
		// Webhook messages bump unread for ALL members — there's no "self".
		return tx.Model(&ConversationMember{}).
			Where("conversation_id = ? AND left_at IS NULL", conversationID).
			Update("unread_count", gorm.Expr("unread_count + 1")).Error
	}); err != nil {
		return uuid.Nil, apperr.New(apperr.CodeInternal, "webhook message persist failed", err)
	}

	// Notifications: webhooks deliver to everyone whose pref isn't "none".
	members, _ := s.repo.ListMembers(ctx, conversationID)
	s.notifier.PublishForMessage(ctx, conv, msg, nil, members)

	view := s.hydrateMessages(ctx, []Message{*msg})
	if len(view) > 0 {
		s.b().MessageCreated(ctx, conversationID, view[0])
	}
	return msg.ID, nil
}

func (s *Service) ListMessages(ctx context.Context, convID uuid.UUID, beforeMessageID uuid.UUID, limit int) ([]MessageView, error) {
	if _, err := s.requireMember(ctx, convID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListMessages(ctx, convID, beforeMessageID, limit)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list messages failed", err)
	}
	return s.hydrateMessages(ctx, rows), nil
}

func (s *Service) EditMessage(ctx context.Context, messageID uuid.UUID, req EditMessageRequest) (*MessageView, error) {
	uid := appctx.UserID(ctx)
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return nil, apperr.New(apperr.CodeNotFound, "message not found", nil)
	}
	if msg.SenderUserID == nil || *msg.SenderUserID != uid {
		return nil, apperr.New(apperr.CodeForbidden, "can only edit your own messages", nil)
	}
	if time.Since(msg.CreatedAt) > editWindowMinutes*time.Minute {
		return nil, apperr.New(apperr.CodeForbidden, "edit window has passed", nil)
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, apperr.New(apperr.CodeValidation, "body cannot be empty", nil)
	}

	// Re-parse mentions; the set may have changed.
	raw := ParseMentions(body)
	resolved, _ := s.resolveMentions(ctx, raw)
	newMentions := BuildMentionsForMessage(messageID, raw, resolved)

	if err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Message{}).Where("id = ?", messageID).Updates(map[string]interface{}{
			"body":      body,
			"edited_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("message_id = ?", messageID).Delete(&MessageMention{}).Error; err != nil {
			return err
		}
		if len(newMentions) > 0 {
			if err := tx.Create(&newMentions).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "edit failed", err)
	}
	updated, _ := s.repo.GetMessage(ctx, messageID)
	views := s.hydrateMessages(ctx, []Message{*updated})
	var out MessageView
	if len(views) == 0 {
		out = MessageView{Message: *updated, Mentions: newMentions}
	} else {
		out = views[0]
		out.Mentions = newMentions
	}
	s.b().MessageUpdated(ctx, updated.ConversationID, out)
	return &out, nil
}

// DeleteMessage soft-deletes a message. Author can always delete their own;
// moderators with comm.message.moderate can delete anyone's. The handler
// gates the moderator path via perm middleware.
func (s *Service) DeleteMessage(ctx context.Context, messageID uuid.UUID, isModerator bool) error {
	uid := appctx.UserID(ctx)
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return apperr.New(apperr.CodeNotFound, "message not found", nil)
	}
	isAuthor := msg.SenderUserID != nil && *msg.SenderUserID == uid
	if !isAuthor && !isModerator {
		return apperr.New(apperr.CodeForbidden, "cannot delete this message", nil)
	}
	if err := s.repo.SoftDeleteMessage(ctx, messageID, ptrUUID(uid)); err != nil {
		return apperr.New(apperr.CodeInternal, "delete failed", err)
	}
	s.b().MessageDeleted(ctx, msg.ConversationID, messageID)
	return nil
}

func (s *Service) MarkRead(ctx context.Context, convID, lastReadMessageID uuid.UUID) error {
	if _, err := s.requireMember(ctx, convID); err != nil {
		return err
	}
	uid := appctx.UserID(ctx)
	if err := s.repo.MarkRead(ctx, convID, uid, lastReadMessageID); err != nil {
		return apperr.New(apperr.CodeInternal, "mark read failed", err)
	}
	s.b().Read(ctx, convID, uid, lastReadMessageID)
	return nil
}

// ── Reactions ─────────────────────────────────────────────────────────────

func (s *Service) AddReaction(ctx context.Context, messageID uuid.UUID, emoji string) error {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return apperr.New(apperr.CodeNotFound, "message not found", nil)
	}
	if _, err := s.requireMember(ctx, msg.ConversationID); err != nil {
		return err
	}
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return apperr.New(apperr.CodeValidation, "emoji is required", nil)
	}
	uid := appctx.UserID(ctx)
	if err := s.repo.AddReaction(ctx, messageID, uid, emoji); err != nil {
		return apperr.New(apperr.CodeInternal, "add reaction failed", err)
	}
	s.b().ReactionAdded(ctx, msg.ConversationID, messageID, uid, emoji)
	return nil
}

func (s *Service) RemoveReaction(ctx context.Context, messageID uuid.UUID, emoji string) error {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return apperr.New(apperr.CodeNotFound, "message not found", nil)
	}
	if _, err := s.requireMember(ctx, msg.ConversationID); err != nil {
		return err
	}
	uid := appctx.UserID(ctx)
	if err := s.repo.RemoveReaction(ctx, messageID, uid, emoji); err != nil {
		return apperr.New(apperr.CodeInternal, "remove reaction failed", err)
	}
	s.b().ReactionRemoved(ctx, msg.ConversationID, messageID, uid, emoji)
	return nil
}

// ── Permission helpers ────────────────────────────────────────────────────

// requireMember returns the conversation + caller's active membership, or a
// 403. The handler still gates `comm.read` via middleware; this is the
// row-level check.
func (s *Service) requireMember(ctx context.Context, convID uuid.UUID) (*Conversation, error) {
	oid := appctx.OrganizationID(ctx)
	uid := appctx.UserID(ctx)
	if oid == uuid.Nil || uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	conv, err := s.repo.GetConversation(ctx, oid, convID)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "conversation not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch conversation failed", err)
	}
	if _, err := s.repo.GetMember(ctx, convID, uid); err != nil {
		return conv, apperr.New(apperr.CodeForbidden, "not a member of this conversation", nil)
	}
	return conv, nil
}

// requireMember overload that also returns the member row.
func (s *Service) requireMemberWithRow(ctx context.Context, convID uuid.UUID) (*Conversation, *ConversationMember, error) {
	conv, err := s.requireMember(ctx, convID)
	if err != nil {
		return nil, nil, err
	}
	mem, err := s.repo.GetMember(ctx, convID, appctx.UserID(ctx))
	if err != nil {
		return conv, nil, apperr.New(apperr.CodeForbidden, "not a member", nil)
	}
	return conv, mem, nil
}

// requireChannelManager returns 403 unless the caller is owner/admin of a
// channel. DMs never have a "manager".
func (s *Service) requireChannelManager(ctx context.Context, convID uuid.UUID) (*Conversation, *ConversationMember, error) {
	conv, mem, err := s.requireMemberWithRow(ctx, convID)
	if err != nil {
		return nil, nil, err
	}
	if conv.Type != "channel" {
		return nil, nil, apperr.New(apperr.CodeValidation, "not a channel", nil)
	}
	if mem.Role != "owner" && mem.Role != "admin" {
		// Super-admins on the tenant escape this check via context.
		if !appctx.IsSuperAdmin(ctx) {
			return nil, nil, apperr.New(apperr.CodeForbidden, "channel manager only", nil)
		}
	}
	return conv, mem, nil
}

// ── Hydration helpers ────────────────────────────────────────────────────

// resolveMentions queries the org user list once and returns identifier→uid.
// Identifiers come from candidateIdentifiers (username, display name, first
// name, email-local). Broadcast tokens skip this entirely.
func (s *Service) resolveMentions(ctx context.Context, raw []RawMention) (map[string]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	users, _, err := s.userSvc.ListInOrg(ctx, tid, oid, user.ListFilter{}, pagination.Params{
		Page: 1, Limit: pagination.MaxLimit,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(users)*3)
	for _, u := range users {
		for _, ident := range candidateIdentifiers(u.DisplayName, u.FirstName, u.LastName, u.Email, u.Username) {
			if _, exists := out[ident]; !exists {
				out[ident] = u.ID
			}
		}
	}
	return out, nil
}

// hydrateMembers looks up each member's user record and copies display
// fields into the view shape. Single-batch — a SELECT per member would be
// terrible for a 100-member channel; in Phase 4 we can add a batch lookup.
func (s *Service) hydrateMembers(ctx context.Context, rows []ConversationMember) []ConversationMemberView {
	out := make([]ConversationMemberView, 0, len(rows))
	for _, m := range rows {
		view := ConversationMemberView{
			ID:                m.ID,
			UserID:            m.UserID,
			Role:              m.Role,
			JoinedAt:          m.JoinedAt,
			LastReadMessageID: m.LastReadMessageID,
			UnreadCount:       m.UnreadCount,
			NotificationPref:  m.NotificationPref,
		}
		if u, err := s.userSvc.GetByID(ctx, m.UserID); err == nil && u != nil {
			view.UserEmail = u.Email
			view.UserDisplayName = u.DisplayName
			view.UserAvatarURL = u.AvatarURL
		}
		out = append(out, view)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].JoinedAt.Before(out[j].JoinedAt)
	})
	return out
}

func (s *Service) hydrateMessages(ctx context.Context, rows []Message) []MessageView {
	if len(rows) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	mentions, _ := s.repo.ListMentionsForMessages(ctx, ids)
	reactions, _ := s.repo.ListReactionsForMessages(ctx, ids)
	mentionsByMsg := groupMentions(mentions)
	reactionsByMsg := groupReactions(reactions)

	out := make([]MessageView, 0, len(rows))
	for _, m := range rows {
		view := MessageView{
			Message:   m,
			Mentions:  mentionsByMsg[m.ID],
			Reactions: reactionsByMsg[m.ID],
		}
		// Soft-deleted: rewrite body for clients.
		if !m.DeletedAt.Time.IsZero() {
			view.Body = "[deleted]"
			view.Attachments = nil
		}
		if m.SenderUserID != nil {
			if u, err := s.userSvc.GetByID(ctx, *m.SenderUserID); err == nil && u != nil {
				view.SenderDisplayName = u.DisplayName
				view.SenderAvatarURL = u.AvatarURL
			}
		}
		out = append(out, view)
	}
	return out
}

func groupMentions(rows []MessageMention) map[uuid.UUID][]MessageMention {
	out := map[uuid.UUID][]MessageMention{}
	for _, m := range rows {
		out[m.MessageID] = append(out[m.MessageID], m)
	}
	return out
}

func groupReactions(rows []MessageReaction) map[uuid.UUID][]MessageReaction {
	out := map[uuid.UUID][]MessageReaction{}
	for _, r := range rows {
		out[r.MessageID] = append(out[r.MessageID], r)
	}
	return out
}

// ── small helpers ─────────────────────────────────────────────────────────

func buildDMKey(a, b uuid.UUID) string {
	if a.String() < b.String() {
		return a.String() + ":" + b.String()
	}
	return b.String() + ":" + a.String()
}

func normaliseSlug(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func isValidSlug(s string) bool {
	if len(s) < 2 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}

func uniqueUUIDs(in []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(in))
	out := make([]uuid.UUID, 0, len(in))
	for _, id := range in {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	if u == uuid.Nil {
		return nil
	}
	id := u
	return &id
}

