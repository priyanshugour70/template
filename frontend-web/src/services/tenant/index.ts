import { api } from "@/lib/client";
import type {
  CreateOrganizationRequest,
  CreateTenantRequest,
  Organization,
  OrganizationListQuery,
  Tenant,
  TenantListQuery,
  UpdateOrganizationRequest,
  UpdateTenantRequest,
} from "@/types/tenant";

export const tenantService = {
  // ── tenants (admin scope) ────────────────────────────────────────────────
  list: (q: TenantListQuery = {}) => api.get<Tenant[]>("/tenants", { query: q }),
  create: (req: CreateTenantRequest) =>
    api.post<{ tenant: Tenant; defaultOrganization: Organization }>("/tenants", req),

  // ── current tenant ───────────────────────────────────────────────────────
  getMine: () => api.get<Tenant>("/tenants/me"),
  updateMine: (req: UpdateTenantRequest) => api.patch<Tenant>("/tenants/me", req),
  archiveMine: () => api.delete<unknown>("/tenants/me"),

  // ── organizations under current tenant ───────────────────────────────────
  listOrganizations: (q: OrganizationListQuery = {}) =>
    api.get<Organization[]>("/tenants/me/organizations", { query: q }),
  createOrganization: (req: CreateOrganizationRequest) =>
    api.post<Organization>("/tenants/me/organizations", req),
  getOrganization: (id: string) => api.get<Organization>(`/organizations/${id}`),
  updateOrganization: (id: string, req: UpdateOrganizationRequest) =>
    api.patch<Organization>(`/organizations/${id}`, req),
  archiveOrganization: (id: string) => api.delete<unknown>(`/organizations/${id}`),
};
