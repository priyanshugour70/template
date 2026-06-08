# src/providers/

App-root React providers mounted by `app/layout.tsx`. Each provider wraps the entire tree and exposes a hook via context.

| File | Purpose |
|------|---------|
| `theme-provider.tsx` | Wraps `next-themes` `ThemeProvider` so `class="dark"` can toggle modes. Also mounts `<ThemeVars/>` and the inline palette bootstrap script. |
| `query-provider.tsx` | TanStack Query `QueryClient` per browser session. |
| `auth-provider.tsx` | Loads the current user from `/api/v1/auth/me`, exposes `useAuth()`. Handles logout + token clearing. |
| `permissions-provider.tsx` | Loads RBAC permissions for the signed-in user, exposes `usePermissions()`. |
| `index.ts` | Re-exports every provider. |

## Ordering

```tsx
<ThemeProvider>            // mounts ThemeVars + reads/writes data-palette
  <QueryProvider>          // outer of any provider that fetches
    <AuthProvider>
      <PermissionsProvider>
        {children}
      </PermissionsProvider>
    </AuthProvider>
  </QueryProvider>
</ThemeProvider>
```

`AuthProvider` depends on the API client, so it must be inside `QueryProvider`.
