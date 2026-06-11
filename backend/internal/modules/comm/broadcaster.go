package comm

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/modules/comm/ws"
)

// wsBroadcaster adapts comm.BroadcasterPort onto ws.Broadcaster. Lives in
// the comm package (not ws/) so it imports ws and not the other way around
// — preventing the cycle that would otherwise form when ws needed to know
// comm's view shapes.
type wsBroadcaster struct {
	out ws.Broadcaster
	log *zap.Logger
}

// NewWSBroadcaster wires the comm-side adapter onto a ready ws Broadcaster
// (typically the Redis pub/sub publisher). Returns a BroadcasterPort the
// comm.Service can hold without importing the ws package directly.
func NewWSBroadcaster(out ws.Broadcaster, log *zap.Logger) BroadcasterPort {
	return &wsBroadcaster{out: out, log: log}
}

// publish serialises the view shape to JSON once so every per-event method
// has identical payload semantics. Errors are logged and swallowed —
// broadcast failures are recoverable: the receiver refetches via REST.
func (b *wsBroadcaster) publish(ctx context.Context, convID uuid.UUID, build func(json.RawMessage) ws.ServerFrame, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		b.log.Debug("comm/ws bcast: marshal failed", zap.Error(err))
		return
	}
	frame := build(payload)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) MessageCreated(ctx context.Context, convID uuid.UUID, view MessageView) {
	b.publish(ctx, convID, func(p json.RawMessage) ws.ServerFrame {
		return ws.NewMessageCreated(convID, p)
	}, view)
}

func (b *wsBroadcaster) MessageUpdated(ctx context.Context, convID uuid.UUID, view MessageView) {
	b.publish(ctx, convID, func(p json.RawMessage) ws.ServerFrame {
		return ws.NewMessageUpdated(convID, p)
	}, view)
}

func (b *wsBroadcaster) MessageDeleted(ctx context.Context, convID, msgID uuid.UUID) {
	frame := ws.NewMessageDeleted(convID, msgID)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) ReactionAdded(ctx context.Context, convID, msgID, userID uuid.UUID, emoji string) {
	frame := ws.NewReactionAdded(convID, msgID, userID, emoji)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) ReactionRemoved(ctx context.Context, convID, msgID, userID uuid.UUID, emoji string) {
	frame := ws.NewReactionRemoved(convID, msgID, userID, emoji)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) Read(ctx context.Context, convID, userID, lastReadMessageID uuid.UUID) {
	frame := ws.NewRead(convID, userID, lastReadMessageID)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) MemberAdded(ctx context.Context, convID uuid.UUID, view ConversationMemberView) {
	b.publish(ctx, convID, func(p json.RawMessage) ws.ServerFrame {
		return ws.NewMemberAdded(convID, view.UserID, p)
	}, view)
}

func (b *wsBroadcaster) MemberRemoved(ctx context.Context, convID, userID uuid.UUID) {
	frame := ws.NewMemberRemoved(convID, userID)
	if err := b.out.BroadcastToConversation(ctx, convID, frame); err != nil {
		b.log.Debug("comm/ws bcast: publish failed", zap.Error(err))
	}
}

func (b *wsBroadcaster) ConversationUpdated(ctx context.Context, conv Conversation) {
	b.publish(ctx, conv.ID, func(p json.RawMessage) ws.ServerFrame {
		return ws.NewConversationUpdated(conv.ID, p)
	}, conv)
}
