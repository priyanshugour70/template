"use client";

import { Copy, Plus, Webhook, X } from "lucide-react";
import { useState } from "react";

import {
  Button,
  Input,
  Label,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui";
import {
  useChannelHooks,
  useCreateHook,
  useRevokeHook,
} from "@/hooks/communication/useCommunication";

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;
  conversationId: string;
}

// Inbound channel webhooks. Token is shown exactly once on creation; the
// list view never reveals it. Matches the Phase 1 backend security model.
export function HooksDrawer({ open, onOpenChange, conversationId }: Props) {
  const { data: hooks = [], isLoading } = useChannelHooks(conversationId);
  const create = useCreateHook(conversationId);
  const revoke = useRevokeHook(conversationId);
  const [name, setName] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [justCreated, setJustCreated] = useState<{ token: string; url: string } | null>(null);

  async function onCreate() {
    if (!name.trim()) return;
    const out = await create.mutateAsync({
      name: name.trim(),
      displayName: displayName.trim() || undefined,
    });
    setJustCreated({ token: out.token, url: out.url });
    setName("");
    setDisplayName("");
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-md">
        <div className="space-y-1.5">
          <SheetTitle className="flex items-center gap-2">
            <Webhook className="h-4 w-4" />
            Inbound webhooks
          </SheetTitle>
          <SheetDescription>
            Generate a URL external services can POST to. Compatible with Slack-shaped payloads.
          </SheetDescription>
        </div>

        <div className="mt-6 space-y-4" data-testid="hooks-panel">
          <div className="space-y-2 rounded border border-border p-3">
            <Label htmlFor="hook-name">Name</Label>
            <Input
              id="hook-name"
              data-testid="hook-name-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Sentry alerts"
            />
            <Label htmlFor="hook-display">Display name (optional)</Label>
            <Input
              id="hook-display"
              data-testid="hook-display-input"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="e.g. Sentry"
            />
            <Button
              size="sm"
              className="w-full"
              onClick={() => void onCreate()}
              disabled={!name.trim() || create.isPending}
              data-testid="hook-create-submit"
            >
              <Plus className="h-3.5 w-3.5 mr-1" />
              Create hook
            </Button>
          </div>

          {justCreated && (
            <div
              className="rounded border border-amber-200 bg-amber-50 dark:bg-amber-950/30 p-3 space-y-2"
              data-testid="hook-secret-banner"
            >
              <div className="text-xs font-semibold text-amber-900 dark:text-amber-200">
                Copy this URL now — it won't be shown again.
              </div>
              <div className="flex gap-1">
                <Input readOnly value={justCreated.url} className="text-xs font-mono" />
                <Button
                  size="icon"
                  variant="outline"
                  onClick={() => navigator.clipboard.writeText(justCreated.url)}
                  aria-label="Copy URL"
                >
                  <Copy className="h-3.5 w-3.5" />
                </Button>
              </div>
              <Button
                size="sm"
                variant="ghost"
                className="w-full"
                onClick={() => setJustCreated(null)}
              >
                Done
              </Button>
            </div>
          )}

          <div className="space-y-1">
            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wide px-1">
              Existing hooks
            </div>
            {isLoading ? (
              <div className="text-xs text-muted-foreground p-2">Loading…</div>
            ) : hooks.length === 0 ? (
              <div className="text-xs text-muted-foreground p-2">No hooks yet.</div>
            ) : (
              <ul className="space-y-1">
                {hooks.map((h) => (
                  <li
                    key={h.id}
                    data-testid="hook-row"
                    className="flex items-center justify-between text-sm rounded border border-border px-2 py-1.5"
                  >
                    <div className="min-w-0">
                      <div className="truncate">{h.name}</div>
                      <div className="text-[11px] text-muted-foreground">
                        {h.useCount} call{h.useCount === 1 ? "" : "s"} ·
                        {h.isActive ? " active" : " revoked"}
                      </div>
                    </div>
                    {h.isActive && (
                      <Button
                        size="icon"
                        variant="ghost"
                        onClick={() => void revoke.mutate(h.id)}
                        aria-label="Revoke"
                      >
                        <X className="h-3.5 w-3.5" />
                      </Button>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
