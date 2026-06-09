"use client";

import { useRouter } from "next/navigation";
import { use } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import {
  useActivateQuotation,
  useDeleteQuotation,
  useQuotation,
} from "@/hooks/billing/useBilling";

import { formatDate, formatMoney } from "../../_components/money";
import { StatusBadge } from "../../_components/status-badge";

export default function QuotationDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const q = useQuotation(id);
  const activateMut = useActivateQuotation();
  const deleteMut = useDeleteQuotation();

  if (q.isLoading) return <Skeleton className="h-96 w-full" />;
  if (!q.data) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-sm text-muted-foreground">
          Quotation not found.
        </CardContent>
      </Card>
    );
  }
  const quot = q.data;

  async function activate() {
    try {
      const res = await activateMut.mutateAsync(id);
      toast.success("Quotation activated", `Invoice ${res.invoice.number} issued`);
      router.push(`/dashboard/billing/invoices/${res.invoice.id}`);
    } catch (e) {
      toast.error("Activation failed", (e as Error).message);
    }
  }

  async function remove() {
    if (!confirm("Delete this draft? This cannot be undone.")) return;
    try {
      await deleteMut.mutateAsync(id);
      toast.success("Quotation deleted");
      router.push("/dashboard/billing/quotations");
    } catch (e) {
      toast.error("Delete failed", (e as Error).message);
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>{quot.number}</CardTitle>
            <CardDescription>
              Expires {formatDate(quot.expiresAt)} · {quot.userCount} users
            </CardDescription>
          </div>
          <StatusBadge status={quot.status} />
        </CardHeader>
        <CardContent>
          <div className="space-y-1">
            {quot.lineItems?.map((l) => (
              <div key={l.featureKey + l.sortOrder} className="flex justify-between text-sm">
                <span>
                  {l.description}
                  {l.quantity > 1 && <> × {l.quantity}</>}
                </span>
                <span className="font-medium">{formatMoney(quot.currency, l.totalCents)}</span>
              </div>
            ))}
          </div>
          <Table className="mt-6">
            <TableHeader>
              <TableRow>
                <TableHead>Subtotal</TableHead>
                <TableHead>CGST</TableHead>
                <TableHead>SGST</TableHead>
                <TableHead>IGST</TableHead>
                <TableHead className="text-right">Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>{formatMoney(quot.currency, quot.subtotalCents)}</TableCell>
                <TableCell>{formatMoney(quot.currency, quot.cgstCents)}</TableCell>
                <TableCell>{formatMoney(quot.currency, quot.sgstCents)}</TableCell>
                <TableCell>{formatMoney(quot.currency, quot.igstCents)}</TableCell>
                <TableCell className="text-right font-semibold">
                  {formatMoney(quot.currency, quot.totalCents)}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {quot.status === "draft" && (
        <div className="flex gap-2">
          <button
            type="button"
            onClick={activate}
            disabled={activateMut.isPending}
            className="rounded-md bg-foreground px-4 py-2 text-sm font-medium text-background disabled:opacity-50"
          >
            {activateMut.isPending ? "Activating…" : "Activate & generate invoice"}
          </button>
          <button
            type="button"
            onClick={remove}
            disabled={deleteMut.isPending}
            className="rounded-md border px-4 py-2 text-sm font-medium"
          >
            Delete draft
          </button>
        </div>
      )}
    </div>
  );
}
