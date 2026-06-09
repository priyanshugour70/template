"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import type { RevenueBucket } from "@/types/dashboard";

import { formatCompactMoney, formatMoney, monthLabel } from "./format";

// RevenueChart — grouped bar chart, issued vs paid per month over the last
// 12 months. Tooltips render full-precision rupees; the y-axis uses compact
// notation (k / L / Cr) so big customers don't blow out the chart width.
export function RevenueChart({ data }: { data: RevenueBucket[] }) {
  const chartData = data.map((d) => ({
    month: monthLabel(d.month),
    issued: d.issuedCents / 100,
    paid: d.paidCents / 100,
  }));

  return (
    <ResponsiveContainer width="100%" height={260}>
      <BarChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" vertical={false} />
        <XAxis dataKey="month" tickLine={false} axisLine={false} className="text-xs fill-muted-foreground" />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickFormatter={(v: number) => formatCompactMoney(v * 100)}
          className="text-xs fill-muted-foreground"
          width={56}
        />
        <Tooltip
          cursor={{ fill: "rgba(0,0,0,0.04)" }}
          contentStyle={{
            background: "var(--background)",
            border: "1px solid var(--border)",
            borderRadius: 8,
            fontSize: 12,
          }}
          formatter={(v: unknown, name) => [formatMoney(Number(v) * 100), name as string]}
        />
        <Legend
          verticalAlign="top"
          height={28}
          iconType="circle"
          iconSize={8}
          wrapperStyle={{ fontSize: 12 }}
        />
        <Bar dataKey="issued" name="Issued" fill="#0ea5e9" radius={[4, 4, 0, 0]} />
        <Bar dataKey="paid" name="Paid" fill="#10b981" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}
