"use client";

/**
 * Active-session client store. Hydrates from session-display cookies on first
 * mount so the dashboard sidebar / header have tenant brand information
 * before any network call returns.
 *
 * Use this for UI state only — the source of truth for auth is the HttpOnly
 * cookies, exercised on every API call by the /api/v1 proxy.
 */

import { create } from "zustand";

import { getSessionSnapshot } from "@/lib/cookies";
import type {
  SessionOrganization,
  SessionTenant,
  SessionUser,
} from "@/lib/cookies/types";

interface SessionState {
  user: SessionUser | null;
  tenant: SessionTenant | null;
  activeOrganization: SessionOrganization | null;
  organizations: SessionOrganization[];
  permissions: string[];
  roles: string[];
  hydrated: boolean;
  hydrate: () => void;
  setUser: (user: SessionUser | null) => void;
  setTenant: (tenant: SessionTenant | null) => void;
  setActiveOrganization: (org: SessionOrganization | null) => void;
  setOrganizations: (orgs: SessionOrganization[]) => void;
  setPermissions: (perms: string[]) => void;
  setRoles: (roles: string[]) => void;
  clear: () => void;
}

const initial = {
  user: null,
  tenant: null,
  activeOrganization: null,
  organizations: [] as SessionOrganization[],
  permissions: [] as string[],
  roles: [] as string[],
};

export const useSessionStore = create<SessionState>((set) => ({
  ...initial,
  hydrated: false,

  hydrate: () => {
    if (typeof window === "undefined") return;
    const snap = getSessionSnapshot();
    set({
      user: snap.user,
      tenant: snap.tenant,
      activeOrganization: snap.activeOrganization,
      organizations: snap.organizations,
      permissions: snap.permissions,
      roles: snap.roles,
      hydrated: true,
    });
  },

  setUser: (user) => set({ user }),
  setTenant: (tenant) => set({ tenant }),
  setActiveOrganization: (activeOrganization) => set({ activeOrganization }),
  setOrganizations: (organizations) => set({ organizations }),
  setPermissions: (permissions) => set({ permissions }),
  setRoles: (roles) => set({ roles }),

  clear: () => set({ ...initial, hydrated: true }),
}));
