package errors

import (
	stderrors "errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// PostgreSQL error codes used by NormalizeMissingTableError + IsUniqueViolation.
// Full list: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUndefinedTable   = "42P01"
	pgUniqueViolation  = "23505"
	pgForeignKeyError  = "23503"
	CodeSchemaNotReady = "SCHEMA_NOT_READY"
)

// IsNotFound returns true when err is a GORM record-not-found error.
func IsNotFound(err error) bool {
	return stderrors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation returns true when err is a PostgreSQL unique constraint failure.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return stderrors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

// IsForeignKeyViolation returns true when err is a PostgreSQL FK failure.
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return stderrors.As(err, &pgErr) && pgErr.Code == pgForeignKeyError
}

// NormalizeMissingTableError turns the PostgreSQL "relation does not exist" (42P01) error
// into a clear AppError so clients see a fixable message after a failed migration instead
// of a generic INTERNAL_ERROR.
func NormalizeMissingTableError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !stderrors.As(err, &pgErr) || pgErr.Code != pgUndefinedTable {
		return err
	}
	return NewWithDetails(CodeSchemaNotReady,
		"database tables are missing. Apply migrations: ./scripts/migrate.sh",
		map[string]interface{}{"pg_sqlstate": pgErr.Code},
		err)
}
