"use client";

import { useMemo } from "react";

import { extractSubdomain, getApexDomain } from "@/lib/tenant/subdomain";

/**
 * Reads the current subdomain from the browser host. Returns null on the apex
 * or on SSR (no window). Client components that need to know "am I on a
 * tenant page" can call this — server components should use
 * getServerSubdomain() from @/lib/tenant/server.
 */
export function useSubdomain(): string | null {
  return useMemo(() => {
    if (typeof window === "undefined") return null;
    return extractSubdomain(window.location.host, getApexDomain());
  }, []);
}
