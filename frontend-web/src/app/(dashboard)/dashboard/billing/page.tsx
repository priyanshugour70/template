"use client";

import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useActiveBilling,
  useInvoices,
  useQuotations,
  useTransactions,
} from "@/hooks/billing/useBilling";

import { formatDate, formatMoney } from "./_components/money";
import { StatusBadge } from "./_components/status-badge";

export default function BillingOverviewPage() {
  const subQ = useActiveBilling();
  const invQ = useInvoices();
  const txQ = useTransactions();
  const quotQ = useQuotations("draft");

  const sub = subQ.data;
  const openInvoices = (invQ.data ?? []).filter((i) => i.status === "open");
  const totalDue = openInvoices.reduce((sum, i) => sum + (i.amountDueCents ?? 0), 0);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Tile
          title="Current plan"
          value={sub?.planCode ?? "—"}
          hint={sub ? <StatusBadge status={sub.status} /> : undefined}
          loading={subQ.isLoading}
          href="/dashboard/billing/subscription"
        />
        <Tile
          title="Next billing"
          value={sub?.nextBillingAt ? formatDate(sub.nextBillingAt) : "—"}
          hint={sub ? formatMoney(sub.currency, sub.totalCents) : ""}
          loading={subQ.isLoading}
        />
        <Tile
          title="Open invoices"
          value={String(openInvoices.length)}
          hint={
            totalDue > 0 && sub
              ? `${formatMoney(sub.currency, totalDue)} due`
              : "All paid"
          }
          loading={invQ.isLoading}
          href="/dashboard/billing/invoices"
        />
        <Tile
          title="Draft quotations"
          value={String((quotQ.data ?? []).length)}
          loading={quotQ.isLoading}
          href="/dashboard/billing/quotations"
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent invoices</CardTitle>
            <CardDescription>Last 5 issued</CardDescription>
          </CardHeader>
          <CardContent>
            {invQ.isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : (invQ.data ?? []).length === 0 ? (
              <p className="text-sm text-muted-foreground">No invoices yet.</p>
            ) : (
              <ul className="divide-y -my-2">
                {(invQ.data ?? []).slice(0, 5).map((inv) => (
                  <li key={inv.id} className="flex items-center justify-between py-2 text-sm">
                    <Link href={`/dashboard/billing/invoices/${inv.id}`} className="hover:underline">
                      {inv.number}
                    </Link>
                    <div className="flex items-center gap-3">
                      <span className="text-muted-foreground">
                        {formatMoney(inv.currency, inv.totalCents)}
                      </span>
                      <StatusBadge status={inv.status} />
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent payments</CardTitle>
            <CardDescription>Last 5 transactions</CardDescription>
          </CardHeader>
          <CardContent>
            {txQ.isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : (txQ.data ?? []).length === 0 ? (
              <p className="text-sm text-muted-foreground">No payments yet.</p>
            ) : (
              <ul className="divide-y -my-2">
                {(txQ.data ?? []).slice(0, 5).map((t) => (
                  <li key={t.id} className="flex items-center justify-between py-2 text-sm">
                    <Link href={`/dashboard/billing/transactions`} className="hover:underline">
                      {t.receiptNumber}
                    </Link>
                    <div className="flex items-center gap-3">
                      <span className="text-muted-foreground">
                        {formatMoney(t.currency, t.amountCents)}
                      </span>
                      <span className="text-xs text-muted-foreground capitalize">{t.method.replace("_", " ")}</span>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function Tile({
  title,
  value,
  hint,
  loading,
  href,
}: {
  title: string;
  value: string;
  hint?: React.ReactNode;
  loading?: boolean;
  href?: string;
}) {
  const body = (
    <Card className={href ? "cursor-pointer hover:bg-accent transition-colors" : ""}>
      <CardHeader className="pb-2">
        <CardDescription className="text-xs uppercase tracking-wide">{title}</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-7 w-24" />
        ) : (
          <div className="text-2xl font-semibold">{value}</div>
        )}
        {hint != null && <div className="mt-2 text-xs text-muted-foreground">{hint}</div>}
      </CardContent>
    </Card>
  );
  return href ? <Link href={href}>{body}</Link> : body;
}
