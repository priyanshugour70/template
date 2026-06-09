"use client";

import {
  AlertCircle,
  Check,
  Copy,
  Mail,
  Monitor,
  PauseCircle,
  PlayCircle,
  Plus,
  Send,
  Shield,
  Trash2,
  Webhook as WebhookIcon,
} from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import {
  useApiKeys,
  useCreateApiKey,
  useRevokeApiKey,
} from "@/hooks/apikey/useApiKeys";
import {
  useChangePasswordMutation,
  useRevokeSession,
  useSessions,
} from "@/hooks/auth/useAuthMutations";
import { useMyTenant, useUpdateMyTenant } from "@/hooks/tenant/useTenantQueries";
import { toast } from "@/hooks/use-toast";
import { useUpdateUser, useUser } from "@/hooks/user/useUserQueries";
import {
  useCreateWebhook,
  useDeleteWebhook,
  useTestFireWebhook,
  useUpdateWebhook,
  useWebhookDeliveries,
  useWebhooks,
} from "@/hooks/webhook/useWebhooks";
import { cn } from "@/lib/cn";
import { useAuth, usePermissions } from "@/providers";
import type { APIKey } from "@/types/apikey";
import type { JSONObject } from "@/types/common";
import type { Webhook } from "@/types/webhook";

const WEBHOOK_EVENTS = [
  "user.created",
  "user.suspended",
  "user.role.assigned",
  "org.created",
  "org.updated",
  "subscription.changed",
  "subscription.cancelled",
  "subscription.paused",
  "invoice.created",
  "invoice.paid",
  "department.created",
  "group.created",
];

// ── profile ───────────────────────────────────────────────────────────────

