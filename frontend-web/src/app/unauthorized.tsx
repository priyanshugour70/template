import { KeyRound, LogIn } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";

/** Rendered when server code calls `unauthorized()` from next/navigation
 * (requires experimental.authInterrupts = true). Returns HTTP 401. */
export default function Unauthorized() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          <KeyRound className="h-7 w-7 text-muted-foreground" strokeWidth={1.5} aria-hidden />
        </div>
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          401 — Sign in required
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Please sign in to continue</h1>
        <p className="mt-3 text-sm text-muted-foreground">
          Your session has expired or you aren&apos;t signed in. Log back in to pick up where
          you left off.
        </p>
        <div className="mt-6 flex items-center justify-center gap-3">
          <Button asChild>
            <Link href="/auth/login">
              <LogIn className="h-4 w-4" />
              Sign in
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
