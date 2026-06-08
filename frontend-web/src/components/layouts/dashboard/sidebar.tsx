"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  Building2,
  CreditCard,
  Home,
  Lock,
  Settings,
  Users,
  type LucideIcon,
} from "lucide-react";

import { useTenant } from "@/providers";
import { usePermissions } from "@/providers";
import { cn } from "@/lib/cn";

interface NavItem {
  href: string;
  label: string;
  icon: LucideIcon;
  permission?: string;
}

const NAV_GROUPS: { title: string; items: NavItem[] }[] = [
  {
    title: "Overview",
    items: [{ href: "/dashboard", label: "Home", icon: Home }],
  },
  {
    title: "Workspace",
    items: [
      { href: "/dashboard/users", label: "Users", icon: Users, permission: "user.list" },
      { href: "/dashboard/roles", label: "Roles & Permissions", icon: Lock, permission: "role.list" },
      { href: "/dashboard/organizations", label: "Organizations", icon: Building2, permission: "org.list" },
    ],
  },
  {
    title: "Billing",
    items: [
      { href: "/dashboard/subscription", label: "Subscription", icon: CreditCard, permission: "subscription.read" },
    ],
  },
  {
    title: "System",
    items: [
      { href: "/dashboard/audit", label: "Audit Log", icon: Activity, permission: "audit.read" },
      { href: "/dashboard/settings", label: "Settings", icon: Settings },
    ],
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { tenant, activeOrganization } = useTenant();
  const { has } = usePermissions();

  return (
    <aside className="hidden md:flex md:flex-col md:fixed md:inset-y-0 md:left-0 md:z-30 md:w-64 md:border-r md:bg-background">
      <div className="flex h-16 items-center gap-3 border-b px-5">
        {tenant?.logoUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={tenant.logoUrl} alt={tenant.name} className="h-8 w-8 rounded-md object-cover" />
        ) : (
          <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center text-primary font-semibold">
            {tenant?.name?.[0]?.toUpperCase() ?? "A"}
          </div>
        )}
        <div className="flex-1 overflow-hidden">
          <div className="text-sm font-semibold truncate">{tenant?.name ?? "Workspace"}</div>
          <div className="text-xs text-muted-foreground truncate">{activeOrganization?.name ?? "—"}</div>
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto p-3">
        {NAV_GROUPS.map((group) => {
          const visible = group.items.filter((it) => !it.permission || has(it.permission));
          if (visible.length === 0) return null;
          return (
            <div key={group.title} className="mb-6">
              <div className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {group.title}
              </div>
              <ul className="space-y-1">
                {visible.map((item) => {
                  const active = pathname === item.href || pathname.startsWith(item.href + "/");
                  const Icon = item.icon;
                  return (
                    <li key={item.href}>
                      <Link
                        href={item.href}
                        className={cn(
                          "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors",
                          active
                            ? "bg-primary/10 text-primary font-medium"
                            : "text-foreground/70 hover:bg-accent hover:text-foreground",
                        )}
                      >
                        <Icon className="h-4 w-4" />
                        <span>{item.label}</span>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          );
        })}
      </nav>

      <div className="p-4 border-t text-xs text-muted-foreground">
        {tenant?.name ?? "App"} · v0.1
      </div>
    </aside>
  );
}
