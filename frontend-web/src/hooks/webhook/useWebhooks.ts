"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { webhookService } from "@/services/webhook";
import type { WebhookCreate, WebhookTestFire, WebhookUpdate } from "@/types/webhook";

const KEY = {
  list: ["webhooks", "list"] as const,
  one: (id: string) => ["webhooks", "one", id] as const,
  deliveries: (id: string) => ["webhooks", "deliveries", id] as const,
};

function bust(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: ["webhooks"] });
}

type PageOpts = { page?: number; limit?: number };

export function useWebhooks(opts: PageOpts = {}) {
  return useQuery({
    queryKey: [...KEY.list, opts] as const,
    queryFn: async () => {
      const res = await webhookService.list(opts);
      if (!res.success) throw new Error(res.error?.message ?? "list failed");
      return {
        items: res.data ?? [],
        total: res.pagination?.total ?? (res.data?.length ?? 0),
        page: res.pagination?.page ?? 1,
        limit: res.pagination?.limit ?? (opts.limit ?? 200),
      };
    },
    placeholderData: (prev) => prev,
  });
}

export function useWebhookDeliveries(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.deliveries(id) : ["webhooks", "deliveries", "_"],
    queryFn: async () => {
      const res = await webhookService.deliveries(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch deliveries failed");
      return res.data ?? [];
    },
  });
}

export function useCreateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: WebhookCreate) => {
      const res = await webhookService.create(body);
      if (!res.success) throw new Error(res.error?.message ?? "create failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useUpdateWebhook(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: WebhookUpdate) => {
      const res = await webhookService.update(id, body);
      if (!res.success) throw new Error(res.error?.message ?? "update failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}

export function useDeleteWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await webhookService.remove(id);
      if (!res.success) throw new Error(res.error?.message ?? "delete failed");
    },
    onSuccess: () => bust(qc),
  });
}

export function useTestFireWebhook(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: WebhookTestFire = {}) => {
      const res = await webhookService.testFire(id, body);
      if (!res.success) throw new Error(res.error?.message ?? "test fire failed");
      return res.data!;
    },
    onSuccess: () => bust(qc),
  });
}
