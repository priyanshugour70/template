/**
 * Next.js 16 Proxy (file convention renamed from middleware). Two
 * responsibilities:
 *
 *   1. Subdomain routing
 *      - Apex   (lssgoo.com)              → marketing, login, signup, handoff.
 *                                           /dashboard/* is bounced to login.
 *      - Tenant (acme.lssgoo.com)         → app shell. Auth-gated as before.
 *      - Reserved (api., www., …)         → www is normalised to apex; others
 *                                           are passed through (most never
 *                                           hit Next.js anyway because they
 *                                           point elsewhere in DNS).
 *
 *   2. Auth gating
 *      - Mirrors the previous proxy behaviour for tenant subdomains.
 *      - Defence in depth: if a tenant subdomain holds a session cookie whose
 *        tenant slug doesn't match the URL host, clear cookies and force
 *        re-login. (Prevents stale-cookie tenant confusion when a user
 *        renames their tenant or hand-edits a cookie.)
 *
 * The resolved subdomain is forwarded to downstream code as the
 * `x-tenant-slug` request header so server components can read it via
 * next/headers without re-parsing the host.
 */

import { NextResponse, type NextRequest } from "next/server";

import { COOKIE_ACCESS, COOKIE_SESSION_TENANT } from "@/lib/cookies/names";
import { extractSubdomain, getApexDomain } from "@/lib/tenant/subdomain";

const APEX_PUBLIC_PATHS = [
  "/",
  "/about",
  "/pricing",
  "/contact",
  "/auth/login",
  "/auth/signup",
  "/auth/forgot-password",
  "/auth/reset-password",
];

const TENANT_PUBLIC_PATHS = [
  "/auth/login",
  "/auth/signup",
  "/auth/forgot-password",
  "/auth/reset-password",
  "/auth/accept-invite",
  "/auth/handoff",
];

// Paths an authed user should never see on a tenant subdomain — login/signup
// only make sense before sign-in.
const AUTHED_BOUNCE = ["/auth/login", "/auth/signup", "/auth/forgot-password"];

function startsWithAny(path: string, list: string[]): boolean {
  return list.some((p) => path === p || path.startsWith(p + "/"));
}

function safeJSON<T>(raw: string | undefined): T | null {
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const host = request.headers.get("host") ?? "";
  const apex = getApexDomain();
  const subdomain = extractSubdomain(host, apex);

  // ── Apex domain (lssgoo.com, www.lssgoo.com) ──────────────────────────────
  if (!subdomain) {
    // /auth has no page — send it somewhere useful.
    if (pathname === "/auth" || pathname === "/auth/") {
      return NextResponse.redirect(new URL("/auth/login", request.url));
    }
    // On the apex, /dashboard/* doesn't exist — tenants live on subdomains.
    if (pathname.startsWith("/dashboard")) {
      return NextResponse.redirect(new URL("/auth/login", request.url));
    }
    // Allow everything else; layouts will handle non-auth pages.
    if (!startsWithAny(pathname, APEX_PUBLIC_PATHS)) {
      // Anything unknown on the apex falls through — Next.js will 404.
      return NextResponse.next();
    }
    return NextResponse.next();
  }

  // ── Tenant subdomain (acme.lssgoo.com) ────────────────────────────────────
  const accessToken = request.cookies.get(COOKIE_ACCESS)?.value;
  const sessionTenantRaw = request.cookies.get(COOKIE_SESSION_TENANT)?.value;
  const sessionTenant = safeJSON<{ slug?: string }>(sessionTenantRaw);

  // Defence in depth: if there's a session cookie for a different tenant on
  // this subdomain, the user is in a confused state — clear and re-login.
  if (
    accessToken &&
    sessionTenant?.slug &&
    sessionTenant.slug.toLowerCase() !== subdomain.toLowerCase()
  ) {
    const res = NextResponse.redirect(new URL("/auth/login", request.url));
    res.cookies.delete(COOKIE_ACCESS);
    res.cookies.delete(COOKIE_SESSION_TENANT);
    return res;
  }

  // /auth has no page on tenant subdomains either.
  if (pathname === "/auth" || pathname === "/auth/") {
    return NextResponse.redirect(
      new URL(accessToken ? "/dashboard" : "/auth/login", request.url),
    );
  }

  const isTenantPublic = startsWithAny(pathname, TENANT_PUBLIC_PATHS);
  if (!isTenantPublic && !accessToken) {
    const loginUrl = new URL("/auth/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }
  if (accessToken && AUTHED_BOUNCE.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  // Forward the resolved tenant slug to downstream code via request headers.
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-tenant-slug", subdomain);
  return NextResponse.next({ request: { headers: requestHeaders } });
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|api).*)"],
};
