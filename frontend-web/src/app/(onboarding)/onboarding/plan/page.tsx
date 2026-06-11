"use client";

import { ArrowLeft, ArrowRight, Check, Sparkles } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useActiveBilling, useStartTrial } from "@/hooks/billing/useBilling";
import { useSetOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { useRequireOwner } from "@/hooks/onboarding/useRequireOwner";
import { toast } from "@/hooks/use-toast";

// Onboarding plan step. The 4-plan-card UI was retired in Phase 10 of the
// billing overhaul. New customers land on a single "Start your trial" CTA —
// they can build a custom plan from /dashboard/billing/plan-builder once
// inside, picking features à la carte. Existing customers see "Continue".
export default function PlanStep() {
  // Owner-only step — invited members get bounced to /onboarding/profile.
  useRequireOwner();
  const router = useRouter();
  const activeQ = useActiveBilling();
  const startTrial = useStartTrial();
  const setState = useSetOnboardingState();

  const [submitting, setSubmitting] = useState(false);
  const hasSub = !!activeQ.data;

  async function go(beginTrial: boolean) {
    setSubmitting(true);
    try {
      if (beginTrial && !hasSub) {
        await startTrial.mutateAsync();
      }
      await setState.mutateAsync({ patch: { step: "done" } });
      router.push("/onboarding/done");
    } catch (e) {
      toast.error("Couldn't continue", e instanceof Error ? e.message : undefined);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="space-y-8">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 5 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">
          {hasSub ? "You're already set up" : "Start your 14-day trial"}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground max-w-md mx-auto">
          {hasSub
            ? `You're on ${activeQ.data?.planCode}. Build a custom plan any time from Billing → Plan builder.`
            : "Get full access to the Starter feature set on the house. Build a custom plan later when you know what fits."}
        </p>
      </div>

      <Card className="max-w-lg mx-auto">
        <CardContent className="p-6 space-y-4">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Sparkles className="h-4 w-4" />
            What's included in your trial
          </div>
          <ul className="space-y-2 text-sm">
            <Bullet>10 user seats</Bullet>
            <Bullet>Core: auth, settings, RBAC, notifications, invites</Bullet>
            <Bullet>14 days free · no card required</Bullet>
            <Bullet>Upgrade with custom features any time</Bullet>
          </ul>
          <Button
            className="w-full"
            onClick={() => go(true)}
            disabled={submitting}
          >
            {submitting
              ? "Setting up…"
              : hasSub
                ? "Continue to dashboard"
                : "Start trial — it's free"}
          </Button>
          <p className="text-center text-xs text-muted-foreground">
            You can build and activate a paid custom plan after onboarding.
          </p>
        </CardContent>
      </Card>

      <div className="flex items-center justify-between max-w-lg mx-auto">
        <Button variant="ghost" onClick={() => router.push("/onboarding/invites")}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <Button
          variant="ghost"
          onClick={() => go(false)}
          disabled={submitting || setState.isPending}
        >
          Decide later
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function Bullet({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-start gap-2">
      <Check className="mt-0.5 h-4 w-4 shrink-0 text-success" />
      <span>{children}</span>
    </li>
  );
}
