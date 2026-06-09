"use client";

import { Home, ShieldOff } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { usePermissions } from "@/providers";

interface PermissionGateProps {
  /** Render children only if the user has ALL of these permissions. */
  required?: string[];
  /** Render children only if the user has AT LEAST ONE of these. */
  anyOf?: string[];
  /** When the gate fails, optionally show this instead of the default 403 panel. */
  fallback?: ReactNode;
  /** Hide the fallback entirely on failure — useful when the gate sits inside
   * a larger conditional and you want a silent skip. */
  silent?: boolean;
  children: ReactNode;
}

/** Client-side permission gate. The sidebar already hides what the user can't
 * use, but a deep-link or bookmark can still land them on a restricted page —
 * this catches that case with a friendly explanation instead of a broken UI. */
export function PermissionGate({
  required,
  anyOf,
  fallback,
  silent,
  children,
}: PermissionGateProps) {
  const { hasAll, hasAny, isSuperAdmin } = usePermissions();

  const ok =
    isSuperAdmin ||
    ((!required || required.length === 0 || hasAll(required)) &&
      (!anyOf || anyOf.length === 0 || hasAny(anyOf)));

  if (ok) return <>{children}</>;
  if (silent) return null;
  if (fallback !== undefined) return <>{fallback}</>;

  return <NoAccessPanel />;
}

/** The inline "need access" panel — same look as forbidden.tsx but embedded
 * inside the dashboard shell (sidebar + header stay visible). */
export function NoAccessPanel() {
  return (
    <Card className="border-warning/40 bg-warning/5">
      <CardContent className="flex flex-col items-center gap-4 py-12 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-warning/15">
          <ShieldOff className="h-6 w-6 text-warning" aria-hidden />
        </div>
        <div>
          <h2 className="text-lg font-semibold">You don&apos;t have access</h2>
          <p className="mt-1 max-w-md text-sm text-muted-foreground">
            Your current role doesn&apos;t allow you to view this page. Ask a workspace owner
            or admin for the right permissions, or head back to the dashboard.
          </p>
        </div>
        <Button asChild>
          <Link href="/dashboard">
            <Home className="h-4 w-4" />
            Back to dashboard
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
}
