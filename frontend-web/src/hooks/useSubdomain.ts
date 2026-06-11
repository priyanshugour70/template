"use client";

import { useEffect, useState } from "react";

import { extractSubdomain, getApexDomain } from "@/lib/tenant/subdomain";

/**
 * Reads the current subdomain from the browser host. Returns null on the apex
 * or on SSR (no window). Client components that need to know "am I on a
 * tenant page" can call this — server components should use
 * getServerSubdomain() from @/lib/tenant/server.
 *
 * The state is intentionally hydrated in a `useEffect` so the FIRST client
 * render returns the same value as SSR (null). Without this guard, any UI
 * that branches on the subdomain would trigger a hydration mismatch:
 *   server → null → renders apex copy
 *   client → "acme" → renders tenant copy
 * React would then complain and discard the SSR tree. By updating after
 * mount we get one extra re-render (cheap) and zero hydration warnings.
 */
export function useSubdomain(): string | null {
  const [sub, setSub] = useState<string | null>(null);
  useEffect(() => {
    if (typeof window === "undefined") return;
    setSub(extractSubdomain(window.location.host, getApexDomain()));
  }, []);
  return sub;
}