export function ProfileSection() {
  const { user, refreshUser } = useAuth();
  const profileQ = useUser(user?.id);
  const profile = profileQ.data;
  const update = useUpdateUser(user?.id ?? "");

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [phone, setPhone] = useState("");
  const [bio, setBio] = useState("");
  const [hydrated, setHydrated] = useState(false);

  if (profile && !hydrated) {
    setFirstName(profile.firstName ?? "");
    setLastName(profile.lastName ?? "");
    setDisplayName(profile.displayName ?? "");
    setPhone(profile.phone ?? "");
    setBio(profile.bio ?? "");
    setHydrated(true);
  }

  if (!user) return null;
  if (profileQ.isLoading) return <Skeleton className="h-64 w-full" />;
  const initials =
    (profile?.firstName?.[0] ?? user.email?.[0] ?? "?").toUpperCase() +
    (profile?.lastName?.[0] ?? "").toUpperCase();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Personal information</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <Avatar className="h-14 w-14">
            {profile?.avatarUrl ? (
              <AvatarImage src={profile.avatarUrl} alt={user.email} />
            ) : null}
            <AvatarFallback className="text-base">{initials}</AvatarFallback>
          </Avatar>
          <div>
            <p className="font-medium">{profile?.displayName ?? user.email}</p>
            <p className="text-xs text-muted-foreground">{user.email}</p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Field label="First name">
            <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
          </Field>
          <Field label="Last name">
            <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
          </Field>
        </div>
        <Field label="Display name">
          <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
        </Field>
        <Field label="Phone">
          <Input value={phone} onChange={(e) => setPhone(e.target.value)} />
        </Field>
        <Field label="Bio">
          <Input value={bio} onChange={(e) => setBio(e.target.value)} />
        </Field>

        <div className="flex justify-end">
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                { firstName, lastName, displayName, phone, bio },
                {
                  onSuccess: async () => {
                    toast.success("Profile saved");
                    await refreshUser();
                  },
                  onError: (e: unknown) =>
                    toast.error("Save failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export function RegionalSection() {
  const { user, refreshUser } = useAuth();
  const profileQ = useUser(user?.id);
  const profile = profileQ.data;
  const update = useUpdateUser(user?.id ?? "");

  const [locale, setLocale] = useState("en-IN");
  const [timezone, setTimezone] = useState("Asia/Kolkata");
  const [hydrated, setHydrated] = useState(false);

  if (profile && !hydrated) {
    setLocale(profile.locale ?? "en-IN");
    setTimezone(profile.timezone ?? "Asia/Kolkata");
    setHydrated(true);
  }

  if (!user || profileQ.isLoading) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Regional</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Locale" hint="BCP 47 (en-US, fr-FR, en-IN…)">
            <Input value={locale} onChange={(e) => setLocale(e.target.value)} />
          </Field>
          <Field label="Timezone" hint="IANA timezone (Asia/Kolkata…)">
            <Input value={timezone} onChange={(e) => setTimezone(e.target.value)} />
          </Field>
        </div>
        <div className="flex justify-end">
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                { locale, timezone },
                {
                  onSuccess: async () => {
                    toast.success("Preferences saved");
                    await refreshUser();
                  },
                  onError: (e: unknown) =>
                    toast.error("Save failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save regional preferences
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

// ── account / password ────────────────────────────────────────────────────

export function AccountSection() {
  const { user } = useAuth();
  const profileQ = useUser(user?.id);
  const profile = profileQ.data;
  const change = useChangePasswordMutation();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");

  const mismatched = next !== "" && confirm !== "" && next !== confirm;
  const tooShort = next !== "" && next.length < 8;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Email</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <Mail className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm">{user?.email}</span>
            {profile?.emailVerifiedAt ? (
              <Badge variant="success">verified</Badge>
            ) : (
              <Badge variant="warning">unverified</Badge>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Change password</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              if (mismatched || tooShort) return;
              try {
                await change.mutateAsync({ currentPassword: current, newPassword: next });
                toast.success("Password updated");
                setCurrent("");
                setNext("");
                setConfirm("");
              } catch (err: unknown) {
                toast.error("Update failed", err instanceof Error ? err.message : undefined);
              }
            }}
            className="max-w-md space-y-3"
          >
            <Field label="Current password">
              <Input
                type="password"
                value={current}
                onChange={(e) => setCurrent(e.target.value)}
                required
              />
            </Field>
            <Field label="New password">
              <Input
                type="password"
                value={next}
                onChange={(e) => setNext(e.target.value)}
                required
              />
              {tooShort && <p className="text-xs text-warning">Use at least 8 characters.</p>}
            </Field>
            <Field label="Confirm new password">
              <Input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
              />
              {mismatched && (
                <p className="text-xs text-destructive">Passwords don&apos;t match.</p>
              )}
            </Field>
            <Button
              type="submit"
              disabled={
                change.isPending || !current || !next || !confirm || mismatched || tooShort
              }
            >
              {change.isPending ? "Updating…" : "Update password"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

// ── MFA ───────────────────────────────────────────────────────────────────

export function MFASection() {
  const { user } = useAuth();
  const profileQ = useUser(user?.id);
  const profile = profileQ.data;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Multi-factor authentication</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Shield className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-sm font-medium">
                {profile?.mfaEnabled ? "MFA enabled" : "MFA disabled"}
              </p>
              <p className="text-xs text-muted-foreground">
                {profile?.mfaEnabled
                  ? "An authenticator app is required for sign-in."
                  : "Add an authenticator app to require a second factor."}
              </p>
            </div>
          </div>
          <Badge variant={profile?.mfaEnabled ? "success" : "muted"}>
            {profile?.mfaEnabled ? "on" : "off"}
          </Badge>
        </div>
      </CardContent>
    </Card>
  );
}

// ── sessions ──────────────────────────────────────────────────────────────

export function SessionsSection() {
  const sessionsQ = useSessions();
  const revoke = useRevokeSession();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Active sessions</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        {sessionsQ.isLoading ? (
          <Skeleton className="m-4 h-16" />
        ) : (sessionsQ.data?.length ?? 0) === 0 ? (
          <p className="p-4 text-sm text-muted-foreground">No active sessions.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Device / client</TableHead>
                <TableHead className="hidden md:table-cell">IP</TableHead>
                <TableHead>Issued</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(sessionsQ.data ?? []).map((s) => (
                <TableRow key={s.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Monitor className="h-4 w-4 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">
                          {s.deviceName ?? s.client ?? "Unknown device"}
                        </p>
                        {s.userAgent && (
                          <p className="truncate text-xs text-muted-foreground max-w-md">
                            {s.userAgent}
                          </p>
                        )}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="hidden font-mono text-xs text-muted-foreground md:table-cell">
                    {s.ip ?? "—"}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {new Date(s.issuedAt).toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={revoke.isPending}
                      onClick={() =>
                        revoke.mutate(s.id, {
                          onSuccess: () => toast.success("Session revoked"),
                          onError: (e: unknown) =>
                            toast.error(
                              "Revoke failed",
                              e instanceof Error ? e.message : undefined,
                            ),
                        })
                      }
                    >
                      Revoke
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}

// ── notifications ─────────────────────────────────────────────────────────

export function NotificationsSection() {
  const { user, refreshUser } = useAuth();
  const profileQ = useUser(user?.id);
  const profile = profileQ.data;
  const update = useUpdateUser(user?.id ?? "");

  const [emailDigest, setEmailDigest] = useState(true);
  const [emailMentions, setEmailMentions] = useState(true);
  const [emailBilling, setEmailBilling] = useState(true);
  const [emailMarketing, setEmailMarketing] = useState(false);
  const [inAppBell, setInAppBell] = useState(true);
  const [hydrated, setHydrated] = useState(false);

  if (profile && !hydrated) {
    const np = (profile.notificationPreferences ?? {}) as Record<string, boolean>;
    setEmailDigest(np.emailDigest ?? true);
    setEmailMentions(np.emailMentions ?? true);
    setEmailBilling(np.emailBilling ?? true);
    setEmailMarketing(np.emailMarketing ?? false);
    setInAppBell(np.inAppBell ?? true);
    setHydrated(true);
  }

  if (!user) return null;
  if (profileQ.isLoading) return <Skeleton className="h-64 w-full" />;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Notifications</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <NotifRow
            label="Email — daily digest"
            description="A morning summary of activity."
            checked={emailDigest}
            onChange={setEmailDigest}
          />
          <NotifRow
            label="Email — mentions & assignments"
            description="When someone tags you or assigns work."
            checked={emailMentions}
            onChange={setEmailMentions}
          />
          <NotifRow
            label="Email — billing"
            description="Invoices, payment failures, plan changes."
            checked={emailBilling}
            onChange={setEmailBilling}
          />
          <NotifRow
            label="Email — product news"
            description="Tips and feature announcements."
            checked={emailMarketing}
            onChange={setEmailMarketing}
          />
          <NotifRow
            label="In-app notification bell"
            description="Show alerts in the header bell."
            checked={inAppBell}
            onChange={setInAppBell}
          />
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button
          disabled={update.isPending}
          onClick={() =>
            update.mutate(
              {
                notificationPreferences: {
                  emailDigest,
                  emailMentions,
                  emailBilling,
                  emailMarketing,
                  inAppBell,
                } as JSONObject,
              },
              {
                onSuccess: async () => {
                  toast.success("Preferences saved");
                  await refreshUser();
                },
                onError: (e: unknown) =>
                  toast.error("Save failed", e instanceof Error ? e.message : undefined),
              },
            )
          }
        >
          Save changes
        </Button>
      </div>
    </div>
  );
}

function NotifRow({
  label,
  description,
  checked,
  onChange,
}: {
  label: string;
  description: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-3 rounded-md border border-border p-3">
      <div className="min-w-0">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onChange} />
    </div>
  );
}

// ── developer (api keys + webhooks) ───────────────────────────────────────

export function DeveloperSection() {
  return (
    <div className="space-y-6">
      <APIKeysSection />
      <WebhooksSection />
    </div>
  );
}

function APIKeysSection() {
  const { has } = usePermissions();
  const keysQ = useApiKeys();
  const revoke = useRevokeApiKey();
  const [creating, setCreating] = useState(false);
  const [createdToken, setCreatedToken] = useState<{ token: string; key: APIKey } | null>(null);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className="text-base">API keys</CardTitle>
        {has("api_key.create") && (
          <Button size="sm" onClick={() => setCreating(true)}>
            <Plus className="h-4 w-4" />
            New key
          </Button>
        )}
      </CardHeader>
      <CardContent className="p-0">
        {keysQ.isLoading ? (
          <Skeleton className="m-4 h-24" />
        ) : (keysQ.data?.length ?? 0) === 0 ? (
          <p className="p-4 text-sm text-muted-foreground">
            No API keys yet. Issue one to access the public API programmatically.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Prefix</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last used</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(keysQ.data ?? []).map((k) => (
                <TableRow key={k.id}>
                  <TableCell className="font-medium">{k.name}</TableCell>
                  <TableCell>
                    <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                      {k.prefix}…
                    </code>
                  </TableCell>
                  <TableCell>
                    {k.revokedAt ? (
                      <Badge variant="danger">revoked</Badge>
                    ) : k.expiresAt && new Date(k.expiresAt) < new Date() ? (
                      <Badge variant="muted">expired</Badge>
                    ) : (
                      <Badge variant="success">active</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {k.lastUsedAt ? new Date(k.lastUsedAt).toLocaleString() : "Never"}
                  </TableCell>
                  <TableCell className="text-right">
                    {!k.revokedAt && has("api_key.delete") && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          if (confirm(`Revoke API key "${k.name}"?`)) {
                            revoke.mutate(k.id, {
                              onSuccess: () => toast.success("Key revoked"),
                              onError: (e: unknown) =>
                                toast.error(
                                  "Revoke failed",
                                  e instanceof Error ? e.message : undefined,
                                ),
                            });
                          }
                        }}
                      >
                        Revoke
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>

      {creating && (
        <CreateAPIKeyDialog
          onClose={() => setCreating(false)}
          onCreated={(token, key) => {
            setCreating(false);
            setCreatedToken({ token, key });
          }}
        />
      )}
      {createdToken && (
        <RevealSecretDialog
          title="Your new API key"
          description="Copy this now. We won't show it again."
          secret={createdToken.token}
          onClose={() => setCreatedToken(null)}
        />
      )}
    </Card>
  );
}

function CreateAPIKeyDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (token: string, key: APIKey) => void;
}) {
  const [name, setName] = useState("");
  const [rateLimit, setRateLimit] = useState("");
  const create = useCreateApiKey();

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New API key</DialogTitle>
          <DialogDescription>
            Used by integrations and CI to call the public API. Scope it narrowly.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Field label="Name">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="GitHub Actions"
            />
          </Field>
          <Field label="Rate limit (requests per minute, optional)">
            <Input
              type="number"
              min={1}
              value={rateLimit}
              onChange={(e) => setRateLimit(e.target.value)}
              placeholder="e.g. 60"
            />
          </Field>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!name || create.isPending}
            onClick={() =>
              create.mutate(
                { name, rateLimitRpm: rateLimit ? parseInt(rateLimit, 10) : undefined },
                {
                  onSuccess: (data) => onCreated(data.token, data.apiKey),
                  onError: (e: unknown) =>
                    toast.error("Create failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            {create.isPending ? "Issuing…" : "Issue key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RevealSecretDialog({
  title,
  description,
  secret,
  onClose,
}: {
  title: string;
  description: string;
  secret: string;
  onClose: () => void;
}) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(secret);
      setCopied(true);
      toast.success("Copied");
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error("Copy failed");
    }
  };
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-2 py-2">
          <div className="flex items-center gap-2 rounded-md border border-warning/40 bg-warning/10 px-3 py-2 text-xs">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>Store this somewhere safe — it can&apos;t be retrieved later.</span>
          </div>
          <div className="flex items-center gap-2 rounded-md border border-border bg-muted/40 p-2">
            <code className="flex-1 break-all font-mono text-xs">{secret}</code>
            <Button variant="outline" size="sm" onClick={copy}>
              {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              {copied ? "Copied" : "Copy"}
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button onClick={onClose}>I&apos;ve copied it</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function WebhooksSection() {
  const { has } = usePermissions();
  const hooksQ = useWebhooks();
  const [creating, setCreating] = useState(false);
  const [createdSecret, setCreatedSecret] = useState<{ secret: string; hook: Webhook } | null>(
    null,
  );
  const [editing, setEditing] = useState<Webhook | null>(null);
  const [deliveriesFor, setDeliveriesFor] = useState<Webhook | null>(null);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className="text-base">Webhooks</CardTitle>
        {has("webhook.create") && (
          <Button size="sm" onClick={() => setCreating(true)}>
            <Plus className="h-4 w-4" />
            New webhook
          </Button>
        )}
      </CardHeader>
      <CardContent className="p-0">
        {hooksQ.isLoading ? (
          <Skeleton className="m-4 h-24" />
        ) : (hooksQ.data?.length ?? 0) === 0 ? (
          <p className="p-4 text-sm text-muted-foreground">
            No webhooks yet. Register a URL to receive event POSTs.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>URL</TableHead>
                <TableHead>Events</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(hooksQ.data ?? []).map((w) => (
                <WebhookRow
                  key={w.id}
                  hook={w}
                  onEdit={() => setEditing(w)}
                  onDeliveries={() => setDeliveriesFor(w)}
                />
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>

      {creating && (
        <CreateWebhookDialog
          onClose={() => setCreating(false)}
          onCreated={(secret, hook) => {
            setCreating(false);
            setCreatedSecret({ secret, hook });
          }}
        />
      )}
      {createdSecret && (
        <RevealSecretDialog
          title={`Webhook secret for ${createdSecret.hook.name}`}
          description="Used to verify the X-Signature-256 header on every delivery."
          secret={createdSecret.secret}
          onClose={() => setCreatedSecret(null)}
        />
      )}
      {editing && <EditWebhookDialog hook={editing} onClose={() => setEditing(null)} />}
      {deliveriesFor && (
        <DeliveriesDialog hook={deliveriesFor} onClose={() => setDeliveriesFor(null)} />
      )}
    </Card>
  );
}

function WebhookRow({
  hook,
  onEdit,
  onDeliveries,
}: {
  hook: Webhook;
  onEdit: () => void;
  onDeliveries: () => void;
}) {
  const { has } = usePermissions();
  const del = useDeleteWebhook();
  const update = useUpdateWebhook(hook.id);
  const test = useTestFireWebhook(hook.id);

  return (
    <TableRow>
      <TableCell>
        <div className="font-medium">{hook.name}</div>
        {hook.description && (
          <div className="text-xs text-muted-foreground">{hook.description}</div>
        )}
      </TableCell>
      <TableCell className="max-w-[260px]">
        <code className="block truncate font-mono text-xs">{hook.url}</code>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {(hook.events ?? []).slice(0, 3).map((e) => (
            <Badge key={e} variant="muted" className="text-[10px]">
              {e}
            </Badge>
          ))}
          {hook.events.length > 3 && (
            <Badge variant="muted" className="text-[10px]">
              +{hook.events.length - 3}
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell>
        {hook.isActive ? (
          <Badge variant="success">active</Badge>
        ) : (
          <Badge variant="muted">paused</Badge>
        )}
      </TableCell>
      <TableCell>
        {hook.lastStatus ? (
          <Badge
            variant={
              hook.lastStatus >= 200 && hook.lastStatus < 300
                ? "success"
                : hook.lastStatus >= 400
                  ? "danger"
                  : "muted"
            }
          >
            {hook.lastStatus}
          </Badge>
        ) : (
          <span className="text-xs text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell className="text-right">
        <div className="inline-flex items-center gap-1">
          {has("webhook.test") && (
            <Button
              variant="ghost"
              size="sm"
              disabled={test.isPending}
              onClick={() =>
                test.mutate(
                  {},
                  {
                    onSuccess: (data) =>
                      toast.success(
                        data.delivery.status === "success"
                          ? `Test delivered — HTTP ${data.delivery.responseStatus ?? "?"}`
                          : `Test ${data.delivery.status}`,
                        data.delivery.errorMessage,
                      ),
                    onError: (e: unknown) =>
                      toast.error("Test failed", e instanceof Error ? e.message : undefined),
                  },
                )
              }
            >
              <Send className="h-4 w-4" />
              Test
            </Button>
          )}
          <Button variant="ghost" size="sm" onClick={onDeliveries}>
            Deliveries
          </Button>
          {has("webhook.update") && (
            <Button
              variant="ghost"
              size="icon"
              disabled={update.isPending}
              onClick={() =>
                update.mutate(
                  { isActive: !hook.isActive },
                  {
                    onSuccess: () =>
                      toast.success(hook.isActive ? "Webhook paused" : "Webhook resumed"),
                  },
                )
              }
              title={hook.isActive ? "Pause" : "Resume"}
            >
              {hook.isActive ? (
                <PauseCircle className="h-4 w-4" />
              ) : (
                <PlayCircle className="h-4 w-4" />
              )}
            </Button>
          )}
          {has("webhook.update") && (
            <Button variant="ghost" size="icon" onClick={onEdit}>
              <WebhookIcon className="h-4 w-4" />
            </Button>
          )}
          {has("webhook.delete") && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                if (confirm(`Delete webhook "${hook.name}"?`)) {
                  del.mutate(hook.id, {
                    onSuccess: () => toast.success("Webhook deleted"),
                    onError: (e: unknown) =>
                      toast.error("Delete failed", e instanceof Error ? e.message : undefined),
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

function CreateWebhookDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (secret: string, hook: Webhook) => void;
}) {
  const create = useCreateWebhook();
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [description, setDescription] = useState("");
  const [events, setEvents] = useState<Set<string>>(new Set());

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>New webhook</DialogTitle>
          <DialogDescription>
            We&apos;ll POST event payloads to this URL with an HMAC signature header.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Field label="Name">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Slack incoming hook"
            />
          </Field>
          <Field label="URL">
            <Input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://hooks.example.com/…"
            />
          </Field>
          <Field label="Description (optional)">
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </Field>
          <Field label="Events">
            <div className="grid grid-cols-2 gap-1 rounded-md border border-border p-2">
              {WEBHOOK_EVENTS.map((e) => (
                <label
                  key={e}
                  className="flex items-center gap-2 rounded px-1.5 py-1 hover:bg-muted/50"
                >
                  <Checkbox
                    checked={events.has(e)}
                    onChange={(ev) => {
                      const checked = ev.target.checked;
                      setEvents((prev) => {
                        const next = new Set(prev);
                        if (checked) next.add(e);
                        else next.delete(e);
                        return next;
                      });
                    }}
                  />
                  <code className="font-mono text-[11px]">{e}</code>
                </label>
              ))}
            </div>
          </Field>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!name || !url || create.isPending}
            onClick={() =>
              create.mutate(
                {
                  name,
                  url,
                  description: description || undefined,
                  events: Array.from(events),
                },
                {
                  onSuccess: (data) => onCreated(data.secret, data.webhook),
                  onError: (e: unknown) =>
                    toast.error("Create failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            {create.isPending ? "Creating…" : "Create webhook"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function EditWebhookDialog({ hook, onClose }: { hook: Webhook; onClose: () => void }) {
  const update = useUpdateWebhook(hook.id);
  const [name, setName] = useState(hook.name);
  const [url, setUrl] = useState(hook.url);
  const [description, setDescription] = useState(hook.description ?? "");
  const [events, setEvents] = useState<Set<string>>(new Set(hook.events));

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit webhook</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="URL">
            <Input value={url} onChange={(e) => setUrl(e.target.value)} />
          </Field>
          <Field label="Description">
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </Field>
          <Field label="Events">
            <div className="grid grid-cols-2 gap-1 rounded-md border border-border p-2">
              {WEBHOOK_EVENTS.map((e) => (
                <label
                  key={e}
                  className="flex items-center gap-2 rounded px-1.5 py-1 hover:bg-muted/50"
                >
                  <Checkbox
                    checked={events.has(e)}
                    onChange={(ev) => {
                      const checked = ev.target.checked;
                      setEvents((prev) => {
                        const next = new Set(prev);
                        if (checked) next.add(e);
                        else next.delete(e);
                        return next;
                      });
                    }}
                  />
                  <code className="font-mono text-[11px]">{e}</code>
                </label>
              ))}
            </div>
          </Field>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                { name, url, description, events: Array.from(events) },
                {
                  onSuccess: () => {
                    toast.success("Webhook updated");
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

function DeliveriesDialog({ hook, onClose }: { hook: Webhook; onClose: () => void }) {
  const dQ = useWebhookDeliveries(hook.id);
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Recent deliveries — {hook.name}</DialogTitle>
          <DialogDescription>Last 50 attempts, newest first.</DialogDescription>
        </DialogHeader>
        <div className="max-h-[60vh] overflow-y-auto">
          {dQ.isLoading ? (
            <Skeleton className="h-32" />
          ) : (dQ.data?.length ?? 0) === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">
              No deliveries yet.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Event</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>HTTP</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(dQ.data ?? []).map((d) => (
                  <TableRow key={d.id}>
                    <TableCell>
                      <code className="font-mono text-xs">{d.event}</code>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          d.status === "success"
                            ? "success"
                            : d.status === "failed"
                              ? "danger"
                              : "muted"
                        }
                      >
                        {d.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {d.responseStatus ?? "—"}
                    </TableCell>
                    <TableCell className="tabular-nums text-xs text-muted-foreground">
                      {d.durationMs != null ? `${d.durationMs} ms` : "—"}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {new Date(d.createdAt).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

// ── tenant ────────────────────────────────────────────────────────────────

export function TenantSection() {
  const tenantQ = useMyTenant();
  const update = useUpdateMyTenant();

  const tenant = tenantQ.data;
  const [name, setName] = useState("");
  const [legalName, setLegalName] = useState("");
  const [description, setDescription] = useState("");
  const [supportEmail, setSupportEmail] = useState("");
  const [supportPhone, setSupportPhone] = useState("");
  const [website, setWebsite] = useState("");
  const [primaryColor, setPrimaryColor] = useState("");
  const [hydrated, setHydrated] = useState(false);

  if (tenant && !hydrated) {
    setName(tenant.name ?? "");
    setLegalName(tenant.legalName ?? "");
    setDescription(tenant.description ?? "");
    setSupportEmail(tenant.supportEmail ?? "");
    setSupportPhone(tenant.supportPhone ?? "");
    setWebsite(tenant.websiteUrl ?? "");
    setPrimaryColor(tenant.primaryColor ?? "");
    setHydrated(true);
  }

  if (tenantQ.isLoading) return <Skeleton className="h-48 w-full" />;
  if (!tenant) return <p className="text-sm text-muted-foreground">No tenant.</p>;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{tenant.name}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Legal name">
            <Input value={legalName} onChange={(e) => setLegalName(e.target.value)} />
          </Field>
        </div>
        <Field label="Description">
          <Input value={description} onChange={(e) => setDescription(e.target.value)} />
        </Field>
        <div className="grid grid-cols-2 gap-3">
          <Field label="Support email">
            <Input
              type="email"
              value={supportEmail}
              onChange={(e) => setSupportEmail(e.target.value)}
            />
          </Field>
          <Field label="Support phone">
            <Input value={supportPhone} onChange={(e) => setSupportPhone(e.target.value)} />
          </Field>
        </div>
        <Field label="Website URL">
          <Input value={website} onChange={(e) => setWebsite(e.target.value)} />
        </Field>
        <Field label="Primary color">
          <div className="flex items-center gap-2">
            <input
              type="color"
              value={primaryColor || "#000000"}
              onChange={(e) => setPrimaryColor(e.target.value)}
              className="h-9 w-12 cursor-pointer rounded border border-input bg-background"
            />
            <Input
              value={primaryColor}
              onChange={(e) => setPrimaryColor(e.target.value)}
              placeholder="#2E7D32"
              className={cn("font-mono")}
            />
          </div>
        </Field>

        <div className="flex justify-end">
          <Button
            disabled={update.isPending}
            onClick={() =>
              update.mutate(
                {
                  name,
                  legalName,
                  description,
                  supportEmail,
                  supportPhone,
                  websiteUrl: website,
                  primaryColor,
                },
                {
                  onSuccess: () => toast.success("Tenant updated"),
                  onError: (e: unknown) =>
                    toast.error("Update failed", e instanceof Error ? e.message : undefined),
                },
              )
            }
          >
            Save changes
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

// ── shared ────────────────────────────────────────────────────────────────

export function Field({
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
