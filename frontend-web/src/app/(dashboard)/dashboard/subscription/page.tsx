"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useActiveSubscription, useFeatureSet, usePlans } from "@/hooks/subscription/useSubscriptionQueries";

function formatPrice(cents: number, currency: string) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: currency || "INR",
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

export default function SubscriptionPage() {
  const activeQ = useActiveSubscription();
  const plansQ = usePlans();
  const featuresQ = useFeatureSet();

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Subscription</h1>
        <p className="text-muted-foreground mt-1">
          Current plan, features, and quota for this organization.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardDescription>Current plan</CardDescription>
          <CardTitle className="capitalize text-3xl">
            {activeQ.data?.planCode ?? "—"}
          </CardTitle>
          {activeQ.data?.status && (
            <Badge variant={activeQ.data.status === "active" ? "success" : "warning"}>
              {activeQ.data.status}
            </Badge>
          )}
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-1">
          {activeQ.data?.currentPeriodEnd && (
            <div>
              Renews on{" "}
              <span className="text-foreground font-medium">
                {new Date(activeQ.data.currentPeriodEnd).toLocaleDateString()}
              </span>
            </div>
          )}
          {featuresQ.data && (
            <div className="pt-3">
              <div className="font-medium text-foreground mb-2">Included</div>
              <div className="flex flex-wrap gap-1">
                {Object.keys(featuresQ.data.features).map((f) => (
                  <Badge key={f} variant="muted">{f}</Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <div>
        <h2 className="text-lg font-semibold mb-4">All plans</h2>
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
          {plansQ.data?.map((p) => (
            <Card key={p.id} className={p.code === activeQ.data?.planCode ? "border-primary" : ""}>
              <CardHeader>
                <CardDescription className="capitalize">{p.code}</CardDescription>
                <CardTitle className="text-xl">{p.name}</CardTitle>
                <div className="text-2xl font-semibold mt-2">
                  {formatPrice(p.priceCents, p.currency)}
                  <span className="text-sm text-muted-foreground font-normal">/{p.billingCycle}</span>
                </div>
              </CardHeader>
              <CardContent className="space-y-2">
                <p className="text-sm text-muted-foreground">{p.description ?? "—"}</p>
                {p.trialDays > 0 && (
                  <Badge variant="muted">{p.trialDays}-day trial</Badge>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
