# src/hooks/

React hooks, grouped by feature. Each hook stays small and composable.

```
hooks/
├── auth/                # useRequireAuth, useLogin, …
├── sample/              # useSamples, useSample, …
└── shared/              # cross-feature helpers (useDebouncedValue, useMediaQuery)
```

A hook should:

- Be a pure function (no global state outside React).
- Compose `services/*`, `stores/*`, or other hooks.
- Never import from `components/*` (that's the wrong direction).
