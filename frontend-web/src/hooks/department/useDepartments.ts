"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { departmentService } from "@/services/department";
import type {
  DepartmentAssignRoles,
  DepartmentCreate,
  DepartmentMove,
  DepartmentUpdate,
} from "@/types/department";

const QK = {
  list: ["departments", "list"] as const,
  tree: ["departments", "tree"] as const,
  one: (id: string) => ["departments", "one", id] as const,
  roles: (id: string) => ["departments", "roles", id] as const,
};

type PageOpts = { page?: number; limit?: number };

export function useDepartments(opts: PageOpts = {}) {
  return useQuery({
    queryKey: [...QK.list, opts] as const,
    queryFn: async () => {
      const res = await departmentService.list(opts);
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

export function useDepartmentTree() {
  return useQuery({
    queryKey: QK.tree,
    queryFn: async () => {
      const res = await departmentService.tree();
      if (!res.success) throw new Error(res.error?.message ?? "tree failed");
      return res.data ?? [];
    },
  });
}

export function useDepartment(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? QK.one(id) : ["departments", "one", "_"],
    queryFn: async () => {
      const res = await departmentService.get(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch failed");
      return res.data!;
    },
  });
}

export function useDepartmentRoles(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: id ? QK.roles(id) : ["departments", "roles", "_"],
    queryFn: async () => {
      const res = await departmentService.listRoles(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch failed");
      return res.data ?? [];
    },
  });
}

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: ["departments"] });
}

export function useCreateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: DepartmentCreate) => departmentService.create(body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useUpdateDepartment(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: DepartmentUpdate) => departmentService.update(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useMoveDepartment(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: DepartmentMove) => departmentService.move(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useDeleteDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => departmentService.remove(id),
    onSuccess: () => invalidateAll(qc),
  });
}

export function useAssignDepartmentRoles(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: DepartmentAssignRoles) => departmentService.assignRoles(id, body),
    onSuccess: () => invalidateAll(qc),
  });
}
