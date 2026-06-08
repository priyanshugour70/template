/**
 * Auth service — talks to /api/auth/* (Next.js Route Handlers) which handle
 * HttpOnly cookie writes for tokens and call the backend on our behalf.
 *
 * Client-side code (login form, logout button, accept-invite flow) imports
 * from here. Tokens are never returned to the browser.
 */

import { api } from "@/lib/client";
import type {
  AcceptInviteRequest,
  ChangePasswordRequest,
  DiscoverResponse,
  ForgotPasswordRequest,
  LoginRequest,
  ResetPasswordRequest,
  SessionResponse,
  SwitchOrgRequest,
  UserSession,
} from "@/types/auth";
import type { UserProfile } from "@/types/user";

const AUTH_BASE = "/api/auth";

export const authService = {
  /** Step 1 of the login flow: returns tenants the email belongs to. */
  discover: (email: string) =>
    api.post<DiscoverResponse>("/discover", { email }, { basePath: AUTH_BASE, skipAuth: true }),

  /** Step 2: posts credentials, sets HttpOnly cookies, returns session display info. */
  login: (req: LoginRequest) =>
    api.post<SessionResponse>("/login", req, { basePath: AUTH_BASE, skipAuth: true }),

  /** Refresh access+refresh tokens (no body — uses HttpOnly cookie). */
  refresh: () => api.post<SessionResponse>("/refresh", undefined, { basePath: AUTH_BASE, skipAuth: true }),

  /** Revoke server-side + drop all auth cookies. */
  logout: () => api.post<{ success: true }>("/logout", undefined, { basePath: AUTH_BASE }),

  /** Pick a different active organization in the current tenant. */
  switchOrg: (req: SwitchOrgRequest) =>
    api.post<SessionResponse>("/switch-org", req, { basePath: AUTH_BASE }),

  /** Current user. Use this to hydrate AuthProvider on mount. */
  me: () => api.get<UserProfile>("/me", { basePath: AUTH_BASE }),

  /** Onboarding from an emailed invite token. */
  acceptInvite: (req: AcceptInviteRequest) =>
    api.post<SessionResponse>("/accept-invite", req, { basePath: AUTH_BASE, skipAuth: true }),

  forgotPassword: (req: ForgotPasswordRequest) =>
    api.post<{ success: true }>("/forgot-password", req, { basePath: AUTH_BASE, skipAuth: true }),

  resetPassword: (req: ResetPasswordRequest) =>
    api.post<{ success: true }>("/reset-password", req, { basePath: AUTH_BASE, skipAuth: true }),

  changePassword: (req: ChangePasswordRequest) =>
    api.post<{ success: true }>("/change-password", req, { basePath: AUTH_BASE }),

  // Sessions list/revoke go through the backend directly via the v1 proxy.
  listSessions: () => api.get<UserSession[]>("/auth/sessions"),
  revokeSession: (jti: string) => api.delete<{ success: true }>(`/auth/sessions/${jti}`),
};
