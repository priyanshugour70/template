import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";
import { readRefreshToken } from "@/lib/cookies/server";

export async function POST() {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    return NextResponse.json(
      { success: false, error: { code: "UNAUTHORIZED", message: "no refresh token" } },
      { status: 401 },
    );
  }

  const upstream = await api.post<BackendSession>(
    "/auth/refresh",
    { refreshToken },
    { skipAuth: true },
  );

  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 401 });
  }

  await persistSession(upstream.data);
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
