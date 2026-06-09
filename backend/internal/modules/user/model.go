package user

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

type User struct {
	model.Base
	Email                   string      `gorm:"type:citext;not null;uniqueIndex" json:"email"`
	EmailVerifiedAt         *time.Time  `                                          json:"emailVerifiedAt,omitempty"`
	PasswordHash            string      `gorm:"column:password_hash"               json:"-"`
	PasswordAlgo            string      `gorm:"not null;default:bcrypt"            json:"-"`
	PasswordChangedAt       *time.Time  `                                           json:"passwordChangedAt,omitempty"`
	MustChangePassword      bool        `                                           json:"mustChangePassword"`
	FirstName               string      `                                           json:"firstName,omitempty"`
	MiddleName              string      `                                           json:"middleName,omitempty"`
	LastName                string      `                                           json:"lastName,omitempty"`
	DisplayName             string      `                                           json:"displayName,omitempty"`
	Username                string      `gorm:"type:citext;uniqueIndex"             json:"username,omitempty"`
	AvatarURL               string      `gorm:"column:avatar_url"                  json:"avatarUrl,omitempty"`
	CoverURL                string      `gorm:"column:cover_url"                   json:"coverUrl,omitempty"`
	Bio                     string      `                                           json:"bio,omitempty"`
	Phone                   string      `                                           json:"phone,omitempty"`
	PhoneVerifiedAt         *time.Time  `                                           json:"phoneVerifiedAt,omitempty"`
	AltEmail                string      `gorm:"type:citext"                        json:"altEmail,omitempty"`
	DateOfBirth             *time.Time  `gorm:"type:date"                          json:"dateOfBirth,omitempty"`
	Gender                  string      `                                           json:"gender,omitempty"`
	JobTitle                string      `                                           json:"jobTitle,omitempty"`
	Department              string      `                                           json:"department,omitempty"`
	EmployeeCode            string      `                                           json:"employeeCode,omitempty"`
	Status                  string      `gorm:"not null;default:invited"           json:"status"`
	Locale                  string      `gorm:"not null;default:en-IN"             json:"locale"`
	Timezone                string      `gorm:"not null;default:Asia/Kolkata"      json:"timezone"`
	Country                 string      `                                           json:"country,omitempty"`
	State                   string      `                                           json:"state,omitempty"`
	City                    string      `                                           json:"city,omitempty"`
	Address                 model.JSONB `gorm:"type:jsonb"                          json:"address,omitempty"`
	Preferences             model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"     json:"preferences"`
	NotificationPreferences model.JSONB `gorm:"column:notification_preferences;type:jsonb;default:'{}'::jsonb" json:"notificationPreferences"`
	Metadata                model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"     json:"metadata"`
	LastLoginAt             *time.Time  `                                           json:"lastLoginAt,omitempty"`
	LastLoginIP             *string     `gorm:"type:inet"                          json:"lastLoginIp,omitempty"`
	LastLoginUserAgent      string      `                                           json:"lastLoginUserAgent,omitempty"`
	FailedLoginCount        int         `gorm:"not null;default:0"                 json:"failedLoginCount"`
	LockedUntil             *time.Time  `                                           json:"lockedUntil,omitempty"`
	MFAEnabled              bool        `gorm:"column:mfa_enabled"                 json:"mfaEnabled"`
	MFASecret               string      `gorm:"column:mfa_secret"                  json:"-"`
	MFARecoveryCodes        model.JSONB `gorm:"column:mfa_recovery_codes;type:jsonb" json:"-"`
	IsSuperAdmin            bool        `gorm:"column:is_super_admin;not null;default:false" json:"isSuperAdmin"`
	PrimaryTenantID         *uuid.UUID  `gorm:"type:uuid"                          json:"primaryTenantId,omitempty"`
	PrimaryOrganizationID   *uuid.UUID  `gorm:"type:uuid"                          json:"primaryOrganizationId,omitempty"`
	SignupSource            string      `                                           json:"signupSource,omitempty"`
	ReferralCode            string      `                                           json:"referralCode,omitempty"`
	MarketingOptIn          bool        `                                           json:"marketingOptIn"`
	TermsAcceptedAt         *time.Time  `                                           json:"termsAcceptedAt,omitempty"`
	TermsVersion            string      `                                           json:"termsVersion,omitempty"`
}

func (User) TableName() string { return "users" }

