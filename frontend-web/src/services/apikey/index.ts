import { api } from "@/lib/client";
import type { APIKey, APIKeyCreate, APIKeyCreateResponse } from "@/types/apikey";

// Backend list is paginated (default 25, max 200). Default to limit=200.
export const apiKeyService = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get<APIKey[]>("/api-keys", { query: { limit: 200, ...params } }),
  create: (body: APIKeyCreate) => api.post<APIKeyCreateResponse>("/api-keys", body),
  revoke: (id: string) => api.delete<null>(`/api-keys/${id}`),
};
