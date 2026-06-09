"use client";

import { AlertCircle, RefreshCw } from "lucide-react";
import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

/** Error boundary inside the dashboard. Preserves the sidebar so the user can
 * navigate away without a full page reload. */
export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Dashboard route error:", error);
  }, [error]);

  return (
    <Card className="border-destructive/30 bg-destructive/5">
      <CardContent className="flex flex-col items-center gap-4 py-12 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
          <AlertCircle className="h-6 w-6 text-destructive" aria-hidden />
        </div>
        <div>
          <h2 className="text-lg font-semibold">Couldn&apos;t load this page</h2>
          <p className="mt-1 max-w-md text-sm text-muted-foreground">
            Something went wrong while fetching data. Try again — if it keeps failing, use the
            sidebar to navigate elsewhere or contact support.
          </p>
          {error.digest && (
            <p className="mt-3 inline-block rounded border border-border bg-muted/60 px-2 py-1 font-mono text-[11px] text-muted-foreground">
              ref: {error.digest}
            </p>
          )}
        </div>
        <Button onClick={() => reset()}>
          <RefreshCw className="h-4 w-4" />
          Try again
        </Button>
      </CardContent>
    </Card>
  );
}
