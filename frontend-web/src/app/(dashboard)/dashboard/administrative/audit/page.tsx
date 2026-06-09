"use client";

import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Download,
  Filter,
  RefreshCw,
  Search,
  Users,
  X,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useAuditLogs,
  useAuditStats,
  useAuditStatusBreakdown,
  useAuditTimeseries,
  useAuditTopActions,
  useAuditTopFailingPaths,
  useAuditTopUsers,
} from "@/hooks/audit/useAuditQueries";
import { PaginationBar } from "@/components/shared/pagination-bar";
import { auditService } from "@/services/audit";
import { cn } from "@/lib/cn";
import type {
  AuditLog,
  AuditStatsFilter,
  AuditTimeInterval,
} from "@/types/audit";

// ── helpers ───────────────────────────────────────────────────────────────

type RangePreset = "24h" | "7d" | "30d";

function presetRange(preset: RangePreset): { from: string; to: string; interval: AuditTimeInterval } {
  const to = new Date();
  const from = new Date(to);
  let interval: AuditTimeInterval = "hour";
  switch (preset) {
    case "24h":
      from.setHours(from.getHours() - 24);
      interval = "hour";
      break;
    case "7d":
      from.setDate(from.getDate() - 7);
      interval = "hour";
      break;
    case "30d":
      from.setDate(from.getDate() - 30);
      interval = "day";
      break;
  }
  return { from: from.toISOString(), to: to.toISOString(), interval };
}

function statusVariant(s?: number) {
  if (!s) return "muted" as const;
  if (s >= 500) return "danger" as const;
  if (s >= 400) return "warning" as const;
  if (s >= 200 && s < 300) return "success" as const;
  return "muted" as const;
}

function methodColor(m?: string) {
  switch (m) {
    case "GET":
      return "bg-blue-500/10 text-blue-700 dark:text-blue-300";
    case "POST":
      return "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
    case "PUT":
    case "PATCH":
      return "bg-amber-500/10 text-amber-700 dark:text-amber-300";
    case "DELETE":
      return "bg-rose-500/10 text-rose-700 dark:text-rose-300";
    default:
      return "bg-muted text-muted-foreground";
  }
}

function formatBucket(iso: string, interval: AuditTimeInterval): string {
  try {
    const d = new Date(iso);
    if (interval === "day" || interval === "week") {
      return d.toLocaleDateString("en-IN", { day: "numeric", month: "short" });
    }
    return d.toLocaleTimeString("en-IN", { hour: "numeric", minute: "2-digit" });
  } catch {
    return iso;
  }
}

// ── main page ─────────────────────────────────────────────────────────────

