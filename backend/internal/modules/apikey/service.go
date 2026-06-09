package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create generates a new plaintext token, stores its SHA-256 hash + first-8
// prefix, and returns the plaintext exactly once. UI must display this to the
// user immediately — no other endpoint will reveal it.
func (s *Service) Create(ctx context.Context, tenantID, orgID uuid.UUID, userID *uuid.UUID, in CreateInput) (*CreateOutput, error) {
	if tenantID == uuid.Nil || orgID == uuid.Nil {
		return nil, apperr.New(apperr.CodeValidation, "tenant + org context required", nil)
	}
	plain, err := generateToken()
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "generate token failed", err)
	}
	hash := hashToken(plain)
	prefix := plain[:8]

	scopes, _ := json.Marshal(in.Scopes)
	if scopes == nil {
		scopes = []byte("[]")
	}

	k := &APIKey{
		UserID:       userID,
		Name:         in.Name,
		Prefix:       prefix,
		TokenHash:    hash,
		Scopes:       scopes,
		RateLimitRPM: in.RateLimitRPM,
		ExpiresAt:    in.ExpiresAt,
	}
	k.TenantID = tenantID
	k.OrganizationID = &orgID
	if err := s.repo.Create(ctx, k); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create api key failed", err)
	}
	return &CreateOutput{APIKey: *k, Token: plain}, nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]APIKey, error) {
	rows, err := s.repo.ListForOrg(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list api keys failed", err)
	}
	return rows, nil
}

func (s *Service) Revoke(ctx context.Context, orgID, id uuid.UUID, by *uuid.UUID) error {
	// Verify ownership before revoking.
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "api key not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load api key failed", err)
	}
	if err := s.repo.Revoke(ctx, id, by); err != nil {
		return apperr.New(apperr.CodeInternal, "revoke api key failed", err)
	}
	return nil
}

// HashToken is exported so middleware (when it's wired) can authenticate by
// hashing the inbound Authorization header value.
func HashToken(plaintext string) string { return hashToken(plaintext) }

func hashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// generateToken returns "sk_" + 32 url-safe hex bytes. Total length 35 chars.
func generateToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(b[:]), nil
}
