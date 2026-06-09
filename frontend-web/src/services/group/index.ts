import { api } from "@/lib/client";
import type {
  Group,
  GroupAddMember,
  GroupAssignRoles,
  GroupCreate,
  GroupMember,
  GroupUpdate,
} from "@/types/group";

export const groupService = {
  list: () => api.get<Group[]>("/groups"),
  get: (id: string) => api.get<Group>(`/groups/${id}`),
  create: (body: GroupCreate) => api.post<Group>("/groups", body),
  update: (id: string, body: GroupUpdate) => api.patch<Group>(`/groups/${id}`, body),
  remove: (id: string) => api.delete<null>(`/groups/${id}`),
  listMembers: (id: string) => api.get<GroupMember[]>(`/groups/${id}/members`),
  addMember: (id: string, body: GroupAddMember) => api.post<null>(`/groups/${id}/members`, body),
  removeMember: (id: string, memberId: string) =>
    api.delete<null>(`/groups/${id}/members/${memberId}`),
  listRoles: (id: string) => api.get<string[]>(`/groups/${id}/roles`),
  assignRoles: (id: string, body: GroupAssignRoles) =>
    api.put<null>(`/groups/${id}/roles`, body),
};
