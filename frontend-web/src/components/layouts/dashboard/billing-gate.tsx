"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "@/providers";
import { useActiveBilling } from "@/hooks/billing/useBilling";

// Trial + active + paused remain usable. expired/cancelled/past_due trigger
// the soft-lock UI — the backend BillingGate middleware still 402s mutations,
// this just gives the user a graceful screen instead of broken pages.
const ACTIVE_STATUSES = new Set(["trial", "active", "paused"]);

// LOCKED_STATUSES surface the BillingInactiveScreen. Everything else falls
// back to "no subscription → onboarding".
const LOCKED_STATUSES = new Set(["expired", "cancelled", "past_due"]);

export function BillingGate({ children }: { children: ReactNode }) {
  const router = useRouter();
  const { user } = useAuth();
  const subQ = useActiveBilling();

  const isSuperAdmin = user?.isSuperAdmin === true;
  const sub = subQ.data;
  const status = sub?.status;
  const hasActive = sub ? ACTIVE_STATUSES.has(sub.status) : false;
  const isLocked = sub ? LOCKED_STATUSES.has(sub.status) : false;

  useEffect(() => {
    if (isSuperAdmin) return;
    if (subQ.isLoading) return;
    // No subscription at all → onboarding picks a plan.
    if (!sub) router.replace("/onboarding/plan");
  }, [isSuperAdmin, subQ.isLoading, sub, router]);

  if (isSuperAdmin) return <>{children}</>;
  if (subQ.isLoading) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-sm text-muted-foreground">Checking billing…</div>
      </div>
    );
  }
  if (isLocked) {
    return <BillingInactiveScreen status={status!} />;
  }
  if (!hasActive) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-sm text-muted-foreground">Redirecting…</div>
      </div>
    );
  }
  return <>{children}</>;
}

function BillingInactiveScreen({ status }: { status: string }) {
  return (
    <div className="min-h-[60vh] flex items-center justify-center p-6">
      <div className="max-w-md w-full text-center space-y-4 rounded-2xl border border-destructive/30 bg-destructive/5 p-8">
        <div className="text-sm font-medium uppercase tracking-wide text-destructive">
          Billing {status}
        </div>
        <h2 className="text-2xl font-semibold">Your account is locked</h2>
        <p className="text-sm text-muted-foreground">
          Settle the open invoice or pick a new plan to restore full access.
          Reads still work — invoices, transactions, and receipts remain
          available so you can finish payment.
        </p>
        <div className="flex justify-center gap-3 pt-2">
          <a
            href="/dashboard/billing/invoices"
            className="rounded-md bg-foreground px-4 py-2 text-sm font-medium text-background"
          >
            View invoices
          </a>
          <a
            href="/dashboard/billing/plan-builder"
            className="rounded-md border px-4 py-2 text-sm font-medium"
          >
            Choose new plan
          </a>
        </div>
      </div>
    </div>
  );
}
