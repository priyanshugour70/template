"use client";

import { Crown, MoreHorizontal, Plus, Shield } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Skeleton } from "@/components/ui/skeleton";
import {
  useArchiveRole,
  useCreateRole,
  usePermissionsCatalog,
  useRolePermissions,
  useRoles,
  useUpdateRole,
} from "@/hooks/rbac/useRBACQueries";
import { usePermissions } from "@/providers";
import type { Permission, Role } from "@/types/rbac";

export default function RolesPage() {
  const rolesQ = useRoles();
  const permsQ = usePermissionsCatalog();
  const { has } = usePermissions();
  const [editing, setEditing] = useState<Role | null>(null);
  const [creating, setCreating] = useState(false);
  const archive = useArchiveRole();

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Roles & Permissions</h1>
          <p className="text-muted-foreground mt-1">
            Roles bundle permissions. Assign roles to members from the user&apos;s detail view.
          </p>
        </div>
        {has("role.create") && (
          <Button onClick={() => setCreating(true)}>
            <Plus className="h-4 w-4" />
            New role
          </Button>
        )}
      </div>

      {rolesQ.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-44" />
          ))}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {(rolesQ.data ?? []).map((r) => (
            <RoleCard
              key={r.id}
              role={r}
              onEdit={() => setEditing(r)}
              onArchive={() => archive.mutate(r.id)}
              canEdit={has("role.update")}
              canDelete={has("role.delete")}
            />
          ))}
        </div>
      )}

      <Dialog open={creating} onOpenChange={setCreating}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create role</DialogTitle>
            <DialogDescription>
              Group permissions into a reusable role that you can assign to members.
            </DialogDescription>
          </DialogHeader>
          <RoleForm
            mode="create"
            permissions={permsQ.data ?? []}
            onCancel={() => setCreating(false)}
            onSaved={() => setCreating(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={!!editing} onOpenChange={(o) => !o && setEditing(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit {editing?.name ?? "role"}</DialogTitle>
            <DialogDescription>
              Update the role&apos;s name, description, and permissions.
            </DialogDescription>
          </DialogHeader>
          {editing && (
            <RoleForm
              mode="edit"
              role={editing}
              permissions={permsQ.data ?? []}
              onCancel={() => setEditing(null)}
              onSaved={() => setEditing(null)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function RoleCard({
  role,
  onEdit,
  onArchive,
  canEdit,
  canDelete,
}: {
  role: Role;
  onEdit: () => void;
  onArchive: () => void;
  canEdit: boolean;
  canDelete: boolean;
}) {
  const permsQ = useRolePermissions(role.id);
  const count = permsQ.data?.length ?? 0;
  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            {role.key === "owner" ? (
              <Crown className="h-4 w-4 text-amber-500" />
            ) : (
              <Shield className="h-4 w-4 text-primary" />
            )}
            <CardTitle className="text-base">{role.name}</CardTitle>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" aria-label="Role actions">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {canEdit && <DropdownMenuItem onSelect={onEdit}>Edit role</DropdownMenuItem>}
              {!role.isSystem && canDelete && <DropdownMenuSeparator />}
              {!role.isSystem && canDelete && (
                <DropdownMenuItem
                  onSelect={onArchive}
                  className="text-destructive focus:text-destructive"
                >
                  Archive role
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="flex items-center gap-1.5 mt-1">
          {role.isSystem && <Badge variant="muted">system</Badge>}
          {role.isDefault && <Badge variant="default">default</Badge>}
        </div>
        {role.description && (
          <p className="text-sm text-muted-foreground mt-2">{role.description}</p>
        )}
      </CardHeader>
      <CardContent className="space-y-1.5 text-sm">
        <div className="flex items-center justify-between text-muted-foreground">
          <span>Permissions</span>
          <span className="font-medium text-foreground">{count}</span>
        </div>
        <div className="flex items-center justify-between text-muted-foreground">
          <span>Priority</span>
          <span className="font-medium text-foreground">{role.priority}</span>
        </div>
      </CardContent>
    </Card>
  );
}

function RoleForm({
  mode,
  role,
  permissions,
  onCancel,
  onSaved,
}: {
  mode: "create" | "edit";
  role?: Role;
  permissions: Permission[];
  onCancel: () => void;
  onSaved: () => void;
}) {
  const [name, setName] = useState(role?.name ?? "");
  const [key, setKey] = useState(role?.key ?? "");
  const [description, setDescription] = useState(role?.description ?? "");
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const existingPerms = useRolePermissions(role?.id);
  useEffect(() => {
    if (mode === "edit" && existingPerms.data) {
      setSelected(new Set(existingPerms.data.map((p) => p.key)));
    }
  }, [mode, existingPerms.data]);

  const create = useCreateRole();
  const update = useUpdateRole(role?.id ?? "");

  const grouped = useMemo(() => {
    const map = new Map<string, Permission[]>();
    for (const p of permissions) {
      const cat = p.category ?? p.resource ?? "other";
      const list = map.get(cat) ?? [];
      list.push(p);
      map.set(cat, list);
    }
    return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b));
  }, [permissions]);

  function toggle(k: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(k)) next.delete(k);
      else next.add(k);
      return next;
    });
  }
  function toggleAll(perms: Permission[]) {
    setSelected((prev) => {
      const next = new Set(prev);
      const allOn = perms.every((p) => next.has(p.key));
      if (allOn) perms.forEach((p) => next.delete(p.key));
      else perms.forEach((p) => next.add(p.key));
      return next;
    });
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const keys = Array.from(selected);
    try {
      if (mode === "create") {
        await create.mutateAsync({
          key: key.trim().toLowerCase(),
          name: name.trim(),
          description: description.trim() || undefined,
          permissionKeys: keys,
        });
      } else if (role) {
        await update.mutateAsync({
          name: name.trim(),
          description: description.trim() || undefined,
          permissionKeys: keys,
        });
      }
      onSaved();
    } catch {
      /* surfaced below */
    }
  }

  const pending = create.isPending || update.isPending;
  const errMsg =
    create.isError ? (create.error as Error).message : update.isError ? (update.error as Error).message : null;

  return (
    <form onSubmit={submit} className="space-y-4 max-h-[70vh] overflow-y-auto">
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="role-name">Name</Label>
          <Input id="role-name" value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <Label htmlFor="role-key">Key</Label>
          <Input
            id="role-key"
            value={key}
            onChange={(e) => setKey(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, "_"))}
            required
            disabled={mode === "edit"}
            placeholder="sales_manager"
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="role-desc">Description</Label>
        <Input
          id="role-desc"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="What can this role do?"
        />
      </div>

      <div>
        <Label className="mb-2 block">Permissions ({selected.size} selected)</Label>
        <div className="space-y-4 max-h-72 overflow-y-auto rounded-md border p-3">
          {grouped.length === 0 && (
            <p className="text-sm text-muted-foreground">No permissions in catalog.</p>
          )}
          {grouped.map(([cat, perms]) => {
            const allOn = perms.every((p) => selected.has(p.key));
            return (
              <div key={cat}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    {cat}
                  </span>
                  <button
                    type="button"
                    onClick={() => toggleAll(perms)}
                    className="text-xs text-primary hover:underline"
                  >
                    {allOn ? "Clear all" : "Select all"}
                  </button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5 pl-1">
                  {perms.map((p) => (
                    <Checkbox
                      key={p.key}
                      checked={selected.has(p.key)}
                      onChange={() => toggle(p.key)}
                      label={
                        <span className="flex items-center gap-1.5">
                          <span className="font-medium">{p.action}</span>
                          {p.isDangerous && (
                            <Badge variant="danger" className="text-[10px] px-1 py-0">
                              danger
                            </Badge>
                          )}
                        </span>
                      }
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {errMsg && <p className="text-sm text-destructive">{errMsg}</p>}

      <DialogFooter>
        <Button type="button" variant="ghost" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={pending || !name || !key}>
          {pending ? "Saving…" : mode === "create" ? "Create role" : "Save changes"}
        </Button>
      </DialogFooter>
    </form>
  );
}
