"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import type { RequestBucket } from "@/types/dashboard";

import { dayLabel, formatNumber } from "./format";

// RequestsChart — 14-day area chart, total requests + 5xx errors stacked
// behind in red so the operator can spot anomaly days at a glance.
export function RequestsChart({ data }: { data: RequestBucket[] }) {
  const chartData = data.map((d) => ({
    day: dayLabel(d.day),
    requests: d.requests,
    errors: d.errors,
  }));

  return (
    <ResponsiveContainer width="100%" height={220}>
      <AreaChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id="reqFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#6366f1" stopOpacity={0.45} />
            <stop offset="100%" stopColor="#6366f1" stopOpacity={0.02} />
          </linearGradient>
          <linearGradient id="errFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#ef4444" stopOpacity={0.5} />
            <stop offset="100%" stopColor="#ef4444" stopOpacity={0.05} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" vertical={false} />
        <XAxis dataKey="day" tickLine={false} axisLine={false} className="text-xs fill-muted-foreground" />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickFormatter={formatNumber}
          width={40}
          className="text-xs fill-muted-foreground"
        />
        <Tooltip
          contentStyle={{
            background: "var(--background)",
            border: "1px solid var(--border)",
            borderRadius: 8,
            fontSize: 12,
          }}
          formatter={(v: unknown, name) => [formatNumber(Number(v)), name as string]}
        />
        <Legend
          verticalAlign="top"
          height={28}
          iconType="circle"
          iconSize={8}
          wrapperStyle={{ fontSize: 12 }}
        />
        <Area
          type="monotone"
          dataKey="requests"
          name="Requests"
          stroke="#6366f1"
          strokeWidth={2}
          fill="url(#reqFill)"
        />
        <Area
          type="monotone"
          dataKey="errors"
          name="5xx errors"
          stroke="#ef4444"
          strokeWidth={1.5}
          fill="url(#errFill)"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
