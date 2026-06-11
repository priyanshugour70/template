/**
 * Derive the backend WebSocket base URL.
 *
 * Precedence:
 *   1. NEXT_PUBLIC_WS_URL — explicit override.
 *   2. Browser host is localhost-ish → ws://localhost:8080 (dev backend).
 *   3. Else → wss://api.<NEXT_PUBLIC_APEX_DOMAIN>.
 *
 * The proxy at /api/v1/* handles REST. WS upgrades skip the proxy and hit
 * the backend directly, so the origin allow-list on the backend must permit
 * the tenant subdomain.
 */
export function commWSBaseURL(): string {
  const override = process.env.NEXT_PUBLIC_WS_URL?.trim();
  if (override) return override.replace(/\/$/, "");

  if (typeof window !== "undefined") {
    const host = window.location.hostname;
    if (host === "localhost" || host.endsWith(".localhost") || host === "lvh.me" || host.endsWith(".lvh.me")) {
      return "ws://localhost:8080";
    }
  }
  const apex = process.env.NEXT_PUBLIC_APEX_DOMAIN ?? "lvh.me";
  return `wss://api.${apex}`;
}
