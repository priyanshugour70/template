import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";

interface ChangeBody {
  currentPassword?: string;
  newPassword?: string;
}

export async function POST(req: Request) {
  let body: ChangeBody;
  try {
    body = (await req.json()) as ChangeBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }
  if (!body.currentPassword || !body.newPassword) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "currentPassword and newPassword are required" } },
      { status: 400 },
    );
  }
  const upstream = await api.post<unknown>("/auth/change-password", body);
  return NextResponse.json(upstream);
}
