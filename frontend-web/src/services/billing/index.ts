import { api } from "@/lib/client";
import type {
  ActivateQuotationResponse,
  CancelSubscriptionRequest,
  CreateQuotationRequest,
  CycleReport,
  Feature,
  FeatureSet,
  Invoice,
  PreviewQuoteRequest,
  Quotation,
  Quote,
  RecordPaymentRequest,
  RecordPaymentResponse,
  Subscription,
  Transaction,
  UpdateBillingRequest,
  UpdateQuotationRequest,
  UsageCounter,
} from "@/types/billing";

// Every billing endpoint lives under /api/v1/billing/*. Phase 10 retired the
// legacy /subscription-plans + /subscriptions/* surface; this service is the
// single source of truth for the client.
export const billingService = {
  // catalog + quote preview
  listFeatures: () => api.get<Feature[]>("/billing/features"),
  previewQuote: (req: PreviewQuoteRequest) =>
    api.post<Quote>("/billing/quotations/preview", req),

  // quotations
  listQuotations: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<Quotation[]>("/billing/quotations", { query: { limit: 100, ...params } }),
  getQuotation: (id: string) => api.get<Quotation>(`/billing/quotations/${id}`),
  createQuotation: (req: CreateQuotationRequest) =>
    api.post<Quotation>("/billing/quotations", req),
  updateQuotation: (id: string, req: UpdateQuotationRequest) =>
    api.patch<Quotation>(`/billing/quotations/${id}`, req),
  deleteQuotation: (id: string) => api.delete<unknown>(`/billing/quotations/${id}`),
  activateQuotation: (id: string) =>
    api.post<ActivateQuotationResponse>(`/billing/quotations/${id}/activate`),

  // active subscription
  getActiveSubscription: () => api.get<Subscription>("/billing/subscription"),
  features: () => api.get<FeatureSet>("/billing/subscription/features"),
  listUsage: () => api.get<UsageCounter[]>("/billing/subscription/usage"),
  cancel: (req: CancelSubscriptionRequest = {}) =>
    api.post<unknown>("/billing/subscription/cancel", req),
  updateBilling: (req: UpdateBillingRequest) =>
    api.patch<Subscription>("/billing/subscription/billing-info", req),
  startTrial: () => api.post<Subscription>("/billing/subscription/start-trial"),

  // invoices
  listInvoices: (limit = 50) =>
    api.get<Invoice[]>("/billing/invoices", { query: { limit } }),
  getInvoice: (id: string) => api.get<Invoice>(`/billing/invoices/${id}`),
  invoicePdfUrl: (id: string, download = false) =>
    `/api/v1/billing/invoices/${id}/pdf${download ? "?download=1" : ""}`,

  // payments + receipts
  recordPayment: (invoiceId: string, req: RecordPaymentRequest) =>
    api.post<RecordPaymentResponse>(`/billing/invoices/${invoiceId}/pay`, req),
  listTransactions: (params?: { page?: number; limit?: number }) =>
    api.get<Transaction[]>("/billing/transactions", { query: { limit: 100, ...params } }),
  getTransaction: (id: string) => api.get<Transaction>(`/billing/transactions/${id}`),
  receiptPdfUrl: (id: string, download = false) =>
    `/api/v1/billing/receipts/${id}/pdf${download ? "?download=1" : ""}`,

  // admin / cycle
  runCycle: (at?: string) => api.post<CycleReport>("/billing/admin/cycle/run", { at }),
  expireTrials: (at?: string) =>
    api.post<{ expired: number }>("/billing/admin/trials/expire", { at }),
};
