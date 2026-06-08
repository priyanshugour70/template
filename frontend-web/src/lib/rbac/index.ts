/**
 * Tiny RBAC helpers shared by middleware and the PermissionsProvider.
 * The authoritative permission list comes from the backend's /auth/me response.
 */

export function hasPermission(set: ReadonlySet<string>, permission: string): boolean {
  return set.has(permission);
}

export function hasAnyPermission(set: ReadonlySet<string>, permissions: readonly string[]): boolean {
  return permissions.some((p) => set.has(p));
}

export function hasAllPermissions(set: ReadonlySet<string>, permissions: readonly string[]): boolean {
  return permissions.every((p) => set.has(p));
}
