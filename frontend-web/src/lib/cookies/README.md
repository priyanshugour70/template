# src/lib/cookies/

Token cookie helpers built on `nookies` so the same code runs server + client.

- `getTokens()` reads `app_access_token` + `app_refresh_token`.
- `setTokens({...}, maxAgeSeconds)` writes both with `SameSite=Lax`, `Secure` in production.
- `clearTokens()` removes both.

Backed by middleware (`src/middleware.ts`) which checks for `app_access_token` and redirects to `/auth/login` when missing on protected routes.
