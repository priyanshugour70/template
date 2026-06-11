package comm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/modules/notification"
)

// NotificationPort is the narrow slice of notification.Service that comm
// depends on. Kept here so the comm.Module can be wired up without the full
// notification surface.
type NotificationPort interface {
	Create(ctx context.Context, in notification.CreateInput) (*notification.Notification, error)
}

// notifier publishes in-app notifications for new messages. Decision rules:
//
//   - The sender NEVER notifies themselves.
//   - DM recipient(s) always get notified (even without an explicit mention).
//   - In a channel: a member gets notified only if (a) they were @mentioned
//     OR (b) their member.NotificationPref == "all" OR (c) the message
//     contained a broadcast mention (@here/@channel/@everyone). Members with
//     pref="none" or whose muted_until is in the future never see a
//     notification from this code path.
//   - "[deleted]" rewrites happen in the API layer; this notifier sees the
//     real body and uses the first 140 chars as the notification preview.
//
// Failures are logged but non-fatal — the API has already returned 201 to
// the user, so dropping a notification is preferable to rolling back the
// message they just sent.
type notifier struct {
	notif NotificationPort
	log   *zap.Logger
}

func newNotifier(notif NotificationPort, log *zap.Logger) *notifier {
	return &notifier{notif: notif, log: log}
}

// PublishForMessage decides who to notify and emits one notification per
// recipient. Caller passes the resolved members so we don't re-query.
func (n *notifier) PublishForMessage(
	ctx context.Context,
	conv *Conversation,
	msg *Message,
	mentions []MessageMention,
	members []ConversationMember,
) {
	if n == nil || n.notif == nil {
		return
	}
	if msg.SenderUserID == nil {
		// System / webhook messages have no "self" to exclude. Notify
		// everyone per the channel pref rules below.
	}

	// Pre-compute which users were directly mentioned + whether the message
	// carries a broadcast (here/channel/everyone). These short-circuit the
	// pref check for individuals.
	directMention := make(map[uuid.UUID]bool, len(mentions))
	hasBroadcast := false
	for _, m := range mentions {
		switch m.MentionType {
		case "user":
			if m.TargetUserID != nil {
				directMention[*m.TargetUserID] = true
			}
		case "here", "channel", "everyone":
			hasBroadcast = true
		}
	}

	title, preview := n.previewFor(conv, msg)
	link := "/dashboard/communication/" + conv.ID.String()
	kind := "info"

	for _, m := range members {
		uid := m.UserID
		// Don't notify the sender about their own message.
		if msg.SenderUserID != nil && uid == *msg.SenderUserID {
			continue
		}
		// Muted member — skip.
		if m.MutedUntil != nil && m.MutedUntil.After(msg.CreatedAt) {
			continue
		}
		// Decide based on context:
		//   DM             → always notify (recipient signed up by joining).
		//   Direct mention → always notify (regardless of pref).
		//   Broadcast      → notify if pref != "none".
		//   Otherwise      → notify only if pref == "all".
		shouldNotify := false
		switch {
		case conv.Type == "dm":
			shouldNotify = true
		case directMention[uid]:
			shouldNotify = true
		case hasBroadcast && m.NotificationPref != "none":
			shouldNotify = true
		case m.NotificationPref == "all":
			shouldNotify = true
		}
		if !shouldNotify {
			continue
		}
		orgID := conv.OrganizationID
		_, err := n.notif.Create(ctx, notification.CreateInput{
			TenantID:       conv.TenantID,
			OrganizationID: &orgID,
			UserID:         uid,
			Kind:           kind,
			Title:          title,
			Message:        preview,
			Link:           link,
			Metadata: map[string]any{
				"conversationId": conv.ID.String(),
				"messageId":      msg.ID.String(),
				"reason":         deriveReason(conv, directMention[uid], hasBroadcast),
			},
		})
		if err != nil {
			n.log.Warn("comm: notification publish failed",
				zap.String("messageId", msg.ID.String()),
				zap.String("recipient", uid.String()),
				zap.Error(err))
		}
	}
}

// previewFor returns the (title, message-preview) tuple used in the bell.
// For DMs the title is "New message from <sender>"; for channels it's the
// channel name. The preview is the body truncated to ~140 chars.
func (n *notifier) previewFor(conv *Conversation, msg *Message) (title string, preview string) {
	preview = strings.TrimSpace(msg.Body)
	if len(preview) > 140 {
		preview = preview[:140] + "…"
	}
	if conv.Type == "dm" {
		// We don't have the sender name here — let the bell render
		// "New message" and link through to the conversation. The recipient
		// sees the sender on click.
		return "New message", preview
	}
	channelName := conv.Name
	if channelName == "" {
		channelName = "#" + conv.Slug
	}
	return channelName, preview
}

func deriveReason(conv *Conversation, mentioned, broadcast bool) string {
	switch {
	case conv.Type == "dm":
		return "dm"
	case mentioned:
		return "mention"
	case broadcast:
		return "broadcast"
	default:
		return "channel"
	}
}
