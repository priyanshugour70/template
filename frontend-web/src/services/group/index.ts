import { api } from "@/lib/client";
import type {
  Group,
  GroupAddMember,
  GroupAssignRoles,
  GroupCreate,
  GroupMember,
  GroupUpdate,
} from "@/types/group";

// Backend list endpoints are now paginated (default 25, max 200). Pass
// limit=200 to preserve current behaviour; callers can override via params.
export const groupService = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get<Group[]>("/groups", { query: { limit: 200, ...params } }),
  get: (id: string) => api.get<Group>(`/groups/${id}`),
  create: (body: GroupCreate) => api.post<Group>("/groups", body),
  update: (id: string, body: GroupUpdate) => api.patch<Group>(`/groups/${id}`, body),
  remove: (id: string) => api.delete<null>(`/groups/${id}`),
  listMembers: (id: string, params?: { page?: number; limit?: number }) =>
    api.get<GroupMember[]>(`/groups/${id}/members`, { query: { limit: 200, ...params } }),
  addMember: (id: string, body: GroupAddMember) => api.post<null>(`/groups/${id}/members`, body),
  removeMember: (id: string, memberId: string) =>
    api.delete<null>(`/groups/${id}/members/${memberId}`),
  listRoles: (id: string) => api.get<string[]>(`/groups/${id}/roles`),
  assignRoles: (id: string, body: GroupAssignRoles) =>
    api.put<null>(`/groups/${id}/roles`, body),
};
