"use client";

import { ChevronDown, ChevronRight, GitBranch, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";
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
  useAssignDepartmentRoles,
  useCreateDepartment,
  useDeleteDepartment,
  useDepartmentRoles,
  useDepartments,
  useDepartmentTree,
  useMoveDepartment,
  useUpdateDepartment,
} from "@/hooks/department/useDepartments";
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";
import { usePermissions } from "@/providers";
import type { Department, DepartmentNode } from "@/types/department";

export default function DepartmentsPage() {
  const treeQ = useDepartmentTree();
  const flatQ = useDepartments();
  const { has } = usePermissions();

  const [creating, setCreating] = useState<{ parentId: string | null } | null>(null);
  const [editing, setEditing] = useState<Department | null>(null);
  const [rolesFor, setRolesFor] = useState<Department | null>(null);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">Departments</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Hierarchical org units. Role grants on a department flow down to descendants — useful for
            attaching default permissions to a whole branch of the org.
          </p>
        </div>
        {has("department.create") && (
          <Button onClick={() => setCreating({ parentId: null })}>
            <Plus className="h-4 w-4" />
            New department
          </Button>
        )}
      </div>

      {treeQ.isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 4 }, (_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : (treeQ.data?.length ?? 0) === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 py-10 text-center">
            <GitBranch className="h-8 w-8 text-muted-foreground" />
            <div>
              <p className="font-medium">No departments yet</p>
              <p className="text-sm text-muted-foreground">
                Create your first department to start organizing roles by team.
              </p>
            </div>
            {has("department.create") && (
              <Button variant="outline" onClick={() => setCreating({ parentId: null })}>
                <Plus className="h-4 w-4" />
                New department
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-2">
            {(treeQ.data ?? []).map((node) => (
              <TreeNode
                key={node.id}
                node={node}
                depth={0}
                onAddChild={(parentId) => setCreating({ parentId })}
                onEdit={setEditing}
                onRoles={setRolesFor}
              />
            ))}
          </CardContent>
        </Card>
      )}

      {creating && (
        <CreateDialog
          parentId={creating.parentId}
          flatDepartments={flatQ.data?.items ?? []}
          onClose={() => setCreating(null)}
        />
      )}
      {editing && (
        <EditDialog
          dept={editing}
          flatDepartments={flatQ.data?.items ?? []}
          onClose={() => setEditing(null)}
        />
      )}
      {rolesFor && <RolesDialog dept={rolesFor} onClose={() => setRolesFor(null)} />}
    </div>
  );
}

