"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { tenantService } from "@/services/tenant";
import type {
  CreateOrganizationRequest,
  OrganizationListQuery,
  UpdateOrganizationRequest,
  UpdateTenantRequest,
} from "@/types/tenant";

const KEY = {
  mine: ["tenant", "mine"] as const,
  orgs: (q: OrganizationListQuery) => ["tenant", "orgs", q] as const,
  org: (id: string) => ["tenant", "org", id] as const,
};

export function useMyTenant() {
  return useQuery({
    queryKey: KEY.mine,
    queryFn: async () => {
      const res = await tenantService.getMine();
      if (!res.success) throw new Error(res.error?.message ?? "fetch tenant failed");
      return res.data!;
    },
  });
}

export function useUpdateMyTenant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateTenantRequest) => {
      const res = await tenantService.updateMine(req);
      if (!res.success) throw new Error(res.error?.message ?? "update tenant failed");
      return res.data!;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY.mine }),
  });
}

export function useOrganizations(q: OrganizationListQuery = {}) {
  return useQuery({
    queryKey: KEY.orgs(q),
    queryFn: async () => {
      const res = await tenantService.listOrganizations(q);
      if (!res.success) throw new Error(res.error?.message ?? "list orgs failed");
      return res.data ?? [];
    },
  });
}

export function useOrganization(id?: string) {
  return useQuery({
    enabled: !!id,
    queryKey: KEY.org(id ?? ""),
    queryFn: async () => {
      const res = await tenantService.getOrganization(id!);
      if (!res.success) throw new Error(res.error?.message ?? "fetch org failed");
      return res.data!;
    },
  });
}

export function useCreateOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateOrganizationRequest) => {
      const res = await tenantService.createOrganization(req);
      if (!res.success) throw new Error(res.error?.message ?? "create org failed");
      return res.data!;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tenant", "orgs"] }),
  });
}

export function useUpdateOrganization(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateOrganizationRequest) => {
      const res = await tenantService.updateOrganization(id, req);
      if (!res.success) throw new Error(res.error?.message ?? "update org failed");
      return res.data!;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY.org(id) });
      qc.invalidateQueries({ queryKey: ["tenant", "orgs"] });
    },
  });
}

export function useArchiveOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await tenantService.archiveOrganization(id);
      if (!res.success) throw new Error(res.error?.message ?? "archive org failed");
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tenant", "orgs"] }),
  });
}
