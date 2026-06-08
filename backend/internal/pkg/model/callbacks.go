package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/pkg/appctx"
)

// RegisterCallbacks installs DB-wide GORM callbacks that auto-populate
// created_by, updated_by, and deleted_by from the request context. Call once
// after opening the GORM connection in bootstrap.
//
// Model hooks (BeforeCreate / BeforeUpdate on Base) cover the struct path.
// These callbacks cover bulk operations (db.Model(&X{}).Updates(...),
// db.Delete(&X{}, "id = ?", id)) where the struct hooks would not fire on a
// fully-populated record.
func RegisterCallbacks(db *gorm.DB) error {
	cb := db.Callback()

	if err := cb.Create().Before("gorm:create").Register("app:set_actor_create", setActorCreate); err != nil {
		return err
	}
	if err := cb.Update().Before("gorm:update").Register("app:set_actor_update", setActorUpdate); err != nil {
		return err
	}
	if err := cb.Delete().Before("gorm:delete").Register("app:set_actor_delete", setActorDelete); err != nil {
		return err
	}
	return nil
}

func setActorCreate(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Schema == nil {
		return
	}
	uid := appctx.UserID(db.Statement.Context)
	if uid == uuid.Nil {
		return
	}
	if _, ok := db.Statement.Schema.FieldsByDBName["created_by"]; ok {
		db.Statement.SetColumn("created_by", uid)
	}
	if _, ok := db.Statement.Schema.FieldsByDBName["updated_by"]; ok {
		db.Statement.SetColumn("updated_by", uid)
	}
}

func setActorUpdate(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Schema == nil {
		return
	}
	uid := appctx.UserID(db.Statement.Context)
	if uid == uuid.Nil {
		return
	}
	if _, ok := db.Statement.Schema.FieldsByDBName["updated_by"]; ok {
		db.Statement.SetColumn("updated_by", uid)
	}
}

func setActorDelete(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Schema == nil {
		return
	}
	uid := appctx.UserID(db.Statement.Context)
	if uid == uuid.Nil {
		return
	}
	if _, ok := db.Statement.Schema.FieldsByDBName["deleted_by"]; ok {
		db.Statement.SetColumn("deleted_by", uid)
	}
}
