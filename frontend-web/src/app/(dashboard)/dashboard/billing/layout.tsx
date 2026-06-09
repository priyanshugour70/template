"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { cn } from "@/lib/cn";

// Top-level tab bar shared across all billing routes. Matches the Settings /
// Administrative pattern so users get a consistent breadcrumb.
const TABS = [
  { href: "/dashboard/billing", label: "Overview", exact: true },
  { href: "/dashboard/billing/subscription", label: "Subscription" },
  { href: "/dashboard/billing/plan-builder", label: "Plan builder" },
  { href: "/dashboard/billing/quotations", label: "Quotations" },
  { href: "/dashboard/billing/invoices", label: "Invoices" },
  { href: "/dashboard/billing/transactions", label: "Transactions" },
];

export default function BillingLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Billing</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Plans, invoices, payments, and tax-compliant receipts.
        </p>
      </div>

      <nav className="flex gap-1 border-b -mb-px overflow-x-auto">
        {TABS.map((t) => {
          const active = t.exact ? pathname === t.href : pathname.startsWith(t.href);
          return (
            <Link
              key={t.href}
              href={t.href}
              className={cn(
                "px-4 py-2 text-sm border-b-2 -mb-px whitespace-nowrap",
                active
                  ? "border-foreground font-medium text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {t.label}
            </Link>
          );
        })}
      </nav>

      <div>{children}</div>
    </div>
  );
}
