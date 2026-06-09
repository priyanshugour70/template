"use client";

import { Pencil, Plus, ShieldCheck, Trash2, UsersRound } from "lucide-react";
import { useMemo, useState } from "react";

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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  useAddGroupMember,
  useAssignGroupRoles,
  useCreateGroup,
  useDeleteGroup,
  useGroupMembers,
  useGroupRoles,
  useGroups,
  useRemoveGroupMember,
  useUpdateGroup,
} from "@/hooks/group/useGroups";
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { useUsers } from "@/hooks/user/useUserQueries";
import { toast } from "@/hooks/use-toast";
import { PaginationBar } from "@/components/shared/pagination-bar";
import { usePermissions } from "@/providers";
import type { Group } from "@/types/group";

export default function GroupsPage() {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);
  const groupsQ = useGroups({ page, limit });
  const { has } = usePermissions();
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Group | null>(null);
  const [membersFor, setMembersFor] = useState<Group | null>(null);
  const [rolesFor, setRolesFor] = useState<Group | null>(null);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">Groups</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Cross-cutting collections of users. Unlike departments, groups don&apos;t inherit — every
            user in the group gets the group&apos;s role grants directly.
          </p>
        </div>
        {has("group.create") && (
          <Button onClick={() => setCreating(true)}>
            <Plus className="h-4 w-4" />
            New group
          </Button>
        )}
      </div>

      {groupsQ.isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 4 }, (_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : (groupsQ.data?.total ?? 0) === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 py-10 text-center">
            <UsersRound className="h-8 w-8 text-muted-foreground" />
            <div>
              <p className="font-medium">No groups yet</p>
              <p className="text-sm text-muted-foreground">
                Groups are great for ad-hoc role bundles like &quot;On-call&quot; or &quot;All
                managers&quot;.
              </p>
            </div>
            {has("group.create") && (
              <Button variant="outline" onClick={() => setCreating(true)}>
                <Plus className="h-4 w-4" />
                New group
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead className="hidden md:table-cell">Slug</TableHead>
                  <TableHead className="hidden md:table-cell">Kind</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(groupsQ.data?.items ?? []).map((g) => (
                  <GroupRow
                    key={g.id}
                    group={g}
                    onEdit={() => setEditing(g)}
                    onMembers={() => setMembersFor(g)}
                    onRoles={() => setRolesFor(g)}
                  />
                ))}
              </TableBody>
            </Table>
          </CardContent>
          <div className="border-t border-border p-3">
            <PaginationBar
              page={page}
              limit={limit}
              total={groupsQ.data?.total ?? 0}
              onPageChange={setPage}
              onLimitChange={(n) => {
                setLimit(n);
                setPage(1);
              }}
            />
          </div>
        </Card>
      )}

      {creating && <CreateDialog onClose={() => setCreating(false)} />}
      {editing && <EditDialog group={editing} onClose={() => setEditing(null)} />}
      {membersFor && <MembersDialog group={membersFor} onClose={() => setMembersFor(null)} />}
      {rolesFor && <RolesDialog group={rolesFor} onClose={() => setRolesFor(null)} />}
    </div>
  );
}

