"use client";

import { Activity, Building2, CreditCard, Users } from "lucide-react";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth, useTenant } from "@/providers";
import { useUsers } from "@/hooks/user/useUserQueries";
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { useActiveSubscription } from "@/hooks/subscription/useSubscriptionQueries";

const CARDS = [
  { href: "/dashboard/users", title: "Users", icon: Users },
  { href: "/dashboard/roles", title: "Roles", icon: Building2 },
  { href: "/dashboard/subscription", title: "Subscription", icon: CreditCard },
  { href: "/dashboard/audit", title: "Audit Log", icon: Activity },
];

export default function DashboardHome() {
  const { user } = useAuth();
  const { tenant, activeOrganization } = useTenant();
  const usersQ = useUsers({ limit: 5 });
  const rolesQ = useRoles();
  const subQ = useActiveSubscription();

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome back{user?.firstName ? `, ${user.firstName}` : ""}
        </h1>
        <p className="text-muted-foreground mt-1">
          Signed in to{" "}
          <span className="font-medium text-foreground">{tenant?.name ?? "—"}</span>
          {activeOrganization ? <> · {activeOrganization.name}</> : null}.
        </p>
      </div>

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Users</CardDescription>
            <CardTitle className="text-3xl">{usersQ.data?.length ?? "—"}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Roles</CardDescription>
            <CardTitle className="text-3xl">{rolesQ.data?.length ?? "—"}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Plan</CardDescription>
            <CardTitle className="text-2xl capitalize">{subQ.data?.planCode ?? "—"}</CardTitle>
          </CardHeader>
          <CardContent>
            {subQ.data?.status ? (
              <Badge variant={subQ.data.status === "active" ? "success" : "warning"}>
                {subQ.data.status}
              </Badge>
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Tenant</CardDescription>
            <CardTitle className="text-2xl">{tenant?.slug ?? "—"}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
        {CARDS.map(({ href, title, icon: Icon }) => (
          <Link key={href} href={href} className="group">
            <Card className="hover:border-primary/40 transition-colors">
              <CardHeader>
                <Icon className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
                <CardTitle className="text-base mt-3">{title}</CardTitle>
                <CardDescription>Manage {title.toLowerCase()}</CardDescription>
              </CardHeader>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
