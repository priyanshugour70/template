"use client";

import {
  AlertTriangle,
  Calendar,
  CheckCircle2,
  CreditCard,
  Download,
  FileText,
  PauseCircle,
  PlayCircle,
  RotateCcw,
  Tag,
  TrendingUp,
  X,
  XCircle,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useActiveSubscription,
  useCancelSubscription,
  useChangePlan,
  useInvoices,
  usePauseSubscription,
  usePlans,
  usePreviewChange,
  useReactivateSubscription,
  useResumeSubscription,
  useUpdateBilling,
  useUsage,
  useValidateCoupon,
} from "@/hooks/subscription/useSubscriptionQueries";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";
import { usePermissions } from "@/providers";
import type {
  BillingCycle,
  Invoice,
  Plan,
  PreviewChangeResponse,
  Subscription,
  UsageCounter,
  ValidateCouponResponse,
} from "@/types/subscription";

// ── formatting ────────────────────────────────────────────────────────────

function formatMoney(cents: number, currency = "INR") {
  const sign = cents < 0 ? "-" : "";
  const abs = Math.abs(cents);
  const major = abs / 100;
  try {
    return (
      sign +
      new Intl.NumberFormat("en-IN", {
        style: "currency",
        currency,
        maximumFractionDigits: 2,
      }).format(major)
    );
  } catch {
    return `${sign}${currency} ${major.toFixed(2)}`;
  }
}

