/**
 * Browser-side cookie helpers. Cannot read HttpOnly cookies — those (access /
 * refresh tokens) are intentionally invisible to client JS for XSS resistance.
 *
 * Use this from client components only. For server contexts (RSC / actions /
 * route handlers) import from "./server" instead.
 */

import {
  COOKIE_PALETTE,
  COOKIE_SESSION_ORG,
  COOKIE_SESSION_ORGS,
  COOKIE_SESSION_PERMS,
  COOKIE_SESSION_ROLES,
  COOKIE_SESSION_TENANT,
  COOKIE_SESSION_USER,
  COOKIE_SIDEBAR_COLLAPSED,
  COOKIE_SIDEBAR_SECTIONS,
  COOKIE_THEME,
  SESSION_DISPLAY_COOKIES,
} from "./names";
import type {
  CookieWriteOptions,
  SessionOrganization,
  SessionSnapshot,
  SessionTenant,
  SessionUser,
} from "./types";

const DEFAULT_MAX_AGE = 60 * 60 * 24 * 30; // 30 days for display cookies.

function parseAll(): Record<string, string> {
  if (typeof document === "undefined") return {};
  const out: Record<string, string> = {};
  const pairs = document.cookie ? document.cookie.split(";") : [];
  for (const pair of pairs) {
    const i = pair.indexOf("=");
    if (i < 0) continue;
    const k = pair.slice(0, i).trim();
    const v = pair.slice(i + 1).trim();
    if (k) out[k] = decodeURIComponent(v);
  }
  return out;
}

function readJSON<T>(key: string): T | null {
  const all = parseAll();
  const raw = all[key];
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

function writeRaw(name: string, value: string, opts: CookieWriteOptions = {}) {
  if (typeof document === "undefined") return;
  const maxAge = opts.maxAgeSeconds ?? DEFAULT_MAX_AGE;
  const sameSite = opts.sameSite ?? "lax";
  const secure = window.location.protocol === "https:";
  const parts = [
    `${name}=${encodeURIComponent(value)}`,
    `Path=/`,
    `Max-Age=${maxAge}`,
    `SameSite=${sameSite[0].toUpperCase() + sameSite.slice(1)}`,
  ];
  if (secure) parts.push("Secure");
  document.cookie = parts.join("; ");
}

function deleteRaw(name: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; Path=/; Max-Age=0; SameSite=Lax`;
}

// ── session display ────────────────────────────────────────────────────────

export function readSessionUser(): SessionUser | null {
  return readJSON<SessionUser>(COOKIE_SESSION_USER);
}

export function readSessionTenant(): SessionTenant | null {
  return readJSON<SessionTenant>(COOKIE_SESSION_TENANT);
}

export function readActiveOrganization(): SessionOrganization | null {
  return readJSON<SessionOrganization>(COOKIE_SESSION_ORG);
}

export function readOrganizations(): SessionOrganization[] {
  return readJSON<SessionOrganization[]>(COOKIE_SESSION_ORGS) ?? [];
}

export function readPermissions(): string[] {
  return readJSON<string[]>(COOKIE_SESSION_PERMS) ?? [];
}

export function readRoles(): string[] {
  return readJSON<string[]>(COOKIE_SESSION_ROLES) ?? [];
}

export function readSessionSnapshot(): SessionSnapshot {
  return {
    user: readSessionUser(),
    tenant: readSessionTenant(),
    activeOrganization: readActiveOrganization(),
    organizations: readOrganizations(),
    permissions: readPermissions(),
    roles: readRoles(),
  };
}

// ── UI preferences ─────────────────────────────────────────────────────────

export function readPalette(): string | null {
  return parseAll()[COOKIE_PALETTE] ?? null;
}

export function writePalette(value: string) {
  writeRaw(COOKIE_PALETTE, value);
}

export function readTheme(): string | null {
  return parseAll()[COOKIE_THEME] ?? null;
}

export function writeTheme(value: string) {
  writeRaw(COOKIE_THEME, value);
}

export function readSidebarCollapsed(): boolean {
  return parseAll()[COOKIE_SIDEBAR_COLLAPSED] === "1";
}

export function writeSidebarCollapsed(collapsed: boolean) {
  writeRaw(COOKIE_SIDEBAR_COLLAPSED, collapsed ? "1" : "0");
}

export function readSidebarSections(): Record<string, boolean> {
  const raw = parseAll()[COOKIE_SIDEBAR_SECTIONS];
  if (!raw) return {};
  try {
    return JSON.parse(raw) as Record<string, boolean>;
  } catch {
    return {};
  }
}

export function writeSidebarSections(state: Record<string, boolean>) {
  writeRaw(COOKIE_SIDEBAR_SECTIONS, JSON.stringify(state));
}

// ── cleanup ────────────────────────────────────────────────────────────────

/**
 * Clears every session-display cookie from the client. Token cookies are
 * HttpOnly and can only be cleared by a server response, so logout must hit
 * /api/auth/logout to fully revoke the session.
 */
export function clearSessionDisplayCookies() {
  for (const name of SESSION_DISPLAY_COOKIES) {
    deleteRaw(name);
  }
}
