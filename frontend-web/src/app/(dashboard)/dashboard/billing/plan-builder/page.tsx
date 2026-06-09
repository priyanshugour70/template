"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "@/hooks/use-toast";
import {
  useCreateQuotation,
  useFeatureCatalog,
  usePreviewQuote,
} from "@/hooks/billing/useBilling";
import type { Feature, FeatureCategory, Quote } from "@/types/billing";

import { formatMoney } from "../_components/money";

const CATEGORY_LABEL: Record<FeatureCategory, string> = {
  core: "Core",
  admin: "Admin & access",
  compliance: "Compliance",
  integrations: "Integrations",
  limits: "Limits & scale",
};

export default function PlanBuilderPage() {
  const router = useRouter();
  const featuresQ = useFeatureCatalog();
  const previewMut = usePreviewQuote();
  const createMut = useCreateQuotation();

  const features = featuresQ.data ?? [];
  const grouped = useMemo(() => groupByCategory(features), [features]);

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [userCount, setUserCount] = useState(10);

  // Seed selection once features load: anything marked starter-default OR core.
  useEffect(() => {
    if (!features.length || selected.size > 0) return;
    const next = new Set<string>();
    for (const f of features) {
      if (f.isCore || f.isStarterDefault) next.add(f.key);
    }
    setSelected(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [features.length]);

  // Debounced live preview — runs on selection/userCount change.
  const [quote, setQuote] = useState<Quote | null>(null);
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
        .then(setQuote)
        .catch(() => setQuote(null));
    }, 300);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, userCount]);

  function toggle(key: string, feature: Feature) {
    if (feature.isCore) return; // core features are always on
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
        // Pull in "requires" chain.
        for (const dep of feature.requires ?? []) next.add(dep);
      }
      return next;
    });
  }

  async function saveDraft() {
    try {
      const res = await createMut.mutateAsync({
        featureKeys: Array.from(selected),
        userCount: Math.max(1, userCount || 1),
      });
      toast.success("Quotation saved", res.number);
      router.push(`/dashboard/billing/quotations/${res.id}`);
    } catch (e) {
      toast.error("Failed to save quotation", (e as Error).message);
    }
  }

  if (featuresQ.isLoading) {
    return <Skeleton className="h-96 w-full" />;
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
      <div className="space-y-6">
        {Object.entries(grouped).map(([cat, list]) => (
          <div key={cat}>
            <h3 className="font-semibold text-sm uppercase tracking-wide text-muted-foreground mb-3">
              {CATEGORY_LABEL[cat as FeatureCategory] ?? cat}
            </h3>
            <div className="grid gap-3 sm:grid-cols-2">
              {list.map((f) => {
                const on = selected.has(f.key);
                const disabled = f.isCore;
                return (
                  <button
                    key={f.id}
                    type="button"
                    onClick={() => toggle(f.key, f)}
                    disabled={disabled}
                    className={`text-left rounded-xl border p-4 transition-all ${
                      on
                        ? "border-foreground bg-foreground/5"
                        : "border-border hover:border-foreground/40"
                    } ${disabled ? "opacity-70 cursor-not-allowed" : ""}`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div>
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
                        {f.isCore && (
                          <div className="text-xs text-muted-foreground">Always on</div>
                        )}
                      </div>
                    </div>
                    {f.requires.length > 0 && (
                      <div className="mt-2 text-xs text-muted-foreground">
                        Requires: {f.requires.join(", ")}
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      <Card className="h-fit sticky top-4">
        <CardHeader>
          <CardTitle className="text-base">Your quote</CardTitle>
          <CardDescription>Live preview · GST applied automatically</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="text-xs uppercase tracking-wide text-muted-foreground">
              Users
            </label>
            <Input
              type="number"
              min={1}
              value={userCount}
              onChange={(e) => setUserCount(parseInt(e.target.value || "1", 10))}
              className="mt-1"
            />
          </div>

          {!quote ? (
            <p className="text-sm text-muted-foreground">Pick at least one feature to see pricing.</p>
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

          <button
            type="button"
            onClick={saveDraft}
            disabled={!quote || createMut.isPending}
            className="w-full rounded-md bg-foreground py-2 text-sm font-medium text-background disabled:opacity-50"
          >
            {createMut.isPending ? "Saving…" : "Save as quotation"}
          </button>
        </CardContent>
      </Card>
    </div>
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
