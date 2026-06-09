"use client";

import { ArrowLeft, ArrowRight, Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSetOnboardingState } from "@/hooks/onboarding/useOnboarding";
import {
  useActiveSubscription,
  useChangePlan,
  usePlans,
} from "@/hooks/subscription/useSubscriptionQueries";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";

function formatMoney(cents: number, currency: string) {
  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: currency || "INR",
      maximumFractionDigits: 0,
    }).format(cents / 100);
  } catch {
    return `${currency} ${(cents / 100).toFixed(0)}`;
  }
}

export default function PlanStep() {
  const router = useRouter();
  const plansQ = usePlans();
  const activeQ = useActiveSubscription();
  const change = useChangePlan();
  const setState = useSetOnboardingState();

  const [picked, setPicked] = useState<string>("");

  const plans = (plansQ.data ?? []).sort((a, b) => a.tier - b.tier);
  const currentCode = activeQ.data?.planCode;

  const next = async (planCode: string) => {
    setPicked(planCode);
    try {
      // Only call ChangePlan if the user picked something different from current
      if (planCode !== currentCode) {
        await change.mutateAsync({ planCode, startImmediately: true });
      }
      await setState.mutateAsync({ patch: { step: "done" } });
      router.push("/onboarding/done");
    } catch (e: unknown) {
      setPicked("");
      toast.error("Couldn't switch plan", e instanceof Error ? e.message : undefined);
    }
  };

  if (plansQ.isLoading) {
    return <Skeleton className="h-64 w-full" />;
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 5 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Pick a plan</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          You can change or cancel any time from Settings → Subscription.
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {plans.map((p) => {
          const isCurrent = p.code === currentCode;
          const isPicking = picked === p.code && (change.isPending || setState.isPending);
          return (
            <Card
              key={p.code}
              className={cn(
                "relative cursor-pointer transition-colors hover:border-primary/50",
                isCurrent && "border-primary",
              )}
              onClick={() => !change.isPending && !setState.isPending && next(p.code)}
            >
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold">{p.name}</h3>
                  {isCurrent && <Badge variant="success">current</Badge>}
                </div>
                <div className="mt-2">
                  <div className="text-2xl font-semibold tabular-nums">
                    {p.priceCents > 0 ? formatMoney(p.priceCents, p.currency) : "Free"}
                  </div>
                  {p.priceCents > 0 && (
                    <div className="text-xs text-muted-foreground">per {p.billingCycle}</div>
                  )}
                  {p.trialDays > 0 && (
                    <div className="mt-1 text-xs text-success">{p.trialDays}-day trial</div>
                  )}
                </div>
                <ul className="mt-3 space-y-1 text-xs">
                  {(p.features ?? []).slice(0, 4).map((f) => (
                    <li key={f} className="flex items-start gap-1.5">
                      <Check className="mt-0.5 h-3 w-3 shrink-0 text-success" />
                      <code className="font-mono">{f}</code>
                    </li>
                  ))}
                  {(p.features?.length ?? 0) > 4 && (
                    <li className="text-muted-foreground">
                      + {(p.features?.length ?? 0) - 4} more
                    </li>
                  )}
                </ul>
                <Button
                  className="mt-4 w-full"
                  variant={isCurrent ? "outline" : "default"}
                  disabled={isPicking}
                >
                  {isPicking ? "Activating…" : isCurrent ? "Continue on this plan" : "Choose"}
                </Button>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={() => router.push("/onboarding/invites")}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <Button
          variant="ghost"
          onClick={async () => {
            await setState.mutateAsync({ patch: { step: "done" } });
            router.push("/onboarding/done");
          }}
          disabled={setState.isPending}
        >
          Decide later
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