export default function AuditPage() {
  const [preset, setPreset] = useState<RangePreset>("24h");
  const [method, setMethod] = useState<string>("");
  const [userEmail, setUserEmail] = useState<string>("");
  const [actionFilter, setActionFilter] = useState<string>("");
  const [pathFilter, setPathFilter] = useState<string>("");
  const [statusGroup, setStatusGroup] = useState<"" | "2xx" | "4xx" | "5xx">("");
  const [selected, setSelected] = useState<AuditLog | null>(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);

  const range = useMemo(() => presetRange(preset), [preset]);

  const statsFilter = useMemo<AuditStatsFilter>(
    () => ({
      from: range.from,
      to: range.to,
      method: method || undefined,
      userEmail: userEmail || undefined,
      action: actionFilter || undefined,
      path: pathFilter || undefined,
    }),
    [range.from, range.to, method, userEmail, actionFilter, pathFilter],
  );

  // List filter for the Activity tab — same shape + status range from the
  // selected status group.
  const statusFrom =
    statusGroup === "2xx" ? 200 : statusGroup === "4xx" ? 400 : statusGroup === "5xx" ? 500 : undefined;
  const statusTo =
    statusGroup === "2xx" ? 299 : statusGroup === "4xx" ? 499 : statusGroup === "5xx" ? 599 : undefined;

  const listQuery = useMemo(
    () => ({
      ...statsFilter,
      statusFrom,
      statusTo,
      page,
      limit,
    }),
    [statsFilter, statusFrom, statusTo, page, limit],
  );

  // Reset to page 1 whenever a filter that affects the row-set changes.
  useEffect(() => {
    setPage(1);
  }, [statsFilter, statusFrom, statusTo, limit]);

  const statsQ = useAuditStats(statsFilter);
  const tsQ = useAuditTimeseries({ ...statsFilter, interval: range.interval });
  const topUsersQ = useAuditTopUsers(statsFilter, 10);
  const topPathsQ = useAuditTopFailingPaths(statsFilter, 10);
  const topActionsQ = useAuditTopActions(statsFilter, 10);
  const statusQ = useAuditStatusBreakdown(statsFilter);
  const logsQ = useAuditLogs(listQuery);

  const activeFilterCount =
    [method, userEmail, actionFilter, pathFilter, statusGroup].filter((v) => v !== "").length;
  const clearFilters = () => {
    setMethod("");
    setUserEmail("");
    setActionFilter("");
    setPathFilter("");
    setStatusGroup("");
  };

  const refetchAll = () => {
    statsQ.refetch();
    tsQ.refetch();
    topUsersQ.refetch();
    topPathsQ.refetch();
    topActionsQ.refetch();
    statusQ.refetch();
    logsQ.refetch();
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">Audit log</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Every API request that hit this tenant, captured asynchronously.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={preset} onValueChange={(v) => setPreset(v as RangePreset)}>
            <SelectTrigger className="w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="24h">Last 24 hours</SelectItem>
              <SelectItem value="7d">Last 7 days</SelectItem>
              <SelectItem value="30d">Last 30 days</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm" onClick={refetchAll} disabled={statsQ.isFetching}>
            <RefreshCw className={cn("h-4 w-4", statsQ.isFetching && "animate-spin")} />
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              const url = auditService.exportUrl(listQuery);
              window.open(url, "_blank");
            }}
          >
            <Download className="h-4 w-4" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Filter sidebar (sticky chip row) */}
      <Card>
        <CardContent className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 lg:grid-cols-5">
          <Select value={method || "_any"} onValueChange={(v) => setMethod(v === "_any" ? "" : v)}>
            <SelectTrigger>
              <SelectValue placeholder="Any method" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_any">Any method</SelectItem>
              {["GET", "POST", "PUT", "PATCH", "DELETE"].map((m) => (
                <SelectItem key={m} value={m}>
                  {m}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            value={statusGroup || "_any"}
            onValueChange={(v) =>
              setStatusGroup(v === "_any" ? "" : (v as "2xx" | "4xx" | "5xx"))
            }
          >
            <SelectTrigger>
              <SelectValue placeholder="Any status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_any">Any status</SelectItem>
              <SelectItem value="2xx">2xx success</SelectItem>
              <SelectItem value="4xx">4xx client error</SelectItem>
              <SelectItem value="5xx">5xx server error</SelectItem>
            </SelectContent>
          </Select>
          <Input
            placeholder="User email"
            value={userEmail}
            onChange={(e) => setUserEmail(e.target.value)}
          />
          <Input
            placeholder="Action (e.g. user.create)"
            value={actionFilter}
            onChange={(e) => setActionFilter(e.target.value)}
          />
          <Input
            placeholder="Path contains…"
            value={pathFilter}
            onChange={(e) => setPathFilter(e.target.value)}
          />
        </CardContent>
        {activeFilterCount > 0 && (
          <CardContent className="flex items-center justify-between border-t border-border px-4 py-2 text-xs text-muted-foreground">
            <span>
              <Filter className="mr-1 inline h-3 w-3" />
              {activeFilterCount} filter{activeFilterCount > 1 ? "s" : ""} active
            </span>
            <Button variant="ghost" size="sm" onClick={clearFilters}>
              <X className="h-3 w-3" />
              Clear
            </Button>
          </CardContent>
        )}
      </Card>

      {/* Stats */}
      {statsQ.isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5">
          <StatCard
            label="Total requests"
            value={(statsQ.data?.totalRequests ?? 0).toLocaleString()}
            icon={Activity}
          />
          <StatCard
            label="Success"
            value={(statsQ.data?.success2xx ?? 0).toLocaleString()}
            icon={CheckCircle2}
            accent="success"
          />
          <StatCard
            label="Error rate"
            value={`${(statsQ.data?.errorRatePct ?? 0).toFixed(2)}%`}
            icon={AlertTriangle}
            accent={
              (statsQ.data?.errorRatePct ?? 0) > 5
                ? "danger"
                : (statsQ.data?.errorRatePct ?? 0) > 1
                  ? "warning"
                  : "muted"
            }
          />
          <StatCard
            label="P95 latency"
            value={`${Math.round(statsQ.data?.p95LatencyMs ?? 0)} ms`}
            icon={Clock}
          />
          <StatCard
            label="Unique users"
            value={(statsQ.data?.uniqueUsers ?? 0).toLocaleString()}
            icon={Users}
          />
        </div>
      )}

      <Tabs defaultValue="overview" className="w-full">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="activity">
            Activity
            {(logsQ.data?.total ?? 0) > 0 && (
              <Badge variant="muted" className="ml-2">
                {logsQ.data?.total.toLocaleString()}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid gap-4">
            {/* Timeseries */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base">Requests over time</CardTitle>
              </CardHeader>
              <CardContent>
                {tsQ.isLoading ? (
                  <Skeleton className="h-64 w-full" />
                ) : (tsQ.data?.length ?? 0) === 0 ? (
                  <p className="py-8 text-center text-sm text-muted-foreground">
                    No activity in the selected range.
                  </p>
                ) : (
                  <div className="h-64 w-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart
                        data={(tsQ.data ?? []).map((b) => ({
                          ...b,
                          label: formatBucket(b.bucket, range.interval),
                        }))}
                      >
                        <defs>
                          <linearGradient id="g-2xx" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="hsl(142,71%,45%)" stopOpacity={0.7} />
                            <stop offset="95%" stopColor="hsl(142,71%,45%)" stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id="g-4xx" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="hsl(38,92%,50%)" stopOpacity={0.7} />
                            <stop offset="95%" stopColor="hsl(38,92%,50%)" stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id="g-5xx" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="hsl(0,72%,55%)" stopOpacity={0.7} />
                            <stop offset="95%" stopColor="hsl(0,72%,55%)" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <XAxis
                          dataKey="label"
                          fontSize={11}
                          tick={{ fill: "var(--muted-foreground)" }}
                        />
                        <YAxis fontSize={11} tick={{ fill: "var(--muted-foreground)" }} />
                        <Tooltip
                          contentStyle={{
                            background: "var(--popover)",
                            border: "1px solid var(--border)",
                            borderRadius: 6,
                            fontSize: 12,
                          }}
                        />
                        <Legend wrapperStyle={{ fontSize: 11 }} />
                        <Area
                          type="monotone"
                          dataKey="success2xx"
                          name="2xx"
                          stackId="1"
                          stroke="hsl(142,71%,45%)"
                          fill="url(#g-2xx)"
                        />
                        <Area
                          type="monotone"
                          dataKey="clientError4xx"
                          name="4xx"
                          stackId="1"
                          stroke="hsl(38,92%,50%)"
                          fill="url(#g-4xx)"
                        />
                        <Area
                          type="monotone"
                          dataKey="serverError5xx"
                          name="5xx"
                          stackId="1"
                          stroke="hsl(0,72%,55%)"
                          fill="url(#g-5xx)"
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                )}
              </CardContent>
            </Card>

            <div className="grid gap-4 lg:grid-cols-2">
              {/* Status donut */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">Status code distribution</CardTitle>
                </CardHeader>
                <CardContent>
                  {statusQ.isLoading ? (
                    <Skeleton className="h-56 w-full" />
                  ) : (statusQ.data?.length ?? 0) === 0 ? (
                    <p className="py-8 text-center text-sm text-muted-foreground">No data</p>
                  ) : (
                    <StatusDonut rows={statusQ.data ?? []} />
                  )}
                </CardContent>
              </Card>

              {/* Top users */}
              <TopBarCard
                title="Top users by activity"
                rows={topUsersQ.data ?? []}
                loading={topUsersQ.isLoading}
                emptyMessage="No user activity yet"
                color="hsl(217,91%,60%)"
              />

              {/* Top failing paths */}
              <TopBarCard
                title="Top failing paths"
                rows={topPathsQ.data ?? []}
                loading={topPathsQ.isLoading}
                emptyMessage="No failures in this range"
                color="hsl(0,72%,55%)"
              />

              {/* Top actions */}
              <TopBarCard
                title="Top actions"
                rows={topActionsQ.data ?? []}
                loading={topActionsQ.isLoading}
                emptyMessage="No labeled actions yet"
                color="hsl(262,83%,58%)"
              />
            </div>
          </div>
        </TabsContent>

        <TabsContent value="activity">
          <ActivityTable
            logs={logsQ.data?.items ?? []}
            loading={logsQ.isLoading}
            onRowClick={setSelected}
          />
          <div className="mt-4">
            <PaginationBar
              page={page}
              limit={limit}
              total={logsQ.data?.total ?? 0}
              onPageChange={setPage}
              onLimitChange={setLimit}
            />
          </div>
        </TabsContent>
      </Tabs>

      {/* Detail dialog */}
      <Dialog open={!!selected} onOpenChange={(o) => !o && setSelected(null)}>
        <DialogContent className="max-w-3xl">
          {selected && (
            <>
              <DialogHeader>
                <DialogTitle className="font-mono">
                  {selected.method} {selected.path}
                </DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <Detail label="Status" value={String(selected.statusCode ?? "—")} />
                  <Detail label="Latency" value={(selected.latencyMs ?? 0) + "ms"} />
                  <Detail label="User" value={selected.userEmail ?? "—"} />
                  <Detail label="IP" value={selected.ip ?? "—"} mono />
                  <Detail
                    label="Time"
                    value={new Date(selected.occurredAt).toLocaleString()}
                  />
                  <Detail label="Correlation ID" value={selected.correlationId ?? "—"} mono />
                  {selected.action && <Detail label="Action" value={selected.action} mono />}
                  {selected.targetType && (
                    <Detail label="Target type" value={selected.targetType} />
                  )}
                </div>
                {selected.requestBody != null && (
                  <Section label="Request body">
                    <pre className="overflow-x-auto whitespace-pre-wrap rounded bg-muted/40 p-3 text-xs">
                      {String(JSON.stringify(selected.requestBody, null, 2))}
                    </pre>
                  </Section>
                )}
                {selected.responseBody != null && (
                  <Section label="Response body">
                    <pre className="overflow-x-auto whitespace-pre-wrap rounded bg-muted/40 p-3 text-xs">
                      {String(JSON.stringify(selected.responseBody, null, 2))}
                    </pre>
                  </Section>
                )}
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ── stat card ─────────────────────────────────────────────────────────────

function StatCard({
  label,
  value,
  icon: Icon,
  accent,
}: {
  label: string;
  value: string;
  icon: typeof Activity;
  accent?: "success" | "warning" | "danger" | "muted";
}) {
  const accentClass = {
    success: "text-success",
    warning: "text-warning",
    danger: "text-destructive",
    muted: "text-muted-foreground",
  }[accent ?? "muted"];
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center gap-1.5 text-xs uppercase tracking-wider text-muted-foreground">
          <Icon className={cn("h-3 w-3", accent && accentClass)} />
          {label}
        </div>
        <div className={cn("mt-1.5 text-2xl font-semibold tabular-nums", accent && accentClass)}>
          {value}
        </div>
      </CardContent>
    </Card>
  );
}

// ── status donut ──────────────────────────────────────────────────────────

const STATUS_COLORS = {
  "2xx": "hsl(142,71%,45%)",
  "3xx": "hsl(217,91%,60%)",
  "4xx": "hsl(38,92%,50%)",
  "5xx": "hsl(0,72%,55%)",
};

function StatusDonut({ rows }: { rows: { key: string; count: number }[] }) {
  // Group by status class for the donut.
  const grouped = useMemo(() => {
    const m: Record<string, number> = { "2xx": 0, "3xx": 0, "4xx": 0, "5xx": 0 };
    for (const r of rows) {
      const code = parseInt(r.key, 10);
      if (code >= 200 && code < 300) m["2xx"] += r.count;
      else if (code >= 300 && code < 400) m["3xx"] += r.count;
      else if (code >= 400 && code < 500) m["4xx"] += r.count;
      else if (code >= 500) m["5xx"] += r.count;
    }
    return Object.entries(m)
      .filter(([, v]) => v > 0)
      .map(([k, v]) => ({ name: k, value: v }));
  }, [rows]);

  if (grouped.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        No status codes recorded.
      </p>
    );
  }
  return (
    <div className="h-56 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={grouped}
            cx="50%"
            cy="50%"
            innerRadius={55}
            outerRadius={85}
            paddingAngle={2}
            dataKey="value"
            label={({ name, value }) => `${name}: ${value}`}
            labelLine={false}
            fontSize={11}
          >
            {grouped.map((entry) => (
              <Cell
                key={entry.name}
                fill={STATUS_COLORS[entry.name as keyof typeof STATUS_COLORS]}
              />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              background: "var(--popover)",
              border: "1px solid var(--border)",
              borderRadius: 6,
              fontSize: 12,
            }}
          />
          <Legend wrapperStyle={{ fontSize: 11 }} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}

// ── top-N bar card ────────────────────────────────────────────────────────

function TopBarCard({
  title,
  rows,
  loading,
  emptyMessage,
  color,
}: {
  title: string;
  rows: { key: string; count: number }[];
  loading: boolean;
  emptyMessage: string;
  color: string;
}) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-56 w-full" />
        ) : rows.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">{emptyMessage}</p>
        ) : (
          <div className="h-56 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                layout="vertical"
                data={rows.map((r) => ({
                  ...r,
                  // Truncate keys so the Y axis stays readable
                  key: r.key.length > 36 ? `…${r.key.slice(-33)}` : r.key,
                }))}
                margin={{ left: 8, right: 16 }}
              >
                <XAxis
                  type="number"
                  fontSize={11}
                  tick={{ fill: "var(--muted-foreground)" }}
                />
                <YAxis
                  type="category"
                  dataKey="key"
                  width={140}
                  fontSize={11}
                  tick={{ fill: "var(--muted-foreground)" }}
                />
                <Tooltip
                  contentStyle={{
                    background: "var(--popover)",
                    border: "1px solid var(--border)",
                    borderRadius: 6,
                    fontSize: 12,
                  }}
                />
                <Bar dataKey="count" fill={color} radius={[0, 3, 3, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ── activity table ────────────────────────────────────────────────────────

function ActivityTable({
  logs,
  loading,
  onRowClick,
}: {
  logs: AuditLog[];
  loading: boolean;
  onRowClick: (l: AuditLog) => void;
}) {
  if (loading) return <Skeleton className="h-64 w-full" />;
  if (logs.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center gap-2 py-10 text-center text-sm text-muted-foreground">
          <Search className="h-8 w-8" />
          No requests match the current filters.
        </CardContent>
      </Card>
    );
  }
  return (
    <Card>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[64px]">Method</TableHead>
              <TableHead>Path</TableHead>
              <TableHead className="w-[80px]">Status</TableHead>
              <TableHead className="hidden md:table-cell">User</TableHead>
              <TableHead className="hidden md:table-cell">Latency</TableHead>
              <TableHead>Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {logs.map((l) => (
              <TableRow
                key={l.id}
                className="cursor-pointer"
                onClick={() => onRowClick(l)}
              >
                <TableCell>
                  <span
                    className={cn(
                      "rounded px-1.5 py-0.5 font-mono text-[10px] font-semibold",
                      methodColor(l.method),
                    )}
                  >
                    {l.method}
                  </span>
                </TableCell>
                <TableCell className="font-mono text-xs">{l.path}</TableCell>
                <TableCell>
                  <Badge variant={statusVariant(l.statusCode)}>
                    {l.statusCode ?? "—"}
                  </Badge>
                </TableCell>
                <TableCell className="hidden text-muted-foreground md:table-cell">
                  {l.userEmail ?? "—"}
                </TableCell>
                <TableCell className="hidden tabular-nums text-muted-foreground md:table-cell">
                  {(l.latencyMs ?? 0).toLocaleString()} ms
                </TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {new Date(l.occurredAt).toLocaleString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function Detail({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn("mt-1", mono ? "break-all font-mono text-xs" : "font-medium")}>
        {value}
      </div>
    </div>
  );
}

function Section({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="mb-1.5 text-xs uppercase tracking-wider text-muted-foreground">
        {label}
      </div>
      {children}
    </div>
  );
}