function TreeNode({
  node,
  depth,
  onAddChild,
  onEdit,
  onRoles,
}: {
  node: DepartmentNode;
  depth: number;
  onAddChild: (parentId: string) => void;
  onEdit: (d: Department) => void;
  onRoles: (d: Department) => void;
}) {
  const [expanded, setExpanded] = useState(true);
  const { has } = usePermissions();
  const del = useDeleteDepartment();
  const hasChildren = (node.children?.length ?? 0) > 0;

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-2 rounded-md px-2 py-2 hover:bg-muted/50 transition-colors",
        )}
        style={{ paddingLeft: 8 + depth * 20 }}
      >
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 shrink-0"
          onClick={() => setExpanded((e) => !e)}
          disabled={!hasChildren}
        >
          {hasChildren ? (
            expanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )
          ) : (
            <span className="block h-4 w-4" />
          )}
        </Button>
        <GitBranch className="h-4 w-4 shrink-0 text-muted-foreground" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium truncate">{node.name}</span>
            {node.isArchived && <Badge variant="muted">archived</Badge>}
            {node.costCenter && (
              <Badge variant="muted" className="font-mono text-xs">
                {node.costCenter}
              </Badge>
            )}
          </div>
          {node.description && (
            <p className="text-xs text-muted-foreground truncate">{node.description}</p>
          )}
        </div>
        <div className="flex items-center gap-1">
          {has("department.assign") && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRoles(node)}
              className="hidden sm:inline-flex"
            >
              <ShieldCheck className="h-4 w-4" />
              Roles
            </Button>
          )}
          {has("department.create") && (
            <Button variant="ghost" size="icon" onClick={() => onAddChild(node.id)} title="Add child">
              <Plus className="h-4 w-4" />
            </Button>
          )}
          {has("department.update") && (
            <Button variant="ghost" size="icon" onClick={() => onEdit(node)} title="Edit">
              <Pencil className="h-4 w-4" />
            </Button>
          )}
          {has("department.delete") && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                if (confirm(`Archive department "${node.name}"?`)) {
                  del.mutate(node.id, {
                    onSuccess: () => toast.success("Department archived"),
                    onError: (e: unknown) =>
                      toast.error(
                        "Archive failed",
                        e instanceof Error ? e.message : undefined,
                      ),
                  });
                }
              }}
              title="Archive"
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          )}
        </div>
      </div>
      {hasChildren && expanded && (
        <div>
          {node.children!.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              depth={depth + 1}
              onAddChild={onAddChild}
              onEdit={onEdit}
              onRoles={onRoles}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function CreateDialog({
  parentId,
  flatDepartments,
  onClose,
}: {
  parentId: string | null;
  flatDepartments: Department[];
  onClose: () => void;
}) {
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [parent, setParent] = useState<string>(parentId ?? "_root");
  const [costCenter, setCostCenter] = useState("");
  const create = useCreateDepartment();

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New department</DialogTitle>
          <DialogDescription>
            Create a unit in the org tree. You can attach role grants later.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 py-2">
          <div className="grid gap-1.5">
            <Label htmlFor="dept-name">Name</Label>
            <Input
              id="dept-name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (!slug)
                  setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, ""));
              }}
              placeholder="Engineering"
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="dept-slug">Slug</Label>
            <Input
              id="dept-slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="engineering"
            />
          </div>
          <div className="grid gap-1.5">
            <Label>Parent department</Label>
            <Select value={parent} onValueChange={setParent}>
              <SelectTrigger>
                <SelectValue placeholder="None (top level)" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_root">None (top level)</SelectItem>
                {flatDepartments.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="dept-cost">Cost center (optional)</Label>
            <Input id="dept-cost" value={costCenter} onChange={(e) => setCostCenter(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="dept-desc">Description (optional)</Label>
            <Input
              id="dept-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() =>
              create.mutate(
                {
                  name,
                  slug,
                  description: description || undefined,
                  costCenter: costCenter || undefined,
                  parentId: parent === "_root" ? null : parent,
                },
                {
                  onSuccess: () => {
                    toast.success("Department created");
                    onClose();
                  },
                  onError: (e: unknown) =>
                    toast.error("Create failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
            disabled={!name || !slug || create.isPending}
          >
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function EditDialog({
  dept,
  flatDepartments,
  onClose,
}: {
  dept: Department;
  flatDepartments: Department[];
  onClose: () => void;
}) {
  const [name, setName] = useState(dept.name);
  const [description, setDescription] = useState(dept.description ?? "");
  const [costCenter, setCostCenter] = useState(dept.costCenter ?? "");
  const [parent, setParent] = useState<string>(dept.parentId ?? "_root");

  const update = useUpdateDepartment(dept.id);
  const move = useMoveDepartment(dept.id);

  // Prevent picking a parent that's a descendant of this dept — backend will
  // reject, but the UI should filter for clarity. Best-effort using flat list.
  const eligibleParents = flatDepartments.filter((d) => d.id !== dept.id);

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit department</DialogTitle>
          <DialogDescription>{dept.slug}</DialogDescription>
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
          <div className="grid gap-1.5">
            <Label>Cost center</Label>
            <Input value={costCenter} onChange={(e) => setCostCenter(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>Parent department</Label>
            <Select value={parent} onValueChange={setParent}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_root">None (top level)</SelectItem>
                {eligibleParents.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={async () => {
              const newParent = parent === "_root" ? null : parent;
              const parentChanged = (dept.parentId ?? null) !== newParent;
              try {
                await update.mutateAsync({
                  name,
                  description: description || undefined,
                  costCenter: costCenter || undefined,
                });
                if (parentChanged) {
                  await move.mutateAsync({ parentId: newParent });
                }
                toast.success("Department updated");
                onClose();
              } catch (e: unknown) {
                toast.error("Update failed", e instanceof Error ? e.message : undefined);
              }
            }}
            disabled={!name || update.isPending || move.isPending}
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RolesDialog({ dept, onClose }: { dept: Department; onClose: () => void }) {
  const rolesQ = useRoles();
  const currentQ = useDepartmentRoles(dept.id);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const assign = useAssignDepartmentRoles(dept.id);

  const initial = useMemo(() => new Set(currentQ.data ?? []), [currentQ.data]);
  // Initialize selection once current loads.
  if (currentQ.data && selected.size === 0 && initial.size > 0) {
    setSelected(initial);
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Department roles</DialogTitle>
          <DialogDescription>
            Roles attached to <span className="font-medium">{dept.name}</span> apply to every
            member of this department and its descendants.
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
