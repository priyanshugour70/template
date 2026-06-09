"use client";

import {
  Activity,
  Building2,
  ChevronRight,
  GitBranch,
  KeyRound,
  Users,
  UsersRound,
} from "lucide-react";
import Link from "next/link";

import { Card, CardContent, CardHeader } from "@/components/ui";
import { usePermissions } from "@/providers";

interface AdminTile {
  href: string;
  title: string;
  description: string;
  icon: typeof Users;
  /** Show only if the user has any of these. */
  anyPermission?: string[];
}

const TILES: AdminTile[] = [
  {
    href: "/dashboard/administrative/users",
    title: "Users",
    description: "Invite, suspend, change roles, and manage team members.",
    icon: Users,
    anyPermission: ["user.list"],
  },
  {
    href: "/dashboard/administrative/roles",
    title: "Roles & Permissions",
    description: "Define which permissions each role grants, and create custom roles.",
    icon: KeyRound,
    anyPermission: ["role.list"],
  },
  {
    href: "/dashboard/administrative/departments",
    title: "Departments",
    description: "Hierarchy of departments — role grants flow down the tree.",
    icon: GitBranch,
    anyPermission: ["department.list"],
  },
  {
    href: "/dashboard/administrative/groups",
    title: "Groups",
    description: "Cross-cutting collections of users for ad-hoc role bundles.",
    icon: UsersRound,
    anyPermission: ["group.list"],
  },
  {
    href: "/dashboard/administrative/organizations",
    title: "Organizations",
    description: "Workspaces inside your tenant. Each has its own users and subscription.",
    icon: Building2,
    anyPermission: ["org.list"],
  },
  {
    href: "/dashboard/administrative/audit",
    title: "Audit log",
    description: "Every API request captured. Investigate access, errors, and changes.",
    icon: Activity,
    anyPermission: ["audit.read"],
  },
];

export default function AdministrativeOverviewPage() {
  const { hasAny } = usePermissions();
  const visible = TILES.filter((t) => !t.anyPermission || hasAny(t.anyPermission));

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Pick a section above, or jump in below. You only see what your roles allow.
      </p>

      {visible.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 py-12 text-center text-sm text-muted-foreground">
            <p>You don&apos;t have access to any administrative areas in this workspace.</p>
            <p className="text-xs">
              Contact a workspace owner to request the roles you need.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {visible.map(({ href, title, description, icon: Icon }) => (
            <Link key={href} href={href} className="group block">
              <Card className="h-full transition-colors hover:bg-muted/30">
                <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2">
                  <div className="flex items-center gap-2">
                    <Icon
                      className="h-4 w-4 text-muted-foreground"
                      strokeWidth={1.5}
                      aria-hidden
                    />
                    <span className="text-sm font-medium group-hover:underline">{title}</span>
                  </div>
                  <ChevronRight
                    className="h-4 w-4 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
                    aria-hidden
                  />
                </CardHeader>
                <CardContent className="pt-0">
                  <p className="text-xs text-muted-foreground">{description}</p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
