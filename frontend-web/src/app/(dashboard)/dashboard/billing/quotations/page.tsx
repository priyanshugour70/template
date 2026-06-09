"use client";

import Link from "next/link";

import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useQuotations } from "@/hooks/billing/useBilling";

import { formatDate, formatMoney } from "../_components/money";
import { StatusBadge } from "../_components/status-badge";

export default function QuotationsListPage() {
  const q = useQuotations();
  if (q.isLoading) return <Skeleton className="h-64 w-full" />;
  const rows = q.data ?? [];
  if (rows.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center space-y-3">
          <p className="text-sm text-muted-foreground">No quotations yet.</p>
          <Link
            href="/dashboard/billing/plan-builder"
            className="inline-block rounded-md bg-foreground px-4 py-2 text-sm font-medium text-background"
          >
            Build your first plan
          </Link>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Number</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Features</TableHead>
            <TableHead>Users</TableHead>
            <TableHead className="text-right">Total</TableHead>
            <TableHead>Expires</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.id} className="cursor-pointer hover:bg-accent">
              <TableCell>
                <Link href={`/dashboard/billing/quotations/${row.id}`} className="hover:underline">
                  {row.number}
                </Link>
              </TableCell>
              <TableCell>
                <StatusBadge status={row.status} />
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">
                {(row.featureKeys ?? []).slice(0, 3).join(", ")}
                {(row.featureKeys ?? []).length > 3 ? ` +${row.featureKeys.length - 3}` : ""}
              </TableCell>
              <TableCell>{row.userCount}</TableCell>
              <TableCell className="text-right font-medium">
                {formatMoney(row.currency, row.totalCents)}
              </TableCell>
              <TableCell>{formatDate(row.expiresAt)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
