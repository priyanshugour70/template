# frontend-web – Directory Structure

Next.js 16 App Router with route groups, shadcn-style components, and feature-folder organization in `src/{actions,services,hooks,stores,types}`.

## Full folder tree

```
frontend-web/
├── public/                       # Static assets served at /
│   └── logo/                     #   Brand logos & marks
├── src/
│   ├── app/                      # Next.js App Router
│   │   ├── (auth)/               #   Public auth routes (login, accept-invite, …)
│   │   ├── (dashboard)/          #   Authenticated app shell
│   │   ├── (public)/             #   Public marketing/storefront pages
│   │   ├── api/v1/[[...path]]/   #   Same-origin proxy to backend API_URL
│   │   ├── layout.tsx            #   Root layout (mounts providers + ThemeVars)
│   │   ├── page.tsx              #   Root redirect
│   │   ├── globals.css           #   Tailwind + @theme inline tokens
│   │   └── favicon.ico
│   ├── actions/                  # Server Actions, one folder per feature
│   ├── components/
│   │   ├── ui/                   #   shadcn/ui primitives (button, dialog, …)
│   │   ├── layouts/              #   Per-route-group shells (dashboard, public)
│   │   ├── shared/               #   Cross-feature components (PageHeader, EmptyState, …)
│   │   ├── features/             #   Feature-specific components (e.g. <SampleTable/>)
│   │   └── providers/            #   React context providers
│   ├── config/                   # Static config: routes, feature flags, env-readers
│   ├── constants/                # Constants without behaviour (regexes, enums)
│   ├── data/                     # Hard-coded data sets (countries, currencies, …)
│   ├── hooks/                    # React hooks, grouped by feature
│   ├── lib/                      # Pure utility modules (no React) — used by RSC + client
│   │   ├── client/               #   HTTP client + interceptors
│   │   ├── cookies/              #   Server + client cookie helpers (nookies)
│   │   ├── rbac/                 #   Permission checks
│   │   ├── routing/              #   Route maps + helpers
│   │   ├── server/               #   Server-only helpers (must use "server-only")
│   │   └── session/              #   Token + user-session helpers
│   ├── middleware.ts             # Next.js edge middleware (auth gate)
│   ├── providers/                # App-root providers (Auth, Permissions, QueryClient, …)
│   ├── services/                 # Domain HTTP services, grouped by feature
│   ├── stores/                   # Zustand stores, grouped by feature
│   ├── theme/                    # Theme tokens + dynamic palette registry
│   ├── types/                    # TypeScript types, grouped by feature
│   └── utils/                    # Pure helpers (fmt, dates, …)
├── docs/                         # Hand-written architectural notes
├── deployments/docker/           # Dockerfile variants
├── scripts/                      # Local automation scripts
├── .github/workflows/            # CI/CD pipelines
├── .githooks/                    # Local git hooks (`make install-hooks`)
├── .env.example
├── amplify.yml                   # AWS Amplify SSR (WEB_COMPUTE) build spec
├── Dockerfile
├── ecosystem.config.cjs          # PM2 process config for production node container
├── Makefile
├── README.md
├── STRUCTURE.md
├── DEPLOYMENT.md
├── eslint.config.mjs
├── next.config.ts
├── postcss.config.mjs
├── tsconfig.json
├── package.json
└── pnpm-lock.yaml
```

## Why this layout

- **Feature folders inside each technical concern** (`services/products/`, `hooks/products/`, `types/products/`) keeps a feature's vertical slice together while preserving the technical grouping. Cmd-clicking `services/products` shows you every domain service.
- **Route groups** (`(auth)`, `(dashboard)`, `(public)`) co-locate route-specific layouts without affecting URLs.
- **Same-origin API proxy** at `src/app/api/v1/[[...path]]/route.ts` keeps the browser on one origin → no CORS, cookie-based auth is simpler.
- **`lib/` is React-free.** Anything that touches `React` belongs in `hooks/`, `providers/`, or `components/`.

## Adding a feature

1. Define types: `src/types/<feature>/`.
2. Add service: `src/services/<feature>/<feature>.service.ts`.
3. (Optional) Store: `src/stores/<feature>/<feature>.store.ts`.
4. Hooks: `src/hooks/<feature>/use<X>.ts`.
5. Components: `src/components/features/<feature>/...`.
6. Routes: `src/app/(<group>)/<feature>/page.tsx` or a route group of your own.

Every folder ships with a `README.md` describing its purpose and a `.keep` so empty folders survive in git.
