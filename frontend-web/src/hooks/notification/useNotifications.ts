"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { notificationService } from "@/services/notification";

const KEY_LIST = ["notifications", "list"] as const;
const KEY_UNREAD = ["notifications", "unread"] as const;

export function useUnreadCount(pollMs = 30_000) {
  return useQuery({
    queryKey: KEY_UNREAD,
    queryFn: async () => {
      const res = await notificationService.unreadCount();
      if (!res.success) throw new Error(res.error?.message ?? "unread count failed");
      return res.data?.unreadCount ?? 0;
    },
    refetchInterval: pollMs,
    staleTime: 10_000,
  });
}

export function useNotifications(enabled: boolean) {
  return useQuery({
    enabled,
    queryKey: KEY_LIST,
    queryFn: async () => {
      const res = await notificationService.list({ limit: 10 });
      if (!res.success) throw new Error(res.error?.message ?? "notifications list failed");
      return res.data ?? [];
    },
    staleTime: 10_000,
  });
}

export function useMarkRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => notificationService.markRead(id),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
}

export function useMarkAllRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => notificationService.markAllRead(),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
}
