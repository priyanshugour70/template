"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { rbacService } from "@/services/rbac";
import type {
  AssignRolesRequest,
  CreateRoleRequest,
  UpdateRoleRequest,
} from "@/types/rbac";

const KEY = {
  permissions: ["rbac", "permissions"] as const,
  roles: ["rbac", "roles"] as const,
  role: (id: string) => ["rbac", "role", id] as const,
  rolePerms: (id: string) => ["rbac", "role", id, "perms"] as const,
  membershipRoles: (mid: string) => ["rbac", "membership", mid, "roles"] as const,
};

export function usePermissionsCatalog() {
  return useQuery({
    queryKey: KEY.permissions,
    queryFn: async () => {
      const res = await rbacService.listPermissions();
      if (!res.success) throw new Error(res.error?.message ?? "permissions failed");
      return res.data ?? [];
    },
  });
}

export function useRoles() {
  return useQuery({
    queryKey: KEY.roles,
    queryFn: async () => {
      const res = await rbacService.listRoles();
      if (!res.success) throw new Error(res.error?.message ?? "roles failed");
      return res.data ?? [];
    },
  });
}

export function useRole(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: KEY.role(id ?? ""),
    queryFn: async () => {
      const res = await rbacService.getRole(id!);
      if (!res.success) throw new Error(res.error?.message ?? "role failed");
      return res.data!;
    },
  });
}

export function useRolePermissions(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: KEY.rolePerms(id ?? ""),
    queryFn: async () => {
      const res = await rbacService.listRolePermissions(id!);
      if (!res.success) throw new Error(res.error?.message ?? "role perms failed");
      return res.data ?? [];
    },
  });
}

export function useCreateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateRoleRequest) => {
      const res = await rbacService.createRole(req);
      if (!res.success) throw new Error(res.error?.message ?? "create role failed");
      return res.data!;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY.roles }),
  });
}

export function useUpdateRole(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateRoleRequest) => {
      const res = await rbacService.updateRole(id, req);
      if (!res.success) throw new Error(res.error?.message ?? "update role failed");
      return res.data!;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY.role(id) });
      qc.invalidateQueries({ queryKey: KEY.roles });
    },
  });
}

export function useArchiveRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await rbacService.archiveRole(id);
      if (!res.success) throw new Error(res.error?.message ?? "archive failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY.roles }),
  });
}

export function useAssignRoles(membershipId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: AssignRolesRequest) => {
      const res = await rbacService.assignRolesToMembership(membershipId, req);
      if (!res.success) throw new Error(res.error?.message ?? "assign roles failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY.membershipRoles(membershipId) }),
  });
}
