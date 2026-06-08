/**
 * Browser helper that invokes the logout route handler. Splits the call out
 * of the cookies module to keep the dependency graph shallow.
 */

import { clearSessionDisplayCookies } from "@/lib/cookies/client";

export async function authClientLogout(): Promise<void> {
  try {
    await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
  } catch {
    // Even if the network call fails, drop the display cookies so the UI
    // forgets the session.
  }
  clearSessionDisplayCookies();
}
