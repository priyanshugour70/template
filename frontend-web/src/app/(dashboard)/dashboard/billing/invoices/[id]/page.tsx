"use client";

import { useRouter } from "next/navigation";
import { use, useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { toast } from "@/hooks/use-toast";
import { billingService } from "@/services/billing";
import { useInvoice, useRecordPayment } from "@/hooks/billing/useBilling";
import type { PaymentMethod } from "@/types/billing";

import { formatDate, formatMoney } from "../../_components/money";
import { StatusBadge } from "../../_components/status-badge";

const METHODS: { value: PaymentMethod; label: string; needsRef: boolean }[] = [
  { value: "cash", label: "Cash", needsRef: false },
  { value: "bank_transfer", label: "Bank transfer", needsRef: true },
  { value: "cheque", label: "Cheque", needsRef: true },
];

export default function InvoiceDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const q = useInvoice(id);
  const payMut = useRecordPayment(id);

  const [method, setMethod] = useState<PaymentMethod>("cash");
  const [reference, setReference] = useState("");

  if (q.isLoading) return <Skeleton className="h-96 w-full" />;
  if (!q.data) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-sm text-muted-foreground">
          Invoice not found.
        </CardContent>
      </Card>
    );
  }
  const inv = q.data;
  const canPay = inv.status === "open" && inv.amountDueCents > 0;
  const needsRef = METHODS.find((m) => m.value === method)?.needsRef ?? false;

  async function pay() {
    try {
      const res = await payMut.mutateAsync({
        method,
        amountCents: inv.amountDueCents,
        reference: needsRef ? reference : undefined,
      });
      toast.success("Payment recorded", `Receipt ${res.transaction.receiptNumber}`);
      router.push(`/dashboard/billing/transactions`);
    } catch (e) {
      toast.error("Payment failed", (e as Error).message);
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>{inv.number}</CardTitle>
            <CardDescription>
              Issued {formatDate(inv.issuedAt)} · Due {inv.dueAt ? formatDate(inv.dueAt) : "—"}
            </CardDescription>
          </div>
          <div className="flex gap-2 items-center">
            <StatusBadge status={inv.status} />
            <a
              href={billingService.invoicePdfUrl(inv.id, true)}
              className="text-xs underline ml-2"
            >
              Download PDF
            </a>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Description</TableHead>
                <TableHead>HSN/SAC</TableHead>
                <TableHead className="text-right">Qty</TableHead>
                <TableHead className="text-right">Unit</TableHead>
                <TableHead className="text-right">Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(inv.lineItems ?? []).map((l, i) => (
                <TableRow key={i}>
                  <TableCell>{l.description}</TableCell>
                  <TableCell className="text-xs">{l.hsnSac ?? "—"}</TableCell>
                  <TableCell className="text-right">{l.quantity}</TableCell>
                  <TableCell className="text-right">
                    {formatMoney(inv.currency, l.unitPriceCents ?? l.unitCents ?? 0)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatMoney(inv.currency, l.totalCents ?? l.amountCents ?? 0)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <div className="mt-6 ml-auto max-w-xs space-y-1 text-sm">
            <Row label="Subtotal" value={formatMoney(inv.currency, inv.subtotalCents)} />
            {inv.cgstCents != null && inv.cgstCents > 0 && (
              <Row label="CGST" value={formatMoney(inv.currency, inv.cgstCents)} />
            )}
            {inv.sgstCents != null && inv.sgstCents > 0 && (
              <Row label="SGST" value={formatMoney(inv.currency, inv.sgstCents)} />
            )}
            {inv.igstCents != null && inv.igstCents > 0 && (
              <Row label="IGST" value={formatMoney(inv.currency, inv.igstCents)} />
            )}
            <Row label="Total" value={formatMoney(inv.currency, inv.totalCents)} bold />
            <Row label="Paid" value={formatMoney(inv.currency, inv.amountPaidCents)} />
            <Row label="Due" value={formatMoney(inv.currency, inv.amountDueCents)} bold />
          </div>
        </CardContent>
      </Card>

      {canPay && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Record payment</CardTitle>
            <CardDescription>
              Manual recording for cash / bank / cheque. Gateway integrations land in Phase 6+.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 gap-2">
              {METHODS.map((m) => (
                <button
                  key={m.value}
                  type="button"
                  onClick={() => setMethod(m.value)}
                  className={`rounded-md border px-3 py-2 text-sm ${
                    method === m.value ? "border-foreground bg-foreground/5" : "border-border"
                  }`}
                >
                  {m.label}
                </button>
              ))}
            </div>
            {needsRef && (
              <div>
                <label className="text-xs uppercase tracking-wide text-muted-foreground">
                  {method === "bank_transfer" ? "Transaction reference (UTR)" : "Cheque number"}
                </label>
                <Input
                  value={reference}
                  onChange={(e) => setReference(e.target.value)}
                  className="mt-1"
                  placeholder={method === "bank_transfer" ? "UTR123456789" : "012345"}
                />
              </div>
            )}
            <div className="flex items-center justify-between border-t pt-4">
              <div>
                <div className="text-xs uppercase tracking-wide text-muted-foreground">
                  Amount to record
                </div>
                <div className="text-2xl font-semibold">
                  {formatMoney(inv.currency, inv.amountDueCents)}
                </div>
              </div>
              <button
                type="button"
                onClick={pay}
                disabled={payMut.isPending || (needsRef && !reference.trim())}
                className="rounded-md bg-foreground px-6 py-2 text-sm font-medium text-background disabled:opacity-50"
              >
                {payMut.isPending ? "Recording…" : "Record payment"}
              </button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function Row({ label, value, bold }: { label: string; value: string; bold?: boolean }) {
  return (
    <div className={`flex justify-between ${bold ? "font-semibold" : ""}`}>
      <span className="text-muted-foreground">{label}</span>
      <span>{value}</span>
    </div>
  );
}
