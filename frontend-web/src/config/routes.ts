/**
 * Canonical route paths. Reference these everywhere instead of hardcoding strings —
 * makes refactoring URLs a single-file change.
 */
export const ROUTES = {
  home: "/",
  login: "/auth/login",
  forgotPassword: "/auth/forgot-password",
  acceptInvite: "/auth/accept-invite",
  dashboard: "/dashboard",
} as const;

export type RouteKey = keyof typeof ROUTES;
