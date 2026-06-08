"use client";

import { useQuery } from "@tanstack/react-query";

import { auditService } from "@/services/audit";
import type { AuditListQuery } from "@/types/audit";

export function useAuditLogs(q: AuditListQuery = {}) {
  return useQuery({
    queryKey: ["audit", "list", q],
    queryFn: async () => {
      const res = await auditService.list(q);
      if (!res.success) throw new Error(res.error?.message ?? "audit list failed");
      return res.data ?? [];
    },
  });
}

export function useAuditLog(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: ["audit", "one", id],
    queryFn: async () => {
      const res = await auditService.get(id!);
      if (!res.success) throw new Error(res.error?.message ?? "audit fetch failed");
      return res.data!;
    },
  });
}
