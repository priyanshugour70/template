package apikey

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// APIKey is the stored record. Plaintext key is never persisted — only its
// SHA-256 hash. The prefix lets the UI show "sk_abc12345…" without storing the
// secret part.
type APIKey struct {
	model.BaseTenant
	UserID       *uuid.UUID  `gorm:"type:uuid"               json:"userId,omitempty"`
	Name         string      `gorm:"not null"                 json:"name"`
	Prefix       string      `gorm:"not null"                 json:"prefix"`
	TokenHash    string      `gorm:"not null;uniqueIndex"    json:"-"`
	Scopes       model.JSONB `gorm:"type:jsonb;default:'[]'::jsonb" json:"scopes"`
	RateLimitRPM *int        `gorm:"column:rate_limit_rpm"   json:"rateLimitRpm,omitempty"`
	LastUsedAt   *time.Time  `                                 json:"lastUsedAt,omitempty"`
	LastUsedIP   *string     `gorm:"type:inet"                json:"lastUsedIp,omitempty"`
	ExpiresAt    *time.Time  `                                 json:"expiresAt,omitempty"`
	RevokedAt    *time.Time  `                                 json:"revokedAt,omitempty"`
	RevokedBy    *uuid.UUID  `gorm:"type:uuid"               json:"revokedBy,omitempty"`
	Metadata     model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (APIKey) TableName() string { return "api_keys" }

// CreateInput is the request to issue a new key.
type CreateInput struct {
	Name         string     `json:"name" binding:"required,min=1,max=100"`
	Scopes       []string   `json:"scopes,omitempty"`
	RateLimitRPM *int       `json:"rateLimitRpm,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
}

// CreateOutput returns the plaintext token exactly once.
type CreateOutput struct {
	APIKey APIKey `json:"apiKey"`
	Token  string `json:"token"` // plaintext sk_… token, shown to user once
}
