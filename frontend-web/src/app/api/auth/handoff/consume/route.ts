import { NextResponse } from "next/server";
import { headers } from "next/headers";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";
import { extractSubdomain, getApexDomain } from "@/lib/tenant/subdomain";

interface Body {
  token?: string;
}

/**
 * Consumes a one-time SSO handoff token (issued by the apex login flow) and
 * persists a real session on this subdomain. Runs ONLY on a tenant subdomain
 * — handoff makes no sense on the apex.
 */
export async function POST(req: Request) {
  let body: Body;
  try {
    body = (await req.json()) as Body;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }
  if (!body.token) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "token is required" } },
      { status: 400 },
    );
  }

  // Refuse to run on the apex. Cookies set here would be apex-scoped and the
  // tenant subdomain would still have no session.
  const h = await headers();
  const subdomain = extractSubdomain(h.get("host"), getApexDomain());
  if (!subdomain) {
    return NextResponse.json(
      {
        success: false,
        error: { code: "WRONG_HOST", message: "handoff must be consumed on a tenant subdomain" },
      },
      { status: 400 },
    );
  }

  const upstream = await api.post<BackendSession>(
    "/auth/handoff/consume",
    { token: body.token },
    { skipAuth: true },
  );
  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 401 });
  }

  // Sanity check: the session we received MUST be for this subdomain's tenant.
  // If somehow a cross-tenant token was consumed, refuse to set cookies.
  if (upstream.data.tenant.slug.toLowerCase() !== subdomain.toLowerCase()) {
    return NextResponse.json(
      {
        success: false,
        error: { code: "TENANT_MISMATCH", message: "handoff is for a different tenant" },
      },
      { status: 400 },
    );
  }

  await persistSession(upstream.data);
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
