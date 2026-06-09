"use client";

import type { ActivityEntry } from "@/types/dashboard";

import { timeAgo } from "./format";

// ActivityFeed — most-recent audit events. Each row shows the user + the
// HTTP verb/path that triggered the audit + how long ago. Status code drives
// the leading dot colour: green 2xx, amber 4xx, red 5xx, grey other.
export function ActivityFeed({ entries }: { entries: ActivityEntry[] }) {
  if (entries.length === 0) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        Nothing to show yet.
      </div>
    );
  }
  return (
    <ul className="divide-y -my-2">
      {entries.map((e, i) => (
        <li key={i} className="flex items-start gap-3 py-2.5 text-xs">
          <span
            className={`mt-1 inline-block w-2 h-2 rounded-full shrink-0 ${dotColor(e.statusCode)}`}
            title={`HTTP ${e.statusCode}`}
          />
          <div className="flex-1 min-w-0">
            <div className="flex items-baseline justify-between gap-3">
              <span className="font-medium truncate">
                {e.action || labelFromPath(e.method, e.path)}
              </span>
              <span className="text-muted-foreground shrink-0">{timeAgo(e.occurredAt)}</span>
            </div>
            <div className="text-muted-foreground truncate">
              {e.userEmail || "anonymous"}
              {e.method && e.path ? ` · ${e.method} ${e.path}` : ""}
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

function dotColor(status: number): string {
  if (status >= 500) return "bg-red-500";
  if (status >= 400) return "bg-amber-500";
  if (status >= 300) return "bg-sky-500";
  if (status >= 200) return "bg-emerald-500";
  return "bg-slate-400";
}

// Fallback label when audit row has no canonical action verb (older rows or
// requests outside the action-tagged set). Surfaces the bare HTTP method +
// resource so the feed still reads as something instead of a blank.
function labelFromPath(method?: string, path?: string): string {
  if (!method || !path) return "Request";
  return `${method} ${path}`;
}
