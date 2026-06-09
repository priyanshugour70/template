"use client";

import { Filter, MoreHorizontal, RefreshCw, UserPlus } from "lucide-react";
import { useMemo, useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useArchiveUser,
  useInviteUser,
  useReactivateUser,
  useSuspendUser,
  useUsers,
} from "@/hooks/user/useUserQueries";
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { useTenant, usePermissions } from "@/providers";
import type { UserStatus } from "@/types/user";

const STATUS_FILTERS: { label: string; value: "" | UserStatus }[] = [
  { label: "All", value: "" },
  { label: "Active", value: "active" },
  { label: "Invited", value: "invited" },
  { label: "Suspended", value: "suspended" },
  { label: "Archived", value: "archived" },
];

function statusVariant(s: string) {
  switch (s) {
    case "active":
      return "success" as const;
    case "invited":
      return "warning" as const;
    case "suspended":
    case "archived":
      return "danger" as const;
    default:
      return "muted" as const;
  }
}

export default function UsersPage() {
  const [status, setStatus] = useState<"" | UserStatus>("");
  const [search, setSearch] = useState("");
  const [invite, setInvite] = useState(false);
  const { has } = usePermissions();
  const { activeOrganization } = useTenant();

  const usersQ = useUsers({
    q: search || undefined,
    status: status || undefined,
    limit: 50,
  });
  const rolesQ = useRoles();
  const suspend = useSuspendUser();
  const reactivate = useReactivateUser();
  const archive = useArchiveUser();
  const inviteM = useInviteUser();

  const counts = useMemo(() => {
    const data = usersQ.data ?? [];
    return {
      total: data.length,
      active: data.filter((u) => u.status === "active").length,
      invited: data.filter((u) => u.status === "invited").length,
      suspended: data.filter((u) => u.status === "suspended").length,
    };
  }, [usersQ.data]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
          <p className="text-muted-foreground mt-1">
            Manage who has access to {activeOrganization?.name ?? "this workspace"}.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => usersQ.refetch()} disabled={usersQ.isFetching}>
            <RefreshCw className={"h-4 w-4 " + (usersQ.isFetching ? "animate-spin" : "")} />
            Refresh
          </Button>
          {has("user.invite") && (
            <Button size="sm" onClick={() => setInvite(true)}>
              <UserPlus className="h-4 w-4" /> Invite user
            </Button>
          )}
        </div>
      </div>

      {/* Stats strip */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatBox label="Total" value={counts.total} />
        <StatBox label="Active" value={counts.active} accent="emerald" />
        <StatBox label="Invited" value={counts.invited} accent="amber" />
        <StatBox label="Suspended" value={counts.suspended} accent="rose" />
      </div>

      {/* Filters */}
      <Card className="overflow-hidden">
        <div className="flex flex-col gap-3 p-4 md:flex-row md:items-center md:justify-between border-b">
          <div className="flex flex-wrap items-center gap-1.5">
            <Filter className="h-4 w-4 text-muted-foreground mr-1" />
            {STATUS_FILTERS.map((f) => (
              <button
                key={f.value}
                onClick={() => setStatus(f.value)}
                className={
                  "px-3 h-8 rounded-md text-xs font-medium transition-colors " +
                  (status === f.value
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:bg-accent")
                }
              >
                {f.label}
              </button>
            ))}
          </div>
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search name or email…"
            className="md:max-w-xs"
          />
        </div>

        <div className="overflow-x-auto">
          {usersQ.isLoading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-14 w-full" />
              ))}
            </div>
          ) : !usersQ.data?.length ? (
            <div className="p-10 text-center text-sm text-muted-foreground">No users found.</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-muted/40 text-xs uppercase tracking-wider text-muted-foreground">
                <tr>
                  <th className="text-left p-3 font-medium">User</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">Job title</th>
                  <th className="text-left p-3 font-medium">Last login</th>
                  <th className="text-right p-3 font-medium w-12"></th>
                </tr>
              </thead>
              <tbody>
                {usersQ.data.map((u) => {
                  const name =
                    u.displayName ||
                    `${u.firstName ?? ""} ${u.lastName ?? ""}`.trim() ||
                    u.email;
                  const initials =
                    (u.firstName?.[0] ?? u.email[0] ?? "?").toUpperCase() +
                    (u.lastName?.[0] ?? "").toUpperCase();
                  return (
                    <tr key={u.id} className="border-t hover:bg-muted/30">
                      <td className="p-3">
                        <div className="flex items-center gap-3 min-w-0">
                          <Avatar className="h-8 w-8">
                            {u.avatarUrl ? <AvatarImage src={u.avatarUrl} alt={name} /> : null}
                            <AvatarFallback>{initials}</AvatarFallback>
                          </Avatar>
                          <div className="min-w-0">
                            <div className="font-medium truncate">{name}</div>
                            <div className="text-xs text-muted-foreground truncate">{u.email}</div>
                          </div>
                        </div>
                      </td>
                      <td className="p-3">
                        <Badge variant={statusVariant(u.status)}>{u.status}</Badge>
                      </td>
                      <td className="p-3 text-muted-foreground">{u.jobTitle ?? "—"}</td>
                      <td className="p-3 text-muted-foreground">
                        {u.lastLoginAt ? new Date(u.lastLoginAt).toLocaleString() : "Never"}
                      </td>
                      <td className="p-3 text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" aria-label="Actions">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            {u.status !== "suspended" && has("user.suspend") && (
                              <DropdownMenuItem onSelect={() => suspend.mutate(u.id)}>
                                Suspend
                              </DropdownMenuItem>
                            )}
                            {u.status === "suspended" && has("user.suspend") && (
                              <DropdownMenuItem onSelect={() => reactivate.mutate(u.id)}>
                                Reactivate
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            {has("user.delete") && (
                              <DropdownMenuItem
                                onSelect={() => archive.mutate(u.id)}
                                className="text-destructive focus:text-destructive"
                              >
                                Archive
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </Card>

      {/* Invite dialog */}
      <Dialog open={invite} onOpenChange={setInvite}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invite a teammate</DialogTitle>
            <DialogDescription>
              We&apos;ll email them an invite link. They&apos;ll set their own password.
            </DialogDescription>
          </DialogHeader>
          <InviteForm
            roles={(rolesQ.data ?? []).map((r) => ({ key: r.key, name: r.name }))}
            orgId={activeOrganization?.id ?? ""}
            onCancel={() => setInvite(false)}
            onSubmit={async (vals) => {
              await inviteM.mutateAsync(vals);
              setInvite(false);
            }}
            pending={inviteM.isPending}
            error={inviteM.isError ? (inviteM.error as Error).message : null}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function StatBox({
  label,
  value,
  accent,
}: {
  label: string;
  value: number;
  accent?: "emerald" | "amber" | "rose";
}) {
  const dot =
    accent === "emerald"
      ? "bg-emerald-500"
      : accent === "amber"
      ? "bg-amber-500"
      : accent === "rose"
      ? "bg-rose-500"
      : "bg-muted-foreground";
  return (
    <Card className="p-4">
      <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-muted-foreground">
        <span className={"h-2 w-2 rounded-full " + dot} />
        {label}
      </div>
      <div className="mt-2 text-2xl font-semibold">{value}</div>
    </Card>
  );
}

function InviteForm(props: {
  roles: { key: string; name: string }[];
  orgId: string;
  onCancel: () => void;
  onSubmit: (v: {
    email: string;
    firstName?: string;
    lastName?: string;
    jobTitle?: string;
    organizationId: string;
    roleKeys: string[];
    message?: string;
  }) => void;
  pending: boolean;
  error: string | null;
}) {
  const [email, setEmail] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [jobTitle, setJobTitle] = useState("");
  const [roleKey, setRoleKey] = useState<string>(props.roles.find((r) => r.key === "member")?.key ?? props.roles[0]?.key ?? "member");
  const [message, setMessage] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        props.onSubmit({
          email: email.trim(),
          firstName: firstName.trim() || undefined,
          lastName: lastName.trim() || undefined,
          jobTitle: jobTitle.trim() || undefined,
          organizationId: props.orgId,
          roleKeys: roleKey ? [roleKey] : [],
          message: message.trim() || undefined,
        });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="firstName">First name</Label>
          <Input id="firstName" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lastName">Last name</Label>
          <Input id="lastName" value={lastName} onChange={(e) => setLastName(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="email">Work email</Label>
        <Input id="email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label htmlFor="jobTitle">Job title (optional)</Label>
        <Input id="jobTitle" value={jobTitle} onChange={(e) => setJobTitle(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label htmlFor="role">Role</Label>
        <select
          id="role"
          value={roleKey}
          onChange={(e) => setRoleKey(e.target.value)}
          className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
        >
          {props.roles.map((r) => (
            <option key={r.key} value={r.key}>
              {r.name}
            </option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label htmlFor="message">Message (optional)</Label>
        <Input id="message" value={message} onChange={(e) => setMessage(e.target.value)} />
      </div>
      {props.error && <p className="text-sm text-destructive">{props.error}</p>}
      <DialogFooter>
        <Button type="button" variant="ghost" onClick={props.onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={props.pending || !email}>
          {props.pending ? "Sending…" : "Send invite"}
        </Button>
      </DialogFooter>
    </form>
  );
}
