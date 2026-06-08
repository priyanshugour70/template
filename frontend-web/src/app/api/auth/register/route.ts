import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";

interface RegisterBody {
  email?: string;
  password?: string;
  firstName?: string;
  lastName?: string;
  organizationName?: string;
  organizationSlug?: string;
}

export async function POST(req: Request) {
  let body: RegisterBody;
  try {
    body = (await req.json()) as RegisterBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }

  const required: (keyof RegisterBody)[] = [
    "email",
    "password",
    "firstName",
    "organizationName",
    "organizationSlug",
  ];
  for (const k of required) {
    if (!body[k]) {
      return NextResponse.json(
        { success: false, error: { code: "INVALID_INPUT", message: `${k} is required` } },
        { status: 400 },
      );
    }
  }

  const upstream = await api.post<BackendSession>("/auth/register", body, { skipAuth: true });
  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 400 });
  }
  await persistSession(upstream.data);
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
