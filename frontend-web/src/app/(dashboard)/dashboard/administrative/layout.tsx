"use client";

import { Activity, Building2, GitBranch, Lock, Users, UsersRound } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { cn } from "@/lib/cn";
import { usePermissions } from "@/providers";

interface AdminTab {
  href: string;
  label: string;
  icon: typeof Users;
  permission?: string;
}

const TABS: AdminTab[] = [
  { href: "/dashboard/administrative/users", label: "Users", icon: Users, permission: "user.list" },
  { href: "/dashboard/administrative/roles", label: "Roles", icon: Lock, permission: "role.list" },
  { href: "/dashboard/administrative/departments", label: "Departments", icon: GitBranch, permission: "department.list" },
  { href: "/dashboard/administrative/groups", label: "Groups", icon: UsersRound, permission: "group.list" },
  { href: "/dashboard/administrative/organizations", label: "Organizations", icon: Building2, permission: "org.list" },
  { href: "/dashboard/administrative/audit", label: "Audit log", icon: Activity, permission: "audit.read" },
];

export default function AdministrativeLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const { has } = usePermissions();

  const visible = TABS.filter((t) => !t.permission || has(t.permission));

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Administrative</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Users, roles, organizations, and activity for this workspace.
        </p>
      </div>

      <div className="border-b border-border">
        <nav className="-mb-px flex flex-wrap gap-x-2 gap-y-1">
          {visible.map((t) => {
            const active = pathname === t.href || pathname.startsWith(`${t.href}/`);
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

      <div>{children}</div>
    </div>
  );
}
