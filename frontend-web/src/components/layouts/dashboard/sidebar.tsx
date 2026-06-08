"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  Building2,
  ChevronDown,
  ChevronsLeft,
  ChevronsRight,
  CreditCard,
  Home,
  Lock,
  Settings,
  Shield,
  Users,
  type LucideIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/cn";
import { usePermissions, useTenant } from "@/providers";
import { useSidebarStore } from "@/stores/ui/sidebar.store";

interface NavItem {
  href: string;
  label: string;
  icon: LucideIcon;
  permission?: string;
  badge?: string;
}

interface NavGroup {
  id: string;
  label: string;
  icon?: LucideIcon;
  items: NavItem[];
}

const GROUPS: NavGroup[] = [
  {
    id: "overview",
    label: "Overview",
    items: [{ href: "/dashboard", label: "Home", icon: Home }],
  },
  {
    id: "administrative",
    label: "Administrative",
    icon: Shield,
    items: [
      { href: "/dashboard/users", label: "Users", icon: Users, permission: "user.list" },
      { href: "/dashboard/roles", label: "Roles & Permissions", icon: Lock, permission: "role.list" },
      { href: "/dashboard/organizations", label: "Organizations", icon: Building2, permission: "org.list" },
      { href: "/dashboard/audit", label: "Audit Log", icon: Activity, permission: "audit.read" },
    ],
  },
  {
    id: "billing",
    label: "Billing",
    icon: CreditCard,
    items: [
      { href: "/dashboard/subscription", label: "Subscription", icon: CreditCard, permission: "subscription.read" },
    ],
  },
  {
    id: "system",
    label: "System",
    items: [{ href: "/dashboard/settings", label: "Settings", icon: Settings }],
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { tenant, activeOrganization } = useTenant();
  const { has } = usePermissions();
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggleCollapsed = useSidebarStore((s) => s.toggleCollapsed);
  const sections = useSidebarStore((s) => s.sections);
  const toggleSection = useSidebarStore((s) => s.toggleSection);

  const visibleGroups = GROUPS.map((g) => ({
    ...g,
    items: g.items.filter((it) => !it.permission || has(it.permission)),
  })).filter((g) => g.items.length > 0);

  return (
    <aside
      className={cn(
        "hidden md:flex md:flex-col md:fixed md:inset-y-0 md:left-0 md:z-30",
        "border-r bg-background transition-[width] duration-200 ease-out",
        collapsed ? "md:w-16" : "md:w-64",
      )}
    >
      {/* Brand */}
      <div className={cn("flex h-16 items-center border-b px-3", collapsed ? "justify-center" : "gap-3 px-5")}>
        {tenant?.logoUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={tenant.logoUrl}
            alt={tenant.name}
            className="h-9 w-9 rounded-md object-cover ring-1 ring-border"
          />
        ) : (
          <div className="h-9 w-9 rounded-md bg-primary/10 flex items-center justify-center text-primary font-semibold">
            {tenant?.name?.[0]?.toUpperCase() ?? "A"}
          </div>
        )}
        {!collapsed && (
          <div className="flex-1 overflow-hidden">
            <div className="text-sm font-semibold truncate">{tenant?.name ?? "Workspace"}</div>
            <div className="text-xs text-muted-foreground truncate">
              {activeOrganization?.name ?? "—"}
            </div>
          </div>
        )}
      </div>

      <nav className="flex-1 overflow-y-auto py-2">
        {visibleGroups.map((group, idx) => {
          const isCollapsed = sections[group.id] === true;
          const GroupIcon = group.icon;
          return (
            <div key={group.id} className={cn("px-2", idx > 0 && "mt-3")}>
              {!collapsed && (
                <button
                  onClick={() => toggleSection(group.id)}
                  className="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors"
                >
                  <span className="flex items-center gap-2">
                    {GroupIcon ? <GroupIcon className="h-3.5 w-3.5" /> : null}
                    {group.label}
                  </span>
                  <ChevronDown
                    className={cn("h-3.5 w-3.5 transition-transform", isCollapsed && "-rotate-90")}
                  />
                </button>
              )}

              {(!collapsed && !isCollapsed) || collapsed ? (
                <ul className="mt-1 space-y-0.5">
                  {group.items.map((item) => {
                    const active = pathname === item.href || pathname.startsWith(item.href + "/");
                    const Icon = item.icon;
                    return (
                      <li key={item.href}>
                        <Link
                          href={item.href}
                          title={collapsed ? item.label : undefined}
                          className={cn(
                            "group flex items-center gap-3 rounded-md text-sm transition-colors",
                            collapsed ? "h-10 w-10 mx-auto justify-center" : "px-3 py-2",
                            active
                              ? "bg-primary/10 text-primary font-medium"
                              : "text-foreground/70 hover:bg-accent hover:text-foreground",
                          )}
                        >
                          <Icon className="h-4 w-4 shrink-0" />
                          {!collapsed && <span className="truncate">{item.label}</span>}
                          {!collapsed && item.badge && (
                            <span className="ml-auto rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary">
                              {item.badge}
                            </span>
                          )}
                        </Link>
                      </li>
                    );
                  })}
                </ul>
              ) : null}
            </div>
          );
        })}
      </nav>

      <Separator />
      <div className={cn("p-2 flex", collapsed ? "justify-center" : "justify-end")}>
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleCollapsed}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          {collapsed ? <ChevronsRight className="h-4 w-4" /> : <ChevronsLeft className="h-4 w-4" />}
        </Button>
      </div>
    </aside>
  );
}