function GroupRow({
  group,
  onEdit,
  onMembers,
  onRoles,
}: {
  group: Group;
  onEdit: () => void;
  onMembers: () => void;
  onRoles: () => void;
}) {
  const { has } = usePermissions();
  const del = useDeleteGroup();
  return (
    <TableRow>
      <TableCell>
        <div className="flex items-center gap-2">
          <UsersRound className="h-4 w-4 text-muted-foreground" />
          <div>
            <p className="font-medium">{group.name}</p>
            {group.description && (
              <p className="text-xs text-muted-foreground">{group.description}</p>
            )}
          </div>
        </div>
      </TableCell>
      <TableCell className="hidden md:table-cell">
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{group.slug}</code>
      </TableCell>
      <TableCell className="hidden md:table-cell">
        <Badge variant="muted">{group.kind}</Badge>
        {group.isArchived && (
          <Badge variant="muted" className="ml-1">
            archived
          </Badge>
        )}
      </TableCell>
      <TableCell className="text-right">
        <div className="inline-flex items-center gap-1">
          {has("group.read") && (
            <Button variant="ghost" size="sm" onClick={onMembers}>
              Members
            </Button>
          )}
          {has("group.assign") && (
            <Button variant="ghost" size="sm" onClick={onRoles}>
              <ShieldCheck className="h-4 w-4" />
              Roles
            </Button>
          )}
          {has("group.update") && (
            <Button variant="ghost" size="icon" onClick={onEdit}>
              <Pencil className="h-4 w-4" />
            </Button>
          )}
          {has("group.delete") && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                if (confirm(`Archive group "${group.name}"?`)) {
                  del.mutate(group.id, {
                    onSuccess: () => toast.success("Group archived"),
                    onError: (e: unknown) =>
                      toast.error("Archive failed", e instanceof Error ? e.message : undefined),
                  });
                }
              }}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

function CreateDialog({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const create = useCreateGroup();
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New group</DialogTitle>
          <DialogDescription>
            Create a flat collection of users — handy for cross-team role bundles.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 py-2">
          <div className="grid gap-1.5">
            <Label>Name</Label>
            <Input
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (!slug)
                  setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, ""));
              }}
              placeholder="On-call rotation"
            />
          </div>
          <div className="grid gap-1.5">
            <Label>Slug</Label>
            <Input value={slug} onChange={(e) => setSlug(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>Description</Label>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!name || !slug || create.isPending}
            onClick={() =>
              create.mutate(
                { name, slug, description: description || undefined },
                {
                  onSuccess: () => {
                    toast.success("Group created");
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error("Create failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function EditDialog({ group, onClose }: { group: Group; onClose: () => void }) {
  const [name, setName] = useState(group.name);
  const [description, setDescription] = useState(group.description ?? "");
  const update = useUpdateGroup(group.id);
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit group</DialogTitle>
          <DialogDescription>{group.slug}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 py-2">
          <div className="grid gap-1.5">
            <Label>Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>Description</Label>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                { name, description: description || undefined },
                {
                  onSuccess: () => {
                    toast.success("Group updated");
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error("Update failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function MembersDialog({ group, onClose }: { group: Group; onClose: () => void }) {
  const membersQ = useGroupMembers(group.id);
  const usersQ = useUsers();
  const groupsQ = useGroups();
  const add = useAddGroupMember(group.id);
  const remove = useRemoveGroupMember(group.id);
  const [pickKind, setPickKind] = useState<"user" | "group">("user");
  const [pickId, setPickId] = useState<string>("");

  const userById = useMemo(() => {
    const m = new Map<string, string>();
    for (const u of usersQ.data?.items ?? []) m.set(u.id, u.displayName ?? u.email);
    return m;
  }, [usersQ.data]);
  const groupById = useMemo(() => {
    const m = new Map<string, string>();
    for (const g of groupsQ.data?.items ?? []) m.set(g.id, g.name);
    return m;
  }, [groupsQ.data]);

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{group.name} — members</DialogTitle>
          <DialogDescription>
            Users get role grants from every group they&apos;re in, including transitively via
            nested groups.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-[1fr_2fr_auto] items-end gap-2 py-2">
          <div className="grid gap-1.5">
            <Label>Kind</Label>
            <Select value={pickKind} onValueChange={(v) => setPickKind(v as "user" | "group")}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="group">Group</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-1.5">
            <Label>{pickKind === "user" ? "User" : "Group"}</Label>
            <Select value={pickId} onValueChange={setPickId}>
              <SelectTrigger>
                <SelectValue placeholder={`Select a ${pickKind}…`} />
              </SelectTrigger>
              <SelectContent>
                {pickKind === "user"
                  ? (usersQ.data?.items ?? []).map((u) => (
                      <SelectItem key={u.id} value={u.id}>
                        {u.displayName ?? u.email}
                      </SelectItem>
                    ))
                  : (groupsQ.data?.items ?? [])
                      .filter((g) => g.id !== group.id)
                      .map((g) => (
                        <SelectItem key={g.id} value={g.id}>
                          {g.name}
                        </SelectItem>
                      ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            disabled={!pickId || add.isPending}
            onClick={() =>
              add.mutate(
                pickKind === "user" ? { userId: pickId } : { groupId: pickId },
                {
                  onSuccess: () => {
                    toast.success("Member added");
                    setPickId("");
                  },
                  onError: (e: unknown) =>
                    toast.error("Add failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Add
          </Button>
        </div>

        <div className="max-h-72 space-y-1 overflow-y-auto py-2">
          {membersQ.isLoading ? (
            <Skeleton className="h-32" />
          ) : (membersQ.data?.length ?? 0) === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">No members yet</p>
          ) : (
            (membersQ.data ?? []).map((m) => {
              const isUser = !!m.memberUserId;
              const label = isUser
                ? userById.get(m.memberUserId!) ?? m.memberUserId
                : groupById.get(m.memberGroupId!) ?? m.memberGroupId;
              return (
                <div key={m.id} className="flex items-center justify-between rounded-md border border-border px-3 py-2">
                  <div className="flex items-center gap-2">
                    <Badge variant="muted">{isUser ? "user" : "group"}</Badge>
                    <span className="text-sm">{label}</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() =>
                      remove.mutate(m.id, {
                        onError: (e: unknown) =>
                          toast.error("Remove failed", e instanceof Error ? e.message : undefined),
                      })
                    }
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              );
            })
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RolesDialog({ group, onClose }: { group: Group; onClose: () => void }) {
  const rolesQ = useRoles();
  const currentQ = useGroupRoles(group.id);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const assign = useAssignGroupRoles(group.id);

  if (currentQ.data && selected.size === 0 && currentQ.data.length > 0) {
    setSelected(new Set(currentQ.data));
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Group roles</DialogTitle>
          <DialogDescription>
            Roles attached to <span className="font-medium">{group.name}</span> apply to every user
            in this group.
          </DialogDescription>
        </DialogHeader>
        <div className="max-h-80 space-y-1.5 overflow-y-auto py-2">
          {rolesQ.isLoading || currentQ.isLoading ? (
            <Skeleton className="h-32 w-full" />
          ) : (
            (rolesQ.data?.items ?? []).map((role) => {
              const checked = selected.has(role.id);
              return (
                <label
                  key={role.id}
                  className="flex items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted/50"
                >
                  <Checkbox
                    checked={checked}
                    onChange={(e) => {
                      const v = e.target.checked;
                      setSelected((prev) => {
                        const next = new Set(prev);
                        if (v) next.add(role.id);
                        else next.delete(role.id);
                        return next;
                      });
                    }}
                  />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium">{role.name}</p>
                    {role.description && (
                      <p className="truncate text-xs text-muted-foreground">{role.description}</p>
                    )}
                  </div>
                </label>
              );
            })
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={assign.isPending}
            onClick={() =>
              assign.mutate(
                { roleIds: Array.from(selected) },
                {
                  onSuccess: () => {
                    toast.success("Roles updated");
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error("Update failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
