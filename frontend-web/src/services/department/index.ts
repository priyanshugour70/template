import { api } from "@/lib/client";
import type {
  Department,
  DepartmentAssignRoles,
  DepartmentCreate,
  DepartmentMove,
  DepartmentNode,
  DepartmentUpdate,
} from "@/types/department";

export const departmentService = {
  list: () => api.get<Department[]>("/departments"),
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
