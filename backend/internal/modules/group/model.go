package group

import (
	"time"

	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// Group is a cross-cutting collection of users / sub-groups.
type Group struct {
	model.BaseTenant
	Slug        string      `gorm:"type:citext;not null"   json:"slug"`
	Name        string      `gorm:"not null"                json:"name"`
	Description string      `                                json:"description,omitempty"`
	Kind        string      `gorm:"not null;default:'custom'" json:"kind"`
	Color       string      `                                json:"color,omitempty"`
	Icon        string      `                                json:"icon,omitempty"`
	IsArchived  bool        `gorm:"not null;default:false" json:"isArchived"`
	Rule        model.JSONB `gorm:"type:jsonb"             json:"rule,omitempty"`
	Metadata    model.JSONB `gorm:"type:jsonb;not null;default:'{}'" json:"metadata,omitempty"`
}

func (Group) TableName() string { return "groups" }

// Member is polymorphic: exactly one of MemberUserID / MemberGroupID is set.
type Member struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID       uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"groupId"`
	MemberUserID  *uuid.UUID `gorm:"type:uuid"                                       json:"memberUserId,omitempty"`
	MemberGroupID *uuid.UUID `gorm:"type:uuid"                                       json:"memberGroupId,omitempty"`
	AddedAt       time.Time  `gorm:"not null;default:now()"                         json:"addedAt"`
	AddedBy       *uuid.UUID `gorm:"type:uuid"                                       json:"addedBy,omitempty"`
}

func (Member) TableName() string { return "group_members" }

type GroupRole struct {
	GroupID uuid.UUID `gorm:"type:uuid;primaryKey;column:group_id" json:"groupId"`
	RoleID  uuid.UUID `gorm:"type:uuid;primaryKey;column:role_id"  json:"roleId"`
}

func (GroupRole) TableName() string { return "group_roles" }

// ── DTOs ───────────────────────────────────────────────────────────────────

type CreateInput struct {
	Slug        string `json:"slug" binding:"required,min=1,max=64"`
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type UpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	IsArchived  *bool   `json:"isArchived,omitempty"`
}

type AddMemberInput struct {
	UserID  *uuid.UUID `json:"userId,omitempty"`
	GroupID *uuid.UUID `json:"groupId,omitempty"`
}

type AssignRolesInput struct {
	RoleIDs []uuid.UUID `json:"roleIds" binding:"required"`
}
