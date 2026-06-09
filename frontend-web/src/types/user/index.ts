import type { BaseEntity, ID, ISODate, JSONObject, PaginatedQuery } from "@/types/common";

export type UserStatus = "active" | "suspended" | "invited" | "pending" | "archived";
export type MembershipStatus = UserStatus;

export interface UserProfile extends BaseEntity {
  email: string;
  emailVerifiedAt?: ISODate;
  firstName?: string;
  middleName?: string;
  lastName?: string;
  displayName?: string;
  username?: string;
  avatarUrl?: string;
  coverUrl?: string;
  bio?: string;
  phone?: string;
  phoneVerifiedAt?: ISODate;
  altEmail?: string;
  dateOfBirth?: ISODate;
  gender?: string;
  jobTitle?: string;
  department?: string;
  employeeCode?: string;
  status: UserStatus;
  locale: string;
  timezone: string;
  country?: string;
  state?: string;
  city?: string;
  address?: JSONObject;
  preferences: JSONObject;
  notificationPreferences: JSONObject;
  metadata: JSONObject;
  lastLoginAt?: ISODate;
  lastLoginIp?: string | null;
  lastLoginUserAgent?: string;
  failedLoginCount: number;
  lockedUntil?: ISODate;
  mfaEnabled: boolean;
  isSuperAdmin: boolean;
  primaryTenantId?: ID;
  primaryOrganizationId?: ID;
  passwordChangedAt?: ISODate;
  mustChangePassword?: boolean;
  termsAcceptedAt?: ISODate;
  termsVersion?: string;
  marketingOptIn?: boolean;
}

export interface Membership extends BaseEntity {
  userId: ID;
  tenantId: ID;
  organizationId: ID;
  status: MembershipStatus;
  isDefault: boolean;
  isOwner: boolean;
  isBillingContact: boolean;
  jobTitle?: string;
  department?: string;
  departmentId?: ID | null;
  employeeCode?: string;
  reportsTo?: ID;
  invitedBy?: ID;
  invitedAt?: ISODate;
  joinedAt?: ISODate;
  lastActiveAt?: ISODate;
  settings: JSONObject;
  metadata: JSONObject;
}

export interface UpdateMembershipRequest {
  departmentId?: ID | null;
  jobTitle?: string;
  department?: string;
  employeeCode?: string;
  reportsTo?: ID | null;
}

export interface BulkUpdateMembershipsRequest {
  membershipIds: ID[];
  patch: UpdateMembershipRequest;
}

export interface BulkUpdateMembershipsResponse {
  updated: number;
  failed?: { membershipId: ID; error: string }[];
}

export interface EffectivePermissionsResponse {
  permissions: string[];
}

// ── requests ───────────────────────────────────────────────────────────────

export interface UpdateUserRequest {
  firstName?: string;
  middleName?: string;
  lastName?: string;
  displayName?: string;
  username?: string;
  avatarUrl?: string;
  coverUrl?: string;
  bio?: string;
  phone?: string;
  altEmail?: string;
  dateOfBirth?: ISODate;
  gender?: string;
  jobTitle?: string;
  department?: string;
  employeeCode?: string;
  locale?: string;
  timezone?: string;
  country?: string;
  state?: string;
  city?: string;
  address?: JSONObject;
  preferences?: JSONObject;
  notificationPreferences?: JSONObject;
  metadata?: JSONObject;
}

export interface UserListQuery extends PaginatedQuery {
  status?: UserStatus;
  role?: string;
  jobTitle?: string;
  department?: string;
  departmentId?: ID;
  mfa?: boolean;
  lastLoginAfter?: ISODate;
  lastLoginBefore?: ISODate;
  createdAfter?: ISODate;
  createdBefore?: ISODate;
}

export interface InviteUserRequest {
  email: string;
  firstName?: string;
  lastName?: string;
  jobTitle?: string;
  department?: string;
  departmentId?: ID;
  organizationId: ID;
  roleKeys?: string[];
  message?: string;
}
