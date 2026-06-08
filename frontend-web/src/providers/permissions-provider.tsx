"use client";

import { createContext, useContext, useMemo, type ReactNode } from "react";

import { useSessionStore } from "@/stores/session/session.store";

interface PermissionsContextValue {
  permissions: Set<string>;
  roles: Set<string>;
  isSuperAdmin: boolean;
  has: (permission: string) => boolean;
  hasAny: (permissions: string[]) => boolean;
  hasAll: (permissions: string[]) => boolean;
  hasRole: (roleKey: string) => boolean;
}

const PermissionsContext = createContext<PermissionsContextValue>({
  permissions: new Set(),
  roles: new Set(),
  isSuperAdmin: false,
  has: () => false,
  hasAny: () => false,
  hasAll: () => false,
  hasRole: () => false,
});

/**
 * Permissions/roles are read from the session-store, which hydrates from the
 * session-display cookies (set by /api/auth/login). Super-admin bypasses all
 * checks.
 */
export function PermissionsProvider({ children }: { children: ReactNode }) {
  const permissions = useSessionStore((s) => s.permissions);
  const roles = useSessionStore((s) => s.roles);
  const isSuperAdmin = useSessionStore((s) => s.user?.isSuperAdmin ?? false);

  const value = useMemo<PermissionsContextValue>(() => {
    const permSet = new Set(permissions);
    const roleSet = new Set(roles);
    return {
      permissions: permSet,
      roles: roleSet,
      isSuperAdmin,
      has: (p) => isSuperAdmin || permSet.has(p),
      hasAny: (ps) => isSuperAdmin || ps.some((p) => permSet.has(p)),
      hasAll: (ps) => isSuperAdmin || ps.every((p) => permSet.has(p)),
      hasRole: (r) => roleSet.has(r),
    };
  }, [permissions, roles, isSuperAdmin]);

  return <PermissionsContext.Provider value={value}>{children}</PermissionsContext.Provider>;
}

export function usePermissions() {
  return useContext(PermissionsContext);
}
