# src/lib/client/

Browser HTTP client. Hits same-origin `/api/v1/*` (proxied by the route handler at `src/app/api/v1/[[...path]]/route.ts`).

- Auto-attaches `Authorization: Bearer <token>` from cookies (override with `skipAuth: true`).
- Standardizes responses to the `ApiResponse<T>` envelope returned by the Go backend (`{ success, data, error, message, timestamp }`).
- Use `api.get / post / put / patch / delete` from feature services; don't import `request` directly from hooks/components.
