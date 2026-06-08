"use client";

import { ArrowUpRight, Check, Crown, Sparkles, X } from "lucide-react";
import { useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useActiveSubscription,
  useChangePlan,
  useFeatureSet,
  usePlans,
} from "@/hooks/subscription/useSubscriptionQueries";
import type { Plan } from "@/types/subscription";

function formatPrice(cents: number, currency: string): string {
  if (cents === 0) return "Free";
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: currency || "INR",
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

function formatLimit(v: number): string {
  if (v === -1) return "Unlimited";
  return v.toLocaleString("en-IN");
}

const LIMIT_LABELS: Record<string, string> = {
  "users.max": "Team members",
  "orgs.max": "Organizations",
  "storage.gb": "Storage (GB)",
  "api.calls.monthly": "API calls / month",
};

const FEATURE_LABELS: Record<string, string> = {
  "user.invite": "Invite users",
  "user.list": "User management",
  "org.read": "View organizations",
  "org.create": "Create organizations",
  "audit.read": "Audit log",
  "audit.export": "Export audit logs",
  "export.csv": "CSV export",
  "export.xlsx": "Excel export",
  webhook: "Webhooks",
  sso: "SSO (Google / Microsoft)",
  saml: "SAML SSO",
  scim: "SCIM provisioning",
  priority_support: "Priority support",
  custom_branding: "Custom branding",
};

function featureLabel(key: string): string {
  return FEATURE_LABELS[key] ?? key.replace(/\./g, " ").replace(/_/g, " ");
}

export default function SubscriptionPage() {
  const activeQ = useActiveSubscription();
  const plansQ = usePlans();
  const featuresQ = useFeatureSet();
  const change = useChangePlan();
  const [confirmPlan, setConfirmPlan] = useState<Plan | null>(null);

  const sortedPlans = useMemo(
    () => [...(plansQ.data ?? [])].sort((a, b) => a.tier - b.tier),
    [plansQ.data],
  );

  const allFeatures = useMemo(() => {
    const set = new Set<string>();
    for (const p of plansQ.data ?? []) for (const f of p.features ?? []) set.add(f);
    return Array.from(set).sort();
  }, [plansQ.data]);

  const allLimits = useMemo(() => {
    const set = new Set<string>();
    for (const p of plansQ.data ?? []) for (const k of Object.keys(p.limits ?? {})) set.add(k);
    return Array.from(set).sort();
  }, [plansQ.data]);

  async function applyPlan() {
    if (!confirmPlan) return;
    try {
      await change.mutateAsync({ planCode: confirmPlan.code, startImmediately: true });
      setConfirmPlan(null);
    } catch {
      /* shown inline below */
    }
  }

  const current = activeQ.data;
  const currentPlanCode = current?.planCode;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Subscription</h1>
          <p className="text-muted-foreground mt-1">
            Manage your plan, compare features, and upgrade or downgrade at any time.
          </p>
        </div>
      </div>

      {/* Current plan summary */}
      {activeQ.isLoading ? (
        <Card>
          <CardContent className="p-6 grid gap-4 md:grid-cols-3">
            <Skeleton className="h-20" />
            <Skeleton className="h-20" />
            <Skeleton className="h-20" />
          </CardContent>
        </Card>
      ) : current ? (
        <Card className="overflow-hidden">
          <div className="grid md:grid-cols-3 divide-y md:divide-y-0 md:divide-x">
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">
                Current plan
              </div>
              <div className="flex items-center gap-2">
                <span className="text-2xl font-semibold capitalize">{current.planCode}</span>
                <Badge variant={current.status === "active" ? "success" : "warning"}>
                  {current.status}
                </Badge>
              </div>
              <div className="mt-2 text-sm text-muted-foreground">
                {formatPrice(current.unitPriceCents, current.currency)} / {current.billingCycle}
              </div>
            </div>
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">
                Renews
              </div>
              <div className="text-lg font-semibold">
                {current.currentPeriodEnd
                  ? new Date(current.currentPeriodEnd).toLocaleDateString()
                  : "—"}
              </div>
              <div className="mt-2 text-sm text-muted-foreground">
                {current.trialEndsAt && new Date(current.trialEndsAt) > new Date()
                  ? `Trial ends ${new Date(current.trialEndsAt).toLocaleDateString()}`
                  : `${current.billingCycle} billing`}
              </div>
            </div>
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">
                Included features
              </div>
              <div className="text-lg font-semibold">
                {featuresQ.data ? Object.keys(featuresQ.data.features).length : 0} features
              </div>
              <div className="mt-2 text-sm text-muted-foreground">
                {featuresQ.data
                  ? Object.keys(featuresQ.data.limits).length + " quotas configured"
                  : "—"}
              </div>
            </div>
          </div>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-6 text-sm text-muted-foreground">
            No active subscription. Pick a plan below to get started.
          </CardContent>
        </Card>
      )}

      {/* Plan cards */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Choose your plan</h2>
        {plansQ.isLoading ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-80" />
            ))}
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {sortedPlans.map((p) => {
              const isCurrent = p.code === currentPlanCode;
              const isUpgrade = current ? p.tier > (sortedPlans.find(x => x.code === currentPlanCode)?.tier ?? -1) : true;
              return (
                <Card
                  key={p.id}
                  className={
                    "relative flex flex-col " +
                    (isCurrent ? "border-primary shadow-md" : "")
                  }
                >
                  {p.isDefault && !isCurrent && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                      <Badge variant="default" className="gap-1">
                        <Sparkles className="h-3 w-3" /> Popular
                      </Badge>
                    </div>
                  )}
                  {isCurrent && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                      <Badge variant="success" className="gap-1">
                        <Check className="h-3 w-3" /> Current
                      </Badge>
                    </div>
                  )}
                  <CardHeader>
                    <div className="flex items-center gap-2">
                      {p.tier >= 3 && <Crown className="h-4 w-4 text-amber-500" />}
                      <CardDescription className="capitalize text-xs uppercase tracking-wider">
                        {p.code}
                      </CardDescription>
                    </div>
                    <CardTitle className="text-2xl">{p.name}</CardTitle>
                    <div className="mt-1">
                      <span className="text-3xl font-semibold">
                        {formatPrice(p.priceCents, p.currency)}
                      </span>
                      <span className="ml-1 text-sm text-muted-foreground">
                        / {p.billingCycle}
                      </span>
                    </div>
                    {p.trialDays > 0 && (
                      <Badge variant="muted" className="w-fit mt-1">
                        {p.trialDays}-day trial
                      </Badge>
                    )}
                    {p.description && (
                      <p className="text-sm text-muted-foreground mt-2">{p.description}</p>
                    )}
                  </CardHeader>
                  <CardContent className="flex-1 space-y-4">
                    <Separator />
                    <ul className="space-y-2 text-sm">
                      {(p.features ?? []).slice(0, 7).map((f) => (
                        <li key={f} className="flex items-start gap-2">
                          <Check className="h-4 w-4 text-emerald-500 mt-0.5 shrink-0" />
                          <span>{featureLabel(f)}</span>
                        </li>
                      ))}
                      {(p.features?.length ?? 0) > 7 && (
                        <li className="text-xs text-muted-foreground">
                          +{(p.features?.length ?? 0) - 7} more features
                        </li>
                      )}
                    </ul>
                    <Separator />
                    <ul className="space-y-1.5 text-sm">
                      {Object.entries(p.limits ?? {}).map(([k, v]) => (
                        <li key={k} className="flex items-center justify-between">
                          <span className="text-muted-foreground">{LIMIT_LABELS[k] ?? k}</span>
                          <span className="font-medium">{formatLimit(Number(v))}</span>
                        </li>
                      ))}
                    </ul>
                  </CardContent>
                  <div className="p-6 pt-0">
                    {isCurrent ? (
                      <Button variant="outline" className="w-full" disabled>
                        Current plan
                      </Button>
                    ) : (
                      <Button
                        variant={isUpgrade ? "default" : "outline"}
                        className="w-full gap-2"
                        onClick={() => setConfirmPlan(p)}
                      >
                        {isUpgrade ? (
                          <>
                            Upgrade <ArrowUpRight className="h-4 w-4" />
                          </>
                        ) : (
                          "Switch plan"
                        )}
                      </Button>
                    )}
                  </div>
                </Card>
              );
            })}
          </div>
        )}
      </div>

      {/* Comparison matrix */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Compare features</h2>
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-muted/40">
                <tr>
                  <th className="text-left p-3 font-medium text-muted-foreground w-1/3">
                    Feature / Quota
                  </th>
                  {sortedPlans.map((p) => (
                    <th key={p.id} className="text-left p-3 font-semibold capitalize">
                      {p.name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {allFeatures.length === 0 && allLimits.length === 0 ? (
                  <tr>
                    <td
                      colSpan={1 + sortedPlans.length}
                      className="p-6 text-center text-muted-foreground"
                    >
                      No comparison data yet.
                    </td>
                  </tr>
                ) : null}
                {allLimits.map((key) => (
                  <tr key={`limit-${key}`} className="border-t">
                    <td className="p-3 text-muted-foreground">
                      {LIMIT_LABELS[key] ?? key}
                    </td>
                    {sortedPlans.map((p) => {
                      const v = (p.limits as Record<string, number> | undefined)?.[key];
                      return (
                        <td key={p.id} className="p-3 font-medium">
                          {v == null ? "—" : formatLimit(Number(v))}
                        </td>
                      );
                    })}
                  </tr>
                ))}
                {allFeatures.map((f) => (
                  <tr key={`feat-${f}`} className="border-t">
                    <td className="p-3 text-muted-foreground">{featureLabel(f)}</td>
                    {sortedPlans.map((p) => {
                      const has = (p.features ?? []).includes(f);
                      return (
                        <td key={p.id} className="p-3">
                          {has ? (
                            <Check className="h-4 w-4 text-emerald-500" />
                          ) : (
                            <X className="h-4 w-4 text-muted-foreground/40" />
                          )}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </div>

      {/* Confirm dialog */}
      <Dialog open={!!confirmPlan} onOpenChange={(o) => !o && setConfirmPlan(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Switch to {confirmPlan?.name}?</DialogTitle>
            <DialogDescription>
              Your subscription will change immediately. You will be billed{" "}
              <span className="font-medium text-foreground">
                {confirmPlan ? formatPrice(confirmPlan.priceCents, confirmPlan.currency) : "—"} /{" "}
                {confirmPlan?.billingCycle}
              </span>{" "}
              starting next cycle. No payment is taken right now (this is a demo).
            </DialogDescription>
          </DialogHeader>
          {change.isError && (
            <p className="text-sm text-destructive">
              {change.error instanceof Error ? change.error.message : "Failed to switch plan."}
            </p>
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setConfirmPlan(null)}>
              Cancel
            </Button>
            <Button onClick={applyPlan} disabled={change.isPending}>
              {change.isPending ? "Switching…" : `Switch to ${confirmPlan?.name}`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
