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
  lastBilledAt?: ISODate;
  cancelAt?: ISODate;
  cancelledAt?: ISODate;
  cancelReason?: string;
  endedAt?: ISODate;
  gateway?: string;
  gatewayCustomerId?: string;
  gatewaySubscriptionId?: string;
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  // Indian-state-aware billing — added in migration 012; drives which GST
  // columns the invoice carries.
  billingState?: string;
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

// ── invoices ───────────────────────────────────────────────────────────────

export type InvoiceStatus = "open" | "paid" | "void" | "uncollectible" | "refunded";

export interface InvoiceLineItem {
  description: string;
  quantity: number;
  // Old legacy shape (Phase 0).
  unitCents?: number;
  amountCents?: number;
  // New richer shape used by the billing module (Phase 3+). Both shapes are
  // optional so JSONB rows from the old plan-change flow still render.
  unitPriceCents?: number;
  totalCents?: number;
  hsnSac?: string;
  featureKey?: string;
  taxableAmountCents?: number;
  tax?: {
    cgstCents: number;
    sgstCents: number;
    igstCents: number;
    cgstPct: number;
    sgstPct: number;
    igstPct: number;
    totalTaxCents: number;
  };
}

export interface Invoice extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  subscriptionId?: ID;
  number: string;
  status: InvoiceStatus;
  currency: string;
  subtotalCents: number;
  discountCents: number;
  taxCents: number;
  totalCents: number;
  amountDueCents: number;
  amountPaidCents: number;
  couponCode?: string;
  description?: string;
  lineItems: InvoiceLineItem[];
  periodStart?: ISODate;
  periodEnd?: ISODate;
  issuedAt: ISODate;
  dueAt?: ISODate;
  paidAt?: ISODate;
  voidedAt?: ISODate;
  gateway?: string;
  gatewayInvoiceId?: string;
  // GST + storage extensions added in migration 012 (Phase 1 of the billing
  // overhaul). Optional in the type so old responses without these fields
  // still type-check.
  hsnSac?: string;
  placeOfSupply?: string;
  cgstCents?: number;
  sgstCents?: number;
  igstCents?: number;
  pdfStorageKey?: string;
  metadata: JSONObject;
}

// ── lifecycle DTOs ─────────────────────────────────────────────────────────

export interface PauseRequest {
  resumeAt?: ISODate;
  reason?: string;
}

export interface UpdateBillingRequest {
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
}

export interface PreviewChangeRequest {
  planCode: string;
  billingCycle?: BillingCycle;
  quantity?: number;
  couponCode?: string;
}

export interface PreviewChangeResponse {
  fromPlanCode: string;
  toPlanCode: string;
  billingCycle: BillingCycle;
  currency: string;
  baseAmountCents: number;
  prorationCents: number;
  couponCode?: string;
  discountCents: number;
  taxCents: number;
  totalDueCents: number;
  effectiveAt: ISODate;
  isUpgrade: boolean;
  unusedDaysRemaining: number;
}

export interface ValidateCouponRequest {
  code: string;
  planCode?: string;
}

export interface ValidateCouponResponse {
  valid: boolean;
  reason?: string;
  code?: string;
  name?: string;
  percentOff?: number;
  amountOffCents?: number;
  currency?: string;
  duration?: "once" | "forever" | "repeating";
}

export interface ChangePlanResponse {
  subscription: Subscription;
  invoice?: Invoice | null;
}
