/**
 * Server-only cookie helpers. Uses Next.js 16 async `cookies()` from
 * next/headers. Tokens are stored HttpOnly + Secure (in production) so client
 * JS can never read them. Session-display cookies are non-HttpOnly so the
 * client can render tenant brand before /me responds.
 *
 * Setting cookies only works inside a Route Handler or Server Function — a
 * pure RSC cannot mutate cookies.
 */

import "server-only";

import { cookies } from "next/headers";

import {
  ALL_APP_COOKIES,
  COOKIE_ACCESS,
  COOKIE_REFRESH,
  COOKIE_SESSION_ORG,
  COOKIE_SESSION_ORGS,
  COOKIE_SESSION_PERMS,
  COOKIE_SESSION_ROLES,
  COOKIE_SESSION_TENANT,
  COOKIE_SESSION_USER,
} from "./names";
import type {
  SessionOrganization,
  SessionSnapshot,
  SessionTenant,
  SessionUser,
  Tokens,
} from "./types";

const DEFAULT_DISPLAY_MAX_AGE = 60 * 60 * 24 * 30; // 30 days
const DEFAULT_ACCESS_MAX_AGE = 60 * 60 * 24 * 30; // 30 days (auto-rotated)
const DEFAULT_REFRESH_MAX_AGE = 60 * 60 * 24 * 30; // 30 days

interface SecureCookieOptions {
  httpOnly: boolean;
  secure: boolean;
  sameSite: "lax" | "strict";
  path: string;
  maxAge: number;
}

function isProd(): boolean {
  return process.env.NODE_ENV === "production";
}

function tokenOpts(maxAge: number): SecureCookieOptions {
  return {
    httpOnly: true,
    secure: isProd(),
    sameSite: "lax",
    path: "/",
    maxAge,
  };
}

function displayOpts(maxAge = DEFAULT_DISPLAY_MAX_AGE): SecureCookieOptions {
  return {
    httpOnly: false,
    secure: isProd(),
    sameSite: "lax",
    path: "/",
    maxAge,
  };
}

// ── read ───────────────────────────────────────────────────────────────────

export async function readTokens(): Promise<Tokens> {
  const c = await cookies();
  return {
    accessToken: c.get(COOKIE_ACCESS)?.value,
    refreshToken: c.get(COOKIE_REFRESH)?.value,
  };
}

export async function readAccessToken(): Promise<string | undefined> {
  return (await cookies()).get(COOKIE_ACCESS)?.value;
}

export async function readRefreshToken(): Promise<string | undefined> {
  return (await cookies()).get(COOKIE_REFRESH)?.value;
}

function safeJSON<T>(raw: string | undefined): T | null {
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

export async function readSessionSnapshotServer(): Promise<SessionSnapshot> {
  const c = await cookies();
  return {
    user: safeJSON<SessionUser>(c.get(COOKIE_SESSION_USER)?.value),
    tenant: safeJSON<SessionTenant>(c.get(COOKIE_SESSION_TENANT)?.value),
    activeOrganization: safeJSON<SessionOrganization>(c.get(COOKIE_SESSION_ORG)?.value),
    organizations: safeJSON<SessionOrganization[]>(c.get(COOKIE_SESSION_ORGS)?.value) ?? [],
    permissions: safeJSON<string[]>(c.get(COOKIE_SESSION_PERMS)?.value) ?? [],
    roles: safeJSON<string[]>(c.get(COOKIE_SESSION_ROLES)?.value) ?? [],
  };
}

// ── write ──────────────────────────────────────────────────────────────────

export async function writeTokens(
  tokens: Required<Tokens>,
  accessMaxAge = DEFAULT_ACCESS_MAX_AGE,
  refreshMaxAge = DEFAULT_REFRESH_MAX_AGE,
): Promise<void> {
  const c = await cookies();
  c.set(COOKIE_ACCESS, tokens.accessToken, tokenOpts(accessMaxAge));
  c.set(COOKIE_REFRESH, tokens.refreshToken, tokenOpts(refreshMaxAge));
}

export async function writeAccessToken(value: string, maxAge = DEFAULT_ACCESS_MAX_AGE): Promise<void> {
  const c = await cookies();
  c.set(COOKIE_ACCESS, value, tokenOpts(maxAge));
}

export async function writeSessionUser(user: SessionUser | null): Promise<void> {
  const c = await cookies();
  if (user === null) {
    c.delete(COOKIE_SESSION_USER);
    return;
  }
  c.set(COOKIE_SESSION_USER, JSON.stringify(user), displayOpts());
}

export async function writeSessionTenant(tenant: SessionTenant | null): Promise<void> {
  const c = await cookies();
  if (tenant === null) {
    c.delete(COOKIE_SESSION_TENANT);
    return;
  }
  c.set(COOKIE_SESSION_TENANT, JSON.stringify(tenant), displayOpts());
}

export async function writeActiveOrganization(org: SessionOrganization | null): Promise<void> {
  const c = await cookies();
  if (org === null) {
    c.delete(COOKIE_SESSION_ORG);
    return;
  }
  c.set(COOKIE_SESSION_ORG, JSON.stringify(org), displayOpts());
}

export async function writeOrganizations(orgs: SessionOrganization[]): Promise<void> {
  const c = await cookies();
  c.set(COOKIE_SESSION_ORGS, JSON.stringify(orgs), displayOpts());
}

export async function writePermissions(perms: string[]): Promise<void> {
  const c = await cookies();
  c.set(COOKIE_SESSION_PERMS, JSON.stringify(perms), displayOpts());
}

export async function writeRoles(roles: string[]): Promise<void> {
  const c = await cookies();
  c.set(COOKIE_SESSION_ROLES, JSON.stringify(roles), displayOpts());
}

// ── delete ─────────────────────────────────────────────────────────────────

export async function clearAllAuthCookies(): Promise<void> {
  const c = await cookies();
  for (const name of ALL_APP_COOKIES) {
    c.delete(name);
  }
}
