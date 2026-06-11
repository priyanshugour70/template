package ws

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis pub/sub channels. Single-instance deployments work without this
// layer; the publisher could call hub.Broadcast directly. We always go
// through Redis so multi-instance behaviour is identical to single-instance
// behaviour — fewer code paths to reason about.
const (
	channelConvPrefix     = "comm:conv:"     // suffix: conversation_id
	channelPresencePrefix = "comm:presence:" // suffix: organization_id
)

// Broadcaster is what the comm service calls to publish events. It abstracts
// over Redis so the service file doesn't import go-redis directly.
type Broadcaster interface {
	BroadcastToConversation(ctx context.Context, convID uuid.UUID, frame ServerFrame) error
	BroadcastToOrganization(ctx context.Context, orgID uuid.UUID, frame ServerFrame, recipients []uuid.UUID) error
	PublishTyping(ctx context.Context, convID, userID uuid.UUID, untilMS int64) error
}

// PubSub is the Redis-backed Broadcaster. It also runs the subscriber loop
// that delivers events received from peer instances into the local hub.
type PubSub struct {
	r   *redis.Client
	hub *Hub
	log *zap.Logger
}

func NewPubSub(r *redis.Client, hub *Hub, log *zap.Logger) *PubSub {
	return &PubSub{r: r, hub: hub, log: log}
}

// ConversationChannel is exposed so other packages can derive the same names.
func ConversationChannel(convID uuid.UUID) string { return channelConvPrefix + convID.String() }

// OrgPresenceChannel returns the pub/sub channel for presence events
// targeting a given organisation.
func OrgPresenceChannel(orgID uuid.UUID) string { return channelPresencePrefix + orgID.String() }

// envelope is the wire format on Redis. The Frame is the user-visible
// payload; Recipients carries explicit user-id targets for org-wide events
// so receiving instances don't have to query membership.
type envelope struct {
	Frame       ServerFrame `json:"frame"`
	Recipients  []uuid.UUID `json:"recipients,omitempty"`
	ExcludeUser *uuid.UUID  `json:"excludeUser,omitempty"`
}

func (p *PubSub) BroadcastToConversation(ctx context.Context, convID uuid.UUID, frame ServerFrame) error {
	return p.publishEnvelope(ctx, ConversationChannel(convID), envelope{Frame: frame})
}

func (p *PubSub) BroadcastToOrganization(ctx context.Context, orgID uuid.UUID, frame ServerFrame, recipients []uuid.UUID) error {
	return p.publishEnvelope(ctx, OrgPresenceChannel(orgID), envelope{Frame: frame, Recipients: recipients})
}

// PublishTyping wraps the typing broadcast and attaches ExcludeUser so the
// typer's own subscribers don't see their indicator echoed back.
func (p *PubSub) PublishTyping(ctx context.Context, convID, userID uuid.UUID, untilMS int64) error {
	frame := NewTyping(convID, userID, time.UnixMilli(untilMS))
	return p.publishEnvelope(ctx, ConversationChannel(convID), envelope{
		Frame:       frame,
		ExcludeUser: &userID,
	})
}

func (p *PubSub) publishEnvelope(ctx context.Context, channel string, env envelope) error {
	if p.r == nil {
		p.log.Debug("comm/ws: pubsub redis nil, dropping", zap.String("channel", channel))
		return nil
	}
	payload, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.r.Publish(ctx, channel, payload).Err()
}

// Run blocks until ctx is cancelled, dispatching incoming Redis messages
// into the local hub. Intended to be started in its own goroutine during
// bootstrap.
func (p *PubSub) Run(ctx context.Context) {
	if p.r == nil {
		p.log.Warn("comm/ws: pubsub redis nil; subscriber NOT running")
		return
	}
	sub := p.r.PSubscribe(ctx, channelConvPrefix+"*", channelPresencePrefix+"*")
	defer sub.Close()
	p.log.Info("comm/ws: pubsub subscriber ready",
		zap.String("pattern1", channelConvPrefix+"*"),
		zap.String("pattern2", channelPresencePrefix+"*"),
	)
	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			p.dispatch(msg.Channel, msg.Payload)
		}
	}
}

func (p *PubSub) dispatch(channel, payload string) {
	var env envelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		p.log.Warn("comm/ws: bad pubsub envelope", zap.String("channel", channel), zap.Error(err))
		return
	}
	switch {
	case strings.HasPrefix(channel, channelConvPrefix):
		idStr := channel[len(channelConvPrefix):]
		convID, err := uuid.Parse(idStr)
		if err != nil {
			return
		}
		p.hub.BroadcastToConversation(convID, env.Frame, env.ExcludeUser)
	case strings.HasPrefix(channel, channelPresencePrefix):
		// Recipients carries the explicit list of user-ids to notify (the
		// org's online members). Skip if empty.
		if len(env.Recipients) > 0 {
			p.hub.BroadcastToUsers(env.Recipients, env.Frame)
		}
	}
}
