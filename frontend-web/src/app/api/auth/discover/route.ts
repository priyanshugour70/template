import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import type { SessionTenant } from "@/lib/cookies/types";

interface DiscoverBody {
  email?: string;
}

export async function POST(req: Request) {
  let body: DiscoverBody;
  try {
    body = (await req.json()) as DiscoverBody;
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

  const upstream = await api.post<{ tenants: SessionTenant[] }>(
    "/auth/discover",
    { email: body.email },
    { skipAuth: true },
  );

  return NextResponse.json(upstream);
}
