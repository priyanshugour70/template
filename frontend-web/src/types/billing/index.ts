import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

// ── Subscription / Plan ───────────────────────────────────────────────────

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
  isCustom?: boolean;
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

export interface CancelSubscriptionRequest {
  reason?: string;
  immediate?: boolean;
}

export interface UpdateBillingRequest {
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
}

export interface FeatureSet {
  planCode: string;
  status: SubscriptionStatus;
  features: Record<string, boolean>;
  limits: Record<string, number>;
}

// ── Invoices ───────────────────────────────────────────────────────────────

export type InvoiceStatus = "open" | "paid" | "void" | "uncollectible" | "refunded";

export interface InvoiceLineItem {
  description: string;
  quantity: number;
  unitCents?: number;
  amountCents?: number;
  unitPriceCents?: number;
  totalCents?: number;
  hsnSac?: string;
  featureKey?: string;
  taxableAmountCents?: number;
  tax?: QuoteTax;
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
  hsnSac?: string;
  placeOfSupply?: string;
  cgstCents?: number;
  sgstCents?: number;
  igstCents?: number;
  pdfStorageKey?: string;
  metadata: JSONObject;
}

// ── Feature catalog ───────────────────────────────────────────────────────

export type FeatureCategory = "core" | "admin" | "compliance" | "integrations" | "limits";

export interface Feature extends BaseEntity {
  key: string;
  name: string;
  description: string;
  category: FeatureCategory;
  basePriceCents: number;
  perUserPriceCents: number;
  includedUsers: number;
  isCore: boolean;
  isStarterDefault: boolean;
  isActive: boolean;
  requires: string[];
  sortOrder: number;
}

// ── Quotation lifecycle ────────────────────────────────────────────────────

export type QuotationStatus = "draft" | "accepted" | "rejected" | "expired";

export interface QuoteTax {
  cgstCents: number;
  sgstCents: number;
  igstCents: number;
  cgstPct: number;
  sgstPct: number;
  igstPct: number;
  totalTaxCents: number;
}

export interface QuoteLine {
  featureKey: string;
  description: string;
  hsnSac: string;
  quantity: number;
  unitPriceCents: number;
  taxableAmountCents: number;
  tax: QuoteTax;
  totalCents: number;
  sortOrder: number;
}

export interface Quote {
  lines: QuoteLine[];
  subtotalCents: number;
  cgstCents: number;
  sgstCents: number;
  igstCents: number;
  totalCents: number;
  currency: string;
  placeOfSupply: string;
  extraUsers: number;
}

export interface PreviewQuoteRequest {
  featureKeys: string[];
  userCount: number;
  customerState?: string;
}

export interface Quotation extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  number: string;
  status: QuotationStatus;
  featureKeys: string[];
  userCount: number;
  subtotalCents: number;
  discountCents: number;
  cgstCents: number;
  sgstCents: number;
  igstCents: number;
  totalCents: number;
  currency: string;
  placeOfSupply?: string;
  lineItems: QuoteLine[];
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  billingState?: string;
  notes?: string;
  expiresAt: ISODate;
  acceptedAt?: ISODate;
  rejectedAt?: ISODate;
  activatedPlanId?: ID;
  activatedSubscriptionId?: ID;
  metadata?: JSONObject;
}

export interface CreateQuotationRequest {
  featureKeys: string[];
  userCount: number;
  customerState?: string;
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  notes?: string;
}

export interface UpdateQuotationRequest {
  featureKeys?: string[];
  userCount?: number;
  customerState?: string;
  billingEmail?: string;
  billingName?: string;
  billingAddress?: JSONObject;
  notes?: string;
}

export interface InvoiceLineRow {
  id: ID;
  invoiceId: ID;
  featureKey?: string;
  description: string;
  hsnSac: string;
  quantity: number;
  unitPriceCents: number;
  taxableAmountCents: number;
  cgstCents: number;
  sgstCents: number;
  igstCents: number;
  totalCents: number;
  sortOrder: number;
  metadata?: JSONObject;
  createdAt: ISODate;
}

export interface ActivateQuotationResponse {
  quotation: Quotation;
  plan: Plan;
  subscription: Subscription;
  invoice: Invoice;
  invoiceLines: InvoiceLineRow[];
}

// ── Payments / Transactions ───────────────────────────────────────────────

export type PaymentMethod = "cash" | "bank_transfer" | "cheque" | "gateway";

export interface Transaction extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  invoiceId: ID;
  receiptNumber: string;
  method: PaymentMethod;
  status: "recorded" | "pending" | "failed" | "refunded";
  amountCents: number;
  currency: string;
  reference?: string;
  gateway?: string;
  gatewayTransactionId?: string;
  paidAt: ISODate;
  refundedAt?: ISODate;
  refundAmountCents: number;
  pdfStorageKey?: string;
  notes?: string;
  metadata?: JSONObject;
}

export interface RecordPaymentRequest {
  method: PaymentMethod;
  amountCents: number;
  reference?: string;
  notes?: string;
}

export interface RecordPaymentResponse {
  transaction: Transaction;
  invoice: Invoice;
  subscription?: Subscription;
  receiptUrl?: string;
}

// ── Admin: cycle ───────────────────────────────────────────────────────────

export interface CycleReport {
  trialsExpired: number;
  invoicesIssued: number;
  errors?: string[];
}
