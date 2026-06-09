"use client";

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
import { useTransactions } from "@/hooks/billing/useBilling";

import { formatDateTime, formatMoney } from "../_components/money";
import { StatusBadge } from "../_components/status-badge";

export default function TransactionsListPage() {
  const q = useTransactions();
  if (q.isLoading) return <Skeleton className="h-64 w-full" />;
  const rows = q.data ?? [];
  if (rows.length === 0) {
    return (
      <Card>
        <div className="py-12 text-center text-sm text-muted-foreground">No payments recorded yet.</div>
      </Card>
    );
  }
  return (
    <Card>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Receipt</TableHead>
            <TableHead>Paid at</TableHead>
            <TableHead>Method</TableHead>
            <TableHead>Reference</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Amount</TableHead>
            <TableHead></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((t) => (
            <TableRow key={t.id} className="hover:bg-accent">
              <TableCell className="font-medium">{t.receiptNumber}</TableCell>
              <TableCell>{formatDateTime(t.paidAt)}</TableCell>
              <TableCell className="capitalize">{t.method.replace("_", " ")}</TableCell>
              <TableCell className="text-xs text-muted-foreground">{t.reference ?? "—"}</TableCell>
              <TableCell>
                <StatusBadge status={t.status} />
              </TableCell>
              <TableCell className="text-right">
                {formatMoney(t.currency, t.amountCents)}
              </TableCell>
              <TableCell>
                <a
                  href={billingService.receiptPdfUrl(t.id, true)}
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
