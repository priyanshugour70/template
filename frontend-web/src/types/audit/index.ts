import type { ID, ISODate, JSONObject, PaginatedQuery } from "@/types/common";

export interface AuditLog {
  id: ID;
  occurredAt: ISODate;
  correlationId?: string;
  tenantId?: ID;
  organizationId?: ID;
  userId?: ID;
  userEmail?: string;
  method?: string;
  path?: string;
  route?: string;
  statusCode?: number;
  latencyMs?: number;
  ip?: string | null;
  userAgent?: string;
  action?: string;
  targetType?: string;
  targetId?: ID;
  errorCode?: string;
  requestHeaders?: Record<string, string>;
  requestBody?: JSONObject | unknown;
  responseHeaders?: Record<string, string>;
  responseBody?: JSONObject | unknown;
  metadata?: JSONObject;
  createdAt: ISODate;
  createdBy?: ID;
}

export interface AuditListQuery extends PaginatedQuery {
  userId?: ID;
  userEmail?: string;
  action?: string;
  targetType?: string;
  targetId?: ID;
  method?: string;
  path?: string;
  statusFrom?: number;
  statusTo?: number;
  from?: ISODate;
  to?: ISODate;
}

export interface AuditStatsSummary {
  totalRequests: number;
  success2xx: number;
  redirect3xx: number;
  clientError4xx: number;
  serverError5xx: number;
  uniqueUsers: number;
  uniquePaths: number;
  avgLatencyMs: number;
  p95LatencyMs: number;
  errorRatePct: number;
}

export interface AuditTimeseriesBucket {
  bucket: string;
  total: number;
  success2xx: number;
  redirect3xx: number;
  clientError4xx: number;
  serverError5xx: number;
}

export interface AuditTopRow {
  key: string;
  label?: string;
  count: number;
}

export type AuditTimeInterval = "minute" | "hour" | "day" | "week";

export interface AuditStatsFilter {
  userId?: ID;
  userEmail?: string;
  action?: string;
  method?: string;
  path?: string;
  from?: ISODate;
  to?: ISODate;
}
