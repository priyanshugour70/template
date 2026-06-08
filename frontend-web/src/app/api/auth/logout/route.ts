import { NextResponse } from "next/server";

import { destroySession } from "@/lib/auth/server-session";
import { readRefreshToken } from "@/lib/cookies/server";

export async function POST() {
  const refreshToken = await readRefreshToken();
  await destroySession(refreshToken);
  return NextResponse.json({ success: true });
}
