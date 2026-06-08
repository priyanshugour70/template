import { api } from "@/lib/client";
import type { Sample, SampleListResponse } from "@/types/sample";

export const sampleService = {
  list: (page = 1, limit = 20, search?: string) =>
    api.get<SampleListResponse>("/samples", { query: { page, limit, search } }),
  get: (id: number) => api.get<Sample>(`/samples/${id}`),
  create: (name: string) => api.post<Sample>("/samples", { name }),
  update: (id: number, patch: Partial<Pick<Sample, "name" | "status">>) =>
    api.patch<Sample>(`/samples/${id}`, patch),
  remove: (id: number) => api.delete<{ ok: true }>(`/samples/${id}`),
};
