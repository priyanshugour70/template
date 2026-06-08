"use client";

import { createContext, useContext, type ReactNode } from "react";

import { useSessionStore } from "@/stores/session/session.store";
import type {
  SessionOrganization,
  SessionTenant,
} from "@/lib/cookies/types";

interface TenantContextValue {
  tenant: SessionTenant | null;
  activeOrganization: SessionOrganization | null;
  organizations: SessionOrganization[];
}

const TenantContext = createContext<TenantContextValue>({
  tenant: null,
  activeOrganization: null,
  organizations: [],
});

/**
 * Read-only view over the session store's tenant slice. Components that need
 * to change the active org should call authService.switchOrg() directly.
 */
export function TenantProvider({ children }: { children: ReactNode }) {
  const tenant = useSessionStore((s) => s.tenant);
  const activeOrganization = useSessionStore((s) => s.activeOrganization);
  const organizations = useSessionStore((s) => s.organizations);

  return (
    <TenantContext.Provider value={{ tenant, activeOrganization, organizations }}>
      {children}
    </TenantContext.Provider>
  );
}

export function useTenant() {
  return useContext(TenantContext);
}
