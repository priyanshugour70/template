"use client";

import { Bell, Building2, Code2, LayoutGrid, Lock, Monitor, User } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { NoAccessPanel } from "@/components/shared/permission-gate";
import { cn } from "@/lib/cn";
import { usePermissions } from "@/providers";

interface SettingsTab {
  href: string;
  label: string;
  icon: typeof User;
  /** Show only if the user has any of these permissions. Omit for everyone. */
  anyPermission?: string[];
}

const TABS: SettingsTab[] = [
  { href: "/dashboard/settings", label: "Overview", icon: LayoutGrid },
  { href: "/dashboard/settings/profile", label: "Profile", icon: User },
  { href: "/dashboard/settings/security", label: "Security", icon: Lock },
  { href: "/dashboard/settings/sessions", label: "Sessions", icon: Monitor },
  { href: "/dashboard/settings/notifications", label: "Notifications", icon: Bell },
  {
    href: "/dashboard/settings/developer",
    label: "Developer",
    icon: Code2,
    anyPermission: ["api_key.list", "webhook.list"],
  },
  {
    href: "/dashboard/settings/tenant",
    label: "Tenant",
    icon: Building2,
    anyPermission: ["tenant.update"],
  },
];

function isActive(pathname: string, href: string) {
  if (href === "/dashboard/settings") return pathname === href;
  return pathname === href || pathname.startsWith(`${href}/`);
}

export default function SettingsLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const { hasAny, isSuperAdmin } = usePermissions();

  const visible = TABS.filter((t) => !t.anyPermission || hasAny(t.anyPermission));
  // Find the tab whose href is the longest prefix of the current path — that's
  // the active sub-route. If that tab requires a permission and the user
  // doesn't hold it, render the No Access panel instead of the children.
  const current = TABS.filter((t) => isActive(pathname, t.href)).reduce<
    (typeof TABS)[number] | undefined
  >((a, b) => (a && a.href.length >= b.href.length ? a : b), undefined);
  const accessDenied =
    current?.anyPermission !== undefined &&
    !isSuperAdmin &&
    !hasAny(current.anyPermission);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Profile, security, sessions, notifications, and workspace administration.
        </p>
      </div>

      <div className="border-b border-border">
        <nav className="-mb-px flex flex-wrap gap-x-2 gap-y-1">
          {visible.map((t) => {
            const active = isActive(pathname, t.href);
            const Icon = t.icon;
            return (
              <Link
                key={t.href}
                href={t.href}
                className={cn(
                  "inline-flex items-center gap-2 border-b-2 px-3 py-2 text-sm font-medium transition-colors",
                  active
                    ? "border-primary text-foreground"
                    : "border-transparent text-muted-foreground hover:border-border hover:text-foreground",
                )}
              >
                <Icon className="h-4 w-4" />
                {t.label}
              </Link>
            );
          })}
        </nav>
      </div>

      <div>{accessDenied ? <NoAccessPanel /> : children}</div>
    </div>
  );
}
