package audit

import (
	"context"
	"encoding/json"
	"net"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/middleware"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
	"github.com/your-org/your-service/internal/queue"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// ── read ───────────────────────────────────────────────────────────────────

func (s *Service) List(ctx context.Context, tenantID *uuid.UUID, orgID *uuid.UUID, filter ListFilter, p pagination.Params) ([]Log, int64, error) {
	rows, total, err := s.repo.List(ctx, tenantID, orgID, filter, p)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list audit logs failed", err)
	}
	return rows, total, nil
}

func (s *Service) Get(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) (*Log, error) {
	row, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "audit log not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch audit log failed", err)
	}
	return row, nil
}

// ── worker consumer ────────────────────────────────────────────────────────

// HandlerFunc returns a queue.Handler that decodes AuditEvent messages and
// inserts them into the audit_log table. Use this in cmd/worker.
func (s *Service) HandlerFunc() queue.Handler {
	return func(ctx context.Context, msg *queue.Message) error {
		var evt middleware.AuditEvent
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			s.log.Warn("audit decode failed", zap.Error(err))
			return nil
		}
		row := Log{
			ID:              evt.ID,
			OccurredAt:      evt.OccurredAt,
			CorrelationID:   evt.CorrelationID,
			TenantID:        evt.TenantID,
			OrganizationID:  evt.OrganizationID,
			UserID:          evt.UserID,
			UserEmail:       evt.UserEmail,
			Method:          evt.Method,
			Path:            evt.Path,
			Route:           evt.Route,
			StatusCode:      evt.StatusCode,
			LatencyMs:       evt.LatencyMs,
			UserAgent:       evt.UserAgent,
			Action:          evt.Action,
			TargetType:      evt.TargetType,
			TargetID:        evt.TargetID,
			ErrorCode:       evt.ErrorCode,
			RequestHeaders:  jsonbFromMap(evt.RequestHeaders),
			RequestBody:     toJSONB(evt.RequestBody),
			ResponseHeaders: jsonbFromMap(evt.ResponseHeaders),
			ResponseBody:    toJSONB(evt.ResponseBody),
		}
		if evt.IP != "" {
			if parsed := net.ParseIP(evt.IP); parsed != nil {
				row.IP = &parsed
			}
		}
		if err := s.repo.Insert(ctx, &row); err != nil {
			s.log.Error("audit insert failed", zap.Error(err), zap.String("correlation_id", evt.CorrelationID))
			return err
		}
		return nil
	}
}

func jsonbFromMap(m map[string]string) []byte {
	if len(m) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func toJSONB(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return nil
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return out
}
