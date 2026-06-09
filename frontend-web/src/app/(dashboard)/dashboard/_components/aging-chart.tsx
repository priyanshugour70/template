"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import type { AgingBucket } from "@/types/dashboard";

import { formatCompactMoney, formatMoney } from "./format";

// AgingChart — open-invoice aging visual. Single-series bar chart, colour
// per bucket so "90+" jumps out red even at a glance.
const COLORS: Record<string, string> = {
  current: "#10b981",
  "1-30": "#84cc16",
  "31-60": "#eab308",
  "61-90": "#f97316",
  "90+": "#ef4444",
};

export function AgingChart({ data }: { data: AgingBucket[] }) {
  const total = data.reduce((sum, d) => sum + d.count, 0);
  if (total === 0) {
    return (
      <div className="h-[180px] flex items-center justify-center text-sm text-muted-foreground">
        No open invoices.
      </div>
    );
  }
  const chartData = data.map((d) => ({
    bucket: d.bucket,
    due: d.totalDueCents / 100,
    count: d.count,
  }));

  return (
    <ResponsiveContainer width="100%" height={180}>
      <BarChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" vertical={false} className="stroke-border" />
        <XAxis dataKey="bucket" tickLine={false} axisLine={false} className="text-xs fill-muted-foreground" />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickFormatter={(v: number) => formatCompactMoney(v * 100)}
          width={56}
          className="text-xs fill-muted-foreground"
        />
        <Tooltip
          cursor={{ fill: "rgba(0,0,0,0.04)" }}
          contentStyle={{
            background: "var(--background)",
            border: "1px solid var(--border)",
            borderRadius: 8,
            fontSize: 12,
          }}
          formatter={(v: unknown, _name, item) => {
            const count = (item as { payload?: { count?: number } } | undefined)?.payload?.count ?? 0;
            return [
              `${formatMoney(Number(v) * 100)} (${count} invoice${count === 1 ? "" : "s"})`,
              "Due",
            ];
          }}
        />
        <Bar dataKey="due" radius={[4, 4, 0, 0]}>
          {chartData.map((d) => (
            <Cell key={d.bucket} fill={COLORS[d.bucket] ?? "#94a3b8"} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
