"use client";

import Link from "next/link";

import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { billingService } from "@/services/billing";
import { useInvoices } from "@/hooks/billing/useBilling";

import { formatDate, formatMoney } from "../_components/money";
import { StatusBadge } from "../_components/status-badge";

export default function InvoicesListPage() {
  const q = useInvoices();
  if (q.isLoading) return <Skeleton className="h-64 w-full" />;
  const rows = q.data ?? [];
  if (rows.length === 0) {
    return (
      <Card>
        <div className="py-12 text-center text-sm text-muted-foreground">No invoices yet.</div>
      </Card>
    );
  }
  return (
    <Card>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Number</TableHead>
            <TableHead>Issued</TableHead>
            <TableHead>Due</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Total</TableHead>
            <TableHead className="text-right">Due</TableHead>
            <TableHead></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((inv) => (
            <TableRow key={inv.id} className="hover:bg-accent">
              <TableCell>
                <Link href={`/dashboard/billing/invoices/${inv.id}`} className="hover:underline">
                  {inv.number}
                </Link>
              </TableCell>
              <TableCell>{formatDate(inv.issuedAt)}</TableCell>
              <TableCell>{inv.dueAt ? formatDate(inv.dueAt) : "—"}</TableCell>
              <TableCell>
                <StatusBadge status={inv.status} />
              </TableCell>
              <TableCell className="text-right">
                {formatMoney(inv.currency, inv.totalCents)}
              </TableCell>
              <TableCell className="text-right font-medium">
                {formatMoney(inv.currency, inv.amountDueCents)}
              </TableCell>
              <TableCell>
                <a
                  href={billingService.invoicePdfUrl(inv.id, true)}
                  className="text-xs underline"
                >
                  PDF
                </a>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
