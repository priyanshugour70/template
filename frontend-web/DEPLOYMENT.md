# Frontend Web — Deployment

Two supported targets:

## 1. AWS Amplify Hosting (SSR / WEB_COMPUTE)

Console settings:

- Framework: **Next.js - SSR** (not static hosting).
- App root: repo root or `frontend-web/` if monorepo.
- Branch deploys:
  - `staging` → staging Amplify branch
  - `main` → production Amplify branch

Required environment variables (Amplify → Environment variables):

```text
API_URL=https://staging-api.example.com   (staging branch override)
API_URL=https://api.example.com           (production)
```

The browser calls same-origin `/api/v1/*`; the route handler at `src/app/api/v1/[[...path]]/route.ts` proxies to `API_URL`.

Build spec: [`amplify.yml`](amplify.yml).

> Amplify console env vars are NOT visible to Next.js SSR at runtime unless they are written into `.env.production` before the build. `amplify.yml` does this — keep that step intact.

## 2. Docker + PM2

`Dockerfile` builds a `node:22-alpine` image with `pnpm`, installs deps, runs `next build`, and starts the server via `pm2-runtime` (`ecosystem.config.cjs`).

```bash
make build    # docker build -t frontend-web .
make up       # run on localhost:3000 with env file
make logs
make down
```

The same image runs anywhere PM2/Node 22 is available (ECS, EC2, k8s). For k8s, wrap PM2 in a single container per pod and front it with a Service / Ingress.

## CI/CD

GitHub Actions (`.github/workflows/web-deploy.yml`) triggers Amplify deploys via `aws amplify start-job` on push to `staging` / `main`. The Amplify branch picks up the commit and runs `amplify.yml`.

## Domains & TLS

Provision domains in Amplify Console → Domain management. Amplify provisions ACM certs and routes traffic. For Docker deploys, terminate TLS at your ingress (ALB, Nginx, CloudFront).
