import { NextResponse } from "next/server";

import { api } from "@/lib/client/server";
import { writeSessionUser } from "@/lib/cookies/server";
import type { SessionUser } from "@/lib/cookies/types";

/**
 * Returns the current user and refreshes the session-display cookie so the
 * sidebar/header reflects the latest profile (avatar updates, name edits)
 * and the onboarding gate sees the right isOwner flag (set by the backend
 * from the active membership).
 */
export async function GET() {
  const upstream = await api.get<SessionUser & { email: string }>("/users/me");
  if (!upstream.success || !upstream.data) {
    return NextResponse.json(upstream, { status: 401 });
  }
  await writeSessionUser({
    id: upstream.data.id,
    email: upstream.data.email,
    displayName: upstream.data.displayName,
    firstName: upstream.data.firstName,
    lastName: upstream.data.lastName,
    avatarUrl: upstream.data.avatarUrl,
    isSuperAdmin: upstream.data.isSuperAdmin,
    isOwner: upstream.data.isOwner,
  });
  return NextResponse.json(upstream);
}
