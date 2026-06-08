"use client";

import type { ReactNode } from "react";

import { useRequireAuth } from "@/hooks/auth/useRequireAuth";

export default function OnboardingLayout({ children }: { children: ReactNode }) {
  const { loading } = useRequireAuth();
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-sm text-muted-foreground">Loading…</div>
      </div>
    );
  }
  return <div className="min-h-screen bg-gradient-to-br from-background to-muted/30">{children}</div>;
}
