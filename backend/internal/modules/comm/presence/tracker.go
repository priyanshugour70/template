// Package presence tracks who's online per organisation using Redis. State
// lives entirely in Redis — no Postgres writes for heartbeats, no shared
// memory across instances. A backend instance restart drops all presence;
// clients reconnect and re-publish their state within a few seconds.
//
// Two key shapes:
//
//	presence:user:<userID>  — string "online" with TTL=PresenceTTL.
//	                          Existence implies online; absence implies offline.
//	org:<orgID>:online      — Redis SET of online userIDs in that org,
//	                          maintained alongside the per-user TTL key so
//	                          we can do bulk-status lookups for member lists.
//
// Heartbeats from the WS layer extend the user TTL. When the TTL lapses,
// background sweep (next tick) removes the user from the org set and emits
// an offline event.
package presence

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// PresenceTTL is how long a user stays "online" without a heartbeat. Longer
// than the WS ping interval so a missed beat doesn't immediately flip them
// to offline; short enough that a half-closed socket clears within a minute.
const PresenceTTL = 45 * time.Second

const (
	statusOnline  = "online"
	statusAway    = "away"
	statusOffline = "offline"
)

// Tracker owns the Redis state and emits presence changes via the supplied
// publisher. Stateless; safe to call from any goroutine.
type Tracker struct {
	r         *redis.Client
	publisher Publisher
}

// Publisher is what the tracker uses to fan-out status changes. The ws
// PubSub satisfies this — declared here as an interface so this package
// doesn't import ws (preventing a cycle).
type Publisher interface {
	BroadcastToOrganization(ctx context.Context, orgID uuid.UUID, frame any, recipients []uuid.UUID) error
}

// frameMaker builds the actual JSON frame. Pulled out so this package can
// stay free of the ws package — the ws layer provides this function at
// construction time.
type frameMaker func(userID uuid.UUID, status string) any

// NewTracker wires the Redis client + the cross-package adapter.
func NewTracker(r *redis.Client) *Tracker { return &Tracker{r: r} }

// SetPublisher installs the publisher used for online/offline broadcasts.
// Separate from NewTracker so the bootstrap can construct tracker first and
// patch in the publisher once the ws layer is built.
func (t *Tracker) SetPublisher(p Publisher) { t.publisher = p }

func userKey(uid uuid.UUID) string { return "presence:user:" + uid.String() }
func orgSetKey(oid uuid.UUID) string { return "presence:org:" + oid.String() + ":online" }

// SetOnline records the user as online and adds them to the org set.
// Idempotent — the user may already be online from a prior connection (other
// tab). Caller MUST also call PublishOnline if the user was previously
// offline (caller decides whether to emit a change event; tracker is silent).
func (t *Tracker) SetOnline(ctx context.Context, userID, orgID uuid.UUID) (wasOffline bool, err error) {
	if t.r == nil {
		return false, errors.New("presence: redis unavailable")
	}
	exists, err := t.r.Exists(ctx, userKey(userID)).Result()
	if err != nil {
		return false, err
	}
	wasOffline = exists == 0
	pipe := t.r.Pipeline()
	pipe.Set(ctx, userKey(userID), statusOnline, PresenceTTL)
	pipe.SAdd(ctx, orgSetKey(orgID), userID.String())
	if _, err := pipe.Exec(ctx); err != nil {
		return wasOffline, err
	}
	return wasOffline, nil
}

// Heartbeat extends the user's online TTL without changing org-set
// membership. Called every few seconds from the WS layer.
func (t *Tracker) Heartbeat(ctx context.Context, userID uuid.UUID) error {
	if t.r == nil {
		return nil
	}
	return t.r.Expire(ctx, userKey(userID), PresenceTTL).Err()
}

// SetOffline removes the user's presence key + org-set membership. Returns
// whether the user was previously online so the caller can decide whether
// to broadcast a change.
func (t *Tracker) SetOffline(ctx context.Context, userID, orgID uuid.UUID) (wasOnline bool, err error) {
	if t.r == nil {
		return false, errors.New("presence: redis unavailable")
	}
	deleted, err := t.r.Del(ctx, userKey(userID)).Result()
	if err != nil {
		return false, err
	}
	wasOnline = deleted > 0
	if err := t.r.SRem(ctx, orgSetKey(orgID), userID.String()).Err(); err != nil {
		return wasOnline, err
	}
	return wasOnline, nil
}

// Status reads the current status of a single user. Returns "offline" if
// the key has expired.
func (t *Tracker) Status(ctx context.Context, userID uuid.UUID) (string, error) {
	if t.r == nil {
		return statusOffline, nil
	}
	v, err := t.r.Get(ctx, userKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return statusOffline, nil
		}
		return statusOffline, err
	}
	if v == "" {
		return statusOffline, nil
	}
	return v, nil
}

// BulkStatus returns a map of userID → status using a single pipeline of
// GETs. Used by GET /members responses so the sidebar can paint presence
// dots without N round-trips. Unknown users are absent from the map (caller
// treats absent as "offline").
func (t *Tracker) BulkStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	out := make(map[uuid.UUID]string, len(userIDs))
	if t.r == nil || len(userIDs) == 0 {
		return out, nil
	}
	pipe := t.r.Pipeline()
	cmds := make([]*redis.StringCmd, len(userIDs))
	for i, uid := range userIDs {
		cmds[i] = pipe.Get(ctx, userKey(uid))
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return out, err
	}
	for i, cmd := range cmds {
		v, err := cmd.Result()
		if err != nil || v == "" {
			out[userIDs[i]] = statusOffline
			continue
		}
		out[userIDs[i]] = v
	}
	return out, nil
}

// OrgOnlineUsers returns the SET of users currently marked online in an org.
// Used to compute the recipient list when broadcasting presence changes —
// only online members care about other members coming online.
func (t *Tracker) OrgOnlineUsers(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error) {
	if t.r == nil {
		return nil, nil
	}
	raw, err := t.r.SMembers(ctx, orgSetKey(orgID)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		if id, err := uuid.Parse(s); err == nil {
			out = append(out, id)
		}
	}
	return out, nil
}

// PresenceTTLSeconds is exposed for handler responses and clients that want
// to schedule their next heartbeat with a buffer.
func PresenceTTLSeconds() int { return int(PresenceTTL / time.Second) }

// ParseMillis is a tiny helper for parsing string-encoded epoch millis
// out of pubsub payloads. Lives here instead of a `util` package because
// it's only used by presence-adjacent code.
func ParseMillis(s string) (time.Time, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(n), nil
}
