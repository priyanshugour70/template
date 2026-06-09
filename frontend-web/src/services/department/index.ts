import { api } from "@/lib/client";
import type {
  Department,
  DepartmentAssignRoles,
  DepartmentCreate,
  DepartmentMove,
  DepartmentNode,
  DepartmentUpdate,
} from "@/types/department";

// Backend list endpoints are now paginated (default 25, max 200). Pass
// limit=200 by default to preserve current behaviour for small tenants;
// callers that want pagination can pass { page, limit } explicitly.
export const departmentService = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get<Department[]>("/departments", { query: { limit: 200, ...params } }),
  tree: () => api.get<DepartmentNode[]>("/departments/tree"),
  get: (id: string) => api.get<Department>(`/departments/${id}`),
  create: (body: DepartmentCreate) => api.post<Department>("/departments", body),
  update: (id: string, body: DepartmentUpdate) => api.patch<Department>(`/departments/${id}`, body),
  move: (id: string, body: DepartmentMove) => api.post<null>(`/departments/${id}/move`, body),
  remove: (id: string) => api.delete<null>(`/departments/${id}`),
  listRoles: (id: string) => api.get<string[]>(`/departments/${id}/roles`),
  assignRoles: (id: string, body: DepartmentAssignRoles) =>
    api.put<null>(`/departments/${id}/roles`, body),
};
