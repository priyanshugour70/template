"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuditLogs } from "@/hooks/audit/useAuditQueries";

function statusVariant(status?: number) {
  if (!status) return "muted";
  if (status >= 500) return "danger";
  if (status >= 400) return "warning";
  if (status >= 200 && status < 300) return "success";
  return "muted";
}

export default function AuditPage() {
  const logsQ = useAuditLogs({ limit: 100 });
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Audit Log</h1>
        <p className="text-muted-foreground mt-1">
          Every API request hitting this tenant, captured asynchronously.
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Recent activity</CardTitle>
        </CardHeader>
        <CardContent>
          {logsQ.isLoading ? (
            <div className="text-sm text-muted-foreground py-8 text-center">Loading…</div>
          ) : !logsQ.data?.length ? (
            <div className="text-sm text-muted-foreground py-8 text-center">
              No audit events yet. They appear here once the worker starts persisting them.
            </div>
          ) : (
            <div className="rounded-md border overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/40">
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
                    <tr key={l.id} className="border-b last:border-0">
                      <td className="p-3 text-muted-foreground">
                        {new Date(l.occurredAt).toLocaleString()}
                      </td>
                      <td className="p-3 font-mono text-xs">{l.method}</td>
                      <td className="p-3 font-mono text-xs">{l.path}</td>
                      <td className="p-3">
                        <Badge variant={statusVariant(l.statusCode)}>{l.statusCode}</Badge>
                      </td>
                      <td className="p-3 text-muted-foreground">{l.userEmail ?? "—"}</td>
                      <td className="p-3 text-muted-foreground">{l.latencyMs}ms</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
