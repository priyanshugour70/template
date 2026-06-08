"use client";

import { useEffect, type ReactNode } from "react";

import { useSessionStore } from "@/stores/session/session.store";

/**
 * Hydrates the session-store from cookies on first client mount. This runs
 * once per page load; subsequent updates flow through actions in the auth
 * service (login / switch-org / logout).
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const hydrate = useSessionStore((s) => s.hydrate);
  useEffect(() => {
    hydrate();
  }, [hydrate]);
  return <>{children}</>;
}
