"use client";

import { Filter, RefreshCw, Search } from "lucide-react";
import { useMemo, useState } from "react";

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
import { Skeleton } from "@/components/ui/skeleton";
import { useAuditLogs } from "@/hooks/audit/useAuditQueries";
import type { AuditLog } from "@/types/audit";

const METHODS = ["", "GET", "POST", "PUT", "PATCH", "DELETE"] as const;

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

export default function AuditPage() {
  const [method, setMethod] = useState<string>("");
  const [search, setSearch] = useState("");
  const [statusGroup, setStatusGroup] = useState<"" | "2xx" | "4xx" | "5xx">("");
  const [selected, setSelected] = useState<AuditLog | null>(null);

  const statusFrom = statusGroup === "2xx" ? 200 : statusGroup === "4xx" ? 400 : statusGroup === "5xx" ? 500 : undefined;
  const statusTo = statusGroup === "2xx" ? 299 : statusGroup === "4xx" ? 499 : statusGroup === "5xx" ? 599 : undefined;

  const logsQ = useAuditLogs({
    method: method || undefined,
    path: search || undefined,
    statusFrom,
    statusTo,
    limit: 200,
  });

  const stats = useMemo(() => {
    const data = logsQ.data ?? [];
    const total = data.length;
    const successes = data.filter((l) => (l.statusCode ?? 0) >= 200 && (l.statusCode ?? 0) < 400).length;
    const failures = data.filter((l) => (l.statusCode ?? 0) >= 400).length;
    const avgLatency = total > 0 ? Math.round(data.reduce((s, l) => s + (l.latencyMs ?? 0), 0) / total) : 0;
    return { total, successes, failures, avgLatency };
  }, [logsQ.data]);

  // Build a 24-bucket histogram by hour of day.
  const histogram = useMemo(() => {
    const buckets = Array.from({ length: 24 }, () => 0);
    for (const l of logsQ.data ?? []) {
      try {
        const h = new Date(l.occurredAt).getHours();
        buckets[h]++;
      } catch {
        /* skip */
      }
    }
    const max = Math.max(1, ...buckets);
    return { buckets, max };
  }, [logsQ.data]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Audit Log</h1>
          <p className="text-muted-foreground mt-1">
            Every API request that hit this tenant, captured asynchronously.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => logsQ.refetch()} disabled={logsQ.isFetching}>
          <RefreshCw className={"h-4 w-4 " + (logsQ.isFetching ? "animate-spin" : "")} />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <Stat label="Events" value={stats.total} />
        <Stat label="Success" value={stats.successes} accent="emerald" />
        <Stat label="Errors" value={stats.failures} accent="rose" />
        <Stat label="Avg latency" value={stats.avgLatency + "ms"} accent="amber" />
      </div>

      {/* Chart */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Requests by hour of day</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-end gap-1 h-32">
            {histogram.buckets.map((v, i) => (
              <div key={i} className="flex-1 flex flex-col items-center justify-end gap-1">
                <div
                  className="w-full rounded-t bg-primary/70 transition-all"
                  style={{ height: `${(v / histogram.max) * 100}%`, minHeight: v > 0 ? "4px" : "0" }}
                  title={`${i}:00 — ${v} events`}
                />
                <span className="text-[10px] text-muted-foreground">
                  {i % 4 === 0 ? i : ""}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Filters */}
      <Card className="overflow-hidden">
        <div className="flex flex-col gap-3 p-4 md:flex-row md:items-center md:justify-between border-b">
          <div className="flex flex-wrap items-center gap-2">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <span className="text-xs text-muted-foreground uppercase tracking-wider">Method:</span>
            {METHODS.map((m) => (
              <button
                key={m || "all"}
                onClick={() => setMethod(m)}
                className={
                  "px-2.5 h-7 rounded-md text-xs font-medium transition-colors " +
                  (method === m
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:bg-accent")
                }
              >
                {m || "All"}
              </button>
            ))}
            <span className="text-xs text-muted-foreground uppercase tracking-wider ml-2">Status:</span>
            {(["", "2xx", "4xx", "5xx"] as const).map((s) => (
              <button
                key={s || "all"}
                onClick={() => setStatusGroup(s)}
                className={
                  "px-2.5 h-7 rounded-md text-xs font-medium transition-colors " +
                  (statusGroup === s
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:bg-accent")
                }
              >
                {s || "All"}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-2 md:max-w-xs md:flex-1">
            <Search className="h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Filter by path…"
              className="flex-1"
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          {logsQ.isLoading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : !logsQ.data?.length ? (
            <div className="p-10 text-center text-sm text-muted-foreground">
              No audit events match these filters. New events show up here within seconds.
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-muted/40 text-xs uppercase tracking-wider text-muted-foreground">
                <tr>
                  <th className="text-left p-3 font-medium">Time</th>
                  <th className="text-left p-3 font-medium">Method</th>
                  <th className="text-left p-3 font-medium">Path</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">User</th>
                  <th className="text-left p-3 font-medium">Latency</th>
                </tr>
              </thead>
              <tbody>
                {logsQ.data.map((l) => (
                  <tr
                    key={l.id}
                    className="border-t hover:bg-muted/30 cursor-pointer"
                    onClick={() => setSelected(l)}
                  >
                    <td className="p-3 text-muted-foreground whitespace-nowrap">
                      {new Date(l.occurredAt).toLocaleString()}
                    </td>
                    <td className="p-3">
                      <span
                        className={
                          "inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-semibold " +
                          methodColor(l.method)
                        }
                      >
                        {l.method}
                      </span>
                    </td>
                    <td className="p-3 font-mono text-xs truncate max-w-[280px]">{l.path}</td>
                    <td className="p-3">
                      <Badge variant={statusVariant(l.statusCode)}>{l.statusCode}</Badge>
                    </td>
                    <td className="p-3 text-muted-foreground truncate max-w-[200px]">
                      {l.userEmail ?? "—"}
                    </td>
                    <td className="p-3 text-muted-foreground">{l.latencyMs}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </Card>

      {/* Detail dialog */}
      <Dialog open={!!selected} onOpenChange={(o) => !o && setSelected(null)}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Event detail</DialogTitle>
          </DialogHeader>
          {selected && (
            <div className="space-y-4 text-sm max-h-[70vh] overflow-y-auto">
              <div className="grid grid-cols-2 gap-3">
                <Detail label="Method" value={selected.method ?? "—"} />
                <Detail label="Path" value={selected.path ?? "—"} mono />
                <Detail label="Status" value={String(selected.statusCode ?? "—")} />
                <Detail label="Latency" value={(selected.latencyMs ?? 0) + "ms"} />
                <Detail label="User" value={selected.userEmail ?? "—"} />
                <Detail label="IP" value={selected.ip ?? "—"} mono />
                <Detail label="Time" value={new Date(selected.occurredAt).toLocaleString()} />
                <Detail label="Correlation ID" value={selected.correlationId ?? "—"} mono />
              </div>
              {selected.requestBody != null && (
                <Section label="Request body">
                  <pre className="text-xs bg-muted/40 rounded p-3 overflow-x-auto whitespace-pre-wrap">
                    {String(JSON.stringify(selected.requestBody, null, 2))}
                  </pre>
                </Section>
              )}
              {selected.responseBody != null && (
                <Section label="Response body">
                  <pre className="text-xs bg-muted/40 rounded p-3 overflow-x-auto whitespace-pre-wrap">
                    {String(JSON.stringify(selected.responseBody, null, 2))}
                  </pre>
                </Section>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Stat({
  label,
  value,
  accent,
}: {
  label: string;
  value: string | number;
  accent?: "emerald" | "amber" | "rose";
}) {
  const dot =
    accent === "emerald"
      ? "bg-emerald-500"
      : accent === "amber"
      ? "bg-amber-500"
      : accent === "rose"
      ? "bg-rose-500"
      : "bg-muted-foreground";
  return (
    <Card className="p-4">
      <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-muted-foreground">
        <span className={"h-2 w-2 rounded-full " + dot} />
        {label}
      </div>
      <div className="mt-2 text-2xl font-semibold">{value}</div>
    </Card>
  );
}

function Detail({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={"mt-1 " + (mono ? "font-mono text-xs break-all" : "font-medium")}>{value}</div>
    </div>
  );
}

function Section({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">{label}</div>
      {children}
    </div>
  );
}
