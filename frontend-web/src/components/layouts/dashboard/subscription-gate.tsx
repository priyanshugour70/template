"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "@/providers";
import { useActiveSubscription } from "@/hooks/subscription/useSubscriptionQueries";

const ACTIVE_STATUSES = new Set(["trial", "active", "past_due", "paused"]);

/**
 * Hard gate around dashboard pages. If the active organisation has no
 * usable subscription, the user is bounced to /onboarding/subscription
 * — they can pick a plan there. Super admins bypass.
 */
export function SubscriptionGate({ children }: { children: ReactNode }) {
  const router = useRouter();
  const { user } = useAuth();
  const subQ = useActiveSubscription();

  const isSuperAdmin = user?.isSuperAdmin === true;
  const sub = subQ.data;
  const hasActive = sub ? ACTIVE_STATUSES.has(sub.status) : false;

  useEffect(() => {
    if (isSuperAdmin) return;
    // Wait for the query to settle before deciding.
    if (subQ.isLoading) return;
    if (!hasActive) router.replace("/onboarding/plan");
  }, [isSuperAdmin, subQ.isLoading, hasActive, router]);

  if (isSuperAdmin) return <>{children}</>;
  if (subQ.isLoading) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-sm text-muted-foreground">Checking subscription…</div>
      </div>
    );
  }
  if (!hasActive) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-sm text-muted-foreground">Redirecting to subscription onboarding…</div>
      </div>
    );
  }
  return <>{children}</>;
}
