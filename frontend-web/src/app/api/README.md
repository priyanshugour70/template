# app/api/

Server-side routes (Route Handlers).

- `v1/[[...path]]/route.ts` — **catch-all proxy** to the backend at `process.env.API_URL`. The browser hits `/api/v1/*` (same origin), this handler forwards to the Go backend. Keeps cookies same-origin and avoids CORS entirely.

Add additional Next.js-served endpoints (webhook receivers, edge functions) under `app/api/` as siblings to `v1/`.
