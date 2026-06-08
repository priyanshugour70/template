import { api } from "@/lib/client";
import type {
  AssignRolesRequest,
  CreateRoleRequest,
  Permission,
  Role,
  UpdateRoleRequest,
} from "@/types/rbac";

export const rbacService = {
  // ── permissions catalog ──────────────────────────────────────────────────
  listPermissions: () => api.get<Permission[]>("/permissions"),

  // ── roles ────────────────────────────────────────────────────────────────
  listRoles: () => api.get<Role[]>("/roles"),
  getRole: (id: string) => api.get<Role>(`/roles/${id}`),
  createRole: (req: CreateRoleRequest) => api.post<Role>("/roles", req),
  updateRole: (id: string, req: UpdateRoleRequest) => api.patch<Role>(`/roles/${id}`, req),
  archiveRole: (id: string) => api.delete<unknown>(`/roles/${id}`),
  listRolePermissions: (id: string) => api.get<Permission[]>(`/roles/${id}/permissions`),

  // ── membership ↔ role ────────────────────────────────────────────────────
  listMembershipRoles: (membershipId: string) =>
    api.get<Role[]>(`/memberships/${membershipId}/roles`),
  assignRolesToMembership: (membershipId: string, req: AssignRolesRequest) =>
    api.post<unknown>(`/memberships/${membershipId}/roles`, req),
  removeRoleFromMembership: (membershipId: string, roleId: string) =>
    api.delete<unknown>(`/memberships/${membershipId}/roles/${roleId}`),
};
