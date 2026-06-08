# internal/pkg/

Shared infrastructure packages used across modules. These are foundational utilities, not domain logic.

## Packages

| Package | Purpose |
|---------|---------|
| `errors` | Typed application errors (`AppError`) with stable error codes (`UNAUTHORIZED`, `NOT_FOUND`, `VALIDATION_ERROR`, …). Includes `IsNotFound()` for GORM record-not-found checks, `IsUniqueViolation` / `IsForeignKeyViolation` for PostgreSQL constraint errors, and `NormalizeMissingTableError()` for `42P01` (relation does not exist). |
| `response` | Standardized Gin HTTP response helpers: `OK()`, `Created()`, `Accepted()`, `NoContent()`, `Fail()`, `Error()`, `PaginatedOK()`. Maps `AppError` codes to HTTP status codes automatically. |
| `logger` | Zap logger factory with structured logging, environment-aware encoders, and daily rotating log files via lumberjack. |
| `mail` | SMTP email sender interface and transactional HTML templates (invitations, password resets, simple notifications). |
| `correlation` | Request correlation ID generation and `context.Context` propagation. |
| `slug` | URL-safe slug normalization. |
| `versioning` | Versioned-entity helpers: next-version-number, deactivate-previous-active, approve-and-activate, soft-delete. |

## Dependency rules

- `pkg/` packages MUST NOT import from `internal/modules/` or `internal/clients/`.
- `pkg/` packages may import from `internal/config/` for typed config structs.
- All modules may import from `pkg/`.
