"use client";

import type { ReactNode } from "react";

import { DashboardLayout } from "@/components/layouts/dashboard/DashboardLayout";
import { OnboardingGate } from "@/components/layouts/dashboard/onboarding-gate";
import { SubscriptionGate } from "@/components/layouts/dashboard/subscription-gate";
import { useRequireAuth } from "@/hooks/auth/useRequireAuth";

export default function DashboardSegmentLayout({ children }: { children: ReactNode }) {
  const { loading } = useRequireAuth();
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-sm text-muted-foreground">Loading…</div>
      </div>
    );
  }
  return (
    <DashboardLayout>
      <OnboardingGate>
        <SubscriptionGate>{children}</SubscriptionGate>
      </OnboardingGate>
    </DashboardLayout>
  );
}