type Membership struct {
	model.Base
	UserID              uuid.UUID   `gorm:"type:uuid;not null;index"        json:"userId"`
	TenantID            uuid.UUID   `gorm:"type:uuid;not null;index"        json:"tenantId"`
	OrganizationID      uuid.UUID   `gorm:"type:uuid;not null;index"        json:"organizationId"`
	Status              string      `gorm:"not null;default:active"         json:"status"`
	IsDefault           bool        `                                         json:"isDefault"`
	IsOwner             bool        `                                         json:"isOwner"`
	IsBillingContact    bool        `gorm:"column:is_billing_contact"        json:"isBillingContact"`
	JobTitle            string      `                                         json:"jobTitle,omitempty"`
	Department          string      `                                         json:"department,omitempty"`
	DepartmentID        *uuid.UUID  `gorm:"type:uuid;column:department_id"   json:"departmentId,omitempty"`
	EmployeeCode        string      `                                         json:"employeeCode,omitempty"`
	ReportsTo           *uuid.UUID  `gorm:"type:uuid"                        json:"reportsTo,omitempty"`
	InvitedBy           *uuid.UUID  `gorm:"type:uuid"                        json:"invitedBy,omitempty"`
	InvitedAt           *time.Time  `                                         json:"invitedAt,omitempty"`
	JoinedAt            *time.Time  `                                         json:"joinedAt,omitempty"`
	LastActiveAt        *time.Time  `                                         json:"lastActiveAt,omitempty"`
	PermissionsCacheKey string      `                                         json:"-"`
	Settings            model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"settings"`
	Metadata            model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"metadata"`
}

func (Membership) TableName() string { return "memberships" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type CreateUserInput struct {
	Email                 string
	FirstName             string
	LastName              string
	MiddleName            string
	DisplayName           string
	PasswordHash          string
	Status                string // 'invited', 'active'
	PrimaryTenantID       *uuid.UUID
	PrimaryOrganizationID *uuid.UUID
	IsSuperAdmin          bool
	Locale                string
	Timezone              string
	Phone                 string
	JobTitle              string
	Department            string
	EmployeeCode          string
}

type CreateMembershipInput struct {
	UserID         uuid.UUID
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	Status         string
	IsDefault      bool
	IsOwner        bool
	JobTitle       string
	Department     string
	EmployeeCode   string
	InvitedBy      *uuid.UUID
}

// UpdateUserRequest — every string field has an upper length bound to prevent
// storage bloat / DoS via 1MB payloads. URL fields use the `url` validator
// which rejects `javascript:` and similar XSS-loaded schemes.
type UpdateUserRequest struct {
	FirstName               *string     `json:"firstName,omitempty"               binding:"omitempty,max=100"`
	MiddleName              *string     `json:"middleName,omitempty"              binding:"omitempty,max=100"`
	LastName                *string     `json:"lastName,omitempty"                binding:"omitempty,max=100"`
	DisplayName             *string     `json:"displayName,omitempty"             binding:"omitempty,max=200"`
	Username                *string     `json:"username,omitempty"                binding:"omitempty,max=64"`
	AvatarURL               *string     `json:"avatarUrl,omitempty"               binding:"omitempty,url,max=2048"`
	CoverURL                *string     `json:"coverUrl,omitempty"                binding:"omitempty,url,max=2048"`
	Bio                     *string     `json:"bio,omitempty"                     binding:"omitempty,max=2000"`
	Phone                   *string     `json:"phone,omitempty"                   binding:"omitempty,max=32"`
	AltEmail                *string     `json:"altEmail,omitempty"                binding:"omitempty,email,max=254"`
	DateOfBirth             *time.Time  `json:"dateOfBirth,omitempty"`
	Gender                  *string     `json:"gender,omitempty"                  binding:"omitempty,max=32"`
	JobTitle                *string     `json:"jobTitle,omitempty"                binding:"omitempty,max=128"`
	Department              *string     `json:"department,omitempty"              binding:"omitempty,max=128"`
	EmployeeCode            *string     `json:"employeeCode,omitempty"            binding:"omitempty,max=64"`
	Locale                  *string     `json:"locale,omitempty"                  binding:"omitempty,max=16"`
	Timezone                *string     `json:"timezone,omitempty"                binding:"omitempty,max=64"`
	Country                 *string     `json:"country,omitempty"                 binding:"omitempty,max=64"`
	State                   *string     `json:"state,omitempty"                   binding:"omitempty,max=64"`
	City                    *string     `json:"city,omitempty"                    binding:"omitempty,max=64"`
	Address                 model.JSONB `json:"address,omitempty"`
	Preferences             model.JSONB `json:"preferences,omitempty"`
	NotificationPreferences model.JSONB `json:"notificationPreferences,omitempty"`
	Metadata                model.JSONB `json:"metadata,omitempty"`
}

type ListFilter struct {
	Status          string
	Search          string
	Role            string // role key, e.g. "owner"
	JobTitle        string
	Department      string // legacy free-text
	DepartmentID    *uuid.UUID
	MFAEnabled      *bool
	LastLoginAfter  *time.Time
	LastLoginBefore *time.Time
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
}

// UpdateMembershipInput is the patch payload for PATCH /users/:id/memberships/:mid.
// Pointer fields = optional; only non-nil fields are applied.
type UpdateMembershipInput struct {
	DepartmentID *uuid.UUID `json:"departmentId,omitempty"`
	JobTitle     *string    `json:"jobTitle,omitempty"`
	Department   *string    `json:"department,omitempty"`
	EmployeeCode *string    `json:"employeeCode,omitempty"`
	ReportsTo    *uuid.UUID `json:"reportsTo,omitempty"`
}
