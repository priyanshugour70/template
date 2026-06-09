"use client";

import type { EndpointBucket } from "@/types/dashboard";

import { formatNumber } from "./format";

// TopEndpoints — compact horizontal-bar list, no recharts (recharts'
// horizontal Bar overdoes it for this list shape). Each row shows method
// pill, route, count, and avg latency. Bar width is each row's count over
// the max count so the user can eyeball the distribution.
export function TopEndpoints({ rows }: { rows: EndpointBucket[] }) {
  if (rows.length === 0) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        No endpoint activity yet.
      </div>
    );
  }
  const max = Math.max(...rows.map((r) => r.count));

  return (
    <ul className="space-y-2">
      {rows.map((r, i) => (
        <li key={r.method + r.route + i} className="space-y-1">
          <div className="flex items-center justify-between gap-3 text-xs">
            <div className="flex items-center gap-2 min-w-0">
              <MethodPill method={r.method} />
              <span className="font-mono truncate" title={r.route}>
                {r.route}
              </span>
            </div>
            <div className="flex items-center gap-3 shrink-0 text-muted-foreground">
              <span className="tabular-nums">{r.avgLatencyMs}ms</span>
              <span className="tabular-nums font-medium text-foreground">
                {formatNumber(r.count)}
              </span>
            </div>
          </div>
          <div className="h-1 w-full rounded-full bg-muted overflow-hidden">
            <div
              className="h-full bg-indigo-500"
              style={{ width: max > 0 ? `${(r.count / max) * 100}%` : "0%" }}
            />
          </div>
        </li>
      ))}
    </ul>
  );
}

function MethodPill({ method }: { method: string }) {
  const colors: Record<string, string> = {
    GET: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400",
    POST: "bg-sky-500/15 text-sky-700 dark:text-sky-400",
    PATCH: "bg-amber-500/15 text-amber-700 dark:text-amber-400",
    PUT: "bg-violet-500/15 text-violet-700 dark:text-violet-400",
    DELETE: "bg-red-500/15 text-red-700 dark:text-red-400",
  };
  return (
    <span
      className={`inline-block w-12 text-center rounded font-mono text-[10px] py-0.5 font-medium ${
        colors[method] ?? "bg-muted text-muted-foreground"
      }`}
    >
      {method}
    </span>
  );
}
