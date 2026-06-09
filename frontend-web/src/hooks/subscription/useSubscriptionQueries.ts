"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { subscriptionService } from "@/services/subscription";
import type {
  CancelSubscriptionRequest,
  ChangePlanRequest,
  PauseRequest,
  PreviewChangeRequest,
  UpdateBillingRequest,
  ValidateCouponRequest,
} from "@/types/subscription";

const KEY = {
  plans: ["subscription", "plans"] as const,
  active: ["subscription", "active"] as const,
  features: ["subscription", "features"] as const,
  usage: ["subscription", "usage"] as const,
  invoices: ["subscription", "invoices"] as const,
  invoice: (id: string) => ["subscription", "invoice", id] as const,
};

function bustAll(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: ["subscription"] });
}

export function usePlans() {
  return useQuery({
    queryKey: KEY.plans,
    queryFn: async () => {
      const res = await subscriptionService.listPlans();
      if (!res.success) throw new Error(res.error?.message ?? "plans failed");
      return res.data ?? [];
    },
  });
}

export function useActiveSubscription() {
  return useQuery({
    queryKey: KEY.active,
    queryFn: async () => {
      const res = await subscriptionService.getActive();
      if (!res.success) throw new Error(res.error?.message ?? "active subscription failed");
      return res.data!;
    },
    retry: false,
  });
}

export function useFeatureSet() {
  return useQuery({
    queryKey: KEY.features,
    queryFn: async () => {
      const res = await subscriptionService.features();
      if (!res.success) throw new Error(res.error?.message ?? "features failed");
      return res.data!;
    },
  });
}

export function useUsage() {
  return useQuery({
    queryKey: KEY.usage,
    queryFn: async () => {
      const res = await subscriptionService.listUsage();
      if (!res.success) throw new Error(res.error?.message ?? "usage failed");
      return res.data ?? [];
    },
  });
}

export function useInvoices() {
  return useQuery({
    queryKey: KEY.invoices,
    queryFn: async () => {
      const res = await subscriptionService.listInvoices();
      if (!res.success) throw new Error(res.error?.message ?? "invoices failed");
      return res.data ?? [];
    },
  });
}

export function useInvoice(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.invoice(id) : ["subscription", "invoice", "_"],
    queryFn: async () => {
      const res = await subscriptionService.getInvoice(id!);
      if (!res.success) throw new Error(res.error?.message ?? "invoice fetch failed");
      return res.data!;
    },
  });
}

export function useChangePlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: ChangePlanRequest) => {
      const res = await subscriptionService.changePlan(req);
      if (!res.success) throw new Error(res.error?.message ?? "change plan failed");
      return res.data!;
    },
    onSuccess: () => bustAll(qc),
  });
}

export function usePreviewChange() {
  return useMutation({
    mutationFn: async (req: PreviewChangeRequest) => {
      const res = await subscriptionService.previewChange(req);
      if (!res.success) throw new Error(res.error?.message ?? "preview failed");
      return res.data!;
    },
  });
}

export function useValidateCoupon() {
  return useMutation({
    mutationFn: async (req: ValidateCouponRequest) => {
      const res = await subscriptionService.validateCoupon(req);
      if (!res.success) throw new Error(res.error?.message ?? "validate coupon failed");
      return res.data!;
    },
  });
}

export function useCancelSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CancelSubscriptionRequest) => {
      const res = await subscriptionService.cancel(req);
      if (!res.success) throw new Error(res.error?.message ?? "cancel failed");
    },
    onSuccess: () => bustAll(qc),
  });
}

export function useReactivateSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await subscriptionService.reactivate();
      if (!res.success) throw new Error(res.error?.message ?? "reactivate failed");
      return res.data!;
    },
    onSuccess: () => bustAll(qc),
  });
}

export function usePauseSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: PauseRequest = {}) => {
      const res = await subscriptionService.pause(req);
      if (!res.success) throw new Error(res.error?.message ?? "pause failed");
      return res.data!;
    },
    onSuccess: () => bustAll(qc),
  });
}

export function useResumeSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await subscriptionService.resume();
      if (!res.success) throw new Error(res.error?.message ?? "resume failed");
      return res.data!;
    },
    onSuccess: () => bustAll(qc),
  });
}

export function useUpdateBilling() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateBillingRequest) => {
      const res = await subscriptionService.updateBilling(req);
      if (!res.success) throw new Error(res.error?.message ?? "update billing failed");
      return res.data!;
    },
    onSuccess: () => bustAll(qc),
  });
}
