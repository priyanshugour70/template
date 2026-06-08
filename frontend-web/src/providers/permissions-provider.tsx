"use client";

import { createContext, useContext, useMemo, type ReactNode } from "react";

import { useAuth } from "./auth-provider";

interface PermissionsContextValue {
  permissions: Set<string>;
  has: (permission: string) => boolean;
  hasAny: (permissions: string[]) => boolean;
  hasAll: (permissions: string[]) => boolean;
}

const PermissionsContext = createContext<PermissionsContextValue>({
  permissions: new Set(),
  has: () => false,
  hasAny: () => false,
  hasAll: () => false,
});

export function PermissionsProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();

  const value = useMemo<PermissionsContextValue>(() => {
    const set = new Set<string>(user?.permissions ?? []);
    return {
      permissions: set,
      has: (p) => set.has(p),
      hasAny: (ps) => ps.some((p) => set.has(p)),
      hasAll: (ps) => ps.every((p) => set.has(p)),
    };
  }, [user]);

  return <PermissionsContext.Provider value={value}>{children}</PermissionsContext.Provider>;
}

export function usePermissions() {
  return useContext(PermissionsContext);
}
