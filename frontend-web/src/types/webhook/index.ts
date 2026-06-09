import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

export type WebhookDeliveryStatus = "pending" | "success" | "failed" | "dropped";

export interface Webhook extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  name: string;
  url: string;
  events: string[];
  isActive: boolean;
  description?: string;
  headers: Record<string, string>;
  lastInvokedAt?: ISODate;
  lastStatus?: number;
  consecutiveFailures: number;
  disabledAt?: ISODate;
  disabledReason?: string;
  metadata: JSONObject;
}

export interface WebhookDelivery {
  id: ID;
  webhookId: ID;
  event: string;
  payload: unknown;
  status: WebhookDeliveryStatus;
  attempt: number;
  responseStatus?: number;
  responseBody?: string;
  errorMessage?: string;
  durationMs?: number;
  deliveredAt?: ISODate;
  nextRetryAt?: ISODate;
  createdAt: ISODate;
}

export interface WebhookCreate {
  name: string;
  url: string;
  events?: string[];
  description?: string;
  headers?: Record<string, string>;
}

export interface WebhookCreateResponse {
  webhook: Webhook;
  /** Plaintext HMAC secret — shown to the user exactly once. */
  secret: string;
}

export interface WebhookUpdate {
  name?: string;
  url?: string;
  events?: string[];
  description?: string;
  headers?: Record<string, string>;
  isActive?: boolean;
}

export interface WebhookTestFire {
  event?: string;
  payload?: Record<string, unknown>;
}

export interface WebhookTestFireResponse {
  delivery: WebhookDelivery;
}
