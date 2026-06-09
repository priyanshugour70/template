"use client";

import { useQuery } from "@tanstack/react-query";

import { dashboardService } from "@/services/dashboard";

// 60s stale time is the sweet spot for the home page: short enough that
// returning users see fresh numbers, long enough that nav back-and-forth
// inside the dashboard doesn't keep refetching. Refetch on window focus is
// disabled — the user actively pulling fresh data via window focus would
// fire 8 SQL queries per visit, which isn't worth the tiny freshness win.
export function useDashboardSummary() {
  return useQuery({
    queryKey: ["dashboard", "summary"],
    queryFn: async () => {
      const res = await dashboardService.getSummary();
      if (!res.success) throw new Error(res.error?.message ?? "dashboard fetch failed");
      return res.data!;
    },
    staleTime: 60_000,
    refetchOnWindowFocus: false,
  });
}
