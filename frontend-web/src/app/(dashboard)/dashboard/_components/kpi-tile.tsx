"use client";

import { ArrowDownRight, ArrowUpRight } from "lucide-react";
import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";

// KPITile is the top-row card on the dashboard. All four tiles share the same
// vertical rhythm — header (title), value (large, single-line truncate),
// optional delta line (always rendered so heights match), optional sub-line.
// The delta arrow + colour are inferred from the sign; neutral grey when 0.
export function KPITile({
  title,
  value,
  loading,
  deltaPct,
  subline,
  href,
}: {
  title: string;
  value: string;
  loading?: boolean;
  deltaPct?: number;
  subline?: string;
  href?: string;
}) {
  const hasDelta = deltaPct !== undefined && deltaPct !== 0;
  const positive = (deltaPct ?? 0) > 0;
  const body = (
    <Card
      className={cn(
        "h-full flex flex-col",
        href && "cursor-pointer hover:bg-accent transition-colors",
      )}
    >
      <CardHeader className="pb-2">
        <CardDescription className="text-xs uppercase tracking-wide">{title}</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between">
        {loading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <div className="text-2xl font-semibold truncate leading-tight" title={value}>
            {value}
          </div>
        )}
        <div className="mt-3 flex items-center justify-between gap-2 text-xs">
          {hasDelta ? (
            <span
              className={cn(
                "inline-flex items-center gap-0.5 font-medium",
                positive ? "text-emerald-600 dark:text-emerald-400" : "text-red-600 dark:text-red-400",
              )}
            >
              {positive ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
              {Math.abs(deltaPct!).toFixed(1)}%
            </span>
          ) : (
            <span className="text-muted-foreground">—</span>
          )}
          {subline ? <span className="text-muted-foreground truncate">{subline}</span> : null}
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
