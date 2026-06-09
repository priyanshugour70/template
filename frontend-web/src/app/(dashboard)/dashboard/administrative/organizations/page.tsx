"use client";

import {
  Archive,
  Building2,
  ExternalLink,
  Mail,
  MapPin,
  Phone,
  Plus,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { JsonEditor } from "@/components/ui/json-editor";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  useArchiveOrganization,
  useCreateOrganization,
  useOrganizations,
  useUpdateOrganization,
} from "@/hooks/tenant/useTenantQueries";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";
import { usePermissions } from "@/providers";
import type {
  Organization,
  UpdateOrganizationRequest,
} from "@/types/tenant";
import type { JSONObject } from "@/types/common";

function slugify(s: string): string {
  return s
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export default function OrganizationsPage() {
  const orgsQ = useOrganizations();
  const { has } = usePermissions();
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Organization | null>(null);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">Organizations</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Workspaces inside your tenant. Each can have its own users, roles, and
            subscription. Click any card to manage every detail.
          </p>
        </div>
        {has("org.create") && (
          <Button onClick={() => setCreating(true)}>
            <Plus className="h-4 w-4" />
            New organization
          </Button>
        )}
      </div>

      {orgsQ.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-60" />
          ))}
        </div>
      ) : !orgsQ.data?.length ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 py-12 text-center text-sm text-muted-foreground">
            <Building2 className="h-8 w-8" />
            <p>No organizations yet.</p>
            {has("org.create") && (
              <Button variant="outline" onClick={() => setCreating(true)}>
                <Plus className="h-4 w-4" />
                Create the first one
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {orgsQ.data.map((o) => (
            <OrgCard key={o.id} org={o} onOpen={() => setEditing(o)} />
          ))}
        </div>
      )}

      {creating && <CreateDialog onClose={() => setCreating(false)} />}
      {editing && (
        <EditSheet
          orgId={editing.id}
          initial={editing}
          onClose={() => setEditing(null)}
        />
      )}
    </div>
  );
}

// ── card ──────────────────────────────────────────────────────────────────

function OrgCard({ org, onOpen }: { org: Organization; onOpen: () => void }) {
  return (
    <Card
      className="cursor-pointer overflow-hidden transition-colors hover:border-primary/40"
      onClick={onOpen}
    >
      <CardHeader>
        <div className="flex items-start gap-3">
          {org.logoUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={org.logoUrl}
              alt={org.name}
              className="h-12 w-12 rounded-md object-cover"
            />
          ) : (
            <div
              className="flex h-12 w-12 items-center justify-center rounded-md"
              style={{
                background: org.primaryColor
                  ? `${org.primaryColor}1a`
                  : undefined,
              }}
            >
              <Building2
                className="h-5 w-5"
                style={{ color: org.primaryColor ?? undefined }}
              />
            </div>
          )}
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="truncate font-semibold">{org.name}</h3>
              {org.isDefault && <Badge>default</Badge>}
            </div>
            <div className="truncate text-xs text-muted-foreground">/{org.slug}</div>
          </div>
          <Badge variant={org.status === "active" ? "success" : "warning"}>
            {org.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {org.description && (
          <p className="line-clamp-2 text-muted-foreground">{org.description}</p>
        )}
        <dl className="space-y-1.5 text-sm">
          {org.industry && <Row label="Industry" value={org.industry} />}
          {org.size && <Row label="Team size" value={org.size} />}
          {(org.city || org.country) && (
            <div className="flex items-center gap-2 text-muted-foreground">
              <MapPin className="h-3.5 w-3.5" />
              <span>{[org.city, org.country].filter(Boolean).join(", ")}</span>
            </div>
          )}
          {org.contactEmail && (
            <div className="flex items-center gap-2 truncate text-muted-foreground">
              <Mail className="h-3.5 w-3.5" />
              <span className="truncate">{org.contactEmail}</span>
            </div>
          )}
          {org.contactPhone && (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Phone className="h-3.5 w-3.5" />
              <span>{org.contactPhone}</span>
            </div>
          )}
          {org.websiteUrl && (
            <div
              className="flex items-center gap-2"
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className="h-3.5 w-3.5 text-muted-foreground" />
              <a
                href={org.websiteUrl}
                target="_blank"
                rel="noreferrer"
                className="truncate text-primary hover:underline"
              >
                {org.websiteUrl}
              </a>
            </div>
          )}
        </dl>
      </CardContent>
    </Card>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="max-w-[60%] truncate font-medium">{value}</span>
    </div>
  );
}

