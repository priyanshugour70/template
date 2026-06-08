import { api } from "@/lib/client";
import type { AuditListQuery, AuditLog } from "@/types/audit";

export const auditService = {
  list: (q: AuditListQuery = {}) => api.get<AuditLog[]>("/audit-logs", { query: q }),
  get: (id: string) => api.get<AuditLog>(`/audit-logs/${id}`),
};
