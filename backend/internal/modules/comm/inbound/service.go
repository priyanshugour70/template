package inbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/hash"
	"github.com/your-org/your-service/internal/pkg/tokens"
)

// CommPort is the slice of comm.Service the inbound module needs. Declared
// here as an interface so comm and inbound stay decoupled (and tests can
// stub a mock easily).
type CommPort interface {
	// SendMessageAsWebhook posts a system/webhook message into the channel
	// identified by hookID's conversation. Returns the persisted message id
	// so the caller can include it in the HTTP response (Slack does the
	// same).
	SendMessageAsWebhook(
		ctx context.Context,
		conversationID uuid.UUID,
		hookID uuid.UUID,
		displayName, iconURL, body string,
		attachments []map[string]interface{},
	) (uuid.UUID, error)
}

type Service struct {
	repo *Repository
	comm CommPort
	// publicBaseURL is what we use to render the "POST here" URL in the
	// create response. Configured via APP_BASE_DOMAIN + APP_FRONTEND_SCHEME,
	// passed in at bootstrap.
	publicBaseURL string
	log           *zap.Logger
}

func NewService(repo *Repository, comm CommPort, publicBaseURL string, log *zap.Logger) *Service {
	return &Service{repo: repo, comm: comm, publicBaseURL: publicBaseURL, log: log}
}

// CreateHook mints a token, stores its hash, and returns the plaintext to
// the caller exactly once. Callers MUST be the channel-manager — the handler
// enforces the perm gate before reaching this function.
func (s *Service) CreateHook(ctx context.Context, conversationID uuid.UUID, req CreateHookRequest) (*CreateHookResponse, error) {
	tid := appctx.TenantID(ctx)
	oid := appctx.OrganizationID(ctx)
	if tid == uuid.Nil || oid == uuid.Nil {
		return nil, apperr.New(apperr.CodeForbidden, "no principal context", nil)
	}
	token, err := tokens.New(32)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "generate token failed", err)
	}
	h := &ChannelHook{
		TenantID:       tid,
		OrganizationID: oid,
		ConversationID: conversationID,
		Name:           strings.TrimSpace(req.Name),
		IconURL:        req.IconURL,
		DisplayName:    req.DisplayName,
		TokenHash:      hash.SHA256(token),
		IsActive:       true,
	}
	if err := s.repo.Create(ctx, h); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create hook failed", err)
	}
	return &CreateHookResponse{
		Hook:  *h,
		Token: token,
		URL:   buildPublicURL(s.publicBaseURL, token),
	}, nil
}

func (s *Service) ListHooks(ctx context.Context, conversationID uuid.UUID) ([]ChannelHook, error) {
	rows, err := s.repo.ListByConversation(ctx, conversationID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list hooks failed", err)
	}
	return rows, nil
}

func (s *Service) RevokeHook(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Revoke(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "revoke hook failed", err)
	}
	return nil
}

// PostFromExternal is the heart of the public endpoint. It:
//   1. Looks up the hook by token hash (no user context — auth IS the token).
//   2. Composes a webhook-typed message via comm.
//   3. Bumps the hook's use counters.
// Returns the created message id so the response can include it.
func (s *Service) PostFromExternal(ctx context.Context, token string, req InboundMessageRequest) (uuid.UUID, error) {
	if token == "" {
		return uuid.Nil, apperr.New(apperr.CodeUnauthorized, "missing token", nil)
	}
	hook, err := s.repo.GetByTokenHash(ctx, hash.SHA256(token))
	if err != nil {
		if IsNotFound(err) {
			return uuid.Nil, apperr.New(apperr.CodeUnauthorized, "invalid token", nil)
		}
		return uuid.Nil, apperr.New(apperr.CodeInternal, "hook lookup failed", err)
	}
	if !hook.IsActive {
		return uuid.Nil, apperr.New(apperr.CodeForbidden, "hook revoked", nil)
	}
	// External callers don't carry user context, so we synthesise one from
	// the hook record. comm.SendMessageAsWebhook reads from this synthetic
	// principal (tenant + org only) when persisting.
	hookCtx := appctx.With(ctx, appctx.Principal{
		TenantID:       hook.TenantID,
		OrganizationID: hook.OrganizationID,
	})
	displayName := req.Username
	if displayName == "" {
		displayName = hook.DisplayName
	}
	if displayName == "" {
		displayName = hook.Name
	}
	iconURL := req.IconURL
	if iconURL == "" {
		iconURL = hook.IconURL
	}
	msgID, err := s.comm.SendMessageAsWebhook(
		hookCtx,
		hook.ConversationID,
		hook.ID,
		displayName, iconURL, req.Text,
		req.Attachments,
	)
	if err != nil {
		return uuid.Nil, err
	}
	_ = s.repo.BumpUse(ctx, hook.ID)
	return msgID, nil
}

// buildPublicURL composes the convenience "POST here" URL shown to admins.
// Falls back to a relative path if no base is configured.
func buildPublicURL(base, token string) string {
	base = strings.TrimRight(base, "/")
	if base == "" {
		return "/api/v1/comm/inbound/" + token
	}
	return fmt.Sprintf("%s/api/v1/comm/inbound/%s", base, token)
}

// AbsoluteHookURL is what the handler shows in list/get responses (without
// the secret — derived from a stable hook id). Reserved for Phase 4 when we
// add a per-hook "regenerate" UI; for now the create response is the only
// place the URL is shown.
func AbsoluteHookURL(base string, hookID uuid.UUID) string {
	_ = time.Time{} // kept here to silence imports — Phase 4 may use timestamps
	base = strings.TrimRight(base, "/")
	return fmt.Sprintf("%s/api/v1/comm/inbound/%s", base, hookID.String())
}
