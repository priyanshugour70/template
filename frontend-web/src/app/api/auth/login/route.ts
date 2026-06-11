import { NextResponse } from "next/server";
import { headers } from "next/headers";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";
import {
  buildTenantUrl,
  extractSubdomain,
  getApexDomain,
} from "@/lib/tenant/subdomain";

interface LoginBody {
  email?: string;
  password?: string;
  tenantId?: string;
}

interface HandoffIssueResponse {
  token: string;
  expiresAt: string;
  tenant: { id: string; slug: string; name: string; logoUrl?: string; primaryColor?: string };
}

export async function POST(req: Request) {
  let body: LoginBody;
  try {
    body = (await req.json()) as LoginBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }

  const { email, password, tenantId } = body;
  if (!email || !password || !tenantId) {
    return NextResponse.json(
      {
        success: false,
        error: { code: "INVALID_INPUT", message: "email, password and tenantId are required" },
      },
      { status: 400 },
    );
  }

  // 1) Authenticate with the backend.
  const upstream = await api.post<BackendSession>(
    "/auth/login",
    { email, password, tenantId },
    { skipAuth: true },
  );

  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 401 });
  }

  const session = upstream.data;
  const h = await headers();
  const apex = getApexDomain();
  const hostSubdomain = extractSubdomain(h.get("host"), apex);

  // 2a) Subdomain match — cookies are already per-subdomain, just persist.
  // This branch handles direct logins at acme.lssgoo.com/auth/login.
  if (hostSubdomain && hostSubdomain.toLowerCase() === session.tenant.slug.toLowerCase()) {
    await persistSession(session);
    const { accessToken: _a, refreshToken: _r, ...publicSession } = session;
    return NextResponse.json({ ...upstream, data: publicSession });
  }

  // 2b) Apex (or cross-tenant) login — mint a one-time handoff token using
  // the access token we just received, then send the client to the tenant
  // subdomain's /auth/handoff page. Tokens we just got are NOT persisted on
  // the apex; only the tenant subdomain holds the real session.
  const handoff = await fetch(`${(process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "")}/api/v1/auth/handoff/issue`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${session.accessToken}`,
    },
    cache: "no-store",
  });
  if (!handoff.ok) {
    return NextResponse.json(
      { success: false, error: { code: "HANDOFF_FAILED", message: "could not issue handoff token" } },
      { status: 500 },
    );
  }
  const handoffBody = (await handoff.json()) as { success: boolean; data?: HandoffIssueResponse };
  if (!handoffBody.success || !handoffBody.data) {
    return NextResponse.json(handoffBody, { status: 500 });
  }

  // Best-effort: revoke the apex's just-issued refresh token. We don't want a
  // long-lived session sitting on the apex with no UI to surface it. The
  // handoff token below carries everything the tenant needs.
  if (session.refreshToken) {
    fetch(`${(process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "")}/api/v1/auth/logout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${session.accessToken}`,
      },
      body: JSON.stringify({ refreshToken: session.refreshToken }),
      cache: "no-store",
    }).catch(() => {});
  }

  const redirectUrl = buildTenantUrl(handoffBody.data.tenant.slug, {
    path: `/auth/handoff?token=${encodeURIComponent(handoffBody.data.token)}`,
  });

  // Return the redirect URL to the client; the login page does
  // window.location.assign(redirect).
  return NextResponse.json({
    success: true,
    data: {
      mode: "handoff",
      tenant: handoffBody.data.tenant,
      redirect: redirectUrl,
    },
  });
}

