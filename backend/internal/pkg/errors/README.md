# internal/pkg/errors/

Application error types and error-code constants.

- `AppError` — structured error with `Code`, `Message`, `Details`, and wrapped `Err`.
- Error codes are string constants the frontend can switch on (`UNAUTHORIZED`, `VALIDATION_ERROR`, `NOT_FOUND`, …).
- Supports `errors.Unwrap()` for error-chain inspection.
- `IsNotFound(err)` — convenience over `errors.Is(err, gorm.ErrRecordNotFound)`.
- `IsUniqueViolation(err)` — true when err is a PostgreSQL `23505` unique constraint error.
- `IsForeignKeyViolation(err)` — true when err is a PostgreSQL `23503` FK error.
- `NormalizeMissingTableError(err)` — converts PostgreSQL `42P01` ("relation does not exist") into a clear `SCHEMA_NOT_READY` `AppError` so deploys with missing migrations fail loudly instead of returning a generic 500.
