# internal/pkg/errors/

Application error types and error-code constants.

- `AppError` — structured error with `Code`, `Message`, `Details`, and wrapped `Err`.
- Error codes are string constants the frontend can switch on (`UNAUTHORIZED`, `VALIDATION_ERROR`, `NOT_FOUND`, …).
- Supports `errors.Unwrap()` for error-chain inspection.
- `IsNotFound(err)` — convenience over `errors.Is(err, gorm.ErrRecordNotFound)`.
- `NormalizeMissingTableError(err)` — converts MySQL 1146 ("table doesn't exist") into a clear `SCHEMA_NOT_READY` `AppError` so deploys with missing migrations fail loudly instead of returning a generic 500.
