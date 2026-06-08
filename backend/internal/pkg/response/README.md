# internal/pkg/response/

Standardized JSON response helpers for all API endpoints.

- `Body` — global response envelope: `{ success, data, error, message, timestamp }`
- `OK()`, `Created()`, `Accepted()`, `NoContent()` — success responses
- `Fail()` — structured error response with code, message, and optional details
- `Error(err)` — converts `AppError` to the appropriate HTTP status + JSON envelope; non-`AppError` becomes a 500 `INTERNAL_ERROR`
- `PaginatedOK()` — paginated list response with `page`, `limit`, `total`, `totalPages`, `hasNext`, `hasPrev`

All timestamps are emitted in Asia/Kolkata (IST). Change `istLocation` if your service runs in a different timezone.
