import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

export type GroupKind = "custom" | "dynamic" | "system";

export interface Group extends BaseEntity {
  tenantId: ID;
  organizationId?: ID | null;
  slug: string;
  name: string;
  description?: string;
  kind: GroupKind;
  color?: string;
  icon?: string;
  isArchived: boolean;
  rule?: JSONObject | null;
  metadata?: JSONObject;
}

export interface GroupMember {
  id: ID;
  groupId: ID;
  memberUserId?: ID | null;
  memberGroupId?: ID | null;
  addedAt: ISODate;
  addedBy?: ID | null;
}

export interface GroupCreate {
  slug: string;
  name: string;
  description?: string;
  kind?: GroupKind;
  color?: string;
  icon?: string;
}

export interface GroupUpdate {
  name?: string;
  description?: string;
  color?: string;
  icon?: string;
  isArchived?: boolean;
}

export interface GroupAddMember {
  userId?: ID;
  groupId?: ID;
}

export interface GroupAssignRoles {
  roleIds: ID[];
}
