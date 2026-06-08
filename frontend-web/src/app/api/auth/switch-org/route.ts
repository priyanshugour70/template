import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { persistSession, type BackendSession } from "@/lib/auth/server-session";

interface SwitchBody {
  organizationId?: string;
}

export async function POST(req: Request) {
  let body: SwitchBody;
  try {
    body = (await req.json()) as SwitchBody;
  } catch {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "invalid json body" } },
      { status: 400 },
    );
  }

  if (!body.organizationId) {
    return NextResponse.json(
      { success: false, error: { code: "INVALID_INPUT", message: "organizationId is required" } },
      { status: 400 },
    );
  }

  const upstream = await api.post<BackendSession>(
    "/auth/switch-org",
    { organizationId: body.organizationId },
  );

  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 400 });
  }

  await persistSession(upstream.data);
  const { accessToken: _a, refreshToken: _r, ...publicSession } = upstream.data;
  return NextResponse.json({ ...upstream, data: publicSession });
}
