import { api } from "@/lib/client";
import type {
  ActivateQuotationResponse,
  CreateQuotationRequest,
  CycleReport,
  Feature,
  PreviewQuoteRequest,
  Quotation,
  Quote,
  RecordPaymentRequest,
  RecordPaymentResponse,
  Transaction,
  UpdateQuotationRequest,
} from "@/types/billing";
import type { Invoice, Subscription } from "@/types/subscription";

// All endpoints below sit under /api/v1/billing/* on the backend. The legacy
// /subscriptions/* aliases still exist while the rest of the app migrates.
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

  // invoices (Phase 4)
  listInvoices: (limit = 50) =>
    api.get<Invoice[]>("/subscriptions/invoices", { query: { limit } }),
  getInvoice: (id: string) => api.get<Invoice>(`/subscriptions/invoices/${id}`),
  invoicePdfUrl: (id: string, download = false) =>
    `/api/v1/billing/invoices/${id}/pdf${download ? "?download=1" : ""}`,

  // payments + receipts (Phase 5)
  recordPayment: (invoiceId: string, req: RecordPaymentRequest) =>
    api.post<RecordPaymentResponse>(`/billing/invoices/${invoiceId}/pay`, req),
  listTransactions: (params?: { page?: number; limit?: number }) =>
    api.get<Transaction[]>("/billing/transactions", { query: { limit: 100, ...params } }),
  getTransaction: (id: string) => api.get<Transaction>(`/billing/transactions/${id}`),
  receiptPdfUrl: (id: string, download = false) =>
    `/api/v1/billing/receipts/${id}/pdf${download ? "?download=1" : ""}`,

  // active subscription (the new flow keeps the legacy endpoints for read paths)
  getActiveSubscription: () => api.get<Subscription>("/subscriptions/active"),

  // admin / cycle (Phase 7)
  runCycle: (at?: string) => api.post<CycleReport>("/billing/admin/cycle/run", { at }),
  expireTrials: (at?: string) =>
    api.post<{ expired: number }>("/billing/admin/trials/expire", { at }),
};
