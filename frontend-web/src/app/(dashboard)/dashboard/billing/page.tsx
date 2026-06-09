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
  const planLabel = formatPlanCode(sub?.planCode);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Tile
          title="Current plan"
          value={planLabel.label}
          hint={
            sub ? (
              <div className="flex items-center gap-2">
                <StatusBadge status={sub.status} />
                {planLabel.sub ? (
                  <span className="font-mono text-[11px] tracking-tight">
                    {planLabel.sub}
                  </span>
                ) : null}
              </div>
            ) : undefined
          }
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
        <Card className="h-full">
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

        <Card className="h-full">
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

// formatPlanCode prettifies the headline shown on the Current plan tile.
// Custom plans materialise as `custom-quo-YYYY-NNNNNN` — too long for the
// tile and meaningless to non-engineers. We show "Custom plan" as the
// headline and surface the quotation number as a small caption underneath
// so the lineage is still visible. Preset plans (free/starter/etc.) get
// title-cased; anything else falls back to the raw code.
function formatPlanCode(code: string | undefined): { label: string; sub?: string } {
  if (!code) return { label: "—" };
  const m = code.match(/^custom-quo-(\d{4}-\d{6})$/i);
  if (m) return { label: "Custom plan", sub: `QUO-${m[1]}` };
  return { label: code.charAt(0).toUpperCase() + code.slice(1) };
}

// Tile is the KPI card. Every tile renders the same vertical rhythm regardless
// of which fields are populated:
//   • header (uppercase title, fixed height)
//   • value (single line, truncated, fixed height)
//   • hint slot (always rendered — `&nbsp;` when empty so the box height
//     stays identical across tiles)
// Without the placeholder hint slot, tiles with no hint shrink and the row
// looks ragged — same reason the "Draft quotations" tile was visibly shorter
// than "Current plan" in the original layout.
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
    <Card
      className={`h-full flex flex-col ${
        href ? "cursor-pointer hover:bg-accent transition-colors" : ""
      }`}
    >
      <CardHeader className="pb-2">
        <CardDescription className="text-xs uppercase tracking-wide">{title}</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between">
        {loading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <div
            className="text-2xl font-semibold truncate leading-tight"
            title={value}
          >
            {value}
          </div>
        )}
        <div className="mt-3 text-xs text-muted-foreground min-h-[1.25rem]">
          {hint ?? " "}
        </div>
      </CardContent>
    </Card>
  );
  return href ? (
    <Link href={href} className="block h-full">
      {body}
    </Link>
  ) : (
    body
  );
}
