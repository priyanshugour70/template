import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";

interface ResetBody {
  token?: string;
  newPassword?: string;
}

export async function POST(req: Request) {
  let body: ResetBody;
  try {
    body = (await req.json()) as ResetBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }
  if (!body.token || !body.newPassword) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "token and newPassword are required" } },
      { status: 400 },
    );
  }
  const upstream = await api.post<unknown>("/auth/reset-password", body, { skipAuth: true });
  return NextResponse.json(upstream);
}
