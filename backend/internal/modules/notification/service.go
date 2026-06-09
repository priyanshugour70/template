package notification

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

var allowedKinds = map[string]bool{
	"info": true, "success": true, "warning": true, "error": true,
}

func normalizeKind(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	if !allowedKinds[k] {
		return "info"
	}
	return k
}

// Create issues a notification. Callers (other modules, the worker, etc.)
// invoke this directly — no public POST route is exposed.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Notification, error) {
	if in.UserID == uuid.Nil || in.TenantID == uuid.Nil {
		return nil, apperr.New(apperr.CodeValidation, "userId and tenantId are required", nil)
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, apperr.New(apperr.CodeValidation, "title is required", nil)
	}
	n := &Notification{
		UserID:  in.UserID,
		Kind:    normalizeKind(in.Kind),
		Title:   in.Title,
		Message: in.Message,
		Link:    in.Link,
	}
	n.TenantID = in.TenantID
	n.OrganizationID = in.OrganizationID
	if in.Metadata != nil {
		if b, err := json.Marshal(in.Metadata); err == nil {
			n.Metadata = b
		}
	}
	if err := s.repo.Insert(ctx, n); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create notification failed", err)
	}
	return n, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, filter ListFilter, p pagination.Params) ([]Notification, int64, error) {
	rows, total, err := s.repo.List(ctx, userID, filter, p)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list notifications failed", err)
	}
	return rows, total, nil
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	n, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return 0, apperr.New(apperr.CodeInternal, "unread count failed", err)
	}
	return n, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.repo.MarkRead(ctx, userID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "notification not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "mark notification read failed", err)
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	n, err := s.repo.MarkAllRead(ctx, userID)
	if err != nil {
		return 0, apperr.New(apperr.CodeInternal, "mark all read failed", err)
	}
	return n, nil
}
