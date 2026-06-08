/**
 * Public cookies API for browser/client code. Server contexts (RSC / route
 * handlers / server actions) should import from "./server" instead — that
 * variant can read AND set HttpOnly token cookies via next/headers.
 *
 * Re-exported here for backward compatibility with existing imports.
 */

export * from "./names";
export * from "./types";
export {
  readSessionUser as getSessionUser,
  readSessionTenant as getSessionTenant,
  readActiveOrganization as getActiveOrganization,
  readOrganizations as getOrganizations,
  readPermissions as getPermissions,
  readRoles as getRoles,
  readSessionSnapshot as getSessionSnapshot,
  readPalette,
  writePalette,
  readTheme,
  writeTheme,
  clearSessionDisplayCookies,
} from "./client";

/**
 * Legacy stubs kept so existing imports of getTokens/clearTokens don't break
 * during the migration. Tokens are now HttpOnly and cannot be read from
 * client code — the new flow is:
 *   - Login posts to /api/auth/login (sets HttpOnly cookies server-side).
 *   - All API calls go through /api/v1/* proxy which attaches the token.
 *   - Logout hits /api/auth/logout to clear cookies + revoke server-side.
 */
import { authClientLogout } from "@/lib/auth/client-logout";

export function getTokens(): { accessToken?: string; refreshToken?: string } {
  // Tokens are now HttpOnly — client code never sees them. This shape is kept
  // to avoid breaking call sites mid-migration; the truthiness check that
  // some code does on `accessToken` will always be undefined here. Callers
  // should switch to the session-display cookies (readSessionUser etc.)
  // to determine "logged in" status from the client.
  return {};
}

export async function clearTokens(): Promise<void> {
  await authClientLogout();
}

export async function setTokens(): Promise<void> {
  // No-op stub. Tokens may only be written by /api/auth/login. This function
  // existed in the pre-HttpOnly template; it is intentionally inert now.
}
