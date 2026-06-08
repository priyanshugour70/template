import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";

interface LoginBody {
  email?: string;
  password?: string;
  tenantId?: string;
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

  const upstream = await api.post<BackendSession>(
    "/auth/login",
    { email, password, tenantId },
    { skipAuth: true },
  );

  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 401 });
  }

  await persistSession(upstream.data);

  // Strip the tokens before returning to the client — they live in HttpOnly cookies now.
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
