import { api } from "@/lib/client";
import type {
  AuditListQuery,
  AuditLog,
  AuditStatsFilter,
  AuditStatsSummary,
  AuditTimeInterval,
  AuditTimeseriesBucket,
  AuditTopRow,
} from "@/types/audit";

export const auditService = {
  list: (q: AuditListQuery = {}) => api.get<AuditLog[]>("/audit-logs", { query: q }),
  get: (id: string) => api.get<AuditLog>(`/audit-logs/${id}`),

  // dashboard aggregations
  stats: (f: AuditStatsFilter = {}) => api.get<AuditStatsSummary>("/audit-logs/stats", { query: f }),
  timeseries: (f: AuditStatsFilter & { interval?: AuditTimeInterval } = {}) =>
    api.get<AuditTimeseriesBucket[]>("/audit-logs/timeseries", { query: f }),
  topUsers: (f: AuditStatsFilter & { limit?: number } = {}) =>
    api.get<AuditTopRow[]>("/audit-logs/top/users", { query: f }),
  topFailingPaths: (f: AuditStatsFilter & { limit?: number } = {}) =>
    api.get<AuditTopRow[]>("/audit-logs/top/failing-paths", { query: f }),
  topActions: (f: AuditStatsFilter & { limit?: number } = {}) =>
    api.get<AuditTopRow[]>("/audit-logs/top/actions", { query: f }),
  statusBreakdown: (f: AuditStatsFilter = {}) =>
    api.get<AuditTopRow[]>("/audit-logs/status-breakdown", { query: f }),

  // CSV export — return the absolute path so the browser opens the proxy
  // route which adds auth cookies, instead of going via the JSON client.
  exportUrl: (f: AuditListQuery = {}) => {
    const qs = new URLSearchParams();
    for (const [k, v] of Object.entries(f)) {
      if (v === undefined || v === null || v === "") continue;
      qs.append(k, String(v));
    }
    const tail = qs.toString();
    return `/api/v1/audit-logs/export.csv${tail ? `?${tail}` : ""}`;
  },
};
