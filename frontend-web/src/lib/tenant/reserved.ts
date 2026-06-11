/**
 * Subdomain slugs that must NEVER be treated as tenant slugs. Keep in sync
 * with backend/internal/modules/auth/reserved_slugs.go — both sides reject
 * registration with one of these and both sides refuse to render the
 * "tenant dashboard" if the current subdomain is reserved (used for infra
 * like api.lssgoo.com / www.lssgoo.com).
 */
export const RESERVED_SUBDOMAINS: ReadonlySet<string> = new Set([
  "api",
  "www",
  "admin",
  "app",
  "auth",
  "mail",
  "smtp",
  "imap",
  "cdn",
  "static",
  "assets",
  "docs",
  "status",
  "support",
  "help",
  "blog",
  "dashboard",
  "console",
  "dev",
  "staging",
  "prod",
  "production",
  "test",
  "root",
  "ws",
  "mx",
  "ftp",
  "pop",
  "webhook",
  "webhooks",
]);

export function isReservedSubdomain(slug: string | null | undefined): boolean {
  if (!slug) return false;
  return RESERVED_SUBDOMAINS.has(slug.toLowerCase().trim());
}
