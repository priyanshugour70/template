import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

// Re-export the subscription / invoice types so we don't break existing consumers
// during the Phase 8 rollout. Phase 10 cleanup deletes `@/types/subscription`.
export type {
  BillingCycle,
  Subscription,
  SubscriptionStatus,
  UsageCounter,
  Plan,
  Invoice,
  InvoiceStatus,
  InvoiceLineItem,
  FeatureSet,
  UpdateBillingRequest,
  CancelSubscriptionRequest,
} from "@/types/subscription";

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

export interface ActivateQuotationResponse {
  quotation: Quotation;
  plan: import("@/types/subscription").Plan;
  subscription: import("@/types/subscription").Subscription;
  invoice: import("@/types/subscription").Invoice;
  invoiceLines: InvoiceLineRow[];
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
  invoice: import("@/types/subscription").Invoice;
  subscription?: import("@/types/subscription").Subscription;
  receiptUrl?: string;
}

// ── Admin: cycle ───────────────────────────────────────────────────────────

export interface CycleReport {
  trialsExpired: number;
  invoicesIssued: number;
  errors?: string[];
}
