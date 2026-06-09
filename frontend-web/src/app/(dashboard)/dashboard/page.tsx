"use client";

import { AlertTriangle, RefreshCcw } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useDashboardSummary } from "@/hooks/dashboard/useDashboard";
import { useAuth, useTenant } from "@/providers";

import { ActivityFeed } from "./_components/activity-feed";
import { AgingChart } from "./_components/aging-chart";
import { KPITile } from "./_components/kpi-tile";
import { RequestsChart } from "./_components/requests-chart";
import { RevenueChart } from "./_components/revenue-chart";
import { StatusDonut } from "./_components/status-donut";
import { TopEndpoints } from "./_components/top-endpoints";
import { formatCompactMoney, formatMoney, formatNumber } from "./_components/format";

// DashboardHome — the operator overview. Built from a single GET
// /api/v1/dashboard/summary call that fans out concurrently on the backend
// (errgroup), so 4 KPIs + 5 charts + a feed all hydrate in one round-trip.
// Each panel handles its own loading + empty state so the page never blanks.
export default function DashboardHome() {
  const { user } = useAuth();
  const { tenant, activeOrganization } = useTenant();
  const q = useDashboardSummary();

  const data = q.data;
  const loading = q.isLoading;

  return (
    <div className="space-y-6">
      {/* Greeting */}
      <div className="flex items-end justify-between flex-wrap gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Welcome back{user?.firstName ? `, ${user.firstName}` : ""}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {tenant?.name ? <span className="font-medium text-foreground">{tenant.name}</span> : "—"}
            {activeOrganization ? <> · {activeOrganization.name}</> : null}
            {data?.generatedAt ? (
              <span className="ml-2 text-xs text-muted-foreground/80">
                · refreshed {new Date(data.generatedAt).toLocaleTimeString()}
              </span>
            ) : null}
          </p>
        </div>
        <button
          type="button"
          onClick={() => q.refetch()}
          disabled={q.isFetching}
          className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs hover:bg-accent disabled:opacity-50"
        >
          <RefreshCcw className={`h-3 w-3 ${q.isFetching ? "animate-spin" : ""}`} />
          {q.isFetching ? "Refreshing…" : "Refresh"}
        </button>
      </div>

      {q.isError && (
        <Card>
          <CardContent className="flex items-center gap-3 py-4 text-sm">
            <AlertTriangle className="h-4 w-4 text-destructive" />
            <span>Couldn&apos;t load dashboard data: {(q.error as Error).message}</span>
          </CardContent>
        </Card>
      )}

      {/* KPI tiles */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KPITile
          title="MRR"
          value={data ? formatMoney(data.kpis.mrrCents) : "—"}
          loading={loading}
          subline="active subscriptions"
          href="/dashboard/billing/subscription"
        />
        <KPITile
          title="Invoiced this month"
          value={data ? formatMoney(data.kpis.invoicedThisMonthCents) : "—"}
          deltaPct={data?.kpis.invoicedDeltaPct}
          loading={loading}
          subline="vs previous month"
          href="/dashboard/billing/invoices"
        />
        <KPITile
          title="Active users · 7d"
          value={data ? formatNumber(data.kpis.activeUsers7d) : "—"}
          deltaPct={data?.kpis.activeUsersDeltaPct}
          loading={loading}
          subline="vs previous week"
          href="/dashboard/administrative/users"
        />
        <KPITile
          title="Outstanding"
          value={data ? formatMoney(data.kpis.outstandingDueCents) : "—"}
          loading={loading}
          subline={data ? `${data.kpis.openInvoiceCount} open invoice${data.kpis.openInvoiceCount === 1 ? "" : "s"}` : ""}
          href="/dashboard/billing/invoices"
        />
      </div>

      {/* Revenue chart — full width because it's the headline business metric */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-base">Revenue trend</CardTitle>
              <CardDescription>Last 12 months · issued vs paid</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-[260px] w-full" />
          ) : data ? (
            <RevenueChart data={data.revenueByMonth} />
          ) : null}
        </CardContent>
      </Card>

      {/* Requests + status donut row */}
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Request activity</CardTitle>
            <CardDescription>Last 14 days · audit-log signals</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-[220px] w-full" />
            ) : data ? (
              <RequestsChart data={data.requestsByDay} />
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">HTTP status mix</CardTitle>
            <CardDescription>Last 14 days</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-[220px] w-full" />
            ) : data ? (
              <StatusDonut data={data.statusBreakdown} />
            ) : null}
          </CardContent>
        </Card>
      </div>

      {/* Top endpoints + invoice aging row */}
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Top endpoints</CardTitle>
            <CardDescription>Most-called routes in the last 14 days</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-[220px] w-full" />
            ) : data ? (
              <TopEndpoints rows={data.topEndpoints} />
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Invoice aging</CardTitle>
            <CardDescription>Open invoices by days past due</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-[180px] w-full" />
            ) : data ? (
              <AgingChart data={data.invoiceAging} />
            ) : null}
            {data && data.kpis.outstandingDueCents > 0 && (
              <div className="text-center text-xs text-muted-foreground mt-3">
                Total due: <span className="font-medium text-foreground">{formatCompactMoney(data.kpis.outstandingDueCents)}</span>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent activity */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Recent activity</CardTitle>
          <CardDescription>Latest audit events</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-[200px] w-full" />
          ) : data ? (
            <ActivityFeed entries={data.recentActivity} />
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
