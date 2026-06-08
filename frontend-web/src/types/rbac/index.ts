import type { BaseEntity, ID, JSONObject } from "@/types/common";

export interface Permission extends BaseEntity {
  key: string;
  resource: string;
  action: string;
  description?: string;
  category?: string;
  isSystem: boolean;
  isDangerous: boolean;
  metadata: JSONObject;
}

export interface Role extends BaseEntity {
  tenantId?: ID;
  organizationId?: ID;
  key: string;
  name: string;
  description?: string;
  isSystem: boolean;
  isDefault: boolean;
  isAssignable: boolean;
  priority: number;
  color?: string;
  icon?: string;
  metadata: JSONObject;
}

export interface CreateRoleRequest {
  key: string;
  name: string;
  description?: string;
  priority?: number;
  color?: string;
  icon?: string;
  permissionKeys?: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  priority?: number;
  color?: string;
  icon?: string;
  isAssignable?: boolean;
  permissionKeys?: string[];
}

export interface AssignRolesRequest {
  roleKeys: string[];
}
