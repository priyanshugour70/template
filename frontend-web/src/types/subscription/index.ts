import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

export type BillingCycle = "monthly" | "quarterly" | "yearly" | "custom" | "one_time";
export type SubscriptionStatus =
  | "trial"
  | "active"
  | "past_due"
  | "cancelled"
  | "expired"
  | "paused"
  | "pending";

export interface Plan extends BaseEntity {
  code: string;
  name: string;
  description?: string;
  tagline?: string;
  tier: number;
  billingCycle: BillingCycle;
  priceCents: number;
  currency: string;
  trialDays: number;
  isActive: boolean;
  isDefault: boolean;
  isPublic: boolean;
  isAddon: boolean;
  features: string[];
  limits: Record<string, number>;
  metadata: JSONObject;
}

export interface Subscription extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  planId: ID;
  planCode: string;
  status: SubscriptionStatus;
  billingCycle: BillingCycle;
  quantity: number;
  unitPriceCents: number;
  discountCents: number;
  taxCents: number;
  totalCents: number;
  currency: string;
  startedAt: ISODate;
  trialStartedAt?: ISODate;
  trialEndsAt?: ISODate;
  currentPeriodStart?: ISODate;
  currentPeriodEnd?: ISODate;
  nextBillingAt?: ISODate;
  cancelAt?: ISODate;
  cancelledAt?: ISODate;
  cancelReason?: string;
  endedAt?: ISODate;
  gateway?: string;
  gatewayCustomerId?: string;
  gatewaySubscriptionId?: string;
  billingEmail?: string;
  features: string[];
  limits: Record<string, number>;
  metadata: JSONObject;
}

export interface UsageCounter extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  subscriptionId?: ID;
  key: string;
  count: number;
  limitValue?: number;
  periodStart: ISODate;
  periodEnd: ISODate;
  lastResetAt?: ISODate;
  metadata: JSONObject;
}

export interface ChangePlanRequest {
  planCode: string;
  billingCycle?: BillingCycle;
  quantity?: number;
  startImmediately?: boolean;
  couponCode?: string;
}

export interface CancelSubscriptionRequest {
  reason?: string;
  immediate?: boolean;
}

export interface FeatureSet {
  planCode: string;
  status: SubscriptionStatus;
  features: Record<string, boolean>;
  limits: Record<string, number>;
}
