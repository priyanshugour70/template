/**
 * Server-only tenant helpers. Reads the subdomain that proxy.ts injected as
 * `x-tenant-slug` so server components and route handlers don't have to
 * re-parse the host header.
 */

import "server-only";

import { headers } from "next/headers";

import { extractSubdomain, getApexDomain } from "./subdomain";

/** Returns the tenant slug for the current request, or null on the apex. */
export async function getServerSubdomain(): Promise<string | null> {
  const h = await headers();
  const fromProxy = h.get("x-tenant-slug");
  if (fromProxy) return fromProxy;
  // Fallback: route handlers in src/app/api/* are excluded from proxy.ts
  // matcher, so the header isn't there — parse the host header ourselves.
  const host = h.get("host");
  return extractSubdomain(host, getApexDomain());
}

/** True when the current request is on the apex (lssgoo.com, not a tenant). */
export async function isApexRequest(): Promise<boolean> {
  return (await getServerSubdomain()) === null;
}
