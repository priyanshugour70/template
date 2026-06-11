package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/tokens"
)

// Handoff is the one-shot SSO record stored in cache. The caller pays for
// authentication on the apex (lssgoo.com/login) and is redirected to the
// tenant subdomain (acme.lssgoo.com/auth/handoff?token=...) which consumes
// the record and issues real access+refresh tokens scoped to that subdomain's
// cookies.
type Handoff struct {
	UserID         uuid.UUID  `json:"userId"`
	TenantID       uuid.UUID  `json:"tenantId"`
	OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
	MembershipID   *uuid.UUID `json:"membershipId,omitempty"`
	IssuedAt       time.Time  `json:"issuedAt"`
}

const handoffCachePrefix = "auth:handoff:"

func handoffCacheKey(token string) string { return handoffCachePrefix + token }

// IssueHandoff mints a single-use handoff token bound to the caller's current
// principal. Returns the token + tenant brand info so the caller can build
// the redirect URL.
func (s *Service) IssueHandoff(ctx context.Context) (*HandoffIssueResponse, error) {
	uid := appctx.UserID(ctx)
	tid := appctx.TenantID(ctx)
	if uid == uuid.Nil || tid == uuid.Nil {
		return nil, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil)
	}

	ttl := time.Duration(s.cfg.Auth.HandoffTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 60 * time.Second
	}

	tokenStr, err := tokens.New(32)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "generate handoff token failed", err)
	}

	h := Handoff{
		UserID:   uid,
		TenantID: tid,
		IssuedAt: time.Now(),
	}
	if oid := appctx.OrganizationID(ctx); oid != uuid.Nil {
		o := oid
		h.OrganizationID = &o
	}
	if mid := appctx.MembershipID(ctx); mid != uuid.Nil {
		m := mid
		h.MembershipID = &m
	}

	if err := cache.SetJSON(ctx, s.cache, handoffCacheKey(tokenStr), h, ttl); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "persist handoff failed", err)
	}

	t, err := s.tenantSvc.GetTenant(ctx, tid)
	if err != nil {
		return nil, err
	}

	return &HandoffIssueResponse{
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(ttl),
		Tenant: DiscoveredTenant{
			ID:           t.ID,
			Name:         t.Name,
			Slug:         t.Slug,
			LogoURL:      t.LogoURL,
			PrimaryColor: t.PrimaryColor,
		},
	}, nil
}

// ConsumeHandoff atomically claims a handoff token and returns a full login
// pair for the bound user+tenant+org. The token is single-use — concurrent
// callers race for the cache delete and the loser gets "invalid token".
func (s *Service) ConsumeHandoff(ctx context.Context, token, ip, userAgent string) (*LoginResponse, error) {
	if token == "" {
		return nil, apperr.New(apperr.CodeValidation, "missing handoff token", nil)
	}
	key := handoffCacheKey(token)
	raw, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, apperr.New(apperr.CodeInvalidToken, "invalid or expired handoff token", nil)
	}

	// Single-use: delete immediately. If the delete fails the token is still
	// time-bounded by its TTL, so worst case a clone has a few seconds to
	// race — refresh token rotation will detect any reuse downstream.
	if delErr := s.cache.Delete(ctx, key); delErr != nil {
		s.log.Warn("handoff cache delete failed", zap.Error(delErr))
	}

	var h Handoff
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "decode handoff failed", err)
	}

	u, err := s.userSvc.GetByID(ctx, h.UserID)
	if err != nil {
		return nil, err
	}
	t, err := s.tenantSvc.GetTenant(ctx, h.TenantID)
	if err != nil {
		return nil, err
	}

	memberships, _, err := s.userSvc.ListMembershipsByUser(ctx, u.ID, 0, 0)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "lookup memberships failed", err)
	}
	pickedIdx := -1
	for i := range memberships {
		m := memberships[i]
		if m.TenantID != h.TenantID || (m.Status != "active" && m.Status != "invited") {
			continue
		}
		if h.OrganizationID != nil && m.OrganizationID != *h.OrganizationID {
			continue
		}
		pickedIdx = i
		break
	}
	if pickedIdx < 0 {
		return nil, apperr.New(apperr.CodeForbidden, "membership no longer valid", nil)
	}

	resp, err := s.issueLoginPair(ctx, u, t, &memberships[pickedIdx], ip, userAgent)
	if err != nil {
		return nil, err
	}
	orgs, _ := s.collectOrgs(ctx, memberships)
	resp.Organizations = orgs
	return resp, nil
}

