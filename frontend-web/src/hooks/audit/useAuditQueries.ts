"use client";

import { useQuery } from "@tanstack/react-query";

import { auditService } from "@/services/audit";
import type {
  AuditListQuery,
  AuditStatsFilter,
  AuditTimeInterval,
} from "@/types/audit";

export function useAuditLogs(q: AuditListQuery = {}) {
  return useQuery({
    queryKey: ["audit", "list", q],
    queryFn: async () => {
      const res = await auditService.list(q);
      if (!res.success) throw new Error(res.error?.message ?? "audit list failed");
      return {
        items: res.data ?? [],
        total: res.pagination?.total ?? (res.data?.length ?? 0),
        page: res.pagination?.page ?? 1,
        limit: res.pagination?.limit ?? (q.limit ?? 25),
      };
    },
    placeholderData: (prev) => prev,
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

export function useAuditStats(f: AuditStatsFilter = {}) {
  return useQuery({
    queryKey: ["audit", "stats", f],
    queryFn: async () => {
      const res = await auditService.stats(f);
      if (!res.success) throw new Error(res.error?.message ?? "stats failed");
      return res.data!;
    },
    refetchInterval: 60_000,
  });
}

export function useAuditTimeseries(f: AuditStatsFilter & { interval?: AuditTimeInterval } = {}) {
  return useQuery({
    queryKey: ["audit", "timeseries", f],
    queryFn: async () => {
      const res = await auditService.timeseries(f);
      if (!res.success) throw new Error(res.error?.message ?? "timeseries failed");
      return res.data ?? [];
    },
    refetchInterval: 60_000,
  });
}

export function useAuditTopUsers(f: AuditStatsFilter = {}, limit = 10) {
  return useQuery({
    queryKey: ["audit", "top-users", f, limit],
    queryFn: async () => {
      const res = await auditService.topUsers({ ...f, limit });
      if (!res.success) throw new Error(res.error?.message ?? "top users failed");
      return res.data ?? [];
    },
  });
}

export function useAuditTopFailingPaths(f: AuditStatsFilter = {}, limit = 10) {
  return useQuery({
    queryKey: ["audit", "top-failing-paths", f, limit],
    queryFn: async () => {
      const res = await auditService.topFailingPaths({ ...f, limit });
      if (!res.success) throw new Error(res.error?.message ?? "top failing paths failed");
      return res.data ?? [];
    },
  });
}

export function useAuditTopActions(f: AuditStatsFilter = {}, limit = 10) {
  return useQuery({
    queryKey: ["audit", "top-actions", f, limit],
    queryFn: async () => {
      const res = await auditService.topActions({ ...f, limit });
      if (!res.success) throw new Error(res.error?.message ?? "top actions failed");
      return res.data ?? [];
    },
  });
}

export function useAuditStatusBreakdown(f: AuditStatsFilter = {}) {
  return useQuery({
    queryKey: ["audit", "status-breakdown", f],
    queryFn: async () => {
      const res = await auditService.statusBreakdown(f);
      if (!res.success) throw new Error(res.error?.message ?? "status breakdown failed");
      return res.data ?? [];
    },
  });
}
