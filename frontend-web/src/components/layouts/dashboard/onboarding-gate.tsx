"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { useAuth } from "@/providers";

/**
 * Bounces unfinished users to /onboarding from any /dashboard route. Super
 * admins bypass — they don't go through tenant onboarding.
 */
export function OnboardingGate({ children }: { children: ReactNode }) {
  const router = useRouter();
  const { user } = useAuth();
  const state = useOnboardingState();

  const isSuperAdmin = user?.isSuperAdmin === true;
  const needsOnboarding = !isSuperAdmin && !state.completed;

  useEffect(() => {
    if (needsOnboarding) router.replace("/onboarding");
  }, [needsOnboarding, router]);

  if (needsOnboarding) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="text-sm text-muted-foreground">Redirecting to onboarding…</div>
      </div>
    );
  }
  return <>{children}</>;
}
