import { api } from "@/lib/client";
import type { InviteUserRequest } from "@/types/user";
import type {
  Membership,
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

  // memberships
  myMemberships: () => api.get<Membership[]>("/users/me/memberships"),
  suspendMembership: (userId: string, membershipId: string) =>
    api.post<unknown>(`/users/${userId}/memberships/${membershipId}/suspend`),
  archiveMembership: (userId: string, membershipId: string) =>
    api.delete<unknown>(`/users/${userId}/memberships/${membershipId}`),

  // invite (routes to /invites on the backend)
  invite: (req: InviteUserRequest) => api.post<unknown>("/invites", req),
};
