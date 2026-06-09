import { api } from "@/lib/client";
import type {
  BulkUpdateMembershipsRequest,
  BulkUpdateMembershipsResponse,
  EffectivePermissionsResponse,
  InviteUserRequest,
  Membership,
  UpdateMembershipRequest,
  UpdateUserRequest,
  UserListQuery,
  UserProfile,
} from "@/types/user";

export const userService = {
  list: (q: UserListQuery = {}) => api.get<UserProfile[]>("/users", { query: q }),
  get: (id: string) => api.get<UserProfile>(`/users/${id}`),
  update: (id: string, req: UpdateUserRequest) => api.patch<UserProfile>(`/users/${id}`, req),
  suspend: (id: string) => api.post<unknown>(`/users/${id}/suspend`),
  reactivate: (id: string) => api.post<unknown>(`/users/${id}/reactivate`),
  archive: (id: string) => api.delete<unknown>(`/users/${id}`),

  // self
  me: () => api.get<UserProfile>("/users/me"),
  updateMe: (req: UpdateUserRequest) => api.patch<UserProfile>("/users/me", req),

  // admin security actions
  forcePasswordReset: (id: string) => api.post<null>(`/users/${id}/force-password-reset`),
  resetMFA: (id: string) => api.post<null>(`/users/${id}/reset-mfa`),
  unlock: (id: string) => api.post<null>(`/users/${id}/unlock`),
  effectivePermissions: (id: string) =>
    api.get<EffectivePermissionsResponse>(`/users/${id}/effective-permissions`),

  // memberships
  myMemberships: () => api.get<Membership[]>("/users/me/memberships"),
  listMemberships: (id: string) => api.get<Membership[]>(`/users/${id}/memberships`),
  updateMembership: (userId: string, membershipId: string, req: UpdateMembershipRequest) =>
    api.patch<Membership>(`/users/${userId}/memberships/${membershipId}`, req),
  suspendMembership: (userId: string, membershipId: string) =>
    api.post<unknown>(`/users/${userId}/memberships/${membershipId}/suspend`),
  archiveMembership: (userId: string, membershipId: string) =>
    api.delete<unknown>(`/users/${userId}/memberships/${membershipId}`),

  // bulk
  bulkUpdateMemberships: (req: BulkUpdateMembershipsRequest) =>
    api.post<BulkUpdateMembershipsResponse>("/users/bulk/memberships", req),

  // invite (routes to /invites on the backend)
  invite: (req: InviteUserRequest) => api.post<unknown>("/invites", req),
};
