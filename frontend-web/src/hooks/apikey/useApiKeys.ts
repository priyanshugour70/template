"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiKeyService } from "@/services/apikey";
import type { APIKeyCreate } from "@/types/apikey";

const KEY = ["apikeys", "list"] as const;

type PageOpts = { page?: number; limit?: number };

export function useApiKeys(opts: PageOpts = {}) {
  return useQuery({
    queryKey: [...KEY, opts] as const,
    queryFn: async () => {
      const res = await apiKeyService.list(opts);
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

export function useCreateApiKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: APIKeyCreate) => {
      const res = await apiKeyService.create(body);
      if (!res.success) throw new Error(res.error?.message ?? "create failed");
      return res.data!;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useRevokeApiKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await apiKeyService.revoke(id);
      if (!res.success) throw new Error(res.error?.message ?? "revoke failed");
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: KEY }),
  });
}
