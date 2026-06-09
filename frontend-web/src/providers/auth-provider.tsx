"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import { getSessionUser, clearSessionDisplayCookies } from "@/lib/cookies";
import { authService } from "@/services/auth";
import { useSessionStore } from "@/stores/session/session.store";
import type { User } from "@/types/auth";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  isAuthenticated: boolean;
  refreshUser: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  loading: true,
  isAuthenticated: false,
  refreshUser: async () => {},
  logout: async () => {},
});

function redirectToLoginAfterAuthFailure() {
  if (typeof window === "undefined") return;
  const { pathname, search } = window.location;
  if (pathname.startsWith("/auth/")) return;
  const currentPath = `${pathname}${search}`;
  const redirect = currentPath ? `?redirect=${encodeURIComponent(currentPath)}` : "";
  window.location.replace(`/auth/login${redirect}`);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  // Hydrate from the session-display cookie so the first paint already has
  // the user shell (avatar, name) — avoids a flash before /me responds.
  const [user, setUser] = useState<User | null>(() => {
    if (typeof window === "undefined") return null;
    const s = getSessionUser();
    return s ? (s as User) : null;
  });
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async () => {
    try {
      const res = await authService.me();
      if (res.success && res.data) {
        setUser(res.data as unknown as User);
        // /api/auth/me rewrote the session-user cookie with the latest fields
        // (including isSuperAdmin). Re-hydrate the session store so the sidebar
        // and permission gates pick up the fresh data — not the stale shape
        // written at login time.
        useSessionStore.getState().hydrate();
      } else {
        clearSessionDisplayCookies();
        setUser(null);
        redirectToLoginAfterAuthFailure();
      }
    } catch {
      clearSessionDisplayCookies();
      setUser(null);
      redirectToLoginAfterAuthFailure();
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const logout = useCallback(async () => {
    await authService.logout().catch(() => {});
    clearSessionDisplayCookies();
    setUser(null);
    if (typeof window !== "undefined") window.location.href = "/auth/login";
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, loading, isAuthenticated: !!user, refreshUser: fetchUser, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