// ── edit sheet ────────────────────────────────────────────────────────────

function EditSheet({
  orgId,
  initial,
  onClose,
}: {
  orgId: string;
  initial: Organization;
  onClose: () => void;
}) {
  // We don't refetch on open — `initial` from the list is the source of truth,
  // and react-query invalidation on save will refresh the list afterward.
  // To always have the freshest values, useOrganization(orgId) would re-fetch,
  // but it's redundant since the list endpoint returns the full row.
  const org = initial;

  const { has } = usePermissions();
  const canEdit = has("org.update");
  const archive = useArchiveOrganization();

  return (
    <Sheet open onOpenChange={(o) => !o && onClose()}>
      <SheetContent
        side="right"
        className="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-2xl"
      >
        <div className="border-b border-border p-6">
          <div className="flex items-start gap-3">
            {org.logoUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={org.logoUrl}
                alt={org.name}
                className="h-12 w-12 rounded-md object-cover"
              />
            ) : (
              <div className="flex h-12 w-12 items-center justify-center rounded-md bg-primary/10">
                <Building2 className="h-5 w-5 text-primary" />
              </div>
            )}
            <div className="min-w-0 flex-1">
              <SheetTitle className="truncate text-lg">{org.name}</SheetTitle>
              <SheetDescription className="truncate text-sm">
                /{org.slug}
              </SheetDescription>
              <div className="mt-1 flex items-center gap-2">
                <Badge variant={org.status === "active" ? "success" : "warning"}>
                  {org.status}
                </Badge>
                {org.isDefault && <Badge>default</Badge>}
              </div>
            </div>
            {has("org.delete") && !org.isDefault && (
              <Button
                variant="ghost"
                size="sm"
                className="text-destructive hover:text-destructive"
                onClick={() => {
                  if (
                    confirm(
                      `Archive organization "${org.name}"? Memberships and data remain but the org will be hidden.`,
                    )
                  ) {
                    archive.mutate(org.id, {
                      onSuccess: () => {
                        toast.success("Organization archived");
                        onClose();
                      },
                      onError: (e: unknown) =>
                        toast.error(
                          "Archive failed",
                          e instanceof Error ? e.message : undefined,
                        ),
                    });
                  }
                }}
              >
                <Archive className="h-4 w-4" />
                Archive
              </Button>
            )}
          </div>
        </div>

        <div className="flex-1 p-6">
          <Tabs defaultValue="general" className="w-full">
            <TabsList className="flex-wrap h-auto">
              <TabsTrigger value="general">General</TabsTrigger>
              <TabsTrigger value="branding">Branding</TabsTrigger>
              <TabsTrigger value="contact">Contact</TabsTrigger>
              <TabsTrigger value="location">Location</TabsTrigger>
              <TabsTrigger value="localization">Localization</TabsTrigger>
              <TabsTrigger value="advanced">Advanced</TabsTrigger>
            </TabsList>

            <TabsContent value="general">
              <GeneralTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
            <TabsContent value="branding">
              <BrandingTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
            <TabsContent value="contact">
              <ContactTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
            <TabsContent value="location">
              <LocationTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
            <TabsContent value="localization">
              <LocalizationTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
            <TabsContent value="advanced">
              <AdvancedTab orgId={orgId} org={org} canEdit={canEdit} />
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  );
}

// ── tabs ──────────────────────────────────────────────────────────────────

interface TabProps {
  orgId: string;
  org: Organization;
  canEdit: boolean;
}

