/**
 * Public env vars (NEXT_PUBLIC_*) are inlined at build time. Server-only env
 * (e.g. API_URL) lives in `src/lib/server/env.ts` and must never be imported
 * by client code.
 */
export const publicEnv = {
  appName: process.env.NEXT_PUBLIC_APP_NAME ?? "App",
  defaultPalette: process.env.NEXT_PUBLIC_DEFAULT_PALETTE ?? "forest-trail",
};
