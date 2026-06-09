/**
 * Next.js 16 Proxy (renamed from middleware). Runs before every matched
 * request and gates protected routes against the HttpOnly access cookie.
 *
 * Note: this is an *optimistic* check — proxy only validates that a cookie
 * exists; the real authorisation happens at the backend via Bearer header on
 * every API call.
 */

import { NextResponse, type NextRequest } from "next/server";

import { COOKIE_ACCESS } from "@/lib/cookies/names";

const publicPaths = [
  "/auth/login",
  "/auth/signup",
  "/auth/forgot-password",
  "/auth/reset-password",
  "/auth/accept-invite",
  "/auth/oauth/callback",
];

// Paths an authenticated user should never see — they exist only to sign in
// or create a new account. Visiting them while logged in bounces to dashboard.
// (reset-password and accept-invite are token-driven and allowed while authed,
// so the user can change their password or accept a workspace invite.)
const authedUserBounceFrom = ["/auth/login", "/auth/signup", "/auth/forgot-password"];

const publicRoots = ["/", "/about", "/pricing", "/contact"];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isPublicPath = publicPaths.some((p) => pathname.startsWith(p));
  const isPublicRoot = publicRoots.includes(pathname);
  const accessToken = request.cookies.get(COOKIE_ACCESS)?.value;

  // Bare /auth has no page — send it somewhere useful instead of 404'ing.
  if (pathname === "/auth" || pathname === "/auth/") {
    return NextResponse.redirect(
      new URL(accessToken ? "/dashboard" : "/auth/login", request.url),
    );
  }

  if (!isPublicPath && !isPublicRoot && !accessToken) {
    const loginUrl = new URL("/auth/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Authed users have no business on sign-in / sign-up / forgot-password.
  if (accessToken && authedUserBounceFrom.some((p) => pathname.startsWith(p))) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|api).*)"],
};
