import { api } from "@/lib/client";
import type { APIKey, APIKeyCreate, APIKeyCreateResponse } from "@/types/apikey";

export const apiKeyService = {
  list: () => api.get<APIKey[]>("/api-keys"),
  create: (body: APIKeyCreate) => api.post<APIKeyCreateResponse>("/api-keys", body),
  revoke: (id: string) => api.delete<null>(`/api-keys/${id}`),
};
