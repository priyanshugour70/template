"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { subscriptionService } from "@/services/subscription";
import type {
  CancelSubscriptionRequest,
  ChangePlanRequest,
} from "@/types/subscription";

const KEY = {
  plans: ["subscription", "plans"] as const,
  active: ["subscription", "active"] as const,
  features: ["subscription", "features"] as const,
  usage: ["subscription", "usage"] as const,
};

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

export function useChangePlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: ChangePlanRequest) => {
      const res = await subscriptionService.changePlan(req);
      if (!res.success) throw new Error(res.error?.message ?? "change plan failed");
      return res.data!;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY.active });
      qc.invalidateQueries({ queryKey: KEY.features });
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
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY.active });
      qc.invalidateQueries({ queryKey: KEY.features });
    },
  });
}
