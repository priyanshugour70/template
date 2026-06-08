import type { BaseEntity, ID, ISODate, JSONObject, PaginatedQuery } from "@/types/common";

export type TenantStatus = "active" | "suspended" | "trial" | "pending" | "archived";
export type OrganizationStatus = "active" | "suspended" | "pending" | "archived";

export interface Tenant extends BaseEntity {
  slug: string;
  name: string;
  legalName?: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  faviconUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
  supportEmail?: string;
  supportPhone?: string;
  websiteUrl?: string;
  status: TenantStatus;
  planCode?: string;
  seatLimit?: number;
  country?: string;
  timezone: string;
  locale: string;
  currency: string;
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  taxId?: string;
  settings: JSONObject;
  features: JSONObject;
  metadata: JSONObject;
  trialEndsAt?: ISODate;
  activatedAt?: ISODate;
  suspendedAt?: ISODate;
  suspensionReason?: string;
  archivedAt?: ISODate;
}

export interface Organization extends BaseEntity {
  tenantId: ID;
  slug: string;
  name: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  coverUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
  websiteUrl?: string;
  contactEmail?: string;
  contactPhone?: string;
  industry?: string;
  size?: string;
  country?: string;
  state?: string;
  city?: string;
  postalCode?: string;
  timezone: string;
  locale: string;
  currency: string;
  status: OrganizationStatus;
  isDefault: boolean;
  settings: JSONObject;
  features: JSONObject;
  metadata: JSONObject;
  address?: JSONObject;
  activatedAt?: ISODate;
  suspendedAt?: ISODate;
  archivedAt?: ISODate;
}

// ── requests ───────────────────────────────────────────────────────────────

export interface CreateTenantRequest {
  slug: string;
  name: string;
  legalName?: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  primaryColor?: string;
  supportEmail?: string;
  supportPhone?: string;
  websiteUrl?: string;
  planCode?: string;
  country?: string;
  timezone?: string;
  locale?: string;
  currency?: string;
  billingEmail?: string;
  taxId?: string;
  adminEmail: string;
  adminFirstName?: string;
  adminLastName?: string;
}

export interface UpdateTenantRequest {
  name?: string;
  legalName?: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  faviconUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
  supportEmail?: string;
  supportPhone?: string;
  websiteUrl?: string;
  country?: string;
  timezone?: string;
  locale?: string;
  currency?: string;
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  taxId?: string;
  settings?: JSONObject;
  features?: JSONObject;
  metadata?: JSONObject;
}

export interface CreateOrganizationRequest {
  slug: string;
  name: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  websiteUrl?: string;
  contactEmail?: string;
  contactPhone?: string;
  industry?: string;
  size?: string;
  country?: string;
  timezone?: string;
  locale?: string;
  currency?: string;
  isDefault?: boolean;
}

export interface UpdateOrganizationRequest {
  name?: string;
  displayName?: string;
  description?: string;
  logoUrl?: string;
  coverUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
  websiteUrl?: string;
  contactEmail?: string;
  contactPhone?: string;
  industry?: string;
  size?: string;
  country?: string;
  state?: string;
  city?: string;
  postalCode?: string;
  timezone?: string;
  locale?: string;
  currency?: string;
  address?: JSONObject;
  settings?: JSONObject;
  features?: JSONObject;
  metadata?: JSONObject;
}

export interface TenantListQuery extends PaginatedQuery {
  status?: TenantStatus;
}

export interface OrganizationListQuery extends PaginatedQuery {
  status?: OrganizationStatus;
}
