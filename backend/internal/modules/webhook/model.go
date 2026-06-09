package webhook

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// Webhook is an outbound HTTP endpoint the platform calls when events fire.
type Webhook struct {
	model.BaseTenant
	Name                string      `gorm:"not null"                json:"name"`
	URL                 string      `gorm:"not null"                json:"url"`
	Events              model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb" json:"events"`
	SecretHash          string      `                                 json:"-"` // sha256 of plaintext; secret returned once on create
	IsActive            bool        `gorm:"not null;default:true"   json:"isActive"`
	Description         string      `                                 json:"description,omitempty"`
	Headers             model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"headers"`
	LastInvokedAt       *time.Time  `                                 json:"lastInvokedAt,omitempty"`
	LastStatus          *int        `gorm:"column:last_status"      json:"lastStatus,omitempty"`
	ConsecutiveFailures int         `gorm:"not null;default:0"      json:"consecutiveFailures"`
	DisabledAt          *time.Time  `                                 json:"disabledAt,omitempty"`
	DisabledReason      string      `                                 json:"disabledReason,omitempty"`
	Metadata            model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (Webhook) TableName() string { return "webhooks" }

// Delivery is one attempt to call the endpoint. The worker writes these as
// events fire; the UI can show recent history.
type Delivery struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WebhookID      uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"webhookId"`
	Event          string     `gorm:"not null"                                        json:"event"`
	Payload        model.JSONB `gorm:"type:jsonb;not null"                            json:"payload"`
	Status         string     `gorm:"not null;default:'pending'"                     json:"status"`
	Attempt        int        `gorm:"not null;default:1"                              json:"attempt"`
	ResponseStatus *int       `gorm:"column:response_status"                          json:"responseStatus,omitempty"`
	ResponseBody   string     `                                                        json:"responseBody,omitempty"`
	ErrorMessage   string     `                                                        json:"errorMessage,omitempty"`
	DurationMs     *int       `gorm:"column:duration_ms"                              json:"durationMs,omitempty"`
	DeliveredAt    *time.Time `                                                        json:"deliveredAt,omitempty"`
	NextRetryAt    *time.Time `                                                        json:"nextRetryAt,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"                          json:"createdAt"`
}

func (Delivery) TableName() string { return "webhook_deliveries" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type CreateInput struct {
	Name        string            `json:"name" binding:"required,min=1,max=100"`
	URL         string            `json:"url" binding:"required,url"`
	Events      []string          `json:"events"`
	Description string            `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type CreateOutput struct {
	Webhook Webhook `json:"webhook"`
	Secret  string  `json:"secret"` // plaintext HMAC secret, shown to user once
}

type UpdateInput struct {
	Name        *string            `json:"name,omitempty"`
	URL         *string            `json:"url,omitempty" binding:"omitempty,url"`
	Events      []string           `json:"events,omitempty"`
	Description *string            `json:"description,omitempty"`
	Headers     map[string]string  `json:"headers,omitempty"`
	IsActive    *bool              `json:"isActive,omitempty"`
}

type TestFireInput struct {
	Event   string                 `json:"event,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type TestFireOutput struct {
	Delivery Delivery `json:"delivery"`
}
