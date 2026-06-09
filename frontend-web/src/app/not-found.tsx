import { Home, Search } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          <Search className="h-7 w-7 text-muted-foreground" strokeWidth={1.5} aria-hidden />
        </div>
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          404
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Page not found</h1>
        <p className="mt-3 text-sm text-muted-foreground">
          The page you&apos;re looking for doesn&apos;t exist or has moved. Double-check the URL,
          or jump back to your dashboard.
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
