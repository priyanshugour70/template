"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { userService } from "@/services/user";
import type {
  InviteUserRequest,
  UpdateUserRequest,
  UserListQuery,
} from "@/types/user";

const KEY = {
  list: (q: UserListQuery) => ["users", "list", q] as const,
  one: (id: string) => ["users", "one", id] as const,
  memberships: ["users", "memberships", "me"] as const,
};

export function useUsers(q: UserListQuery = {}) {
  return useQuery({
    queryKey: KEY.list(q),
    queryFn: async () => {
      const res = await userService.list(q);
      if (!res.success) throw new Error(res.error?.message ?? "list users failed");
      return res.data ?? [];
    },
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
