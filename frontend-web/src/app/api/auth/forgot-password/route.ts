import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";

export async function POST(req: Request) {
  let body: { email?: string };
  try {
    body = (await req.json()) as { email?: string };
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }
  if (!body.email) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "email is required" } },
      { status: 400 },
    );
  }
  // Backend always 202s — never reveals whether the email exists.
  const upstream = await api.post<unknown>("/auth/forgot-password", { email: body.email }, { skipAuth: true });
  return NextResponse.json({ success: true, message: upstream.message });
}
