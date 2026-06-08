# src/services/

Domain HTTP services. One folder per feature. Each file wraps `lib/client.api` calls into a small typed module that hooks and Server Actions consume.

```
services/
├── auth/index.ts           # authService.login, authService.me, authService.logout
├── sample/index.ts         # sampleService.list, sampleService.create, …
└── …
```

A service:

- Is the **only** place feature endpoints are referenced. Hooks and components never call `api.get("/some-endpoint")` directly.
- Returns the typed `ApiResponse<T>` envelope; callers decide how to surface errors.
- Stays free of React (no hooks, no JSX) so it works in Server Actions, route handlers, and tests.
