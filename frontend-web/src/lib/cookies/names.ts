/**
 * Cookie names used across the app. Centralized so server, client, and proxy
 * never drift on key naming.
 */

// Auth — HttpOnly. Never readable by client JS.
export const COOKIE_ACCESS = "app_access_token";
export const COOKIE_REFRESH = "app_refresh_token";

// Session display — non-HttpOnly. Readable by client so the dashboard sidebar
// can paint tenant brand on first render without an API roundtrip. Contents
// are NOT secrets — names, logos, IDs.
export const COOKIE_SESSION_USER = "app_session_user";
export const COOKIE_SESSION_TENANT = "app_session_tenant";
export const COOKIE_SESSION_ORG = "app_session_org";
export const COOKIE_SESSION_ORGS = "app_session_orgs";
export const COOKIE_SESSION_PERMS = "app_session_perms";
export const COOKIE_SESSION_ROLES = "app_session_roles";

// UI preferences — non-HttpOnly. Persisted across logins.
export const COOKIE_PALETTE = "app_palette";
export const COOKIE_THEME = "app_theme";
export const COOKIE_LOCALE = "app_locale";

// All cookies the app owns. Used by logout to clear everything.
export const ALL_APP_COOKIES = [
  COOKIE_ACCESS,
  COOKIE_REFRESH,
  COOKIE_SESSION_USER,
  COOKIE_SESSION_TENANT,
  COOKIE_SESSION_ORG,
  COOKIE_SESSION_ORGS,
  COOKIE_SESSION_PERMS,
  COOKIE_SESSION_ROLES,
] as const;

// Session-display cookies that should be cleared on logout but preserve UI prefs.
export const SESSION_DISPLAY_COOKIES = [
  COOKIE_SESSION_USER,
  COOKIE_SESSION_TENANT,
  COOKIE_SESSION_ORG,
  COOKIE_SESSION_ORGS,
  COOKIE_SESSION_PERMS,
  COOKIE_SESSION_ROLES,
] as const;
