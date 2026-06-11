"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { userService } from "@/services/user";
import type {
  BulkUpdateMembershipsRequest,
  InviteUserRequest,
  UpdateMembershipRequest,
  UpdateUserRequest,
  UserListQuery,
} from "@/types/user";

const KEY = {
  list: (q: UserListQuery) => ["users", "list", q] as const,
  one: (id: string) => ["users", "one", id] as const,
  memberships: ["users", "memberships", "me"] as const,
  userMemberships: (id: string) => ["users", "memberships", id] as const,
  permissions: (id: string) => ["users", "permissions", id] as const,
};

export function useUsers(q: UserListQuery = {}) {
  return useQuery({
    queryKey: KEY.list(q),
    queryFn: async () => {
      const res = await userService.list(q);
      if (!res.success) throw new Error(res.error?.message ?? "list users failed");
      return {
        items: res.data ?? [],
        total: res.pagination?.total ?? (res.data?.length ?? 0),
        page: res.pagination?.page ?? 1,
        limit: res.pagination?.limit ?? (q.limit ?? 25),
      };
    },
    placeholderData: (prev) => prev,
  });
}

export function useUser(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: KEY.one(id ?? ""),
    queryFn: async () => {
      const res = await userService.get(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch user failed");
      return res.data!;
    },
  });
}

export function useUpdateUser(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateUserRequest) => {
      const res = await userService.update(id, req);
      if (!res.success) throw new Error(res.error?.message ?? "update user failed");
      return res.data!;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY.one(id) });
      qc.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });
}

/**
 * Self-update via PATCH /users/me. Unlike useUpdateUser this hits the
 * permission-free `me` endpoint — any authed user can edit their own
 * profile, no `user.update` RBAC needed. Use this for onboarding profile
 * step, settings/profile page, and anywhere a user edits themselves.
 */
export function useUpdateMe() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateUserRequest) => {
      const res = await userService.updateMe(req);
      if (!res.success) throw new Error(res.error?.message ?? "update profile failed");
      return res.data!;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useSuspendUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.suspend(id);
      if (!res.success) throw new Error(res.error?.message ?? "suspend failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useReactivateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.reactivate(id);
      if (!res.success) throw new Error(res.error?.message ?? "reactivate failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useArchiveUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.archive(id);
      if (!res.success) throw new Error(res.error?.message ?? "archive failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useInviteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: InviteUserRequest) => {
      const res = await userService.invite(req);
      if (!res.success) throw new Error(res.error?.message ?? "invite failed");
      return res.data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useMyMemberships() {
  return useQuery({
    queryKey: KEY.memberships,
    queryFn: async () => {
      const res = await userService.myMemberships();
      if (!res.success) throw new Error(res.error?.message ?? "memberships failed");
      return res.data ?? [];
    },
  });
}

// ── admin actions ────────────────────────────────────────────────────────

export function useUserMemberships(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.userMemberships(id) : ["users", "memberships", "_"],
    queryFn: async () => {
      const res = await userService.listMemberships(id!);
      if (!res.success) throw new Error(res.error?.message ?? "memberships failed");
      return res.data ?? [];
    },
  });
}

export function useEffectivePermissions(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? KEY.permissions(id) : ["users", "permissions", "_"],
    queryFn: async () => {
      const res = await userService.effectivePermissions(id!);
      if (!res.success) throw new Error(res.error?.message ?? "permissions failed");
      return res.data?.permissions ?? [];
    },
  });
}

export function useForcePasswordReset() {
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.forcePasswordReset(id);
      if (!res.success) throw new Error(res.error?.message ?? "force reset failed");
    },
  });
}

export function useResetMFA() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.resetMFA(id);
      if (!res.success) throw new Error(res.error?.message ?? "reset MFA failed");
    },
    onSuccess: (_d, id) => {
      void qc.invalidateQueries({ queryKey: KEY.one(id) });
      void qc.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });
}

export function useUnlockUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await userService.unlock(id);
      if (!res.success) throw new Error(res.error?.message ?? "unlock failed");
    },
    onSuccess: (_d, id) => {
      void qc.invalidateQueries({ queryKey: KEY.one(id) });
      void qc.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });
}

export function useUpdateMembership(userId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (args: { membershipId: string; patch: UpdateMembershipRequest }) => {
      const res = await userService.updateMembership(userId, args.membershipId, args.patch);
      if (!res.success) throw new Error(res.error?.message ?? "update membership failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.userMemberships(userId) });
      void qc.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });
}

export function useBulkUpdateMemberships() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: BulkUpdateMembershipsRequest) => {
      const res = await userService.bulkUpdateMemberships(req);
      if (!res.success) throw new Error(res.error?.message ?? "bulk update failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
