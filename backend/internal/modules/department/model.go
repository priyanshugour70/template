package department

import (
	"github.com/google/uuid"

	"github.com/your-org/your-service/internal/pkg/model"
)

// Department is a hierarchical org unit. Single tree per organization.
type Department struct {
	model.BaseTenant
	ParentID      *uuid.UUID  `gorm:"type:uuid;index"        json:"parentId,omitempty"`
	Slug          string      `gorm:"type:citext;not null"   json:"slug"`
	Name          string      `gorm:"not null"                json:"name"`
	Description   string      `                                json:"description,omitempty"`
	CostCenter    string      `gorm:"column:cost_center"     json:"costCenter,omitempty"`
	ManagerUserID *uuid.UUID  `gorm:"type:uuid"               json:"managerUserId,omitempty"`
	Color         string      `                                json:"color,omitempty"`
	Icon          string      `                                json:"icon,omitempty"`
	IsArchived    bool        `gorm:"not null;default:false" json:"isArchived"`
	SortOrder     int         `gorm:"not null;default:0"     json:"sortOrder"`
	Metadata      model.JSONB `gorm:"type:jsonb;not null;default:'{}'" json:"metadata,omitempty"`
}

func (Department) TableName() string { return "departments" }

// Closure is the precomputed (ancestor, descendant, depth) triple.
type Closure struct {
	AncestorID   uuid.UUID `gorm:"type:uuid;primaryKey;column:ancestor_id"   json:"ancestorId"`
	DescendantID uuid.UUID `gorm:"type:uuid;primaryKey;column:descendant_id" json:"descendantId"`
	Depth        int       `gorm:"not null"                                   json:"depth"`
}

func (Closure) TableName() string { return "department_closure" }

// DeptRole is a role grant attached to a department; applies to all members
// of the department and (by closure) its descendants.
type DeptRole struct {
	DepartmentID uuid.UUID `gorm:"type:uuid;primaryKey;column:department_id" json:"departmentId"`
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey;column:role_id"       json:"roleId"`
}

func (DeptRole) TableName() string { return "department_roles" }

// CreateInput / UpdateInput / MoveInput are the public DTOs.
type CreateInput struct {
	ParentID      *uuid.UUID `json:"parentId,omitempty"`
	Slug          string     `json:"slug" binding:"required,min=1,max=64"`
	Name          string     `json:"name" binding:"required,min=1,max=128"`
	Description   string     `json:"description,omitempty"`
	CostCenter    string     `json:"costCenter,omitempty"`
	ManagerUserID *uuid.UUID `json:"managerUserId,omitempty"`
	Color         string     `json:"color,omitempty"`
	Icon          string     `json:"icon,omitempty"`
	SortOrder     int        `json:"sortOrder,omitempty"`
}

type UpdateInput struct {
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CostCenter    *string    `json:"costCenter,omitempty"`
	ManagerUserID *uuid.UUID `json:"managerUserId,omitempty"`
	Color         *string    `json:"color,omitempty"`
	Icon          *string    `json:"icon,omitempty"`
	IsArchived    *bool      `json:"isArchived,omitempty"`
	SortOrder     *int       `json:"sortOrder,omitempty"`
}

type MoveInput struct {
	ParentID *uuid.UUID `json:"parentId"`
}

type AssignRolesInput struct {
	RoleIDs []uuid.UUID `json:"roleIds" binding:"required"`
}

// Node is the tree-view DTO returned by the tree endpoint.
type Node struct {
	Department
	Children []*Node `json:"children,omitempty"`
}
