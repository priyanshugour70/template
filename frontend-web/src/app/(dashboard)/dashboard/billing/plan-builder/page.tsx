"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "@/hooks/use-toast";
import {
  useActivateQuotation,
  useCreateQuotation,
  useFeatureCatalog,
  usePreviewQuote,
} from "@/hooks/billing/useBilling";
import type { Feature, FeatureCategory, Quote } from "@/types/billing";

import { formatMoney } from "../_components/money";

const CATEGORY_LABEL: Record<FeatureCategory, string> = {
  core: "Core (always included)",
  admin: "Admin & access",
  compliance: "Compliance",
  integrations: "Integrations",
  limits: "Limits & scale",
};

// User-count slider stops every +5 users up to 100, then +25 up to 500.
// Picked these because most customers land below 100 and the granularity at the
// top end stops mattering — we'd rather make the common case feel precise.
function buildSliderStops(): number[] {
  const stops: number[] = [];
  for (let n = 1; n <= 100; n += 1) stops.push(n);
  for (let n = 125; n <= 500; n += 25) stops.push(n);
  return stops;
}

export default function PlanBuilderPage() {
  const router = useRouter();
  const featuresQ = useFeatureCatalog();
  const previewMut = usePreviewQuote();
  const createMut = useCreateQuotation();
  const activateMut = useActivateQuotation();

  const features = featuresQ.data ?? [];
  const grouped = useMemo(() => groupByCategory(features), [features]);
  const byKey = useMemo(() => Object.fromEntries(features.map((f) => [f.key, f])), [features]);
  const stops = useMemo(buildSliderStops, []);

  // Reverse-dependency map: feature → features that require it. Drives the
  // "this will also disable X" warning on toggle-off.
  const dependents = useMemo(() => {
    const m: Record<string, string[]> = {};
    for (const f of features) {
      for (const dep of f.requires ?? []) {
        (m[dep] ??= []).push(f.key);
      }
    }
    return m;
  }, [features]);

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [userCount, setUserCount] = useState(10);
  const [submitting, setSubmitting] = useState<"draft" | "activate" | null>(null);

  // Seed selection once features load: core + starter_default features on by
  // default so the customer sees a working quote immediately.
  useEffect(() => {
    if (!features.length || selected.size > 0) return;
    const next = new Set<string>();
    for (const f of features) {
      if (f.isCore || f.isStarterDefault) next.add(f.key);
    }
    setSelected(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [features.length]);

  // Per-user feature for the cents-per-user readout on the slider.
  const extraUsersFeature = byKey["extra_users"];
  const perUserCents = extraUsersFeature?.perUserPriceCents ?? 0;
  const includedUsers = useMemo(() => {
    // Add up included_users across selected features (starter_bundle bundles 10).
    let total = 0;
    for (const key of selected) {
      total += byKey[key]?.includedUsers ?? 0;
    }
    return total;
  }, [selected, byKey]);

  // Debounced live preview — runs on selection/userCount change.
  const [quote, setQuote] = useState<Quote | null>(null);
  const [quoteError, setQuoteError] = useState<string | null>(null);
  useEffect(() => {
    if (selected.size === 0) {
      setQuote(null);
      return;
    }
    const handle = setTimeout(() => {
      previewMut
        .mutateAsync({
          featureKeys: Array.from(selected),
          userCount: Math.max(1, userCount || 1),
        })
        .then((q) => {
          setQuote(q);
          setQuoteError(null);
        })
        .catch((e) => {
          setQuote(null);
          setQuoteError((e as Error).message);
        });
    }, 300);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, userCount]);

  function toggle(key: string, feature: Feature) {
    if (feature.isCore) return; // core is always on, no toggle
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        // Toggling OFF — find selected dependents, warn before removing them.
        const downstream = (dependents[key] ?? []).filter((d) => next.has(d));
        if (downstream.length > 0) {
          const ok = confirm(
            `Removing "${feature.name}" will also disable:\n  • ${downstream
              .map((d) => byKey[d]?.name ?? d)
              .join("\n  • ")}\n\nContinue?`,
          );
          if (!ok) return prev;
          for (const d of downstream) next.delete(d);
        }
        next.delete(key);
      } else {
        // Toggling ON — pull in the requires chain (transitively).
        const queue = [key];
        while (queue.length) {
          const k = queue.shift()!;
          if (next.has(k)) continue;
          next.add(k);
          for (const dep of byKey[k]?.requires ?? []) {
            if (!next.has(dep)) queue.push(dep);
          }
        }
      }
      return next;
    });
  }

  // Non-core selected feature count gates "Save / Activate". Plain starter
  // alone is fine since it's the published preset.
  const nonCoreSelected = useMemo(
    () => Array.from(selected).filter((k) => !byKey[k]?.isCore),
    [selected, byKey],
  );
  const canSubmit = nonCoreSelected.length > 0 && !!quote && !submitting;

  async function persist(): Promise<{ id: string } | null> {
    return await createMut.mutateAsync({
      featureKeys: Array.from(selected),
      userCount: Math.max(1, userCount || 1),
    });
  }

  async function saveDraft() {
    setSubmitting("draft");
    try {
      const res = await persist();
      if (!res) return;
      toast.success("Quotation saved", res.id);
      router.push(`/dashboard/billing/quotations/${res.id}`);
    } catch (e) {
      toast.error("Failed to save quotation", (e as Error).message);
    } finally {
      setSubmitting(null);
    }
  }

  // Activate now = persist draft + immediately call /activate. Lands on the
  // newly-issued invoice so the user can record payment in the same flow.
  async function activateNow() {
    setSubmitting("activate");
    try {
      const draft = await persist();
      if (!draft) return;
      const activation = await activateMut.mutateAsync(draft.id);
      toast.success("Plan activated", `Invoice ${activation.invoice.number} issued`);
      router.push(`/dashboard/billing/invoices/${activation.invoice.id}`);
    } catch (e) {
      toast.error("Activation failed", (e as Error).message);
    } finally {
      setSubmitting(null);
    }
  }

  if (featuresQ.isLoading) {
    return <Skeleton className="h-96 w-full" />;
  }

  return (
    <TooltipProvider>
      <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
        <div className="space-y-6">
          {Object.entries(grouped).map(([cat, list]) => (
            <div key={cat}>
              <h3 className="font-semibold text-sm uppercase tracking-wide text-muted-foreground mb-3">
                {CATEGORY_LABEL[cat as FeatureCategory] ?? cat}
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                {list.map((f) => {
                  const on = selected.has(f.key);
                  const downstream = (dependents[f.key] ?? []).filter((d) =>
                    selected.has(d),
                  );
                  const card = (
                    <button
                      key={f.id}
                      type="button"
                      onClick={() => toggle(f.key, f)}
                      disabled={f.isCore}
                      className={`text-left w-full rounded-xl border p-4 transition-all ${
                        on
                          ? "border-foreground bg-foreground/5"
                          : "border-border hover:border-foreground/40"
                      } ${f.isCore ? "opacity-70 cursor-not-allowed" : ""}`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <div className="font-medium">{f.name}</div>
                          <div className="text-xs text-muted-foreground mt-0.5">
                            {f.description}
                          </div>
                        </div>
                        <div className="text-right shrink-0">
                          {f.basePriceCents > 0 && (
                            <div className="text-sm font-medium">
                              {formatMoney("INR", f.basePriceCents)}/mo
                            </div>
                          )}
                          {f.perUserPriceCents > 0 && (
                            <div className="text-xs text-muted-foreground">
                              +{formatMoney("INR", f.perUserPriceCents)}/user
                            </div>
                          )}
                          {f.includedUsers > 0 && (
                            <div className="text-xs text-muted-foreground">
                              Includes {f.includedUsers} users
                            </div>
                          )}
                          {f.isCore && (
                            <div className="text-xs text-muted-foreground">Always on</div>
                          )}
                        </div>
                      </div>
                      {f.requires.length > 0 && (
                        <div className="mt-2 text-xs text-muted-foreground">
                          Requires:{" "}
                          {f.requires.map((r) => byKey[r]?.name ?? r).join(", ")}
                        </div>
                      )}
                      {on && downstream.length > 0 && !f.isCore && (
                        <div className="mt-2 text-xs text-amber-600 dark:text-amber-400">
                          Required by:{" "}
                          {downstream.map((d) => byKey[d]?.name ?? d).join(", ")}
                        </div>
                      )}
                    </button>
                  );
                  return f.isCore ? (
                    <Tooltip key={f.id} delayDuration={150}>
                      <TooltipTrigger asChild>{card}</TooltipTrigger>
                      <TooltipContent side="top">
                        Core features ship with every plan.
                      </TooltipContent>
                    </Tooltip>
                  ) : (
                    card
                  );
                })}
              </div>
            </div>
          ))}
        </div>

        <Card className="h-fit lg:sticky lg:top-4">
          <CardHeader>
            <CardTitle className="text-base">Your quote</CardTitle>
            <CardDescription>
              Live preview · CGST + SGST applied for intra-state customers
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="flex items-end justify-between">
                <label className="text-xs uppercase tracking-wide text-muted-foreground">
                  Users
                </label>
                <div className="text-right">
                  <div className="font-semibold text-lg">{userCount}</div>
                  {perUserCents > 0 && (
                    <div className="text-xs text-muted-foreground">
                      {Math.max(0, userCount - includedUsers)} extra · {" "}
                      {formatMoney("INR", perUserCents)}/user
                    </div>
                  )}
                </div>
              </div>
              <input
                type="range"
                min={0}
                max={stops.length - 1}
                value={Math.max(0, stops.indexOf(userCount))}
                onChange={(e) => setUserCount(stops[parseInt(e.target.value, 10)] ?? 1)}
                className="w-full mt-2 accent-foreground"
              />
              <div className="flex items-center gap-2 mt-1">
                <span className="text-[10px] text-muted-foreground">1</span>
                <Input
                  type="number"
                  min={1}
                  max={1000}
                  value={userCount}
                  onChange={(e) => setUserCount(parseInt(e.target.value || "1", 10))}
                  className="h-7 text-xs flex-1"
                />
                <span className="text-[10px] text-muted-foreground">500</span>
              </div>
            </div>

            {quoteError ? (
              <p className="text-sm text-destructive">{quoteError}</p>
            ) : !quote ? (
              <p className="text-sm text-muted-foreground">
                Pick at least one feature to see pricing.
              </p>
            ) : (
              <div className="space-y-2 text-sm">
                {quote.lines.map((l) => (
                  <div key={l.featureKey + l.sortOrder} className="flex justify-between">
                    <span className="text-muted-foreground truncate pr-2">
                      {l.description}
                      {l.quantity > 1 && <> × {l.quantity}</>}
                    </span>
                    <span>{formatMoney(quote.currency, l.totalCents)}</span>
                  </div>
                ))}
                <div className="border-t pt-2 flex justify-between">
                  <span className="text-muted-foreground">Subtotal</span>
                  <span>{formatMoney(quote.currency, quote.subtotalCents)}</span>
                </div>
                {quote.cgstCents > 0 && (
                  <>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>CGST 9%</span>
                      <span>{formatMoney(quote.currency, quote.cgstCents)}</span>
                    </div>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>SGST 9%</span>
                      <span>{formatMoney(quote.currency, quote.sgstCents)}</span>
                    </div>
                  </>
                )}
                {quote.igstCents > 0 && (
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>IGST 18%</span>
                    <span>{formatMoney(quote.currency, quote.igstCents)}</span>
                  </div>
                )}
                <div className="border-t pt-2 flex justify-between font-semibold text-base">
                  <span>Total / month</span>
                  <span>{formatMoney(quote.currency, quote.totalCents)}</span>
                </div>
              </div>
            )}

            {!canSubmit && nonCoreSelected.length === 0 && (
              <p className="text-xs text-amber-600 dark:text-amber-400">
                Pick at least one feature beyond core to save or activate.
              </p>
            )}

            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={saveDraft}
                disabled={!canSubmit}
                className="rounded-md border py-2 text-sm font-medium disabled:opacity-50"
              >
                {submitting === "draft" ? "Saving…" : "Save draft"}
              </button>
              <button
                type="button"
                onClick={activateNow}
                disabled={!canSubmit}
                className="rounded-md bg-foreground py-2 text-sm font-medium text-background disabled:opacity-50"
              >
                {submitting === "activate" ? "Activating…" : "Activate now"}
              </button>
            </div>
          </CardContent>
        </Card>
      </div>
    </TooltipProvider>
  );
}

function groupByCategory(features: Feature[]): Record<FeatureCategory, Feature[]> {
  const out: Record<FeatureCategory, Feature[]> = {
    core: [],
    admin: [],
    compliance: [],
    integrations: [],
    limits: [],
  };
  for (const f of features) {
    if (!f.isActive) continue;
    (out[f.category] ?? (out[f.category] = [])).push(f);
  }
  for (const k of Object.keys(out) as FeatureCategory[]) {
    out[k].sort((a, b) => a.sortOrder - b.sortOrder);
  }
  return out;
}
