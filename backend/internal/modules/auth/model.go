package auth

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

type Invite struct {
	model.Base
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index"     json:"tenantId"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index"     json:"organizationId"`
	Email          string         `gorm:"type:citext;not null"          json:"email"`
	FirstName      string         `                                      json:"firstName,omitempty"`
	LastName       string         `                                      json:"lastName,omitempty"`
	JobTitle       string         `                                      json:"jobTitle,omitempty"`
	Department     string         `                                      json:"department,omitempty"`
	TokenHash      []byte         `gorm:"type:bytea;not null;uniqueIndex" json:"-"`
	RoleIDs        pq.StringArray `gorm:"type:uuid[]"                   json:"roleIds,omitempty"`
	InvitedBy      *uuid.UUID     `gorm:"type:uuid"                     json:"invitedBy,omitempty"`
	Message        string         `                                      json:"message,omitempty"`
	Status         string         `gorm:"not null;default:pending"      json:"status"`
	ExpiresAt      time.Time      `gorm:"not null"                       json:"expiresAt"`
	AcceptedAt     *time.Time     `                                      json:"acceptedAt,omitempty"`
	AcceptedBy     *uuid.UUID     `gorm:"type:uuid"                     json:"acceptedBy,omitempty"`
	RevokedAt      *time.Time     `                                      json:"revokedAt,omitempty"`
	RevokedBy      *uuid.UUID     `gorm:"type:uuid"                     json:"revokedBy,omitempty"`
	ResendCount    int            `gorm:"not null;default:0"            json:"resendCount"`
	LastResentAt   *time.Time     `                                      json:"lastResentAt,omitempty"`
	IP             *net.IP        `gorm:"type:inet"                     json:"ip,omitempty"`
	UserAgent      string         `                                      json:"userAgent,omitempty"`
	Metadata       model.JSONB    `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (Invite) TableName() string { return "invites" }

type PasswordResetToken struct {
	model.Base
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	TokenHash []byte     `gorm:"type:bytea;not null;uniqueIndex" json:"-"`
	IP        *net.IP    `gorm:"type:inet"                json:"ip,omitempty"`
	UserAgent string     `                                  json:"userAgent,omitempty"`
	ExpiresAt time.Time  `gorm:"not null"                  json:"expiresAt"`
	UsedAt    *time.Time `                                  json:"usedAt,omitempty"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

type RefreshToken struct {
	model.Base
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"   json:"userId"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index"   json:"tenantId"`
	OrganizationID *uuid.UUID `gorm:"type:uuid"                  json:"organizationId,omitempty"`
	MembershipID   *uuid.UUID `gorm:"type:uuid"                  json:"membershipId,omitempty"`
	TokenHash      []byte     `gorm:"type:bytea;not null;index"  json:"-"`
	FamilyID       uuid.UUID  `gorm:"type:uuid;not null;index"   json:"familyId"`
	ParentID       *uuid.UUID `gorm:"type:uuid"                  json:"parentId,omitempty"`
	DeviceID       string     `                                    json:"deviceId,omitempty"`
	DeviceName     string     `                                    json:"deviceName,omitempty"`
	Client         string     `                                    json:"client,omitempty"`
	IP             *net.IP    `gorm:"type:inet"                  json:"ip,omitempty"`
	UserAgent      string     `                                    json:"userAgent,omitempty"`
	IssuedAt       time.Time  `gorm:"not null;default:now()"     json:"issuedAt"`
	ExpiresAt      time.Time  `gorm:"not null"                    json:"expiresAt"`
	LastUsedAt     *time.Time `                                    json:"lastUsedAt,omitempty"`
	RevokedAt      *time.Time `                                    json:"revokedAt,omitempty"`
	RevokedReason  string     `                                    json:"revokedReason,omitempty"`
	Metadata       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// ── DTOs (request) ─────────────────────────────────────────────────────────

type DiscoverRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Email    string    `json:"email"    binding:"required,email"`
	Password string    `json:"password" binding:"required,min=8"`
	TenantID uuid.UUID `json:"tenantId" binding:"required"`
}

type SwitchOrgRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

type InviteRequest struct {
	Email          string    `json:"email" binding:"required,email"`
	FirstName      string    `json:"firstName,omitempty"`
	LastName       string    `json:"lastName,omitempty"`
	JobTitle       string    `json:"jobTitle,omitempty"`
	Department     string    `json:"department,omitempty"`
	OrganizationID uuid.UUID `json:"organizationId" binding:"required"`
	RoleKeys       []string  `json:"roleKeys,omitempty"`
	Message        string    `json:"message,omitempty"`
}

type AcceptInviteRequest struct {
	Token     string `json:"token"     binding:"required"`
	FirstName string `json:"firstName" binding:"required,min=1,max=80"`
	LastName  string `json:"lastName,omitempty"`
	Password  string `json:"password"  binding:"required,min=8,max=128"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"       binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=128"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword"     binding:"required,min=8,max=128"`
}

// ── DTOs (response) ────────────────────────────────────────────────────────

type DiscoveredTenant struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	LogoURL      string    `json:"logoUrl,omitempty"`
	PrimaryColor string    `json:"primaryColor,omitempty"`
}

type LoginResponse struct {
	AccessToken           string             `json:"accessToken"`
	RefreshToken          string             `json:"refreshToken"`
	AccessTokenExpiresAt  time.Time          `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time          `json:"refreshTokenExpiresAt"`
	TokenType             string             `json:"tokenType"`
	User                  UserSummary        `json:"user"`
	Tenant                DiscoveredTenant   `json:"tenant"`
	ActiveOrganization    *OrgSummary        `json:"activeOrganization,omitempty"`
	Organizations         []OrgSummary       `json:"organizations"`
}

type OrgSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	LogoURL     string    `json:"logoUrl,omitempty"`
	IsDefault   bool      `json:"isDefault"`
	Status      string    `json:"status"`
}

type UserSummary struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName,omitempty"`
	FirstName   string    `json:"firstName,omitempty"`
	LastName    string    `json:"lastName,omitempty"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
}
