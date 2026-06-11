"use client";

import { Check, Sparkles, Loader2 } from "lucide-react";
import { useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  useActiveBilling,
  useCreateQuotationFromPlan,
  usePlans,
  useStartTrial,
} from "@/hooks/billing/useBilling";
import { toast } from "@/hooks/use-toast";
import { useAuth, useTenant } from "@/providers";
import type { Plan } from "@/types/billing";

// Enterprise tier is special-cased to also show a "Talk to sales" CTA — these
// deals are typically negotiated, not self-served. Plan codes are matched
// case-insensitively against this set.
const ENTERPRISE_PLAN_CODES = new Set(["enterprise"]);

// Sales contact for the Enterprise mailto:. Easy to swap out later when you
// hook up a proper contact form.
const SALES_EMAIL = "sales@lssgoo.com";

export default function SubscriptionOnboardingPage() {
  const { tenant, activeOrganization } = useTenant();
  const { user } = useAuth();
  const activeQ = useActiveBilling();
  const plansQ = usePlans();
  const startTrial = useStartTrial();
  const fromPlan = useCreateQuotationFromPlan();

  const hasSub = !!activeQ.data;
  const [pending, setPending] = useState<string | null>(null);

  // Sort plans by tier ascending — Free, Starter, Pro, Enterprise. Custom
  // plans (negative tier or marked is_custom) get pushed to the end. Free is
  // always first if it exists.
  const plans = useMemo<Plan[]>(() => {
    const list = plansQ.data ?? [];
    return [...list].sort((a, b) => {
      const aCustom = a.isCustom || a.code.startsWith("custom-");
      const bCustom = b.isCustom || b.code.startsWith("custom-");
      if (aCustom !== bCustom) return aCustom ? 1 : -1;
      if (a.tier !== b.tier) return a.tier - b.tier;
      return a.priceCents - b.priceCents;
    });
  }, [plansQ.data]);

  async function pickFree(plan: Plan) {
    setPending(plan.code);
    try {
      await startTrial.mutateAsync();
      window.location.assign("/dashboard");
    } catch (e) {
      toast.error("Couldn't start your plan", e instanceof Error ? e.message : undefined);
      setPending(null);
    }
  }

  async function pickPaid(plan: Plan) {
    setPending(plan.code);
    try {
      const q = await fromPlan.mutateAsync({ planCode: plan.code, userCount: 1 });
      // Land the user on the quotation page so they can review pricing and
      // pay. /dashboard is gated by the BillingGate for mutations, but
      // /dashboard/billing/quotations is read-only and unrestricted.
      window.location.assign(`/dashboard/billing/quotations/${q.id}`);
    } catch (e) {
      toast.error("Couldn't create your quote", e instanceof Error ? e.message : undefined);
      setPending(null);
    }
  }

  // Already-subscribed branch: short-circuits the plan grid entirely.
  if (hasSub) {
    return (
      <div className="container mx-auto max-w-2xl px-6 py-16">
        <Header step="Step 3 of 3" title="You're already set up">
          Hi {user?.firstName ?? "there"}! Your workspace{" "}
          <span className="font-medium text-foreground">
            {tenant?.name ?? activeOrganization?.name ?? "your workspace"}
          </span>{" "}
          is already on a plan — continue to the dashboard. You can change plans any time
          from Billing.
        </Header>
        <Card>
          <CardHeader>
            <CardDescription className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" /> Current plan
            </CardDescription>
            <CardTitle className="text-xl mt-2">{activeQ.data?.planCode}</CardTitle>
          </CardHeader>
          <CardContent>
            <Button className="w-full" onClick={() => window.location.assign("/dashboard")}>
              Continue to dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-6xl px-6 py-16">
      <Header step="Step 3 of 3" title="Pick a plan">
        Hi {user?.firstName ?? "there"}! You&apos;ve created{" "}
        <span className="font-medium text-foreground">
          {tenant?.name ?? activeOrganization?.name ?? "your workspace"}
        </span>
        . Choose the plan that fits — paid plans land you on a quote you can review and pay
        in one go. Free starts immediately.
      </Header>

      {plansQ.isLoading ? (
        <div className="flex items-center justify-center py-20 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading plans…
        </div>
      ) : plans.length === 0 ? (
        <div className="text-center text-muted-foreground py-12">
          No plans are configured yet. Ask an admin to publish a plan, or{" "}
          <button
            className="text-primary underline"
            onClick={() => pickFree({ code: "", priceCents: 0 } as Plan)}
          >
            start a default trial
          </button>
          .
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {plans.map((p) => (
            <PlanCard
              key={p.id}
              plan={p}
              pending={pending === p.code}
              disabled={pending !== null && pending !== p.code}
              onFree={() => pickFree(p)}
              onPaid={() => pickPaid(p)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function Header({
  step,
  title,
  children,
}: {
  step: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="text-center mb-10">
      <Badge variant="muted">{step}</Badge>
      <h1 className="mt-4 text-3xl font-semibold tracking-tight">{title}</h1>
      <p className="mt-2 text-muted-foreground max-w-2xl mx-auto">{children}</p>
    </div>
  );
}

function PlanCard({
  plan,
  pending,
  disabled,
  onFree,
  onPaid,
}: {
  plan: Plan;
  pending: boolean;
  disabled: boolean;
  onFree: () => void;
  onPaid: () => void;
}) {
  const isFree = plan.priceCents === 0 && !ENTERPRISE_PLAN_CODES.has(plan.code.toLowerCase());
  const isEnterprise = ENTERPRISE_PLAN_CODES.has(plan.code.toLowerCase());
  const isCustom = plan.isCustom || plan.code.startsWith("custom-");

  return (
    <Card
      className={`flex flex-col ${plan.isDefault ? "border-primary shadow-sm" : ""}`}
    >
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-xl">{plan.name}</CardTitle>
          {plan.isDefault && <Badge variant="success">Default</Badge>}
          {isCustom && !plan.isDefault && <Badge variant="muted">Custom</Badge>}
        </div>
        {plan.tagline && (
          <CardDescription>{plan.tagline}</CardDescription>
        )}
        <Price priceCents={plan.priceCents} currency={plan.currency} cycle={plan.billingCycle} />
        {plan.trialDays > 0 && (
          <div className="text-xs text-muted-foreground mt-1">
            {plan.trialDays}-day free trial
          </div>
        )}
      </CardHeader>
      <CardContent className="flex-1 flex flex-col">
        <ul className="space-y-2 text-sm flex-1">
          {(plan.features ?? []).slice(0, 8).map((f) => (
            <li key={f} className="flex items-start gap-2">
              <Check className="h-4 w-4 text-emerald-500 mt-0.5" />
              <span className="break-all">{prettyFeature(f)}</span>
            </li>
          ))}
          {(plan.features?.length ?? 0) > 8 && (
            <li className="text-xs text-muted-foreground pl-6">
              + {(plan.features?.length ?? 0) - 8} more
            </li>
          )}
        </ul>

        <div className="mt-6 space-y-2">
          {isFree ? (
            <Button
              className="w-full"
              disabled={disabled || pending}
              onClick={onFree}
            >
              {pending ? "Starting…" : "Start now"}
            </Button>
          ) : (
            <Button
              className="w-full"
              disabled={disabled || pending}
              onClick={onPaid}
            >
              {pending ? "Building quote…" : isEnterprise ? "Get a quote" : "Choose plan"}
            </Button>
          )}
          {isEnterprise && (
            <Button
              variant="outline"
              className="w-full"
              asChild
            >
              <a
                href={`mailto:${SALES_EMAIL}?subject=${encodeURIComponent(
                  `Enterprise plan enquiry — ${plan.name}`,
                )}`}
              >
                Talk to sales
              </a>
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function Price({
  priceCents,
  currency,
  cycle,
}: {
  priceCents: number;
  currency: string;
  cycle: string;
}) {
  if (priceCents === 0) {
    return <div className="mt-2 text-3xl font-semibold">Free</div>;
  }
  // priceCents in this codebase is rupee * 100 for INR (and equivalent
  // minor-units for other currencies). Format with the Intl API and append a
  // /mo or /yr badge.
  const amount = priceCents / 100;
  const formatted = new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: currency || "INR",
    maximumFractionDigits: 0,
  }).format(amount);
  const suffix = cycle === "yearly" ? "/yr" : cycle === "weekly" ? "/wk" : "/mo";
  return (
    <div className="mt-2">
      <span className="text-3xl font-semibold">{formatted}</span>
      <span className="text-sm text-muted-foreground ml-1">{suffix}</span>
    </div>
  );
}

// Feature keys are dotted machine names like "audit.export". Render them as
// "Audit export" so users can actually read the plan benefits.
function prettyFeature(key: string): string {
  return key
    .split(".")
    .map((part) => part.replace(/_/g, " "))
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" · ");
}
