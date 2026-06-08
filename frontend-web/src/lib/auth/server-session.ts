/**
 * Server-side helpers for persisting and clearing the session across cookies
 * + invoking backend auth endpoints from a Route Handler.
 *
 * Centralised here so every /api/auth/* handler writes cookies the same way.
 */

import "server-only";

import { api } from "@/lib/client/server";
import {
  clearAllAuthCookies,
  writeActiveOrganization,
  writeOrganizations,
  writePermissions,
  writeRoles,
  writeSessionTenant,
  writeSessionUser,
  writeTokens,
} from "@/lib/cookies/server";
import type {
  SessionOrganization,
  SessionTenant,
  SessionUser,
} from "@/lib/cookies/types";

/** Shape of the backend LoginResponse — kept narrow on purpose. */
export interface BackendSession {
  accessToken: string;
  refreshToken: string;
  accessTokenExpiresAt: string;
  refreshTokenExpiresAt: string;
  tokenType?: string;
  user: SessionUser;
  tenant: SessionTenant;
  activeOrganization?: SessionOrganization | null;
  organizations?: SessionOrganization[];
}

function expiryToMaxAge(iso: string, fallback: number): number {
  const ms = new Date(iso).getTime() - Date.now();
  if (!Number.isFinite(ms) || ms <= 0) return fallback;
  return Math.floor(ms / 1000);
}

/**
 * Persist a fresh session received from the backend. Tokens go to HttpOnly
 * cookies; display info goes to non-HttpOnly cookies so the client can paint
 * tenant brand before /me responds.
 *
 * Permissions/roles are NOT in the backend login response yet — the calling
 * handler may fetch them and call writePermissions/writeRoles separately.
 */
export async function persistSession(session: BackendSession): Promise<void> {
  const accessMax = expiryToMaxAge(session.accessTokenExpiresAt, 60 * 15);
  const refreshMax = expiryToMaxAge(session.refreshTokenExpiresAt, 60 * 60 * 24 * 7);

  await writeTokens(
    { accessToken: session.accessToken, refreshToken: session.refreshToken },
    accessMax,
    refreshMax,
  );
  await writeSessionUser(session.user);
  await writeSessionTenant(session.tenant);
  await writeActiveOrganization(session.activeOrganization ?? null);
  await writeOrganizations(session.organizations ?? []);
}

/** Logs out: revoke server-side then drop every cookie we own. */
export async function destroySession(refreshToken?: string): Promise<void> {
  if (refreshToken) {
    // Best-effort. Backend logout is rate-limited and tolerant of missing tokens.
    await api.post("/auth/logout", { refreshToken }, { skipAuth: true });
  }
  await clearAllAuthCookies();
}

/**
 * Hydrate permissions + roles for the active org and persist them as display
 * cookies. The backend currently does not have a single endpoint that returns
 * the resolved (user, org) permission set, so this is a stub that the caller
 * can replace by calling the appropriate endpoint(s) when they're added.
 */
export async function persistPermissions(permissions: string[], roles: string[]): Promise<void> {
  await writePermissions(permissions);
  await writeRoles(roles);
}
