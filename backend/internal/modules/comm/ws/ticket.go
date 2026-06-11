package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/your-org/your-service/internal/pkg/tokens"
)

// Ticket is the principal context bound to a WS session at the moment the
// ticket was issued. Stays valid for the lifetime of the connection — we
// don't re-check tenant membership mid-stream, on the assumption that
// connection lifetimes are short relative to RBAC changes. A user kicked
// from the org has their WS dropped when they next try to send.
type Ticket struct {
	UserID         uuid.UUID `json:"u"`
	TenantID       uuid.UUID `json:"t"`
	OrganizationID uuid.UUID `json:"o"`
	IssuedAt       time.Time `json:"i"`
}

// TicketTTL is how long a ticket is valid between issue and consume. Long
// enough to handle a slow page load + WS connect; short enough that a leaked
// query-string token (logs, referrer headers) is useless within seconds.
const TicketTTL = 60 * time.Second

// TicketStore issues + consumes single-use WS auth tokens via Redis. The
// raw token is the cryptographic random string — only its lookup key is
// derived from a fixed prefix. Storage value is the Ticket as JSON.
type TicketStore struct {
	r *redis.Client
}

func NewTicketStore(r *redis.Client) *TicketStore { return &TicketStore{r: r} }

func ticketKey(token string) string { return "comm:ws:ticket:" + token }

// Issue mints a fresh ticket, persists it, and returns the raw token to the
// caller. The raw token is the ONLY way to consume the ticket — nothing
// else (including the caller) should retain it. Lifetime is TicketTTL.
func (s *TicketStore) Issue(ctx context.Context, t Ticket) (token string, expiresAt time.Time, err error) {
	if t.UserID == uuid.Nil || t.TenantID == uuid.Nil || t.OrganizationID == uuid.Nil {
		return "", time.Time{}, errors.New("ticket: missing principal")
	}
	if s.r == nil {
		return "", time.Time{}, errors.New("ticket: redis unavailable")
	}
	tok, err := tokens.New(32)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("ticket: random: %w", err)
	}
	t.IssuedAt = time.Now()
	payload, err := json.Marshal(t)
	if err != nil {
		return "", time.Time{}, err
	}
	exp := time.Now().Add(TicketTTL)
	if err := s.r.Set(ctx, ticketKey(tok), payload, TicketTTL).Err(); err != nil {
		return "", time.Time{}, fmt.Errorf("ticket: persist: %w", err)
	}
	return tok, exp, nil
}

// Consume atomically reads and deletes the ticket. Returns ErrTicketNotFound
// if the token never existed, expired, or was already consumed (the GETDEL
// returns empty in all three cases).
func (s *TicketStore) Consume(ctx context.Context, token string) (*Ticket, error) {
	if token == "" {
		return nil, ErrTicketNotFound
	}
	if s.r == nil {
		return nil, errors.New("ticket: redis unavailable")
	}
	raw, err := s.r.GetDel(ctx, ticketKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("ticket: lookup: %w", err)
	}
	var t Ticket
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, fmt.Errorf("ticket: decode: %w", err)
	}
	return &t, nil
}

// ErrTicketNotFound is returned by Consume for missing / expired / already-
// used tickets. Handlers map this to a 401 — never leak distinguishing info.
var ErrTicketNotFound = errors.New("ws: ticket not found")
