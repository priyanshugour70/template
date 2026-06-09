import { api } from "@/lib/client";
import type {
  Notification,
  NotificationListQuery,
  UnreadCountResponse,
} from "@/types/notification";

export const notificationService = {
  list: (q: NotificationListQuery = {}) =>
    api.get<Notification[]>("/notifications", { query: q }),
  unreadCount: () => api.get<UnreadCountResponse>("/notifications/unread-count"),
  markRead: (id: string) => api.post<null>(`/notifications/${id}/read`),
  markAllRead: () => api.post<{ updated: number }>("/notifications/mark-all-read"),
};
