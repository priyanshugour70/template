"use client";

import {
  ChevronDown,
  ChevronUp,
  Filter,
  KeyRound,
  Lock,
  LockOpen,
  MoreHorizontal,
  RotateCcw,
  ShieldCheck,
  UserPlus,
  X,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { PaginationBar } from "@/components/shared/pagination-bar";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Sheet, SheetContent, SheetDescription, SheetTitle } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useDepartments } from "@/hooks/department/useDepartments";
import { useAssignRoles, useRoles } from "@/hooks/rbac/useRBACQueries";
import { toast } from "@/hooks/use-toast";
import {
  useArchiveUser,
  useBulkUpdateMemberships,
  useEffectivePermissions,
  useForcePasswordReset,
  useInviteUser,
  useReactivateUser,
  useResetMFA,
  useSuspendUser,
  useUnlockUser,
  useUpdateMembership,
  useUpdateUser,
  useUserMemberships,
  useUsers,
} from "@/hooks/user/useUserQueries";
import { cn } from "@/lib/cn";
import { usePermissions, useTenant } from "@/providers";
import type {
  Membership,
  UpdateMembershipRequest,
  UserListQuery,
  UserProfile,
  UserStatus,
} from "@/types/user";

// ── helpers ────────────────────────────────────────────────────────────────

const STATUS_OPTIONS: { label: string; value: "" | UserStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Active", value: "active" },
  { label: "Invited", value: "invited" },
  { label: "Suspended", value: "suspended" },
  { label: "Archived", value: "archived" },
];

