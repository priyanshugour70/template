"use client";

import { ChevronRight, Home } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/cn";

const PRETTY: Record<string, string> = {
  dashboard: "Dashboard",
  users: "Users",
  roles: "Roles & Permissions",
  organizations: "Organizations",
  subscription: "Subscription",
  audit: "Audit Log",
  settings: "Settings",
  onboarding: "Onboarding",
  auth: "Auth",
  login: "Sign in",
  signup: "Sign up",
};

function prettify(seg: string): string {
  return PRETTY[seg] ?? seg.replace(/-/g, " ").replace(/^\w/, (c) => c.toUpperCase());
}

export function Breadcrumbs({ className }: { className?: string }) {
  const pathname = usePathname();
  const parts = pathname.split("/").filter(Boolean);

  if (parts.length === 0) return null;

  const crumbs = parts.map((part, i) => {
    const href = "/" + parts.slice(0, i + 1).join("/");
    return { href, label: prettify(part), last: i === parts.length - 1 };
  });

  return (
    <nav aria-label="Breadcrumb" className={cn("flex items-center gap-1.5 text-sm", className)}>
      <Link href="/dashboard" className="text-muted-foreground hover:text-foreground transition-colors">
        <Home className="h-4 w-4" />
      </Link>
      {crumbs.map((c, i) => (
        <span key={c.href} className="flex items-center gap-1.5">
          <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />
          {c.last || i === crumbs.length - 1 ? (
            <span className="text-foreground font-medium">{c.label}</span>
          ) : (
            <Link href={c.href} className="text-muted-foreground hover:text-foreground transition-colors">
              {c.label}
            </Link>
          )}
        </span>
      ))}
    </nav>
  );
}
