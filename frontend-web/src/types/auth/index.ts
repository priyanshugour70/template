import type { SessionOrganization, SessionTenant, SessionUser } from "@/lib/cookies/types";
import type { ID, ISODate, JSONObject } from "@/types/common";

/** Discovered tenant shape from /auth/discover and /auth/login. */
export interface DiscoveredTenant {
  id: ID;
  name: string;
  slug: string;
  logoUrl?: string;
  primaryColor?: string;
}

export interface DiscoverResponse {
  tenants: DiscoveredTenant[];
}

export interface LoginRequest {
  tenantId: ID;
  email: string;
  password: string;
}

export interface SwitchOrgRequest {
  organizationId: ID;
}

export interface AcceptInviteRequest {
  token: string;
  firstName: string;
  lastName?: string;
  password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

/** Returned by /api/auth/login (tokens stripped — they live in HttpOnly cookies). */
export interface SessionResponse {
  user: SessionUser;
  tenant: SessionTenant;
  activeOrganization?: SessionOrganization | null;
  organizations: SessionOrganization[];
  tokenType?: string;
  accessTokenExpiresAt?: ISODate;
  refreshTokenExpiresAt?: ISODate;
}

export interface UserSession {
  id: ID;
  userId: ID;
  tenantId: ID;
  organizationId?: ID | null;
  membershipId?: ID | null;
  deviceId?: string;
  deviceName?: string;
  client?: string;
  ip?: string | null;
  userAgent?: string;
  issuedAt: ISODate;
  expiresAt: ISODate;
  lastUsedAt?: ISODate;
  revokedAt?: ISODate | null;
  metadata?: JSONObject;
}

export type {
  SessionUser,
  SessionTenant,
  SessionOrganization,
} from "@/lib/cookies/types";

/** Legacy User shape kept for backward-compat with existing imports. */
export interface User extends SessionUser {
  permissions?: string[];
  roles?: string[];
}
