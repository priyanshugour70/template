"use client";

import { Check, Sparkles } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useActiveBilling, useStartTrial } from "@/hooks/billing/useBilling";
import { toast } from "@/hooks/use-toast";
import { useAuth, useTenant } from "@/providers";

// Onboarding subscription step. The multi-plan card UI was retired in Phase 10
// of the billing overhaul. Replaced by a single "Start trial" path because
// every customer now starts on the same trial and builds a custom plan
// post-onboarding from /dashboard/billing/plan-builder.
export default function SubscriptionOnboardingPage() {
  const { tenant, activeOrganization } = useTenant();
  const { user } = useAuth();
  const activeQ = useActiveBilling();
  const startTrial = useStartTrial();
  const [submitting, setSubmitting] = useState(false);
  const hasSub = !!activeQ.data;

  async function go(beginTrial: boolean) {
    setSubmitting(true);
    try {
      if (beginTrial && !hasSub) {
        await startTrial.mutateAsync();
      }
      window.location.assign("/dashboard");
    } catch (e) {
      toast.error("Couldn't continue", e instanceof Error ? e.message : undefined);
      setSubmitting(false);
    }
  }

  return (
    <div className="container mx-auto max-w-3xl px-6 py-16">
      <div className="text-center mb-10">
        <Badge variant="muted">Step 3 of 3</Badge>
        <h1 className="mt-4 text-3xl font-semibold tracking-tight">
          {hasSub ? "You're already set up" : "Start your 14-day trial"}
        </h1>
        <p className="mt-2 text-muted-foreground max-w-md mx-auto">
          Hi {user?.firstName ?? "there"}! You&apos;ve created{" "}
          <span className="font-medium text-foreground">
            {tenant?.name ?? activeOrganization?.name ?? "your workspace"}
          </span>
          .{" "}
          {hasSub
            ? "Continue to the dashboard — you can change plans any time from Billing."
            : "Get full Starter access with no payment for 14 days. Build a custom plan whenever you're ready."}
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardDescription className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" />
              {hasSub ? "Current plan" : "Trial includes"}
            </CardDescription>
            {hasSub && <Badge variant="success">{activeQ.data?.status}</Badge>}
          </div>
          <CardTitle className="text-xl mt-2">
            {hasSub ? activeQ.data?.planCode : "Starter trial"}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <ul className="space-y-2 text-sm">
            <Bullet>10 user seats</Bullet>
            <Bullet>Core: auth, settings, RBAC, notifications, invites</Bullet>
            <Bullet>14 days free · no card required</Bullet>
            <Bullet>Build custom plan from plan-builder when ready</Bullet>
          </ul>
          <Button
            className="w-full"
            disabled={submitting}
            onClick={() => go(true)}
          >
            {submitting
              ? "Setting up…"
              : hasSub
                ? "Continue to dashboard"
                : "Start trial"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

function Bullet({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-center gap-2">
      <Check className="h-4 w-4 text-emerald-500" />
      <span>{children}</span>
    </li>
  );
}
