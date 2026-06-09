"use client";

import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useActiveBilling } from "@/hooks/billing/useBilling";

import { formatDate, formatMoney } from "../_components/money";
import { StatusBadge } from "../_components/status-badge";

export default function BillingSubscriptionPage() {
  const subQ = useActiveBilling();
  const sub = subQ.data;

  if (subQ.isLoading) {
    return <Skeleton className="h-64 w-full" />;
  }
  if (!sub) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-sm text-muted-foreground">
          No active subscription.{" "}
          <Link href="/dashboard/billing/plan-builder" className="underline">
            Build a plan
          </Link>{" "}
          to get started.
        </CardContent>
      </Card>
    );
  }

  const features = sub.features ?? [];
  const limits = sub.limits ?? {};

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>{sub.planCode}</CardTitle>
            <CardDescription>
              {sub.billingCycle} · started {formatDate(sub.startedAt)}
            </CardDescription>
          </div>
          <StatusBadge status={sub.status} />
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2 text-sm">
            <Row label="Subtotal" value={formatMoney(sub.currency, sub.unitPriceCents * (sub.quantity || 1))} />
            <Row label="Tax" value={formatMoney(sub.currency, sub.taxCents)} />
            <Row label="Total / cycle" value={formatMoney(sub.currency, sub.totalCents)} />
            <Row label="Currency" value={sub.currency} />
            <Row label="Current period start" value={formatDate(sub.currentPeriodStart)} />
            <Row label="Current period end" value={formatDate(sub.currentPeriodEnd)} />
            <Row label="Next billing" value={formatDate(sub.nextBillingAt)} />
            <Row label="Trial ends" value={sub.trialEndsAt ? formatDate(sub.trialEndsAt) : "—"} />
            <Row label="Billing email" value={sub.billingEmail ?? "—"} />
            <Row label="Billing state" value={sub.billingState ?? "—"} />
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Plan features</CardTitle>
          <CardDescription>
            Resolved at activation — catalog price changes don't retroactively re-bill.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {features.length === 0 ? (
            <p className="text-sm text-muted-foreground">No features attached.</p>
          ) : (
            <ul className="flex flex-wrap gap-2">
              {features.map((f) => (
                <li
                  key={f}
                  className="rounded-full bg-muted px-3 py-1 text-xs font-mono"
                >
                  {f}
                </li>
              ))}
            </ul>
          )}
          {Object.keys(limits).length > 0 && (
            <dl className="grid grid-cols-2 sm:grid-cols-4 gap-4 mt-6 text-sm">
              {Object.entries(limits).map(([k, v]) => (
                <Row key={k} label={k} value={String(v)} />
              ))}
            </dl>
          )}
        </CardContent>
      </Card>

      <div className="flex gap-2">
        <Link
          href="/dashboard/billing/plan-builder"
          className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
        >
          Build a new plan
        </Link>
        <Link
          href="/dashboard/billing/invoices"
          className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
        >
          View invoices
        </Link>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs uppercase tracking-wide text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium">{value}</dd>
    </div>
  );
}
