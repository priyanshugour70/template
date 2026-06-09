import { api } from "@/lib/client";
import type {
  AssignRolesRequest,
  CreateRoleRequest,
  Permission,
  Role,
  UpdateRoleRequest,
} from "@/types/rbac";

// Backend list endpoints are paginated (default 25, max 200). Default to
// limit=200 so existing UI keeps showing all items in typical tenants.
export const rbacService = {
  // ── permissions catalog ──────────────────────────────────────────────────
  listPermissions: (params?: { page?: number; limit?: number }) =>
    api.get<Permission[]>("/permissions", { query: { limit: 200, ...params } }),

  // ── roles ────────────────────────────────────────────────────────────────
  listRoles: (params?: { page?: number; limit?: number }) =>
    api.get<Role[]>("/roles", { query: { limit: 200, ...params } }),
  getRole: (id: string) => api.get<Role>(`/roles/${id}`),
  createRole: (req: CreateRoleRequest) => api.post<Role>("/roles", req),
  updateRole: (id: string, req: UpdateRoleRequest) => api.patch<Role>(`/roles/${id}`, req),
  archiveRole: (id: string) => api.delete<unknown>(`/roles/${id}`),
  listRolePermissions: (id: string, params?: { page?: number; limit?: number }) =>
    api.get<Permission[]>(`/roles/${id}/permissions`, { query: { limit: 200, ...params } }),

  // ── membership ↔ role ────────────────────────────────────────────────────
  listMembershipRoles: (membershipId: string) =>
    api.get<Role[]>(`/memberships/${membershipId}/roles`),
  assignRolesToMembership: (membershipId: string, req: AssignRolesRequest) =>
    api.post<unknown>(`/memberships/${membershipId}/roles`, req),
  removeRoleFromMembership: (membershipId: string, roleId: string) =>
    api.delete<unknown>(`/memberships/${membershipId}/roles/${roleId}`),
};
