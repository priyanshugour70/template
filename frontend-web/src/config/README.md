# src/config/

Static, build-time configuration: route maps, feature flags, environment readers, third-party SDK keys (public).

- One file per concern: `routes.ts`, `features.ts`, `env.ts`.
- Never put secrets here — anything imported into client code is shipped to the browser. Server-only secrets stay in `process.env` and are read from `src/lib/server/`.
