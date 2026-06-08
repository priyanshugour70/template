# src/lib/rbac/

Pure helpers for permission checks. Used by:

- `providers/permissions-provider.tsx` (`usePermissions()` hook)
- `middleware.ts` for coarse-grained route gating
- Server-side route handlers that need to assert permissions before doing work

Adding new helpers: keep them pure (`Set<string>` in, `boolean` out). Never call `useAuth()` from this module — it's intentionally React-free.
