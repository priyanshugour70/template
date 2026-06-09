import { Home, ShieldOff } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";

/** Rendered when server code calls `forbidden()` from next/navigation
 * (requires experimental.authInterrupts = true). Returns HTTP 403. */
export default function Forbidden() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-warning/10">
          <ShieldOff className="h-7 w-7 text-warning" strokeWidth={1.5} aria-hidden />
        </div>
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          403 — Access denied
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">You don&apos;t have access</h1>
        <p className="mt-3 text-sm text-muted-foreground">
          Your account is signed in, but the role you currently hold doesn&apos;t allow you to
          view this page. Ask a workspace owner or admin for the right permissions.
        </p>
        <div className="mt-6 flex items-center justify-center gap-3">
          <Button asChild>
            <Link href="/dashboard">
              <Home className="h-4 w-4" />
              Back to dashboard
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
