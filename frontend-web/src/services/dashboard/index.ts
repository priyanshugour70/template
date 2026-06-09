import { api } from "@/lib/client";
import type { DashboardSummary } from "@/types/dashboard";

export const dashboardService = {
  getSummary: () => api.get<DashboardSummary>("/dashboard/summary"),
};
