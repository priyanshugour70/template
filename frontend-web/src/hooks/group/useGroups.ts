"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { groupService } from "@/services/group";
import type {
  GroupAddMember,
  GroupAssignRoles,
  GroupCreate,
  GroupUpdate,
} from "@/types/group";

const QK = {
  list: ["groups", "list"] as const,
  one: (id: string) => ["groups", "one", id] as const,
  members: (id: string) => ["groups", "members", id] as const,
  roles: (id: string) => ["groups", "roles", id] as const,
};

type PageOpts = { page?: number; limit?: number };

export function useGroups(opts: PageOpts = {}) {
  return useQuery({
    queryKey: [...QK.list, opts] as const,
    queryFn: async () => {
      const res = await groupService.list(opts);
      if (!res.success) throw new Error(res.error?.message ?? "list failed");
      return {
        items: res.data ?? [],
        total: res.pagination?.total ?? (res.data?.length ?? 0),
        page: res.pagination?.page ?? 1,
        limit: res.pagination?.limit ?? (opts.limit ?? 200),
      };
    },
    placeholderData: (prev) => prev,
  });
}

export function useGroup(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? QK.one(id) : ["groups", "one", "_"],
    queryFn: async () => {
      const res = await groupService.get(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch failed");
      return res.data!;
    },
  });
}

export function useGroupMembers(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? QK.members(id) : ["groups", "members", "_"],
    queryFn: async () => {
      const res = await groupService.listMembers(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch failed");
      return res.data ?? [];
    },
  });
}

export function useGroupRoles(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? QK.roles(id) : ["groups", "roles", "_"],
    queryFn: async () => {
      const res = await groupService.listRoles(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch failed");
      return res.data ?? [];
    },
  });
}

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: ["groups"] });
}

export function useCreateGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: GroupCreate) => groupService.create(body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useUpdateGroup(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: GroupUpdate) => groupService.update(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useDeleteGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => groupService.remove(id),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useAddGroupMember(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: GroupAddMember) => groupService.addMember(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useRemoveGroupMember(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (memberId: string) => groupService.removeMember(id, memberId),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useAssignGroupRoles(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: GroupAssignRoles) => groupService.assignRoles(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}
