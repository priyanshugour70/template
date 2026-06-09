import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

export interface APIKey extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  userId?: ID;
  name: string;
  prefix: string;
  scopes: string[];
  rateLimitRpm?: number;
  lastUsedAt?: ISODate;
  lastUsedIp?: string;
  expiresAt?: ISODate;
  revokedAt?: ISODate;
  revokedBy?: ID;
  metadata: JSONObject;
}

export interface APIKeyCreate {
  name: string;
  scopes?: string[];
  rateLimitRpm?: number;
  expiresAt?: ISODate;
}

export interface APIKeyCreateResponse {
  apiKey: APIKey;
  /** Plaintext token — shown to the user exactly once. */
  token: string;
}