function useOrgPatch(orgId: string, initial: UpdateOrganizationRequest) {
  // local form state + dirty tracking
  const [draft, setDraft] = useState<UpdateOrganizationRequest>(initial);
  // Reset when org/orgId changes
  useEffect(() => {
    setDraft(initial);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orgId]);

  const update = useUpdateOrganization(orgId);

  // Compute dirty by shallow string compare of relevant keys
  const dirty = useMemo(() => {
    for (const k of Object.keys(draft) as (keyof UpdateOrganizationRequest)[]) {
      if (!shallowEqualValue(draft[k], initial[k])) return true;
    }
    return false;
  }, [draft, initial]);

  const save = (onDone?: () => void) =>
    update.mutate(draft, {
      onSuccess: () => {
        toast.success("Organization updated");
        onDone?.();
      },
      onError: (e: unknown) =>
        toast.error("Update failed", e instanceof Error ? e.message : undefined),
    });

  return { draft, setDraft, dirty, saving: update.isPending, save };
}

function shallowEqualValue(a: unknown, b: unknown) {
  if (a == null && b == null) return true;
  if (typeof a === "object" || typeof b === "object") {
    return JSON.stringify(a ?? null) === JSON.stringify(b ?? null);
  }
  return a === b;
}

function SaveBar({
  dirty,
  saving,
  onSave,
  canEdit,
  onReset,
}: {
  dirty: boolean;
  saving: boolean;
  onSave: () => void;
  canEdit: boolean;
  onReset?: () => void;
}) {
  if (!canEdit) return null;
  return (
    <div className="sticky bottom-0 -mx-6 mt-6 flex items-center justify-end gap-2 border-t border-border bg-background/80 px-6 py-3 backdrop-blur">
      {onReset && (
        <Button variant="ghost" disabled={!dirty || saving} onClick={onReset}>
          Discard
        </Button>
      )}
      <Button disabled={!dirty || saving} onClick={onSave}>
        {saving ? "Saving…" : "Save changes"}
      </Button>
    </div>
  );
}