function formatDate(s?: string | null) {
  if (!s) return "—";
  try {
    return new Date(s).toLocaleDateString("en-IN", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  } catch {
    return s;
  }
}

function statusVariant(s: string) {
  switch (s) {
    case "active":
      return "success" as const;
    case "trial":
      return "default" as const;
    case "paused":
      return "warning" as const;
    case "past_due":
    case "cancelled":
    case "expired":
      return "danger" as const;
    default:
      return "muted" as const;
  }
}

const LIMIT_LABELS: { key: string; label: string }[] = [
  { key: "users.max", label: "Users" },
  { key: "orgs.max", label: "Organizations" },
  { key: "storage.gb", label: "Storage (GB)" },
  { key: "api.calls.monthly", label: "API calls (monthly)" },
];

// ── main page ─────────────────────────────────────────────────────────────

export default function SubscriptionPage() {
  const { has } = usePermissions();
  const activeQ = useActiveSubscription();
  const plansQ = usePlans();
  const usageQ = useUsage();
  const invoicesQ = useInvoices();

  const [switchTo, setSwitchTo] = useState<Plan | null>(null);
  const [cancelling, setCancelling] = useState(false);
  const [showBilling, setShowBilling] = useState(false);
  const [invoiceView, setInvoiceView] = useState<Invoice | null>(null);

  const sub = activeQ.data;
  const currentPlan = useMemo(
    () => (plansQ.data ?? []).find((p) => p.code === sub?.planCode) ?? null,
    [plansQ.data, sub?.planCode],
  );

  if (activeQ.isLoading || plansQ.isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Subscription</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Plan, billing, usage and invoices for this organization.
        </p>
      </div>

      {sub ? (
        <CurrentPlanCard
          subscription={sub}
          plan={currentPlan}
          onCancel={() => setCancelling(true)}
          onShowBilling={() => setShowBilling(true)}
          canManage={has("subscription.update")}
          canCancel={has("subscription.cancel")}
          canPause={has("subscription.pause")}
        />
      ) : (
        <NoSubscriptionCard
          plans={plansQ.data ?? []}
          onPick={(p) => setSwitchTo(p)}
          canManage={has("subscription.update")}
        />
      )}

      <Tabs defaultValue="usage" className="w-full">
        <TabsList>
          <TabsTrigger value="usage">Usage</TabsTrigger>
          <TabsTrigger value="plans">Compare plans</TabsTrigger>
          <TabsTrigger value="invoices">
            Invoices
            {(invoicesQ.data?.length ?? 0) > 0 && (
              <Badge variant="muted" className="ml-2">
                {invoicesQ.data?.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="usage">
          <UsagePanel
            sub={sub ?? null}
            plan={currentPlan}
            usage={usageQ.data ?? []}
            loading={usageQ.isLoading}
          />
        </TabsContent>
        <TabsContent value="plans">
          <PlansComparison
            plans={plansQ.data ?? []}
            currentCode={sub?.planCode}
            onPick={(p) => setSwitchTo(p)}
            canManage={has("subscription.update")}
          />
        </TabsContent>
        <TabsContent value="invoices">
          <InvoicesPanel
            invoices={invoicesQ.data ?? []}
            loading={invoicesQ.isLoading}
            onView={setInvoiceView}
          />
        </TabsContent>
      </Tabs>

      {switchTo && (
        <PlanSwitchDialog
          target={switchTo}
          current={sub ?? null}
          onClose={() => setSwitchTo(null)}
        />
      )}
      {cancelling && sub && (
        <CancelDialog sub={sub} onClose={() => setCancelling(false)} />
      )}
      {showBilling && sub && (
        <BillingDialog sub={sub} onClose={() => setShowBilling(false)} />
      )}
      {invoiceView && (
        <InvoiceDialog invoice={invoiceView} onClose={() => setInvoiceView(null)} />
      )}
    </div>
  );
}

// ── current plan card ─────────────────────────────────────────────────────

function CurrentPlanCard({
  subscription,
  plan,
  onCancel,
  onShowBilling,
  canManage,
  canCancel,
  canPause,
}: {
  subscription: Subscription;
  plan: Plan | null;
  onCancel: () => void;
  onShowBilling: () => void;
  canManage: boolean;
  canCancel: boolean;
  canPause: boolean;
}) {
  const pause = usePauseSubscription();
  const resume = useResumeSubscription();
  const reactivate = useReactivateSubscription();

  const isPaused = subscription.status === "paused";
  const cancelScheduled = !!subscription.cancelAt;
  const isCancelled =
    subscription.status === "cancelled" || subscription.status === "expired";
  const trialActive = subscription.status === "trial";

  return (
    <Card>
      <CardContent className="space-y-5 p-6">
        <div className="flex flex-wrap items-start gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-xl font-semibold tracking-tight">
                {plan?.name ?? subscription.planCode}
              </h2>
              <Badge variant={statusVariant(subscription.status)}>{subscription.status}</Badge>
              {cancelScheduled && !isCancelled && (
                <Badge variant="warning">Ending {formatDate(subscription.cancelAt)}</Badge>
              )}
              {subscription.gateway && (
                <Badge variant="muted">
                  {subscription.gateway}
                  {subscription.gatewaySubscriptionId ? " · synced" : " · pending"}
                </Badge>
              )}
            </div>
            {plan?.tagline && (
              <p className="mt-1 text-sm text-muted-foreground">{plan.tagline}</p>
            )}
          </div>
          <div className="text-right">
            <div className="text-2xl font-semibold">
              {formatMoney(subscription.unitPriceCents, subscription.currency)}
            </div>
            <div className="text-xs text-muted-foreground">
              per {subscription.billingCycle.replace("_", " ")}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <Metric
            icon={Calendar}
            label="Period ends"
            value={formatDate(subscription.currentPeriodEnd)}
          />
          <Metric
            icon={CreditCard}
            label="Next bill"
            value={
              cancelScheduled
                ? "—"
                : subscription.nextBillingAt
                  ? formatDate(subscription.nextBillingAt)
                  : "—"
            }
          />
          <Metric
            icon={Tag}
            label="Discount"
            value={subscription.discountCents > 0 ? "applied" : "—"}
          />
          <Metric
            icon={TrendingUp}
            label="Trial ends"
            value={trialActive ? formatDate(subscription.trialEndsAt) : "—"}
          />
        </div>

        {/* Action row */}
        <div className="flex flex-wrap items-center gap-2 border-t border-border pt-4">
          {canManage && !isPaused && !isCancelled && (
            <Button variant="outline" onClick={onShowBilling}>
              <CreditCard className="h-4 w-4" />
              Billing info
            </Button>
          )}
          {canPause && !isPaused && !isCancelled && !cancelScheduled && (
            <Button
              variant="outline"
              disabled={pause.isPending}
              onClick={() =>
                pause.mutate(
                  {},
                  {
                    onSuccess: () => toast.success("Subscription paused"),
                    onError: (e: unknown) =>
                      toast.error("Pause failed", e instanceof Error ? e.message : undefined),
                  },
                )
              }
            >
              <PauseCircle className="h-4 w-4" />
              Pause
            </Button>
          )}
          {canPause && isPaused && (
            <Button
              variant="outline"
              disabled={resume.isPending}
              onClick={() =>
                resume.mutate(undefined, {
                  onSuccess: () => toast.success("Subscription resumed"),
                  onError: (e: unknown) =>
                    toast.error("Resume failed", e instanceof Error ? e.message : undefined),
                })
              }
            >
              <PlayCircle className="h-4 w-4" />
              Resume
            </Button>
          )}
          {canManage && cancelScheduled && !isCancelled && (
            <Button
              variant="outline"
              disabled={reactivate.isPending}
              onClick={() =>
                reactivate.mutate(undefined, {
                  onSuccess: () => toast.success("Subscription reactivated"),
                  onError: (e: unknown) =>
                    toast.error("Reactivate failed", e instanceof Error ? e.message : undefined),
                })
              }
            >
              <RotateCcw className="h-4 w-4" />
              Resume billing
            </Button>
          )}
          {canCancel && !cancelScheduled && !isCancelled && (
            <Button
              variant="ghost"
              className="ml-auto text-destructive hover:text-destructive"
              onClick={onCancel}
            >
              <XCircle className="h-4 w-4" />
              Cancel subscription
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function Metric({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Calendar;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-md border border-border bg-muted/30 p-3">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Icon className="h-3 w-3" />
        {label}
      </div>
      <div className="mt-1 truncate text-sm font-medium">{value}</div>
    </div>
  );
}

function NoSubscriptionCard({
  plans,
  onPick,
  canManage,
}: {
  plans: Plan[];
  onPick: (p: Plan) => void;
  canManage: boolean;
}) {
  const defaultPlan = plans.find((p) => p.isDefault) ?? plans[0];
  return (
    <Card>
      <CardContent className="space-y-3 p-6">
        <div className="flex items-start gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 text-warning" />
          <div>
            <h2 className="text-lg font-semibold">No active subscription</h2>
            <p className="text-sm text-muted-foreground">
              Pick a plan below or in the &quot;Compare plans&quot; tab to start.
            </p>
          </div>
        </div>
        {canManage && defaultPlan && (
          <Button onClick={() => onPick(defaultPlan)}>
            Start on {defaultPlan.name}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

// ── usage gauges ──────────────────────────────────────────────────────────

function UsagePanel({
  sub,
  plan,
  usage,
  loading,
}: {
  sub: Subscription | null;
  plan: Plan | null;
  usage: UsageCounter[];
  loading: boolean;
}) {
  const usageByKey = useMemo(() => {
    const m = new Map<string, number>();
    for (const u of usage) m.set(u.key, u.count);
    return m;
  }, [usage]);

  const limits = (sub?.limits as Record<string, number>) ?? plan?.limits ?? {};

  if (loading) return <Skeleton className="h-48 w-full" />;
  if (!sub) {
    return (
      <Card>
        <CardContent className="p-6 text-sm text-muted-foreground">
          Start a subscription to track usage against quota.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {LIMIT_LABELS.map(({ key, label }) => {
        const limit = limits[key] ?? -1;
        const current = usageByKey.get(key) ?? 0;
        return <UsageGauge key={key} label={label} current={current} limit={limit} />;
      })}
    </div>
  );
}

function UsageGauge({
  label,
  current,
  limit,
}: {
  label: string;
  current: number;
  limit: number;
}) {
  const unlimited = limit === -1;
  const pct = unlimited ? 0 : limit > 0 ? Math.min(100, (current / limit) * 100) : 0;
  const over80 = !unlimited && pct >= 80;
  const over100 = !unlimited && current >= limit;
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-baseline justify-between">
          <p className="text-sm font-medium">{label}</p>
          {over100 && (
            <Badge variant="danger" className="text-[10px]">
              over limit
            </Badge>
          )}
          {!over100 && over80 && (
            <Badge variant="warning" className="text-[10px]">
              approaching
            </Badge>
          )}
          {unlimited && (
            <Badge variant="muted" className="text-[10px]">
              unlimited
            </Badge>
          )}
        </div>
        <div className="mt-2 flex items-baseline gap-1.5">
          <span className="text-2xl font-semibold tabular-nums">
            {current.toLocaleString()}
          </span>
          {!unlimited && (
            <span className="text-sm text-muted-foreground">
              / {limit.toLocaleString()}
            </span>
          )}
        </div>
        {!unlimited && (
          <div className="mt-2 h-2 overflow-hidden rounded-full bg-muted">
            <div
              className={cn(
                "h-full transition-all",
                over100 ? "bg-destructive" : over80 ? "bg-warning" : "bg-primary",
              )}
              style={{ width: `${pct}%` }}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ── plans comparison ──────────────────────────────────────────────────────

function PlansComparison({
  plans,
  currentCode,
  onPick,
  canManage,
}: {
  plans: Plan[];
  currentCode?: string;
  onPick: (p: Plan) => void;
  canManage: boolean;
}) {
  const allFeatures = useMemo(() => {
    const seen = new Set<string>();
    const out: string[] = [];
    for (const p of plans) {
      for (const f of p.features ?? []) {
        if (!seen.has(f)) {
          seen.add(f);
          out.push(f);
        }
      }
    }
    return out;
  }, [plans]);

  const sorted = useMemo(() => [...plans].sort((a, b) => a.tier - b.tier), [plans]);

  return (
    <Card>
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="min-w-[180px]">Feature</TableHead>
                {sorted.map((p) => (
                  <TableHead key={p.code} className="min-w-[170px]">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 font-semibold text-foreground">
                        {p.name}
                        {p.code === currentCode && <Badge variant="success">current</Badge>}
                      </div>
                      <div className="text-xs font-normal text-muted-foreground">
                        {p.priceCents > 0
                          ? `${formatMoney(p.priceCents, p.currency)} / ${p.billingCycle}`
                          : "Free"}
                        {p.trialDays > 0 && ` · ${p.trialDays}d trial`}
                      </div>
                    </div>
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {allFeatures.map((f) => (
                <TableRow key={f}>
                  <TableCell className="font-mono text-xs">{f}</TableCell>
                  {sorted.map((p) => (
                    <TableCell key={p.code}>
                      {(p.features ?? []).includes(f) ? (
                        <CheckCircle2 className="h-4 w-4 text-success" />
                      ) : (
                        <X className="h-4 w-4 text-muted-foreground/40" />
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
              {LIMIT_LABELS.map(({ key, label }) => (
                <TableRow key={key}>
                  <TableCell className="text-xs text-muted-foreground">{label}</TableCell>
                  {sorted.map((p) => {
                    const v = p.limits?.[key];
                    return (
                      <TableCell key={p.code} className="tabular-nums">
                        {v == null ? "—" : v === -1 ? "Unlimited" : v.toLocaleString()}
                      </TableCell>
                    );
                  })}
                </TableRow>
              ))}
              <TableRow>
                <TableCell />
                {sorted.map((p) => (
                  <TableCell key={p.code}>
                    {p.code === currentCode ? (
                      <Badge variant="muted">Current plan</Badge>
                    ) : (
                      <Button
                        size="sm"
                        variant={canManage ? "default" : "outline"}
                        disabled={!canManage}
                        onClick={() => onPick(p)}
                      >
                        {currentCode ? "Switch" : "Choose"}
                      </Button>
                    )}
                  </TableCell>
                ))}
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}

// ── invoices ──────────────────────────────────────────────────────────────

function InvoicesPanel({
  invoices,
  loading,
  onView,
}: {
  invoices: Invoice[];
  loading: boolean;
  onView: (i: Invoice) => void;
}) {
  if (loading) return <Skeleton className="h-32 w-full" />;
  if (invoices.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center gap-2 py-10 text-center text-sm text-muted-foreground">
          <FileText className="h-8 w-8" />
          <p>No invoices yet.</p>
          <p className="text-xs">Invoices are generated automatically on plan changes.</p>
        </CardContent>
      </Card>
    );
  }
  return (
    <Card>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Number</TableHead>
              <TableHead>Issued</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Total</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {invoices.map((inv) => (
              <TableRow key={inv.id} className="cursor-pointer" onClick={() => onView(inv)}>
                <TableCell className="font-mono text-xs">{inv.number}</TableCell>
                <TableCell>{formatDate(inv.issuedAt)}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      inv.status === "paid"
                        ? "success"
                        : inv.status === "open"
                          ? "warning"
                          : "muted"
                    }
                  >
                    {inv.status}
                  </Badge>
                </TableCell>
                <TableCell className="tabular-nums">
                  {formatMoney(inv.totalCents, inv.currency)}
                </TableCell>
                <TableCell className="text-right" onClick={(e) => e.stopPropagation()}>
                  <Button variant="ghost" size="sm" onClick={() => onView(inv)}>
                    <FileText className="h-4 w-4" />
                    View
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function InvoiceDialog({ invoice, onClose }: { invoice: Invoice; onClose: () => void }) {
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="font-mono">{invoice.number}</DialogTitle>
          <DialogDescription>
            Issued {formatDate(invoice.issuedAt)} · Status{" "}
            <Badge
              variant={
                invoice.status === "paid"
                  ? "success"
                  : invoice.status === "open"
                    ? "warning"
                    : "muted"
              }
              className="ml-1"
            >
              {invoice.status}
            </Badge>
          </DialogDescription>
        </DialogHeader>

        {invoice.description && (
          <p className="text-sm text-muted-foreground">{invoice.description}</p>
        )}

        <div className="mt-2 border-t border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Item</TableHead>
                <TableHead className="text-right">Qty</TableHead>
                <TableHead className="text-right">Unit</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(invoice.lineItems ?? []).map((li, idx) => (
                <TableRow key={idx}>
                  <TableCell>{li.description}</TableCell>
                  <TableCell className="text-right tabular-nums">{li.quantity}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatMoney(li.unitCents, invoice.currency)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatMoney(li.amountCents, invoice.currency)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>

        <div className="mt-3 space-y-1 border-t border-border pt-3 text-sm">
          <SummaryRow
            label="Subtotal"
            value={formatMoney(invoice.subtotalCents, invoice.currency)}
          />
          {invoice.discountCents > 0 && (
            <SummaryRow
              label={`Discount${invoice.couponCode ? ` (${invoice.couponCode})` : ""}`}
              value={`-${formatMoney(invoice.discountCents, invoice.currency)}`}
              muted
            />
          )}
          {invoice.taxCents > 0 && (
            <SummaryRow
              label="Tax"
              value={formatMoney(invoice.taxCents, invoice.currency)}
              muted
            />
          )}
          <SummaryRow
            label="Total"
            value={formatMoney(invoice.totalCents, invoice.currency)}
            bold
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() =>
              toast.info(
                "PDF download coming soon",
                "Wire to a PDF service in a future patch.",
              )
            }
          >
            <Download className="h-4 w-4" />
            Download
          </Button>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function SummaryRow({
  label,
  value,
  muted,
  bold,
}: {
  label: string;
  value: string;
  muted?: boolean;
  bold?: boolean;
}) {
  return (
    <div className="flex items-center justify-between">
      <span
        className={cn("text-sm", muted && "text-muted-foreground", bold && "font-semibold")}
      >
        {label}
      </span>
      <span className={cn("text-sm tabular-nums", bold && "font-semibold")}>{value}</span>
    </div>
  );
}

// ── plan switch dialog ────────────────────────────────────────────────────

function PlanSwitchDialog({
  target,
  current,
  onClose,
}: {
  target: Plan;
  current: Subscription | null;
  onClose: () => void;
}) {
  const [cycle, setCycle] = useState<BillingCycle>(target.billingCycle);
  const [quantity, setQuantity] = useState(current?.quantity ?? 1);
  const [coupon, setCoupon] = useState("");
  const [appliedCoupon, setAppliedCoupon] = useState<ValidateCouponResponse | null>(null);

  const previewMutation = usePreviewChange();
  const validateCoupon = useValidateCoupon();
  const change = useChangePlan();

  // Recompute preview when inputs change.
  useEffect(() => {
    previewMutation.mutate({
      planCode: target.code,
      billingCycle: cycle,
      quantity,
      couponCode: appliedCoupon?.valid ? appliedCoupon.code : undefined,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cycle, quantity, appliedCoupon?.code]);

  const preview = previewMutation.data ?? null;

  const apply = () => {
    change.mutate(
      {
        planCode: target.code,
        billingCycle: cycle,
        quantity,
        couponCode: appliedCoupon?.valid ? appliedCoupon.code : undefined,
        startImmediately: true,
      },
      {
        onSuccess: (data) => {
          toast.success(
            "Plan changed",
            data.invoice
              ? `Invoice ${data.invoice.number} created — ${formatMoney(data.invoice.totalCents, data.invoice.currency)}`
              : undefined,
          );
          onClose();
        },
        onError: (e: unknown) =>
          toast.error("Change failed", e instanceof Error ? e.message : undefined),
      },
    );
  };

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Switch to {target.name}</DialogTitle>
          <DialogDescription>
            Review proration and any active coupon before confirming.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <div className="grid grid-cols-2 gap-3">
            <Field label="Billing cycle">
              <Select value={cycle} onValueChange={(v) => setCycle(v as BillingCycle)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="monthly">Monthly</SelectItem>
                  <SelectItem value="quarterly">Quarterly</SelectItem>
                  <SelectItem value="yearly">Yearly</SelectItem>
                </SelectContent>
              </Select>
            </Field>
            <Field label="Quantity">
              <Input
                type="number"
                min={1}
                value={quantity}
                onChange={(e) => setQuantity(Math.max(1, Number(e.target.value) || 1))}
              />
            </Field>
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs uppercase tracking-wider text-muted-foreground">
              Coupon
            </Label>
            <div className="flex gap-2">
              <Input
                placeholder="WELCOME20"
                value={coupon}
                onChange={(e) => setCoupon(e.target.value.toUpperCase())}
                disabled={!!appliedCoupon?.valid}
              />
              {appliedCoupon?.valid ? (
                <Button
                  variant="outline"
                  onClick={() => {
                    setAppliedCoupon(null);
                    setCoupon("");
                  }}
                >
                  Remove
                </Button>
              ) : (
                <Button
                  variant="outline"
                  disabled={!coupon || validateCoupon.isPending}
                  onClick={() =>
                    validateCoupon.mutate(
                      { code: coupon, planCode: target.code },
                      {
                        onSuccess: (data) => {
                          setAppliedCoupon(data);
                          if (!data.valid) {
                            toast.error("Coupon invalid", data.reason);
                          } else {
                            toast.success("Coupon applied", data.name);
                          }
                        },
                      },
                    )
                  }
                >
                  Apply
                </Button>
              )}
            </div>
            {appliedCoupon?.valid && (
              <p className="text-xs text-success">
                {appliedCoupon.name} —{" "}
                {appliedCoupon.percentOff != null
                  ? `${appliedCoupon.percentOff}% off`
                  : appliedCoupon.amountOffCents != null
                    ? `${formatMoney(appliedCoupon.amountOffCents, appliedCoupon.currency || target.currency)} off`
                    : "discount"}
                {appliedCoupon.duration === "repeating" && " (repeating)"}
              </p>
            )}
            {appliedCoupon && !appliedCoupon.valid && (
              <p className="text-xs text-destructive">{appliedCoupon.reason}</p>
            )}
          </div>

          <div className="rounded-md border border-border bg-muted/40 p-4 text-sm">
            <p className="text-xs uppercase tracking-wider text-muted-foreground">
              Proration preview
            </p>
            {previewMutation.isPending ? (
              <Skeleton className="mt-2 h-24 w-full" />
            ) : preview ? (
              <ProrationView preview={preview} />
            ) : (
              <p className="mt-2 text-muted-foreground">
                Preview unavailable. Try changing inputs.
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button disabled={change.isPending || !preview} onClick={apply}>
            {change.isPending
              ? "Switching…"
              : preview
                ? `Confirm — ${formatMoney(preview.totalDueCents, preview.currency)}`
                : "Confirm"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ProrationView({ preview }: { preview: PreviewChangeResponse }) {
  return (
    <div className="mt-2 space-y-1.5">
      <SummaryRow
        label={`${preview.toPlanCode} (${preview.billingCycle})`}
        value={formatMoney(preview.baseAmountCents, preview.currency)}
      />
      {preview.prorationCents !== 0 && (
        <SummaryRow
          label={`Proration${preview.unusedDaysRemaining ? ` (${preview.unusedDaysRemaining}d unused)` : ""}`}
          value={formatMoney(preview.prorationCents, preview.currency)}
          muted
        />
      )}
      {preview.discountCents > 0 && (
        <SummaryRow
          label={`Coupon${preview.couponCode ? ` (${preview.couponCode})` : ""}`}
          value={`-${formatMoney(preview.discountCents, preview.currency)}`}
          muted
        />
      )}
      <SummaryRow
        label="Tax"
        value={formatMoney(preview.taxCents, preview.currency)}
        muted
      />
      <div className="border-t border-border pt-1.5">
        <SummaryRow
          label="Total due now"
          value={formatMoney(preview.totalDueCents, preview.currency)}
          bold
        />
      </div>
      <p className="pt-1 text-xs text-muted-foreground">
        {preview.isUpgrade
          ? "Upgrade — applies immediately"
          : "Downgrade — applies immediately"}
      </p>
    </div>
  );
}

// ── cancel ────────────────────────────────────────────────────────────────

const CANCEL_REASONS = [
  { value: "too_expensive", label: "Too expensive" },
  { value: "missing_feature", label: "Missing a feature I need" },
  { value: "found_alternative", label: "Found a better alternative" },
  { value: "no_longer_needed", label: "No longer needed" },
  { value: "technical_issues", label: "Technical issues" },
  { value: "other", label: "Other" },
];

function CancelDialog({ sub, onClose }: { sub: Subscription; onClose: () => void }) {
  const [reasonKey, setReasonKey] = useState<string>("no_longer_needed");
  const [note, setNote] = useState("");
  const [immediate, setImmediate] = useState(false);
  const cancel = useCancelSubscription();

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel subscription</DialogTitle>
          <DialogDescription>
            {immediate
              ? "Service will end immediately. You won't be billed again."
              : `Service continues until ${formatDate(sub.currentPeriodEnd)}. No refund for the remainder.`}
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-md border border-warning/40 bg-warning/10 p-3 text-sm">
          <p className="font-medium">Before you go…</p>
          <p className="mt-1 text-xs text-muted-foreground">
            If billing is the concern, consider <span className="font-medium">Pause</span> instead.
            We&apos;ll keep your data and you can resume any time.
          </p>
        </div>

        <div className="space-y-3 py-2">
          <Field label="Why are you cancelling?">
            <Select value={reasonKey} onValueChange={setReasonKey}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {CANCEL_REASONS.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </Field>
          <Field label="Anything else? (optional)">
            <Input value={note} onChange={(e) => setNote(e.target.value)} />
          </Field>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={immediate}
              onChange={(e) => setImmediate(e.target.checked)}
              className="h-4 w-4"
            />
            Cancel immediately (lose remaining period)
          </label>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Keep subscription
          </Button>
          <Button
            variant="destructive"
            disabled={cancel.isPending}
            onClick={() =>
              cancel.mutate(
                {
                  reason: note ? `${reasonKey}: ${note}` : reasonKey,
                  immediate,
                },
                {
                  onSuccess: () => {
                    toast.success(
                      immediate ? "Cancelled immediately" : "Cancellation scheduled",
                    );
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error(
                      "Cancel failed",
                      e instanceof Error ? e.message : undefined,
                    ),
                },
              )
            }
          >
            {cancel.isPending ? "Cancelling…" : "Cancel subscription"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ── billing ──────────────────────────────────────────────────────────────

function BillingDialog({ sub, onClose }: { sub: Subscription; onClose: () => void }) {
  const [email, setEmail] = useState(sub.billingEmail ?? "");
  const [name, setName] = useState(
    (sub as unknown as { billingName?: string }).billingName ?? "",
  );
  const update = useUpdateBilling();
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Billing information</DialogTitle>
          <DialogDescription>
            Used on invoices and billing-related notifications.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Field label="Billing email">
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="billing@example.com"
            />
          </Field>
          <Field label="Billing name">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Acme Inc"
            />
          </Field>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                { billingEmail: email, billingName: name },
                {
                  onSuccess: () => {
                    toast.success("Billing info updated");
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error(
                      "Update failed",
                      e instanceof Error ? e.message : undefined,
                    ),
                },
              )
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs uppercase tracking-wider text-muted-foreground">{label}</Label>
      {children}
    </div>
  );
}
