import type { BaseEntity, ID, JSONObject } from "@/types/common";

export interface Department extends BaseEntity {
  tenantId: ID;
  organizationId?: ID | null;
  parentId?: ID | null;
  slug: string;
  name: string;
  description?: string;
  costCenter?: string;
  managerUserId?: ID | null;
  color?: string;
  icon?: string;
  isArchived: boolean;
  sortOrder: number;
  metadata?: JSONObject;
}

export interface DepartmentNode extends Department {
  children?: DepartmentNode[];
}

export interface DepartmentCreate {
  parentId?: ID | null;
  slug: string;
  name: string;
  description?: string;
  costCenter?: string;
  managerUserId?: ID | null;
  color?: string;
  icon?: string;
  sortOrder?: number;
}

export interface DepartmentUpdate {
  name?: string;
  description?: string;
  costCenter?: string;
  managerUserId?: ID | null;
  color?: string;
  icon?: string;
  isArchived?: boolean;
  sortOrder?: number;
}

export interface DepartmentMove {
  parentId: ID | null;
}

export interface DepartmentAssignRoles {
  roleIds: ID[];
}
