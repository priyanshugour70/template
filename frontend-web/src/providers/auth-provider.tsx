"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import { clearTokens, getTokens } from "@/lib/cookies";
import { authService } from "@/services/auth";
import type { User } from "@/types/auth";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  isAuthenticated: boolean;
  refreshUser: () => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  loading: true,
  isAuthenticated: false,
  refreshUser: async () => {},
  logout: () => {},
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
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async () => {
    const { accessToken } = getTokens();
    if (!accessToken) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      const res = await authService.me();
      if (res.success && res.data) {
        setUser(res.data);
      } else {
        clearTokens();
        setUser(null);
        redirectToLoginAfterAuthFailure();
      }
    } catch {
      clearTokens();
      setUser(null);
      redirectToLoginAfterAuthFailure();
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const logout = useCallback(() => {
    const { refreshToken } = getTokens();
    if (refreshToken) authService.logout(refreshToken).catch(() => {});
    clearTokens();
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
