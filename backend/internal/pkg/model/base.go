// Package model provides Base and BaseTenant mixins that every table embeds for
// consistent audit columns and soft-delete behavior across the entire backend.
//
// Every domain model must embed Base (single-tenant tables) or BaseTenant
// (tenant-scoped tables). Hooks auto-populate created_by/updated_by/deleted_by
// from appctx on the request context; see RegisterCallbacks for the bulk-update
// and soft-delete actor capture.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/appctx"
)

// Base — every table embeds this. Soft-delete via gorm.DeletedAt.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"not null;default:now()"                          json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"                          json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                            json:"deletedAt,omitempty" swaggerignore:"true"`
	CreatedBy *uuid.UUID     `gorm:"type:uuid"                                        json:"createdBy,omitempty"`
	UpdatedBy *uuid.UUID     `gorm:"type:uuid"                                        json:"updatedBy,omitempty"`
	DeletedBy *uuid.UUID     `gorm:"type:uuid"                                        json:"deletedBy,omitempty"`
}

// BeforeCreate stamps a fresh UUID (if empty) and captures the actor.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if uid := appctx.UserID(tx.Statement.Context); uid != uuid.Nil {
		if b.CreatedBy == nil {
			id := uid
			b.CreatedBy = &id
		}
		if b.UpdatedBy == nil {
			id := uid
			b.UpdatedBy = &id
		}
	}
	return nil
}

// BeforeUpdate captures the actor on row-level updates.
func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	if uid := appctx.UserID(tx.Statement.Context); uid != uuid.Nil {
		id := uid
		b.UpdatedBy = &id
	}
	return nil
}

// BaseTenant — every tenant-scoped table embeds this. OrganizationID is nullable
// for tables that belong to a tenant but not a specific org (e.g. tenant-wide
// settings).
type BaseTenant struct {
	Base
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenantId"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index"           json:"organizationId,omitempty"`
}
