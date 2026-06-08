package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB stores arbitrary JSON in a PostgreSQL jsonb column. Use this for
// settings / metadata / features fields throughout the app — keeps SQL access
// uniform while letting Go code marshal whatever it needs.
type JSONB json.RawMessage

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("model.JSONB: unsupported scan type")
	}
	out := make([]byte, len(b))
	copy(out, b)
	*j = out
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("model.JSONB: UnmarshalJSON on nil")
	}
	out := make([]byte, len(b))
	copy(out, b)
	*j = out
	return nil
}

// GormDataType signals GORM AutoMigrate to use jsonb.
func (JSONB) GormDataType() string { return "jsonb" }
