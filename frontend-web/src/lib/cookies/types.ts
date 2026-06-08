/**
 * Shapes of every JSON-encoded cookie the app writes. Kept lean — these are
 * sent on every request and rendered on first paint, so only carry what
 * sidebar / header / role gates actually need.
 */

export interface Tokens {
  accessToken?: string;
  refreshToken?: string;
}

export interface SessionUser {
  id: string;
  email: string;
  displayName?: string;
  firstName?: string;
  lastName?: string;
  avatarUrl?: string;
  isSuperAdmin?: boolean;
}

export interface SessionTenant {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  faviconUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
}

export interface SessionOrganization {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  isDefault?: boolean;
  status?: string;
}

export interface SessionSnapshot {
  user: SessionUser | null;
  tenant: SessionTenant | null;
  activeOrganization: SessionOrganization | null;
  organizations: SessionOrganization[];
  permissions: string[];
  roles: string[];
}

export interface CookieWriteOptions {
  /** Override the default 7-day max age. */
  maxAgeSeconds?: number;
  /** Override the default Lax SameSite. */
  sameSite?: "lax" | "strict" | "none";
}
