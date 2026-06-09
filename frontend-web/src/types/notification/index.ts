import type { BaseEntity, ID, ISODate, JSONObject, PaginatedQuery } from "@/types/common";

export type NotificationKind = "info" | "success" | "warning" | "error";

export interface Notification extends BaseEntity {
  tenantId: ID;
  organizationId?: ID | null;
  userId: ID;
  kind: NotificationKind;
  title: string;
  message?: string;
  link?: string;
  isRead: boolean;
  readAt?: ISODate | null;
  metadata?: JSONObject;
}

export interface NotificationListQuery extends PaginatedQuery {
  unread?: boolean;
}

export interface UnreadCountResponse {
  unreadCount: number;
}
