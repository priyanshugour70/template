"use client";

import { Hash, Lock, MessageSquare, Plus } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui";
import { useConversations, useCreateChannel } from "@/hooks/communication/useCommunication";
import { cn } from "@/lib/cn";

import { NewChannelDialog } from "./NewChannelDialog";

// ChannelSidebar is the left rail of the comm page. Lists channels the user
// is a member of; opens a dialog to create a new one.
export function ChannelSidebar() {
  const pathname = usePathname();
  const { data: conversations = [], isLoading } = useConversations();
  const createChannel = useCreateChannel();
  const [dialogOpen, setDialogOpen] = useState(false);

  const channels = conversations.filter((c) => c.type === "channel");
  const dms = conversations.filter((c) => c.type === "dm");

  return (
    <aside className="w-64 shrink-0 border-r border-border bg-sidebar h-full flex flex-col">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h2 className="text-sm font-semibold">Channels</h2>
        <Button
          size="icon"
          variant="ghost"
          className="h-6 w-6"
          onClick={() => setDialogOpen(true)}
          data-testid="new-channel-trigger"
          aria-label="Create channel"
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto py-2">
        {isLoading ? (
          <div className="px-4 py-2 text-xs text-muted-foreground">Loading…</div>
        ) : channels.length === 0 ? (
          <div className="px-4 py-2 text-xs text-muted-foreground">
            No channels yet. Create one to start chatting.
          </div>
        ) : (
          <nav className="px-2 space-y-0.5" aria-label="Channels">
            {channels.map((c) => {
              const href = `/dashboard/communication/${c.id}`;
              const active = pathname === href;
              return (
                <Link
                  key={c.id}
                  href={href}
                  data-testid={`channel-link-${c.slug}`}
                  className={cn(
                    "flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent",
                    active && "bg-accent text-accent-foreground font-medium",
                  )}
                >
                  {c.isPrivate ? (
                    <Lock className="h-3.5 w-3.5 text-muted-foreground" />
                  ) : (
                    <Hash className="h-3.5 w-3.5 text-muted-foreground" />
                  )}
                  <span className="truncate">{c.slug}</span>
                </Link>
              );
            })}
          </nav>
        )}

        {dms.length > 0 && (
          <>
            <div className="px-4 mt-4 mb-1 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
              Direct messages
            </div>
            <nav className="px-2 space-y-0.5" aria-label="Direct messages">
              {dms.map((c) => {
                const href = `/dashboard/communication/${c.id}`;
                const active = pathname === href;
                return (
                  <Link
                    key={c.id}
                    href={href}
                    className={cn(
                      "flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent",
                      active && "bg-accent text-accent-foreground font-medium",
                    )}
                  >
                    <MessageSquare className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="truncate">{c.name ?? "DM"}</span>
                  </Link>
                );
              })}
            </nav>
          </>
        )}
      </div>

      <NewChannelDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onCreate={async (slug, name) => {
          const out = await createChannel.mutateAsync({ slug, name });
          setDialogOpen(false);
          if (typeof window !== "undefined") {
            window.location.assign(`/dashboard/communication/${out.id}`);
          }
        }}
        pending={createChannel.isPending}
      />
    </aside>
  );
}
