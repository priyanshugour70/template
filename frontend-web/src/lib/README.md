# src/lib/

Pure utility modules. **No React.** Anything that imports `react` belongs in `hooks/` or `components/`.

| Folder | Purpose |
|--------|---------|
| `cn.ts` | `cn()` — Tailwind class-name concatenation (clsx + tailwind-merge). |
| `client/` | Browser HTTP client. Wraps `fetch`, attaches Authorization header from cookies, returns the typed `ApiResponse<T>` envelope. |
| `cookies/` | Read/write/clear auth cookies via `nookies` (works server + client). |
| `rbac/` | Permission-check helpers shared by middleware and providers. |
| `routing/` | Route-map helpers, builders for parameterized routes, redirect helpers. |
| `server/` | Server-only helpers. **Must start with `import "server-only";`** so accidental client imports break the build. |
| `session/` | Token + user-session helpers used by both `services/auth` and the auth provider. |
