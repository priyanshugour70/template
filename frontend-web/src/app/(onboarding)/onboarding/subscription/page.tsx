"use client";

import { Check } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuth, useTenant } from "@/providers";
import { useChangePlan, usePlans } from "@/hooks/subscription/useSubscriptionQueries";

function formatPrice(cents: number, currency: string): string {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: currency || "INR",
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

export default function SubscriptionOnboardingPage() {
  const { tenant, activeOrganization } = useTenant();
  const { user } = useAuth();
  const plansQ = usePlans();
  const change = useChangePlan();
  const [picked, setPicked] = useState<string | null>(null);

  async function selectPlan(code: string) {
    setPicked(code);
    try {
      await change.mutateAsync({ planCode: code, startImmediately: true });
      window.location.assign("/dashboard");
    } catch {
      setPicked(null);
    }
  }

  return (
    <div className="container mx-auto max-w-6xl px-6 py-16">
      <div className="text-center mb-12">
        <Badge variant="muted">Step 3 of 3</Badge>
        <h1 className="mt-4 text-3xl font-semibold tracking-tight">Pick a plan to get started</h1>
        <p className="mt-2 text-muted-foreground">
          Hi {user?.firstName ?? "there"}! You&apos;ve created{" "}
          <span className="font-medium text-foreground">{tenant?.name ?? activeOrganization?.name ?? "your workspace"}</span>.
          Pick a plan to unlock the dashboard. You can always change or cancel later.
        </p>
      </div>

      {plansQ.isLoading && (
        <div className="text-center text-muted-foreground">Loading plans…</div>
      )}

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
        {plansQ.data?.map((p) => (
          <Card key={p.id} className={picked === p.code ? "border-primary ring-2 ring-primary/30" : ""}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardDescription className="capitalize">{p.code}</CardDescription>
                {p.isDefault && <Badge variant="muted">popular</Badge>}
              </div>
              <CardTitle className="text-xl">{p.name}</CardTitle>
              <div className="mt-2 text-3xl font-semibold">
                {formatPrice(p.priceCents, p.currency)}
                <span className="ml-1 text-sm font-normal text-muted-foreground">
                  /{p.billingCycle}
                </span>
              </div>
              {p.trialDays > 0 && (
                <Badge variant="success" className="w-fit mt-2">
                  {p.trialDays}-day trial
                </Badge>
              )}
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground">{p.description}</p>
              <ul className="space-y-1.5 text-sm">
                {(p.features ?? []).slice(0, 6).map((f) => (
                  <li key={f} className="flex items-center gap-2">
                    <Check className="h-4 w-4 text-emerald-500" />
                    <span className="text-foreground/80">{f}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
            <CardFooter>
              <Button
                className="w-full"
                variant={p.isDefault ? "default" : "outline"}
                disabled={change.isPending}
                onClick={() => selectPlan(p.code)}
              >
                {change.isPending && picked === p.code
                  ? "Activating…"
                  : p.priceCents === 0
                  ? "Get started"
                  : "Start trial"}
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

      {change.isError && (
        <p className="mt-6 text-center text-sm text-destructive">
          {change.error instanceof Error ? change.error.message : "Activation failed"}
        </p>
      )}
    </div>
  );
}
