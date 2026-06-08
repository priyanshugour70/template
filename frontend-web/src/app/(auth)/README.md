# app/(auth)/

Public auth routes — `/auth/login`, `/auth/forgot-password`, `/auth/reset-password`, `/auth/accept-invite/[token]`, `/auth/oauth/callback`.

The route group is exempted from the auth gate in `middleware.ts` via the `publicPaths` list. Add new public auth URLs to both places (this folder + `publicPaths`).

A shared layout (`(auth)/layout.tsx`) can host the brand panel + form container.
