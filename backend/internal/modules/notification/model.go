package notification

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// Notification is a per-user dashboard notification.
type Notification struct {
	model.BaseTenant
	UserID   uuid.UUID   `gorm:"type:uuid;not null;index"           json:"userId"`
	Kind     string      `gorm:"not null;default:'info'"            json:"kind"`
	Title    string      `gorm:"not null"                           json:"title"`
	Message  string      `                                           json:"message,omitempty"`
	Link     string      `                                           json:"link,omitempty"`
	IsRead   bool        `gorm:"column:is_read;not null;default:false" json:"isRead"`
	ReadAt   *time.Time  `                                           json:"readAt,omitempty"`
	Metadata model.JSONB `gorm:"type:jsonb;not null;default:'{}'"   json:"metadata,omitempty"`
}

func (Notification) TableName() string { return "notifications" }

// CreateInput is the service-level input for issuing a new notification.
type CreateInput struct {
	TenantID       uuid.UUID
	OrganizationID *uuid.UUID
	UserID         uuid.UUID
	Kind           string
	Title          string
	Message        string
	Link           string
	Metadata       map[string]any
}

// ListFilter is the read-side filter for the bell list.
type ListFilter struct {
	UnreadOnly bool
}
