"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/providers";

/** Redirects to /auth/login if the user is not authenticated. Returns auth state. */
export function useRequireAuth() {
  const router = useRouter();
  const { user, loading, isAuthenticated } = useAuth();

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      const path = typeof window !== "undefined" ? window.location.pathname + window.location.search : "";
      const redirect = path ? `?redirect=${encodeURIComponent(path)}` : "";
      router.replace(`/auth/login${redirect}`);
    }
  }, [loading, isAuthenticated, router]);

  return { user, loading, isAuthenticated };
}
