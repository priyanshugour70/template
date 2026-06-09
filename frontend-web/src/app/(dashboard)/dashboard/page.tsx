"use client";

import {
  Activity,
  ArrowUpRight,
  Building2,
  CreditCard,
  Lock,
  TrendingUp,
  Users,
} from "lucide-react";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth, useTenant } from "@/providers";
import { useActiveSubscription, useFeatureSet } from "@/hooks/subscription/useSubscriptionQueries";
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { useUsers } from "@/hooks/user/useUserQueries";

export default function DashboardHome() {
  const { user } = useAuth();
  const { tenant, activeOrganization } = useTenant();
  const usersQ = useUsers({ limit: 100 });
  const rolesQ = useRoles();
  const subQ = useActiveSubscription();
  const featuresQ = useFeatureSet();

  const userCount = usersQ.data?.length ?? 0;
  const roleCount = rolesQ.data?.length ?? 0;
  const planLimit = featuresQ.data?.limits["users.max"];
  const usagePct =
    planLimit != null && planLimit > 0 ? Math.min(100, (userCount / planLimit) * 100) : 0;

  return (
    <div className="space-y-8">
      {/* Greeting */}
      <div className="flex items-center justify-between flex-wrap gap-4">
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
        <Link href="/dashboard/subscription">
          <Badge variant="muted" className="cursor-pointer hover:bg-accent gap-1.5">
            <CreditCard className="h-3.5 w-3.5" />
            {subQ.data?.planCode ?? "—"} plan
            <ArrowUpRight className="h-3 w-3" />
          </Badge>
        </Link>
      </div>

      {/* Top metrics */}
      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
        <Metric
          label="Team members"
          value={userCount}
          icon={Users}
          loading={usersQ.isLoading}
          hint={planLimit != null && planLimit > 0 ? `of ${planLimit} on ${subQ.data?.planCode ?? "plan"}` : undefined}
          progress={planLimit != null && planLimit > 0 ? usagePct : undefined}
          href="/dashboard/administrative/users"
        />
        <Metric
          label="Roles"
          value={roleCount}
          icon={Lock}
          loading={rolesQ.isLoading}
          hint="Owner, Admin, Member by default"
          href="/dashboard/administrative/roles"
        />
        <Metric
          label="Subscription"
          value={subQ.data?.planCode ?? "—"}
          icon={CreditCard}
          loading={subQ.isLoading}
          hint={subQ.data?.status}
          href="/dashboard/subscription"
        />
        <Metric
          label="Active features"
          value={featuresQ.data ? Object.keys(featuresQ.data.features).length : 0}
          icon={TrendingUp}
          loading={featuresQ.isLoading}
          hint="from your current plan"
        />
      </div>

      {/* Quick links */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Jump to</h2>
        <div className="grid gap-3 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
          <QuickLink href="/dashboard/administrative/users" icon={Users} title="Manage users" sub="Invite teammates, suspend, archive" />
          <QuickLink
            href="/dashboard/administrative/roles"
            icon={Lock}
            title="Roles & permissions"
            sub="Bundle permissions, assign to members"
          />
          <QuickLink
            href="/dashboard/administrative/organizations"
            icon={Building2}
            title="Organizations"
            sub="Manage workspaces inside your tenant"
          />
          <QuickLink
            href="/dashboard/administrative/audit"
            icon={Activity}
            title="Audit log"
            sub="Every API call, with filters and search"
          />
        </div>
      </div>

      {/* Tenant card */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Workspace details</h2>
        <Card>
          <div className="grid md:grid-cols-3 divide-y md:divide-y-0 md:divide-x">
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">Tenant</div>
              <div className="text-lg font-semibold">{tenant?.name ?? "—"}</div>
              <div className="text-sm text-muted-foreground mt-1">/{tenant?.slug ?? "—"}</div>
            </div>
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">Active organization</div>
              <div className="text-lg font-semibold">{activeOrganization?.name ?? "—"}</div>
              <div className="text-sm text-muted-foreground mt-1">/{activeOrganization?.slug ?? "—"}</div>
            </div>
            <div className="p-6">
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-1.5">Signed in as</div>
              <div className="text-lg font-semibold">{user?.displayName ?? user?.email}</div>
              <div className="text-sm text-muted-foreground mt-1">{user?.email}</div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}

function Metric({
  label,
  value,
  icon: Icon,
  hint,
  href,
  loading,
  progress,
}: {
  label: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  hint?: string;
  href?: string;
  loading?: boolean;
  progress?: number;
}) {
  const card = (
    <Card className={"group transition-shadow " + (href ? "hover:shadow-md cursor-pointer" : "")}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardDescription>{label}</CardDescription>
          <Icon className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
        </div>
        {loading ? (
          <Skeleton className="h-9 w-24 mt-1" />
        ) : (
          <CardTitle className="text-3xl capitalize">{value}</CardTitle>
        )}
      </CardHeader>
      {(hint || progress != null) && (
        <CardContent className="space-y-2">
          {hint && <div className="text-xs text-muted-foreground">{hint}</div>}
          {progress != null && (
            <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
              <div className="h-full bg-primary rounded-full" style={{ width: `${progress}%` }} />
            </div>
          )}
        </CardContent>
      )}
    </Card>
  );
  return href ? <Link href={href}>{card}</Link> : card;
}

function QuickLink({
  href,
  icon: Icon,
  title,
  sub,
}: {
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  sub: string;
}) {
  return (
    <Link href={href} className="group">
      <Card className="hover:border-primary/40 transition-colors">
        <CardHeader>
          <div className="flex items-center justify-between">
            <Icon className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
            <ArrowUpRight className="h-4 w-4 text-muted-foreground/50 group-hover:text-primary transition-colors" />
          </div>
          <CardTitle className="text-base mt-3">{title}</CardTitle>
          <CardDescription>{sub}</CardDescription>
        </CardHeader>
      </Card>
    </Link>
  );
}
