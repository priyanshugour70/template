"use client";

import { AlertCircle, Home, RefreshCw } from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";

import { Button } from "@/components/ui/button";

/** Root-level error boundary. Catches any unhandled error thrown during render
 * in routes that don't have a more specific error.tsx. */
export default function RootError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Unhandled app error:", error);
  }, [error]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
          <AlertCircle className="h-7 w-7 text-destructive" strokeWidth={1.5} aria-hidden />
        </div>
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Something went wrong
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">We hit an unexpected error</h1>
        <p className="mt-3 text-sm text-muted-foreground">
          The page couldn&apos;t load. Try again, or head back to the dashboard. If this keeps
          happening, contact support with the reference code below.
        </p>
        {error.digest && (
          <p className="mt-4 inline-block rounded border border-border bg-muted/60 px-2 py-1 font-mono text-[11px] text-muted-foreground">
            ref: {error.digest}
          </p>
        )}
        <div className="mt-6 flex items-center justify-center gap-3">
          <Button onClick={() => reset()}>
            <RefreshCw className="h-4 w-4" />
            Try again
          </Button>
          <Button variant="outline" asChild>
            <Link href="/dashboard">
              <Home className="h-4 w-4" />
              Dashboard
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
