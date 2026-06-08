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
