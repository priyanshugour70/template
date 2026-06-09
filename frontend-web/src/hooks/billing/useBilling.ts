"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { billingService } from "@/services/billing";
import type {
  CreateQuotationRequest,
  PreviewQuoteRequest,
  RecordPaymentRequest,
  UpdateQuotationRequest,
} from "@/types/billing";

// Single root key — mutations bust everything billing-related.
const ROOT = "billing" as const;
const KEY = {
  features: [ROOT, "features"] as const,
  active: [ROOT, "subscription", "active"] as const,
  quotations: (status?: string) => [ROOT, "quotations", status ?? "_all"] as const,
  quotation: (id: string) => [ROOT, "quotation", id] as const,
  invoices: [ROOT, "invoices"] as const,
  invoice: (id: string) => [ROOT, "invoice", id] as const,
  transactions: [ROOT, "transactions"] as const,
  transaction: (id: string) => [ROOT, "transaction", id] as const,
};

function bust(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: [ROOT] });
}

// ── catalog + preview ─────────────────────────────────────────────────────

export function useFeatureCatalog() {
  return useQuery({
    queryKey: KEY.features,
    queryFn: async () => {
      const res = await billingService.listFeatures();
      if (!res.success) throw new Error(res.error?.message ?? "features failed");
      return res.data ?? [];
    },
    staleTime: 10 * 60_000,
  });
}

export function usePreviewQuote() {
  return useMutation({
    mutationFn: async (req: PreviewQuoteRequest) => {
      const res = await billingService.previewQuote(req);
      if (!res.success) throw new Error(res.error?.message ?? "preview failed");
      return res.data!;
    },
  });
}

// ── subscription ──────────────────────────────────────────────────────────

export function useActiveBilling() {
  return useQuery({
    queryKey: KEY.active,
    queryFn: async () => {
      const res = await billingService.getActiveSubscription();
      if (!res.success) throw new Error(res.error?.message ?? "active failed");
      return res.data!;
    },
    retry: false,
  });
}

export function useFeatureSet() {
  return useQuery({
    queryKey: [ROOT, "subscription", "features"],
    queryFn: async () => {
      const res = await billingService.features();
      if (!res.success) throw new Error(res.error?.message ?? "features failed");
      return res.data!;
    },
  });
}

export function useStartTrial() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await billingService.startTrial();
      if (!res.success) throw new Error(res.error?.message ?? "start trial failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useCancelBilling() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: { reason?: string; immediate?: boolean } = {}) => {
      const res = await billingService.cancel(req);
      if (!res.success) throw new Error(res.error?.message ?? "cancel failed");
    },
    onSuccess: () => bust(qc),
  });
}

export function useUpdateBillingInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: import("@/types/billing").UpdateBillingRequest) => {
      const res = await billingService.updateBilling(req);
      if (!res.success) throw new Error(res.error?.message ?? "update billing failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

// ── quotations ────────────────────────────────────────────────────────────

export function useQuotations(status?: string) {
  return useQuery({
    queryKey: KEY.quotations(status),
    queryFn: async () => {
      const res = await billingService.listQuotations({ status });
      if (!res.success) throw new Error(res.error?.message ?? "quotations failed");
      return res.data ?? [];
    },
  });
}

export function useQuotation(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.quotation(id) : [ROOT, "quotation", "_"],
    queryFn: async () => {
      const res = await billingService.getQuotation(id!);
      if (!res.success) throw new Error(res.error?.message ?? "quotation failed");
      return res.data!;
    },
  });
}

export function useCreateQuotation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateQuotationRequest) => {
      const res = await billingService.createQuotation(req);
      if (!res.success) throw new Error(res.error?.message ?? "create quotation failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useUpdateQuotation(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateQuotationRequest) => {
      const res = await billingService.updateQuotation(id, req);
      if (!res.success) throw new Error(res.error?.message ?? "update quotation failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useDeleteQuotation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await billingService.deleteQuotation(id);
      if (!res.success) throw new Error(res.error?.message ?? "delete quotation failed");
    },
    onSuccess: () => bust(qc),
  });
}

export function useActivateQuotation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await billingService.activateQuotation(id);
      if (!res.success) throw new Error(res.error?.message ?? "activate quotation failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

// ── invoices ──────────────────────────────────────────────────────────────

export function useInvoices() {
  return useQuery({
    queryKey: KEY.invoices,
    queryFn: async () => {
      const res = await billingService.listInvoices();
      if (!res.success) throw new Error(res.error?.message ?? "invoices failed");
      return res.data ?? [];
    },
  });
}

export function useInvoice(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.invoice(id) : [ROOT, "invoice", "_"],
    queryFn: async () => {
      const res = await billingService.getInvoice(id!);
      if (!res.success) throw new Error(res.error?.message ?? "invoice failed");
      return res.data!;
    },
  });
}

// ── payments ──────────────────────────────────────────────────────────────

export function useRecordPayment(invoiceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: RecordPaymentRequest) => {
      const res = await billingService.recordPayment(invoiceId, req);
      if (!res.success) throw new Error(res.error?.message ?? "record payment failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useTransactions() {
  return useQuery({
    queryKey: KEY.transactions,
    queryFn: async () => {
      const res = await billingService.listTransactions();
      if (!res.success) throw new Error(res.error?.message ?? "transactions failed");
      return res.data ?? [];
    },
  });
}

export function useTransaction(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.transaction(id) : [ROOT, "transaction", "_"],
    queryFn: async () => {
      const res = await billingService.getTransaction(id!);
      if (!res.success) throw new Error(res.error?.message ?? "transaction failed");
      return res.data!;
    },
  });
}
