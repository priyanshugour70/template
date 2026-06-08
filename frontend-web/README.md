# frontend-web

Next.js 16 + React 19 + Tailwind v4 web frontend template with **shadcn/ui** primitives, **5 runtime-switchable color palettes** (Lss Goo design tokens), Zustand stores, Zod validation, and a same-origin API proxy.

> Read [`AGENTS.md`](AGENTS.md) before assuming any Next.js API — this is Next.js 16 with breaking changes. Look in `node_modules/next/dist/docs/` for the authoritative docs.

## Stack

- **Next.js 16** (App Router, React Compiler enabled)
- **React 19**
- **TypeScript 5** strict
- **Tailwind v4** with `@tailwindcss/postcss`
- **shadcn/ui** primitives over Radix
- **next-themes** for light/dark
- **Dynamic palettes** via `data-palette` attribute on `<html>`
- **Zustand** for client state
- **Zod** for runtime validation
- **TanStack Query** for data fetching/cache
- **TanStack Table** for tables
- **lucide-react** for icons
- **nookies** for cookie I/O (server + client)

## Quick start

```bash
pnpm install
cp .env.example .env.local
pnpm dev
```

App: <http://localhost:3000>

## Theme system (dynamic palettes)

The palette is controlled at runtime by `data-palette="…"` on `<html>` (set on first paint by an inline script, mutated reactively by `usePaletteStore`). `class="dark"` (managed by `next-themes`) toggles light/dark mode.

All 5 palettes × 2 modes are emitted as CSS variables by `ThemeVars` in `<head>`. Tailwind utilities (`bg-primary`, `text-foreground`, …) read those vars via `@theme inline` in `globals.css`, so the entire UI re-tints with a single attribute change — no JS re-render, no flash.

Available palettes:

| ID | Vibe |
|----|------|
| `sunset-horizon` | Golden hour · romantic destinations |
| `forest-trail` (default) | Mountain hikes · eco · slow travel |
| `tropical-paradise` | Beach holidays · islands · vacation mode |
| `mountain-mist` | Premium alpine · editorial · curated stays |
| `desert-dunes` | Heritage routes · Rajasthan · road trips |

To add a palette: append a `PaletteDef` entry to `PALETTES` in `src/theme/palettes.ts`. The shape is enforced by TypeScript.

## Project structure

See [`STRUCTURE.md`](STRUCTURE.md).

## Same-origin API proxy

The browser hits `/api/v1/*` on this Next.js host. The route handler in `src/app/api/v1/[[...path]]/route.ts` proxies to the backend at `API_URL`. This avoids CORS and keeps auth cookies same-origin.

## Deployment

- **AWS Amplify Hosting (SSR / WEB_COMPUTE)** — see `amplify.yml` and [`DEPLOYMENT.md`](DEPLOYMENT.md).
- **Docker / PM2** — see `Dockerfile`, `ecosystem.config.cjs`, and `Makefile`.

## Scripts

| Script | Purpose |
|--------|---------|
| `pnpm dev` | Next.js dev server |
| `pnpm build` | Production build |
| `pnpm start` | Run the built server |
| `pnpm lint` | ESLint |
| `pnpm analyze` | `ANALYZE=true next build` (bundle analyzer) |
