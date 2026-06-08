package audit

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// Log mirrors the audit_log partitioned table. Note: no soft-delete here —
// audit rows are append-only and never deleted from the application.
type Log struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OccurredAt      time.Time   `gorm:"not null;default:now()" json:"occurredAt"`
	CorrelationID   string      `                                json:"correlationId,omitempty"`
	TenantID        *uuid.UUID  `gorm:"type:uuid;index"          json:"tenantId,omitempty"`
	OrganizationID  *uuid.UUID  `gorm:"type:uuid;index"          json:"organizationId,omitempty"`
	UserID          *uuid.UUID  `gorm:"type:uuid;index"          json:"userId,omitempty"`
	UserEmail       string      `gorm:"type:citext"              json:"userEmail,omitempty"`
	Method          string      `                                 json:"method,omitempty"`
	Path            string      `                                 json:"path,omitempty"`
	Route           string      `                                 json:"route,omitempty"`
	StatusCode      int         `gorm:"column:status_code"       json:"statusCode,omitempty"`
	LatencyMs       int64       `gorm:"column:latency_ms"        json:"latencyMs,omitempty"`
	IP              *string     `gorm:"type:inet"                json:"ip,omitempty"`
	UserAgent       string      `                                 json:"userAgent,omitempty"`
	Action          string      `                                 json:"action,omitempty"`
	TargetType      string      `                                 json:"targetType,omitempty"`
	TargetID        *uuid.UUID  `gorm:"type:uuid"                json:"targetId,omitempty"`
	ErrorCode       string      `                                 json:"errorCode,omitempty"`
	RequestHeaders  model.JSONB `gorm:"type:jsonb"               json:"requestHeaders,omitempty"`
	RequestBody     model.JSONB `gorm:"type:jsonb"               json:"requestBody,omitempty"`
	ResponseHeaders model.JSONB `gorm:"type:jsonb"               json:"responseHeaders,omitempty"`
	ResponseBody    model.JSONB `gorm:"type:jsonb"               json:"responseBody,omitempty"`
	Metadata        model.JSONB `gorm:"type:jsonb"               json:"metadata,omitempty"`
	CreatedAt       time.Time   `gorm:"not null;default:now()"   json:"createdAt"`
	CreatedBy       *uuid.UUID  `gorm:"type:uuid"                json:"createdBy,omitempty"`
}

func (Log) TableName() string { return "audit_log" }

// ListFilter is the read-side filter.
type ListFilter struct {
	UserID       *uuid.UUID
	UserEmail    string
	Action       string
	TargetType   string
	TargetID     *uuid.UUID
	Method       string
	Path         string
	StatusFrom   int
	StatusTo     int
	OccurredFrom *time.Time
	OccurredTo   *time.Time
	Search       string
}
