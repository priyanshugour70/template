/**
 * Same-origin proxy: browser → /api/v1/* → backend at API_URL.
 *
 * Responsibilities:
 *   1. Forward the request body / headers untouched.
 *   2. Attach the bearer token from the HttpOnly access cookie.
 *   3. On 401 + a present refresh cookie, attempt one rotation and retry.
 *   4. On refresh failure, clear every auth cookie and let the 401 surface.
 *
 * Keeping refresh inside the proxy means client code never sees tokens or
 * deals with race conditions during rotation — concurrent 401s queue up
 * naturally because each proxy invocation is independent.
 */

import {
  clearAllAuthCookies,
  readRefreshToken,
  writeTokens,
} from "@/lib/cookies/server";
import { COOKIE_ACCESS } from "@/lib/cookies/names";
import { cookies } from "next/headers";

const BACKEND = (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "");

type Ctx = { params: Promise<{ path?: string[] }> };

interface RefreshResponse {
  success: boolean;
  data?: {
    accessToken: string;
    refreshToken: string;
    accessTokenExpiresAt?: string;
    refreshTokenExpiresAt?: string;
  };
}

function expiryToMaxAge(iso: string | undefined, fallback: number): number {
  if (!iso) return fallback;
  const ms = new Date(iso).getTime() - Date.now();
  if (!Number.isFinite(ms) || ms <= 0) return fallback;
  return Math.floor(ms / 1000);
}

async function attemptRefresh(): Promise<boolean> {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) return false;
  try {
    const res = await fetch(`${BACKEND}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ refreshToken }),
      cache: "no-store",
    });
    if (!res.ok) return false;
    const body = (await res.json()) as RefreshResponse;
    if (!body.success || !body.data?.accessToken || !body.data?.refreshToken) return false;
    await writeTokens(
      { accessToken: body.data.accessToken, refreshToken: body.data.refreshToken },
      expiryToMaxAge(body.data.accessTokenExpiresAt, 60 * 15),
      expiryToMaxAge(body.data.refreshTokenExpiresAt, 60 * 60 * 24 * 7),
    );
    return true;
  } catch {
    return false;
  }
}

async function forward(req: Request, target: string, accessToken: string | undefined): Promise<Response> {
  const headers = new Headers(req.headers);
  headers.delete("host");
  headers.delete("content-length");
  headers.delete("cookie");
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  const init: RequestInit & { duplex?: "half" } = {
    method: req.method,
    headers,
    redirect: "manual",
    body: ["GET", "HEAD"].includes(req.method) ? undefined : req.body,
    duplex: "half",
    cache: "no-store",
  };

  return fetch(target, init);
}

async function proxy(req: Request, { params }: Ctx) {
  const { path = [] } = await params;
  const url = new URL(req.url);
  const target = `${BACKEND}/api/v1/${path.join("/")}${url.search}`;

  const cookieStore = await cookies();
  let accessToken = cookieStore.get(COOKIE_ACCESS)?.value;

  let upstream = await forward(req, target, accessToken);

  if (upstream.status === 401 && !path.join("/").startsWith("auth/")) {
    // Try one rotation, then retry. Body would have been consumed by the
    // first fetch, so refresh-on-retry is only safe for idempotent methods
    // or bodyless requests. For body-carrying requests, fall through with
    // the original 401 — client will refresh + replay.
    if (["GET", "HEAD"].includes(req.method)) {
      const rotated = await attemptRefresh();
      if (rotated) {
        const fresh = (await cookies()).get(COOKIE_ACCESS)?.value;
        accessToken = fresh;
        upstream = await forward(req, target, accessToken);
      } else {
        await clearAllAuthCookies();
      }
    } else {
      // Best-effort refresh so the next request from the client succeeds.
      const rotated = await attemptRefresh();
      if (!rotated) await clearAllAuthCookies();
    }
  }

  // Strip transfer-encoding / connection headers that Node fetch may set
  // and that browsers don't need.
  const respHeaders = new Headers(upstream.headers);
  respHeaders.delete("transfer-encoding");
  respHeaders.delete("connection");

  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: respHeaders,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
export const OPTIONS = proxy;
