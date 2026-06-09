import { api } from "@/lib/client";
import type {
  Webhook,
  WebhookCreate,
  WebhookCreateResponse,
  WebhookDelivery,
  WebhookTestFire,
  WebhookTestFireResponse,
  WebhookUpdate,
} from "@/types/webhook";

export const webhookService = {
  list: () => api.get<Webhook[]>("/webhooks"),
  get: (id: string) => api.get<Webhook>(`/webhooks/${id}`),
  create: (body: WebhookCreate) => api.post<WebhookCreateResponse>("/webhooks", body),
  update: (id: string, body: WebhookUpdate) => api.patch<Webhook>(`/webhooks/${id}`, body),
  remove: (id: string) => api.delete<null>(`/webhooks/${id}`),
  deliveries: (id: string) => api.get<WebhookDelivery[]>(`/webhooks/${id}/deliveries`),
  testFire: (id: string, body: WebhookTestFire = {}) =>
    api.post<WebhookTestFireResponse>(`/webhooks/${id}/test`, body),
};
