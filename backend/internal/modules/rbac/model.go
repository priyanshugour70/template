package rbac

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// ── DB models ──────────────────────────────────────────────────────────────

type Permission struct {
	model.Base
	Key         string      `gorm:"not null;uniqueIndex"            json:"key"`
	Resource    string      `gorm:"not null;index"                   json:"resource"`
	Action      string      `gorm:"not null"                         json:"action"`
	Description string      `                                         json:"description,omitempty"`
	Category    string      `                                         json:"category,omitempty"`
	IsSystem    bool        `gorm:"not null;default:true"            json:"isSystem"`
	IsDangerous bool        `gorm:"not null;default:false"           json:"isDangerous"`
	Metadata    model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"metadata"`
}

func (Permission) TableName() string { return "permissions" }

type Role struct {
	model.Base
	TenantID       *uuid.UUID  `gorm:"type:uuid;index"                  json:"tenantId,omitempty"`
	OrganizationID *uuid.UUID  `gorm:"type:uuid;index"                  json:"organizationId,omitempty"`
	Key            string      `gorm:"not null"                          json:"key"`
	Name           string      `gorm:"not null"                          json:"name"`
	Description    string      `                                          json:"description,omitempty"`
	IsSystem       bool        `gorm:"not null;default:false"           json:"isSystem"`
	IsDefault      bool        `gorm:"not null;default:false"           json:"isDefault"`
	IsAssignable   bool        `gorm:"not null;default:true"            json:"isAssignable"`
	Priority       int         `gorm:"not null;default:0"               json:"priority"`
	Color          string      `                                          json:"color,omitempty"`
	Icon           string      `                                          json:"icon,omitempty"`
	Metadata       model.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"   json:"metadata"`
}

func (Role) TableName() string { return "roles" }

type RolePermission struct {
	RoleID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"roleId"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"permissionId"`
	GrantedAt    time.Time  `gorm:"not null"              json:"grantedAt"`
	GrantedBy    *uuid.UUID `gorm:"type:uuid"             json:"grantedBy,omitempty"`
}

func (RolePermission) TableName() string { return "role_permissions" }

type MembershipRole struct {
	MembershipID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"membershipId"`
	RoleID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"roleId"`
	GrantedAt    time.Time  `gorm:"not null"              json:"grantedAt"`
	GrantedBy    *uuid.UUID `gorm:"type:uuid"             json:"grantedBy,omitempty"`
	ExpiresAt    *time.Time `                              json:"expiresAt,omitempty"`
}

func (MembershipRole) TableName() string { return "membership_roles" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Key           string   `json:"key" binding:"required,min=1,max=64"`
	Name          string   `json:"name" binding:"required,min=1,max=128"`
	Description   string   `json:"description,omitempty"`
	Priority      int      `json:"priority,omitempty"`
	Color         string   `json:"color,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	PermissionKeys []string `json:"permissionKeys,omitempty"`
}

type UpdateRoleRequest struct {
	Name          *string  `json:"name,omitempty" binding:"omitempty,min=1,max=128"`
	Description   *string  `json:"description,omitempty"`
	Priority      *int     `json:"priority,omitempty"`
	Color         *string  `json:"color,omitempty"`
	Icon          *string  `json:"icon,omitempty"`
	IsAssignable  *bool    `json:"isAssignable,omitempty"`
	PermissionKeys []string `json:"permissionKeys,omitempty"` // if non-nil, replaces existing perms
}

type AssignRolesRequest struct {
	RoleKeys []string `json:"roleKeys" binding:"required,min=1"`
}

// SystemRoleSeed defines the role created automatically when an organization
// is provisioned. Initial set: Owner, Admin, Member.
type SystemRoleSeed struct {
	Key           string
	Name          string
	Description   string
	IsDefault     bool
	Priority      int
	PermissionKey func(catalog []string) []string // selector against the seeded catalog
}

func DefaultSystemRoles() []SystemRoleSeed {
	return []SystemRoleSeed{
		{
			Key:         "owner",
			Name:        "Owner",
			Description: "Full access to everything in the organization.",
			Priority:    100,
			PermissionKey: func(catalog []string) []string {
				return catalog
			},
		},
		{
			Key:         "admin",
			Name:        "Admin",
			Description: "Administrative access. Cannot delete the tenant or top-level org.",
			Priority:    80,
			PermissionKey: func(catalog []string) []string {
				blocked := map[string]bool{"tenant.delete": true, "org.delete": true, "user.impersonate": true}
				out := make([]string, 0, len(catalog))
				for _, k := range catalog {
					if !blocked[k] {
						out = append(out, k)
					}
				}
				return out
			},
		},
		{
			Key:         "member",
			Name:        "Member",
			Description: "Read-only baseline access.",
			IsDefault:   true,
			Priority:    10,
			PermissionKey: func(catalog []string) []string {
				out := []string{}
				for _, k := range catalog {
					if isReadOnly(k) {
						out = append(out, k)
					}
				}
				return out
			},
		},
	}
}

func isReadOnly(key string) bool {
	// any *.read / *.list permission, plus a couple of friendly self perms.
	for _, suffix := range []string{".read", ".list"} {
		if len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}