function Field({
  label,
  children,
  hint,
}: {
  label: string;
  children: React.ReactNode;
  hint?: string;
}) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs uppercase tracking-wider text-muted-foreground">
        {label}
      </Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function GeneralTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    name: org.name,
    displayName: org.displayName ?? "",
    description: org.description ?? "",
    industry: org.industry ?? "",
    size: org.size ?? "",
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);

  return (
    <div className="space-y-4">
      <Field label="Name">
        <Input
          value={draft.name ?? ""}
          onChange={(e) => setDraft({ ...draft, name: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Display name" hint="Shown in the sidebar and emails. Defaults to Name if blank.">
        <Input
          value={draft.displayName ?? ""}
          onChange={(e) => setDraft({ ...draft, displayName: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Description">
        <Input
          value={draft.description ?? ""}
          onChange={(e) => setDraft({ ...draft, description: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label="Industry" hint="SaaS, Retail, Healthcare…">
          <Input
            value={draft.industry ?? ""}
            onChange={(e) => setDraft({ ...draft, industry: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
        <Field label="Team size" hint="1-10, 11-50, 50-200, 200+">
          <Input
            value={draft.size ?? ""}
            onChange={(e) => setDraft({ ...draft, size: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
      </div>

      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

function BrandingTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    logoUrl: org.logoUrl ?? "",
    coverUrl: org.coverUrl ?? "",
    primaryColor: org.primaryColor ?? "",
    secondaryColor: org.secondaryColor ?? "",
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);

  return (
    <div className="space-y-4">
      <Field label="Logo URL">
        <Input
          placeholder="https://…/logo.png"
          value={draft.logoUrl ?? ""}
          onChange={(e) => setDraft({ ...draft, logoUrl: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Cover URL">
        <Input
          placeholder="https://…/cover.jpg"
          value={draft.coverUrl ?? ""}
          onChange={(e) => setDraft({ ...draft, coverUrl: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label="Primary color">
          <ColorInput
            value={draft.primaryColor ?? ""}
            onChange={(v) => setDraft({ ...draft, primaryColor: v })}
            disabled={!canEdit}
          />
        </Field>
        <Field label="Secondary color">
          <ColorInput
            value={draft.secondaryColor ?? ""}
            onChange={(v) => setDraft({ ...draft, secondaryColor: v })}
            disabled={!canEdit}
          />
        </Field>
      </div>

      {/* Live preview */}
      <div className="rounded-md border border-border p-4">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Preview</p>
        <div className="mt-3 flex items-center gap-3">
          {draft.logoUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={draft.logoUrl}
              alt="logo preview"
              className="h-12 w-12 rounded-md object-cover"
            />
          ) : (
            <div
              className="flex h-12 w-12 items-center justify-center rounded-md"
              style={{
                background: draft.primaryColor
                  ? `${draft.primaryColor}1a`
                  : undefined,
              }}
            >
              <Building2
                className="h-5 w-5"
                style={{ color: draft.primaryColor || undefined }}
              />
            </div>
          )}
          <div className="text-sm font-semibold">{org.name}</div>
        </div>
      </div>

      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

function ColorInput({
  value,
  onChange,
  disabled,
}: {
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center gap-2">
      <input
        type="color"
        value={value || "#000000"}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className={cn(
          "h-9 w-12 cursor-pointer rounded border border-input bg-background",
          disabled && "cursor-not-allowed opacity-50",
        )}
      />
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="#2E7D32"
        disabled={disabled}
        className="font-mono"
      />
    </div>
  );
}

function ContactTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    contactEmail: org.contactEmail ?? "",
    contactPhone: org.contactPhone ?? "",
    websiteUrl: org.websiteUrl ?? "",
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);
  return (
    <div className="space-y-4">
      <Field label="Contact email">
        <Input
          type="email"
          value={draft.contactEmail ?? ""}
          onChange={(e) => setDraft({ ...draft, contactEmail: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Contact phone">
        <Input
          value={draft.contactPhone ?? ""}
          onChange={(e) => setDraft({ ...draft, contactPhone: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Website URL">
        <Input
          placeholder="https://acme.example"
          value={draft.websiteUrl ?? ""}
          onChange={(e) => setDraft({ ...draft, websiteUrl: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

function LocationTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    country: org.country ?? "",
    state: org.state ?? "",
    city: org.city ?? "",
    postalCode: org.postalCode ?? "",
    address: (org.address ?? {}) as JSONObject,
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <Field label="Country">
          <Input
            value={draft.country ?? ""}
            onChange={(e) => setDraft({ ...draft, country: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
        <Field label="State / region">
          <Input
            value={draft.state ?? ""}
            onChange={(e) => setDraft({ ...draft, state: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Field label="City">
          <Input
            value={draft.city ?? ""}
            onChange={(e) => setDraft({ ...draft, city: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
        <Field label="Postal code">
          <Input
            value={draft.postalCode ?? ""}
            onChange={(e) => setDraft({ ...draft, postalCode: e.target.value })}
            disabled={!canEdit}
          />
        </Field>
      </div>
      <Field
        label="Full address (JSON)"
        hint='Free-form structured address. Shape is up to you, e.g. {"line1":"…","line2":"…"}'
      >
        <JsonEditor
          value={draft.address ?? {}}
          onChange={(v) => setDraft({ ...draft, address: v as JSONObject })}
          disabled={!canEdit}
          rows={6}
        />
      </Field>
      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

function LocalizationTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    timezone: org.timezone,
    locale: org.locale,
    currency: org.currency,
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);

  return (
    <div className="space-y-4">
      <Field label="Timezone" hint="IANA name, e.g. Asia/Kolkata, America/Los_Angeles">
        <Input
          value={draft.timezone ?? ""}
          onChange={(e) => setDraft({ ...draft, timezone: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Locale" hint="BCP 47 tag, e.g. en-IN, en-US, fr-FR">
        <Input
          value={draft.locale ?? ""}
          onChange={(e) => setDraft({ ...draft, locale: e.target.value })}
          disabled={!canEdit}
        />
      </Field>
      <Field label="Currency" hint="ISO 4217 code, e.g. INR, USD, EUR">
        <Input
          value={draft.currency ?? ""}
          onChange={(e) =>
            setDraft({ ...draft, currency: e.target.value.toUpperCase() })
          }
          disabled={!canEdit}
        />
      </Field>
      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

function AdvancedTab({ orgId, org, canEdit }: TabProps) {
  const initial: UpdateOrganizationRequest = {
    settings: (org.settings ?? {}) as JSONObject,
    features: (org.features ?? {}) as JSONObject,
    metadata: (org.metadata ?? {}) as JSONObject,
  };
  const { draft, setDraft, dirty, saving, save } = useOrgPatch(orgId, initial);

  return (
    <div className="space-y-5">
      <Field
        label="Settings"
        hint="App-level configuration. Read by your code; never shown to end users."
      >
        <JsonEditor
          value={draft.settings ?? {}}
          onChange={(v) => setDraft({ ...draft, settings: v as JSONObject })}
          disabled={!canEdit}
          rows={6}
        />
      </Field>
      <Field
        label="Features"
        hint="Feature flags / entitlement toggles for this organization."
      >
        <JsonEditor
          value={draft.features ?? {}}
          onChange={(v) => setDraft({ ...draft, features: v as JSONObject })}
          disabled={!canEdit}
          rows={6}
        />
      </Field>
      <Field
        label="Metadata"
        hint="Free-form key/value bag for integrations and reporting."
      >
        <JsonEditor
          value={draft.metadata ?? {}}
          onChange={(v) => setDraft({ ...draft, metadata: v as JSONObject })}
          disabled={!canEdit}
          rows={6}
        />
      </Field>
      <SaveBar
        canEdit={canEdit}
        dirty={dirty}
        saving={saving}
        onSave={() => save()}
        onReset={() => setDraft(initial)}
      />
    </div>
  );
}

// ── create dialog ─────────────────────────────────────────────────────────

function CreateDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateOrganization();
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [industry, setIndustry] = useState("");
  const [size, setSize] = useState("");
  const [website, setWebsite] = useState("");
  const [email, setEmail] = useState("");
  const [country, setCountry] = useState("");
  const [city, setCity] = useState("");

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>New organization</DialogTitle>
          <DialogDescription>
            Create another workspace inside your tenant. You can edit every other
            field after creating.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            create.mutate(
              {
                name: name.trim(),
                slug: slug || slugify(name),
                description: description.trim() || undefined,
                industry: industry.trim() || undefined,
                size: size.trim() || undefined,
                websiteUrl: website.trim() || undefined,
                contactEmail: email.trim() || undefined,
                country: country.trim() || undefined,
              },
              {
                onSuccess: () => {
                  toast.success("Organization created");
                  onClose();
                },
                onError: (e: unknown) =>
                  toast.error(
                    "Create failed",
                    e instanceof Error ? e.message : undefined,
                  ),
              },
            );
            // unused
            void city;
          }}
          className="space-y-3 py-2"
        >
          <div className="grid grid-cols-2 gap-3">
            <Field label="Name">
              <Input
                required
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (!slug) setSlug(slugify(e.target.value));
                }}
              />
            </Field>
            <Field label="Slug">
              <Input
                required
                value={slug}
                onChange={(e) => setSlug(slugify(e.target.value))}
              />
            </Field>
          </div>
          <Field label="Description">
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Industry">
              <Input value={industry} onChange={(e) => setIndustry(e.target.value)} />
            </Field>
            <Field label="Team size">
              <Input value={size} onChange={(e) => setSize(e.target.value)} />
            </Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Website">
              <Input value={website} onChange={(e) => setWebsite(e.target.value)} />
            </Field>
            <Field label="Contact email">
              <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            </Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Country">
              <Input value={country} onChange={(e) => setCountry(e.target.value)} />
            </Field>
            <Field label="City">
              <Input value={city} onChange={(e) => setCity(e.target.value)} />
            </Field>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name || create.isPending}>
              {create.isPending ? "Creating…" : "Create organization"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