const MFA_OPTIONS = [
  { label: "MFA any", value: "" },
  { label: "MFA on", value: "true" },
  { label: "MFA off", value: "false" },
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

function displayName(u: UserProfile) {
  return (
    u.displayName ||
    `${u.firstName ?? ""} ${u.lastName ?? ""}`.trim() ||
    u.email
  );
}

function initials(u: UserProfile) {
  return (
    (u.firstName?.[0] ?? u.email[0] ?? "?").toUpperCase() +
    (u.lastName?.[0] ?? "").toUpperCase()
  );
}

type SortField = "created_at" | "last_login_at" | "email" | "status";

// ── main page ─────────────────────────────────────────────────────────────

export default function UsersPage() {
  const { has } = usePermissions();
  const { activeOrganization } = useTenant();
  const rolesQ = useRoles();
  const deptsQ = useDepartments();

  // filters
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<string>("");
  const [roleKey, setRoleKey] = useState<string>("");
  const [departmentId, setDepartmentId] = useState<string>("");
  const [mfa, setMfa] = useState<string>("");

  // sort
  const [sortField, setSortField] = useState<SortField>("created_at");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");

  // selection
  const [selected, setSelected] = useState<Set<string>>(new Set());

  // dialogs
  const [invite, setInvite] = useState(false);
  const [detailUser, setDetailUser] = useState<UserProfile | null>(null);
  const [bulkOpen, setBulkOpen] = useState<"assignRole" | "changeDept" | "suspend" | null>(null);
  // pagination
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);

  const query = useMemo<UserListQuery>(
    () => ({
      q: search || undefined,
      status: (status || undefined) as UserStatus | undefined,
      role: roleKey || undefined,
      departmentId: departmentId || undefined,
      mfa: mfa === "" ? undefined : mfa === "true",
      sort: `${sortDir === "desc" ? "-" : ""}${sortField}`,
      page,
      limit,
    }),
    [search, status, roleKey, departmentId, mfa, sortField, sortDir, page, limit],
  );

  // Reset to page 1 whenever a filter changes (otherwise you'd land on an empty
  // page when the new result-set is shorter).
  useEffect(() => {
    setPage(1);
  }, [search, status, roleKey, departmentId, mfa, sortField, sortDir, limit]);

  const usersQ = useUsers(query);
  const users = usersQ.data?.items ?? [];
  const usersTotal = usersQ.data?.total ?? 0;

  const allSelected = users.length > 0 && users.every((u) => selected.has(u.id));
  const toggleAll = () => {
    if (allSelected) setSelected(new Set());
    else setSelected(new Set(users.map((u) => u.id)));
  };
  const toggleOne = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleSort = (field: SortField) => {
    if (sortField === field) setSortDir(sortDir === "asc" ? "desc" : "asc");
    else {
      setSortField(field);
      setSortDir("desc");
    }
  };

  const clearFilters = () => {
    setSearch("");
    setStatus("");
    setRoleKey("");
    setDepartmentId("");
    setMfa("");
  };

  const activeFilters =
    [search, status, roleKey, departmentId, mfa].filter((v) => v !== "").length;

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">Users</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Everyone in this workspace. Click a row to manage roles, security, and preferences.
          </p>
        </div>
        {has("user.invite") && (
          <Button onClick={() => setInvite(true)}>
            <UserPlus className="h-4 w-4" />
            Invite teammate
          </Button>
        )}
      </div>

      {/* Filter bar */}
      <Card>
        <CardContent className="grid grid-cols-1 gap-3 p-4 sm:grid-cols-2 lg:grid-cols-6">
          <div className="lg:col-span-2">
            <Input
              placeholder="Search by name or email…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger>
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              {STATUS_OPTIONS.map((o) => (
                <SelectItem key={o.value || "_all"} value={o.value || "_all"}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={roleKey || "_any"} onValueChange={(v) => setRoleKey(v === "_any" ? "" : v)}>
            <SelectTrigger>
              <SelectValue placeholder="Any role" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_any">Any role</SelectItem>
              {(rolesQ.data?.items ?? []).map((r) => (
                <SelectItem key={r.id} value={r.key}>
                  {r.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            value={departmentId || "_any"}
            onValueChange={(v) => setDepartmentId(v === "_any" ? "" : v)}
          >
            <SelectTrigger>
              <SelectValue placeholder="Any department" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_any">Any department</SelectItem>
              {(deptsQ.data?.items ?? []).map((d) => (
                <SelectItem key={d.id} value={d.id}>
                  {d.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={mfa || "_any"} onValueChange={(v) => setMfa(v === "_any" ? "" : v)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {MFA_OPTIONS.map((o) => (
                <SelectItem key={o.value || "_any"} value={o.value || "_any"}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
        {activeFilters > 0 && (
          <CardContent className="flex items-center justify-between border-t border-border px-4 py-2 text-xs text-muted-foreground">
            <span>
              <Filter className="mr-1 inline h-3 w-3" />
              {activeFilters} filter{activeFilters > 1 ? "s" : ""} active
            </span>
            <Button variant="ghost" size="sm" onClick={clearFilters}>
              <X className="h-3 w-3" />
              Clear
            </Button>
          </CardContent>
        )}
      </Card>

      {/* Bulk action bar */}
      {selected.size > 0 && (
        <Card>
          <CardContent className="flex flex-wrap items-center gap-2 px-4 py-2">
            <span className="text-sm">
              <strong>{selected.size}</strong> selected
            </span>
            <div className="ml-auto flex flex-wrap items-center gap-2">
              {has("user.assign_role") && (
                <Button variant="outline" size="sm" onClick={() => setBulkOpen("assignRole")}>
                  Assign role
                </Button>
              )}
              {has("user.update") && (
                <Button variant="outline" size="sm" onClick={() => setBulkOpen("changeDept")}>
                  Change department
                </Button>
              )}
              {has("user.suspend") && (
                <Button variant="outline" size="sm" onClick={() => setBulkOpen("suspend")}>
                  Suspend
                </Button>
              )}
              <Button variant="ghost" size="sm" onClick={() => setSelected(new Set())}>
                <X className="h-4 w-4" />
                Clear
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {usersQ.isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }, (_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : users.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-10 text-center text-sm text-muted-foreground">
              No users match the current filters.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-10">
                    <Checkbox
                      aria-label="Select all"
                      checked={allSelected}
                      onChange={toggleAll}
                    />
                  </TableHead>
                  <SortableHead
                    field="email"
                    label="User"
                    sortField={sortField}
                    sortDir={sortDir}
                    onClick={toggleSort}
                  />
                  <SortableHead
                    field="status"
                    label="Status"
                    sortField={sortField}
                    sortDir={sortDir}
                    onClick={toggleSort}
                  />
                  <TableHead className="hidden md:table-cell">Title</TableHead>
                  <TableHead className="hidden md:table-cell">MFA</TableHead>
                  <SortableHead
                    field="last_login_at"
                    label="Last login"
                    sortField={sortField}
                    sortDir={sortDir}
                    onClick={toggleSort}
                  />
                  <TableHead className="w-10 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((u) => (
                  <UserRow
                    key={u.id}
                    user={u}
                    selected={selected.has(u.id)}
                    onToggleSelect={() => toggleOne(u.id)}
                    onClickRow={() => setDetailUser(u)}
                  />
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
        <div className="border-t border-border p-3">
          <PaginationBar
            page={page}
            limit={limit}
            total={usersTotal}
            onPageChange={setPage}
            onLimitChange={setLimit}
          />
        </div>
      </Card>

      {/* Dialogs */}
      {invite && (
        <InviteDialog
          orgId={activeOrganization?.id ?? ""}
          onClose={() => setInvite(false)}
        />
      )}
      {detailUser && (
        <UserDetailSheet
          user={detailUser}
          onClose={() => setDetailUser(null)}
        />
      )}
      {bulkOpen === "assignRole" && (
        <BulkAssignRoleDialog
          userIds={Array.from(selected)}
          onClose={() => setBulkOpen(null)}
          onDone={() => {
            setBulkOpen(null);
            setSelected(new Set());
          }}
        />
      )}
      {bulkOpen === "changeDept" && (
        <BulkChangeDeptDialog
          userIds={Array.from(selected)}
          onClose={() => setBulkOpen(null)}
          onDone={() => {
            setBulkOpen(null);
            setSelected(new Set());
          }}
        />
      )}
      {bulkOpen === "suspend" && (
        <BulkSuspendDialog
          userIds={Array.from(selected)}
          onClose={() => setBulkOpen(null)}
          onDone={() => {
            setBulkOpen(null);
            setSelected(new Set());
          }}
        />
      )}
    </div>
  );
}

// ── table primitives ──────────────────────────────────────────────────────

function SortableHead({
  field,
  label,
  sortField,
  sortDir,
  onClick,
}: {
  field: SortField;
  label: string;
  sortField: SortField;
  sortDir: "asc" | "desc";
  onClick: (f: SortField) => void;
}) {
  const active = sortField === field;
  return (
    <TableHead>
      <button
        type="button"
        onClick={() => onClick(field)}
        className={cn(
          "inline-flex items-center gap-1 text-xs font-medium transition-colors",
          active ? "text-foreground" : "text-muted-foreground hover:text-foreground",
        )}
      >
        {label}
        {active &&
          (sortDir === "asc" ? (
            <ChevronUp className="h-3 w-3" />
          ) : (
            <ChevronDown className="h-3 w-3" />
          ))}
      </button>
    </TableHead>
  );
}

function UserRow({
  user,
  selected,
  onToggleSelect,
  onClickRow,
}: {
  user: UserProfile;
  selected: boolean;
  onToggleSelect: () => void;
  onClickRow: () => void;
}) {
  const { has } = usePermissions();
  const suspend = useSuspendUser();
  const reactivate = useReactivateUser();
  const archive = useArchiveUser();
  const name = displayName(user);

  return (
    <TableRow className="cursor-pointer" onClick={onClickRow}>
      <TableCell onClick={(e) => e.stopPropagation()}>
        <Checkbox aria-label="Select row" checked={selected} onChange={onToggleSelect} />
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-3 min-w-0">
          <Avatar className="h-8 w-8">
            {user.avatarUrl ? <AvatarImage src={user.avatarUrl} alt={name} /> : null}
            <AvatarFallback>{initials(user)}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <div className="truncate font-medium">{name}</div>
            <div className="truncate text-xs text-muted-foreground">{user.email}</div>
          </div>
        </div>
      </TableCell>
      <TableCell>
        <Badge variant={statusVariant(user.status)}>{user.status}</Badge>
      </TableCell>
      <TableCell className="hidden text-muted-foreground md:table-cell">
        {user.jobTitle ?? "—"}
      </TableCell>
      <TableCell className="hidden md:table-cell">
        {user.mfaEnabled ? (
          <Badge variant="success">on</Badge>
        ) : (
          <Badge variant="muted">off</Badge>
        )}
      </TableCell>
      <TableCell className="text-muted-foreground">
        {user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleDateString() : "Never"}
      </TableCell>
      <TableCell className="text-right" onClick={(e) => e.stopPropagation()}>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Actions">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onSelect={onClickRow}>Manage…</DropdownMenuItem>
            {user.status !== "suspended" && has("user.suspend") && (
              <DropdownMenuItem
                onSelect={() =>
                  suspend.mutate(user.id, {
                    onSuccess: () => toast.success("User suspended"),
                    onError: (e: unknown) =>
                      toast.error("Suspend failed", e instanceof Error ? e.message : undefined),
                  })
                }
              >
                Suspend
              </DropdownMenuItem>
            )}
            {user.status === "suspended" && has("user.suspend") && (
              <DropdownMenuItem
                onSelect={() =>
                  reactivate.mutate(user.id, {
                    onSuccess: () => toast.success("User reactivated"),
                  })
                }
              >
                Reactivate
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            {has("user.delete") && (
              <DropdownMenuItem
                onSelect={() =>
                  archive.mutate(user.id, {
                    onSuccess: () => toast.success("User archived"),
                  })
                }
                className="text-destructive focus:text-destructive"
              >
                Archive
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
}

// ── detail sheet ──────────────────────────────────────────────────────────

function UserDetailSheet({
  user,
  onClose,
}: {
  user: UserProfile;
  onClose: () => void;
}) {
  return (
    <Sheet open onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-2xl">
        <div className="border-b border-border p-6">
          <div className="flex items-center gap-3">
            <Avatar className="h-12 w-12">
              {user.avatarUrl ? <AvatarImage src={user.avatarUrl} alt={displayName(user)} /> : null}
              <AvatarFallback className="text-base">{initials(user)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <SheetTitle className="truncate text-lg">{displayName(user)}</SheetTitle>
              <SheetDescription className="truncate text-sm text-muted-foreground">
                {user.email}
              </SheetDescription>
              <div className="mt-1 flex items-center gap-2">
                <Badge variant={statusVariant(user.status)}>{user.status}</Badge>
                {user.mfaEnabled && <Badge variant="success">MFA on</Badge>}
                {user.mustChangePassword && <Badge variant="warning">password reset pending</Badge>}
                {user.lockedUntil && new Date(user.lockedUntil) > new Date() && (
                  <Badge variant="danger">locked</Badge>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="flex-1 p-6">
          <Tabs defaultValue="profile" className="w-full">
            <TabsList>
              <TabsTrigger value="profile">Profile</TabsTrigger>
              <TabsTrigger value="security">Security</TabsTrigger>
              <TabsTrigger value="membership">Membership</TabsTrigger>
              <TabsTrigger value="preferences">Preferences</TabsTrigger>
            </TabsList>

            <TabsContent value="profile">
              <ProfileTab user={user} />
            </TabsContent>
            <TabsContent value="security">
              <SecurityTab user={user} />
            </TabsContent>
            <TabsContent value="membership">
              <MembershipTab user={user} />
            </TabsContent>
            <TabsContent value="preferences">
              <PreferencesTab user={user} />
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function ProfileTab({ user }: { user: UserProfile }) {
  const { has } = usePermissions();
  const canEdit = has("user.update");
  const update = useUpdateUser(user.id);

  const [firstName, setFirstName] = useState(user.firstName ?? "");
  const [lastName, setLastName] = useState(user.lastName ?? "");
  const [displayNameValue, setDisplayNameValue] = useState(user.displayName ?? "");
  const [jobTitle, setJobTitle] = useState(user.jobTitle ?? "");
  const [phone, setPhone] = useState(user.phone ?? "");
  const [bio, setBio] = useState(user.bio ?? "");

  const dirty =
    firstName !== (user.firstName ?? "") ||
    lastName !== (user.lastName ?? "") ||
    displayNameValue !== (user.displayName ?? "") ||
    jobTitle !== (user.jobTitle ?? "") ||
    phone !== (user.phone ?? "") ||
    bio !== (user.bio ?? "");

  return (
    <div className="space-y-4">
      <Field label="First name">
        <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} disabled={!canEdit} />
      </Field>
      <Field label="Last name">
        <Input value={lastName} onChange={(e) => setLastName(e.target.value)} disabled={!canEdit} />
      </Field>
      <Field label="Display name">
        <Input
          value={displayNameValue}
          onChange={(e) => setDisplayNameValue(e.target.value)}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Job title">
        <Input value={jobTitle} onChange={(e) => setJobTitle(e.target.value)} disabled={!canEdit} />
      </Field>
      <Field label="Phone">
        <Input value={phone} onChange={(e) => setPhone(e.target.value)} disabled={!canEdit} />
      </Field>
      <Field label="Bio">
        <Input value={bio} onChange={(e) => setBio(e.target.value)} disabled={!canEdit} />
      </Field>

      {canEdit && (
        <div className="flex justify-end gap-2 pt-2">
          <Button
            disabled={!dirty || update.isPending}
            onClick={() =>
              update.mutate(
                {
                  firstName,
                  lastName,
                  displayName: displayNameValue,
                  jobTitle,
                  phone,
                  bio,
                },
                {
                  onSuccess: () => toast.success("Profile updated"),
                  onError: (e: unknown) =>
                    toast.error("Update failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save changes
          </Button>
        </div>
      )}
    </div>
  );
}

function SecurityTab({ user }: { user: UserProfile }) {
  const { has } = usePermissions();
  const canEdit = has("user.update");
  const forceReset = useForcePasswordReset();
  const resetMFA = useResetMFA();
  const unlock = useUnlockUser();
  const suspend = useSuspendUser();
  const reactivate = useReactivateUser();

  if (!canEdit) {
    return <p className="text-sm text-muted-foreground">You don&apos;t have permission to manage this user&apos;s security.</p>;
  }

  const isLocked = user.lockedUntil && new Date(user.lockedUntil) > new Date();

  return (
    <div className="space-y-3">
      <SecurityRow
        icon={KeyRound}
        title="Force password reset"
        subtitle="User will be prompted to change their password on next login."
        action={
          <Button
            variant="outline"
            size="sm"
            disabled={forceReset.isPending}
            onClick={() =>
              forceReset.mutate(user.id, {
                onSuccess: () => toast.success("Password reset requested"),
                onError: (e: unknown) =>
                  toast.error("Failed", e instanceof Error ? e.message : undefined),
              })
            }
          >
            Force reset
          </Button>
        }
      />
      <SecurityRow
        icon={ShieldCheck}
        title="Reset multi-factor authentication"
        subtitle={user.mfaEnabled ? "Clears the MFA secret and recovery codes." : "MFA is currently disabled."}
        action={
          <Button
            variant="outline"
            size="sm"
            disabled={resetMFA.isPending || !user.mfaEnabled}
            onClick={() =>
              resetMFA.mutate(user.id, {
                onSuccess: () => toast.success("MFA reset"),
                onError: (e: unknown) =>
                  toast.error("Failed", e instanceof Error ? e.message : undefined),
              })
            }
          >
            Reset MFA
          </Button>
        }
      />
      <SecurityRow
        icon={isLocked ? Lock : LockOpen}
        title={isLocked ? "Unlock account" : "Account not locked"}
        subtitle={
          isLocked
            ? `Locked until ${new Date(user.lockedUntil!).toLocaleString()}`
            : `${user.failedLoginCount} failed login(s) tracked.`
        }
        action={
          <Button
            variant="outline"
            size="sm"
            disabled={unlock.isPending || (!isLocked && user.failedLoginCount === 0)}
            onClick={() =>
              unlock.mutate(user.id, {
                onSuccess: () => toast.success("Account unlocked"),
                onError: (e: unknown) =>
                  toast.error("Failed", e instanceof Error ? e.message : undefined),
              })
            }
          >
            <RotateCcw className="h-4 w-4" />
            Unlock
          </Button>
        }
      />
      {has("user.suspend") && (
        <SecurityRow
          icon={user.status === "suspended" ? RotateCcw : Lock}
          title={user.status === "suspended" ? "Reactivate user" : "Suspend user"}
          subtitle={
            user.status === "suspended"
              ? "Allow this user to sign in again."
              : "Block sign-ins. User keeps memberships."
          }
          action={
            user.status === "suspended" ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  reactivate.mutate(user.id, {
                    onSuccess: () => toast.success("User reactivated"),
                  })
                }
              >
                Reactivate
              </Button>
            ) : (
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  suspend.mutate(user.id, {
                    onSuccess: () => toast.success("User suspended"),
                  })
                }
              >
                Suspend
              </Button>
            )
          }
        />
      )}
    </div>
  );
}

function SecurityRow({
  icon: Icon,
  title,
  subtitle,
  action,
}: {
  icon: typeof KeyRound;
  title: string;
  subtitle: string;
  action: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-border p-3">
      <Icon className="mt-0.5 h-5 w-5 text-muted-foreground" />
      <div className="flex-1">
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-muted-foreground">{subtitle}</p>
      </div>
      {action}
    </div>
  );
}

function MembershipTab({ user }: { user: UserProfile }) {
  const { activeOrganization } = useTenant();
  const membershipsQ = useUserMemberships(user.id);
  const deptsQ = useDepartments();
  const rolesQ = useRoles();
  const permsQ = useEffectivePermissions(user.id);
  const updateMembership = useUpdateMembership(user.id);

  // Find the membership for the active org — that's what we let admins edit here.
  const activeMembership = useMemo<Membership | undefined>(
    () =>
      (membershipsQ.data ?? []).find(
        (m) => m.organizationId === activeOrganization?.id,
      ),
    [membershipsQ.data, activeOrganization?.id],
  );

  return (
    <div className="space-y-5">
      {membershipsQ.isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : (
        <>
          {activeMembership ? (
            <div className="space-y-4 rounded-md border border-border p-4">
              <div className="text-sm font-medium">In this organization</div>
              <Field label="Department">
                <Select
                  value={activeMembership.departmentId ?? "_none"}
                  onValueChange={(v) =>
                    updateMembership.mutate(
                      {
                        membershipId: activeMembership.id,
                        patch: {
                          departmentId:
                            v === "_none"
                              ? "00000000-0000-0000-0000-000000000000"
                              : v,
                        } as UpdateMembershipRequest,
                      },
                      {
                        onSuccess: () => toast.success("Department updated"),
                        onError: (e: unknown) =>
                          toast.error("Failed", e instanceof Error ? e.message : undefined),
                      },
                    )
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="No department" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">No department</SelectItem>
                    {(deptsQ.data?.items ?? []).map((d) => (
                      <SelectItem key={d.id} value={d.id}>
                        {d.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>
              <RolesEditor
                membership={activeMembership}
                allRoles={rolesQ.data?.items ?? []}
              />
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No membership in the active organization.
            </p>
          )}

          {/* Effective permissions */}
          <div className="rounded-md border border-border p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium">Effective permissions</p>
              <Badge variant="muted">{(permsQ.data ?? []).length}</Badge>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Union of direct, department, and group role grants.
            </p>
            {permsQ.isLoading ? (
              <Skeleton className="mt-3 h-16 w-full" />
            ) : (
              <div className="mt-3 flex max-h-48 flex-wrap gap-1.5 overflow-y-auto">
                {(permsQ.data ?? []).map((p) => (
                  <code
                    key={p}
                    className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]"
                  >
                    {p}
                  </code>
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}

function RolesEditor({
  membership,
  allRoles,
}: {
  membership: Membership;
  allRoles: { id: string; key: string; name: string }[];
}) {
  // We don't have a hook for listMembershipRoles in scope here; rely on the
  // assign mutation. Read current roles from a separate fetch when needed.
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set());
  const assign = useAssignRoles(membership.id);

  return (
    <Field label="Roles in this org">
      <div className="space-y-1.5">
        {allRoles.map((r) => {
          const checked = selectedKeys.has(r.key);
          return (
            <label
              key={r.id}
              className="flex items-center gap-2 rounded-md px-2 py-1 hover:bg-muted/50"
            >
              <Checkbox
                checked={checked}
                onChange={(e) => {
                  setSelectedKeys((prev) => {
                    const next = new Set(prev);
                    if (e.target.checked) next.add(r.key);
                    else next.delete(r.key);
                    return next;
                  });
                }}
              />
              <span className="text-sm">{r.name}</span>
            </label>
          );
        })}
        <Button
          size="sm"
          variant="outline"
          disabled={selectedKeys.size === 0 || assign.isPending}
          onClick={() =>
            assign.mutate(
              { roleKeys: Array.from(selectedKeys) },
              {
                onSuccess: () => toast.success("Roles updated"),
                onError: (e: unknown) =>
                  toast.error("Failed", e instanceof Error ? e.message : undefined),
              },
            )
          }
        >
          Apply role changes
        </Button>
        <p className="text-[11px] text-muted-foreground">
          Replaces this user&apos;s direct role assignment in the current organization.
          Department- and group-based grants remain in effect.
        </p>
      </div>
    </Field>
  );
}

function PreferencesTab({ user }: { user: UserProfile }) {
  const { has } = usePermissions();
  const canEdit = has("user.update");
  const update = useUpdateUser(user.id);

  const [locale, setLocale] = useState(user.locale);
  const [timezone, setTimezone] = useState(user.timezone);
  const [marketingOptIn, setMarketingOptIn] = useState(user.marketingOptIn ?? false);

  const dirty =
    locale !== user.locale ||
    timezone !== user.timezone ||
    marketingOptIn !== (user.marketingOptIn ?? false);

  return (
    <div className="space-y-4">
      <Field label="Locale">
        <Input value={locale} onChange={(e) => setLocale(e.target.value)} disabled={!canEdit} />
      </Field>
      <Field label="Timezone">
        <Input value={timezone} onChange={(e) => setTimezone(e.target.value)} disabled={!canEdit} />
      </Field>
      <div className="flex items-start justify-between gap-3 rounded-md border border-border p-3">
        <div className="min-w-0">
          <p className="text-sm font-medium">Marketing opt-in</p>
          <p className="text-xs text-muted-foreground">Receive product news and tips.</p>
        </div>
        <Switch
          checked={marketingOptIn}
          onCheckedChange={setMarketingOptIn}
          disabled={!canEdit}
        />
      </div>

      {canEdit && (
        <div className="flex justify-end pt-2">
          <Button
            disabled={!dirty || update.isPending}
            onClick={() =>
              update.mutate(
                { locale, timezone },
                {
                  onSuccess: () => toast.success("Preferences saved"),
                },
              )
            }
          >
            Save changes
          </Button>
        </div>
      )}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs uppercase tracking-wider text-muted-foreground">{label}</Label>
      {children}
    </div>
  );
}

// ── bulk action dialogs ───────────────────────────────────────────────────

function BulkAssignRoleDialog({
  userIds,
  onClose,
  onDone,
}: {
  userIds: string[];
  onClose: () => void;
  onDone: () => void;
}) {
  // Backend doesn't expose a bulk role-assign by userId — would need a per-user
  // membership lookup. Keep this UI but note the limitation.
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Assign role to {userIds.length} users</DialogTitle>
          <DialogDescription>
            Bulk role-assign by user is not yet wired — assign roles individually from each user&apos;s detail drawer.
            (Tracked for a future patch on the bulk endpoint.)
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Got it
          </Button>
          <Button onClick={onDone} disabled>
            Apply
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function BulkChangeDeptDialog({
  userIds,
  onClose,
  onDone,
}: {
  userIds: string[];
  onClose: () => void;
  onDone: () => void;
}) {
  // We need to resolve userIds → membershipIds for the active org. Use the
  // user list query (it's already cached) via useUsers — but we don't have
  // the membership IDs in that payload. So we fetch memberships per user
  // and pick the one for the active org. Cheap n*1 calls; fine for bulk size.
  const { activeOrganization } = useTenant();
  const deptsQ = useDepartments();
  const [departmentId, setDepartmentId] = useState<string>("_none");
  const bulk = useBulkUpdateMemberships();
  const [resolving, setResolving] = useState(false);
  const [resolvedIds, setResolvedIds] = useState<string[]>([]);

  const resolveMemberships = useCallback(async () => {
    if (!activeOrganization?.id) return [];
    setResolving(true);
    try {
      const out: string[] = [];
      for (const uid of userIds) {
        const res = await fetch(`/api/v1/users/${uid}/memberships`, {
          credentials: "include",
        }).then((r) => r.json());
        if (res?.success && Array.isArray(res.data)) {
          const m = res.data.find(
            (x: Membership) => x.organizationId === activeOrganization.id,
          );
          if (m) out.push(m.id);
        }
      }
      setResolvedIds(out);
      return out;
    } finally {
      setResolving(false);
    }
  }, [userIds, activeOrganization?.id]);

  const apply = async () => {
    const ids = resolvedIds.length ? resolvedIds : await resolveMemberships();
    if (ids.length === 0) {
      toast.error("No memberships resolved for the selected users");
      return;
    }
    bulk.mutate(
      {
        membershipIds: ids,
        patch: {
          departmentId:
            departmentId === "_none"
              ? "00000000-0000-0000-0000-000000000000"
              : departmentId,
        },
      },
      {
        onSuccess: (data) => {
          toast.success(
            `Updated ${data.updated} membership${data.updated === 1 ? "" : "s"}`,
            data.failed?.length ? `${data.failed.length} failed` : undefined,
          );
          onDone();
        },
        onError: (e: unknown) =>
          toast.error("Bulk update failed", e instanceof Error ? e.message : undefined),
      },
    );
  };

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Change department for {userIds.length} users</DialogTitle>
          <DialogDescription>
            Applies to each user&apos;s membership in the active organization.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-2 py-2">
          <Label>Department</Label>
          <Select value={departmentId} onValueChange={setDepartmentId}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_none">No department</SelectItem>
              {(deptsQ.data?.items ?? []).map((d) => (
                <SelectItem key={d.id} value={d.id}>
                  {d.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={apply} disabled={resolving || bulk.isPending}>
            {resolving || bulk.isPending ? "Applying…" : "Apply"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function BulkSuspendDialog({
  userIds,
  onClose,
  onDone,
}: {
  userIds: string[];
  onClose: () => void;
  onDone: () => void;
}) {
  const suspend = useSuspendUser();

  const apply = async () => {
    let ok = 0;
    let failed = 0;
    for (const id of userIds) {
      try {
        await suspend.mutateAsync(id);
        ok++;
      } catch {
        failed++;
      }
    }
    toast.success(`${ok} suspended`, failed > 0 ? `${failed} failed` : undefined);
    onDone();
  };

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Suspend {userIds.length} users</DialogTitle>
          <DialogDescription>
            Suspended users can&apos;t sign in but keep their memberships. You can reactivate them later.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={apply} disabled={suspend.isPending}>
            Suspend
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ── invite dialog ─────────────────────────────────────────────────────────

function InviteDialog({ orgId, onClose }: { orgId: string; onClose: () => void }) {
  const rolesQ = useRoles();
  const deptsQ = useDepartments();
  const invite = useInviteUser();

  const [email, setEmail] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [jobTitle, setJobTitle] = useState("");
  const [roleKey, setRoleKey] = useState<string>("member");
  const [departmentId, setDepartmentId] = useState<string>("_none");
  const [message, setMessage] = useState("");

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite a teammate</DialogTitle>
          <DialogDescription>
            They&apos;ll receive an email link to set their password.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            invite.mutate(
              {
                email: email.trim(),
                firstName: firstName.trim() || undefined,
                lastName: lastName.trim() || undefined,
                jobTitle: jobTitle.trim() || undefined,
                departmentId: departmentId === "_none" ? undefined : departmentId,
                organizationId: orgId,
                roleKeys: [roleKey],
                message: message.trim() || undefined,
              },
              {
                onSuccess: () => {
                  toast.success("Invite sent", `Emailed ${email}`);
                  onClose();
                },
                onError: (e: unknown) =>
                  toast.error("Invite failed", e instanceof Error ? e.message : undefined),
              },
            );
          }}
          className="space-y-3 py-2"
        >
          <div className="grid grid-cols-2 gap-3">
            <Field label="First name">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
            <Field label="Last name">
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </Field>
          </div>
          <Field label="Work email">
            <Input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </Field>
          <Field label="Job title">
            <Input value={jobTitle} onChange={(e) => setJobTitle(e.target.value)} />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Role">
              <Select value={roleKey} onValueChange={setRoleKey}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {(rolesQ.data?.items ?? []).map((r) => (
                    <SelectItem key={r.key} value={r.key}>
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field label="Department">
              <Select value={departmentId} onValueChange={setDepartmentId}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none">None</SelectItem>
                  {(deptsQ.data?.items ?? []).map((d) => (
                    <SelectItem key={d.id} value={d.id}>
                      {d.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </div>
          <Field label="Message (optional)">
            <Input value={message} onChange={(e) => setMessage(e.target.value)} />
          </Field>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={!email || invite.isPending}>
              {invite.isPending ? "Sending…" : "Send invite"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
