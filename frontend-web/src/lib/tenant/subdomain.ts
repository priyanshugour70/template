/**
 * Subdomain parsing used by both server (proxy.ts, route handlers) and client
 * (login page when the URL needs to be reconstructed). Pure function — no
 * Next.js or Node imports — safe in Edge/Node/browser.
 *
 * The apex domain is configured via NEXT_PUBLIC_APEX_DOMAIN (e.g. "lssgoo.com"
 * in prod, "lvh.me" in dev). Hosts that don't end in the apex are treated as
 * apex itself (defensive fallback so a misconfigured host doesn't accidentally
 * masquerade as a tenant).
 */

import { isReservedSubdomain } from "./reserved";

export const DEFAULT_APEX = "lvh.me";

export function getApexDomain(): string {
  return (process.env.NEXT_PUBLIC_APEX_DOMAIN ?? DEFAULT_APEX).toLowerCase();
}

/**
 * extractSubdomain returns the tenant slug for a given host, or null when the
 * host is the apex / www / a reserved subdomain.
 *
 *   acme.lssgoo.com     → "acme"
 *   www.lssgoo.com      → null
 *   lssgoo.com          → null
 *   api.lssgoo.com      → null  (reserved)
 *   acme.lvh.me:3000    → "acme"   (dev)
 *   localhost           → null
 *   localhost:3000      → null
 *   acme.localhost:3000 → "acme"   (Chrome/Firefox resolve *.localhost)
 */
export function extractSubdomain(host: string | null | undefined, apex: string = getApexDomain()): string | null {
  if (!host) return null;
  // Strip the port — host headers in dev look like "acme.lvh.me:3000".
  const bareHost = host.split(":")[0].toLowerCase();
  if (!bareHost) return null;

  // Apex itself, or no dotted prefix at all.
  if (bareHost === apex) return null;

  // Localhost without subdomain.
  if (bareHost === "localhost") return null;

  // Tenant must end with ".<apex>" — otherwise it's not one of ours.
  const suffix = "." + apex;
  if (!bareHost.endsWith(suffix)) {
    // Localhost convenience: Chrome resolves "*.localhost" automatically. Treat
    // anything ending in ".localhost" as a tenant subdomain in dev.
    if (bareHost.endsWith(".localhost")) {
      const sub = bareHost.slice(0, -".localhost".length);
      return sub && !isReservedSubdomain(sub) ? sub : null;
    }
    return null;
  }

  const sub = bareHost.slice(0, -suffix.length);
  if (!sub || sub.includes(".")) return null; // No multi-level subdomains.
  if (isReservedSubdomain(sub)) return null;
  return sub;
}

/**
 * buildTenantUrl returns the full origin for a given tenant slug. Used when
 * redirecting to the tenant subdomain after apex login.
 *
 *   buildTenantUrl("acme")  → "https://acme.lssgoo.com"
 *   buildTenantUrl("acme")  → "http://acme.lvh.me:3000"   (dev)
 */
export function buildTenantUrl(
  slug: string,
  opts: { apex?: string; path?: string; port?: string | null } = {},
): string {
  const apex = (opts.apex ?? getApexDomain()).toLowerCase();
  const isDev = apex === "lvh.me" || apex === "localhost" || apex.endsWith(".localhost") || apex.endsWith(".test");
  const scheme = isDev ? "http" : "https";
  const port = opts.port === undefined ? (isDev ? ":3000" : "") : opts.port ? `:${opts.port}` : "";
  const path = opts.path ?? "";
  return `${scheme}://${slug}.${apex}${port}${path}`;
}

/**
 * buildApexUrl returns the apex origin. Used to redirect from a tenant
 * subdomain back to lssgoo.com/login when the session is invalid.
 */
export function buildApexUrl(opts: { apex?: string; path?: string; port?: string | null } = {}): string {
  const apex = (opts.apex ?? getApexDomain()).toLowerCase();
  const isDev = apex === "lvh.me" || apex === "localhost" || apex.endsWith(".localhost") || apex.endsWith(".test");
  const scheme = isDev ? "http" : "https";
  const port = opts.port === undefined ? (isDev ? ":3000" : "") : opts.port ? `:${opts.port}` : "";
  const path = opts.path ?? "";
  return `${scheme}://${apex}${port}${path}`;
}
