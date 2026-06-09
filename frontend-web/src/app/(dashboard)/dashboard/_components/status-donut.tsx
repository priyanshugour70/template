"use client";

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

import type { StatusSlice } from "@/types/dashboard";

import { formatNumber } from "./format";

const COLORS: Record<string, string> = {
  "2xx": "#10b981",
  "3xx": "#0ea5e9",
  "4xx": "#f59e0b",
  "5xx": "#ef4444",
  other: "#94a3b8",
};

// StatusDonut — read-at-a-glance HTTP status distribution. The big number in
// the middle is the total; the slices show the split. Legend renders below
// the chart as a compact key.
export function StatusDonut({ data }: { data: StatusSlice[] }) {
  const total = data.reduce((sum, d) => sum + d.count, 0);
  const errorPct = total > 0
    ? (((data.find((d) => d.group === "5xx")?.count ?? 0) +
       (data.find((d) => d.group === "4xx")?.count ?? 0)) /
        total) * 100
    : 0;

  if (total === 0) {
    return (
      <div className="h-[220px] flex items-center justify-center text-sm text-muted-foreground">
        No requests in the last 14 days.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="relative">
        <ResponsiveContainer width="100%" height={180}>
          <PieChart>
            <Pie
              data={data}
              dataKey="count"
              nameKey="group"
              innerRadius={56}
              outerRadius={80}
              paddingAngle={2}
              stroke="var(--background)"
              strokeWidth={2}
            >
              {data.map((slice) => (
                <Cell key={slice.group} fill={COLORS[slice.group] ?? "#94a3b8"} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                background: "var(--background)",
                border: "1px solid var(--border)",
                borderRadius: 8,
                fontSize: 12,
              }}
              formatter={(v: unknown, name) => [formatNumber(Number(v)), name as string]}
            />
          </PieChart>
        </ResponsiveContainer>
        <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
          <div className="text-2xl font-semibold">{formatNumber(total)}</div>
          <div className="text-[10px] uppercase tracking-wide text-muted-foreground">
            requests
          </div>
        </div>
      </div>
      <div className="flex flex-wrap justify-center gap-x-3 gap-y-1 text-xs">
        {data.map((slice) => (
          <div key={slice.group} className="flex items-center gap-1.5">
            <span
              className="inline-block w-2 h-2 rounded-full"
              style={{ background: COLORS[slice.group] ?? "#94a3b8" }}
            />
            <span className="font-medium">{slice.group}</span>
            <span className="text-muted-foreground">{formatNumber(slice.count)}</span>
          </div>
        ))}
      </div>
      {errorPct > 1 && (
        <div className="text-center text-xs text-amber-600 dark:text-amber-400">
          {errorPct.toFixed(1)}% error rate over the window
        </div>
      )}
    </div>
  );
}
