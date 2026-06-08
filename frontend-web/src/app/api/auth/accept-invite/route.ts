import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";

interface AcceptBody {
  token?: string;
  firstName?: string;
  lastName?: string;
  password?: string;
}

export async function POST(req: Request) {
  let body: AcceptBody;
  try {
    body = (await req.json()) as AcceptBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }

  if (!body.token || !body.firstName || !body.password) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "token, firstName, password required" } },
      { status: 400 },
    );
  }

  const upstream = await api.post<BackendSession>(
    "/auth/accept-invite",
    body,
    { skipAuth: true },
  );

  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 400 });
  }

  await persistSession(upstream.data);
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
