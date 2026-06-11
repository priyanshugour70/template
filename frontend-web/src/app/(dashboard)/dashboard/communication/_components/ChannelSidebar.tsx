"use client";

import { Hash, Lock, MessageSquare, Plus } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { Badge, Button } from "@/components/ui";
import { useCommSocket } from "@/hooks/communication/useCommSocket";
import {
  useConversations,
  useCreateChannel,
  useCreateDM,
} from "@/hooks/communication/useCommunication";
import { usePresenceTracker } from "@/hooks/communication/usePresence";
import { cn } from "@/lib/cn";
import type { ConversationListItem } from "@/types/communication";

import { NewChannelDialog } from "./NewChannelDialog";
import { NewDMDialog } from "./NewDMDialog";

// Sidebar split into three sections:
//   1. Recent      → top 5 conversations across both types by last_message_at
//   2. Channels    → all channels alphabetically
//   3. DMs         → all DMs by last_message_at
//
// Unread badges and presence dots live alongside each row. Presence tracker
// boots once at the sidebar level because the sidebar is mounted whenever
// the comm route is open.
export function ChannelSidebar() {
  usePresenceTracker();
  const pathname = usePathname();
  const { data: conversations = [], isLoading } = useConversations();
  const createChannel = useCreateChannel();
  const createDM = useCreateDM();
  const [channelDialog, setChannelDialog] = useState(false);
  const [dmDialog, setDmDialog] = useState(false);

  // Keep the sidebar in sync with cross-conversation WS chatter so unread
  // badges and recent-order respond to messages that arrive in conversations
  // the user isn't currently looking at. Debounced via setTimeout so a burst
  // of frames doesn't trigger N invalidations.
  const qc = useQueryClient();
  const { socket } = useCommSocket();
  useEffect(() => {
    let pending: ReturnType<typeof setTimeout> | null = null;
    const schedule = () => {
      if (pending) return;
      pending = setTimeout(() => {
        pending = null;
        void qc.invalidateQueries({ queryKey: ["communication", "conversations"] });
      }, 250);
    };
    const unsub = socket.onFrame((f) => {
      if (f.type === "message.created" || f.type === "read" || f.type === "conversation.updated") {
        schedule();
      }
    });
    return () => {
      unsub();
      if (pending) clearTimeout(pending);
    };
  }, [socket, qc]);

  const channels = useMemo(
    () =>
      conversations
        .filter((c) => c.type === "channel")
        .sort((a, b) => (a.slug ?? "").localeCompare(b.slug ?? "")),
    [conversations],
  );
  const dms = useMemo(
    () => conversations.filter((c) => c.type === "dm"),
    [conversations],
  );
  const recent = useMemo(
    () =>
      [...conversations]
        .filter((c) => !!c.lastMessageAt)
        .sort((a, b) => (b.lastMessageAt ?? "").localeCompare(a.lastMessageAt ?? ""))
        .slice(0, 5),
    [conversations],
  );

  return (
    <aside
      className="w-64 shrink-0 border-r border-border bg-sidebar h-full flex flex-col"
      data-testid="comm-sidebar"
    >
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h2 className="text-sm font-semibold">Messages</h2>
        <div className="flex items-center gap-1">
          <Button
            size="icon"
            variant="ghost"
            className="h-6 w-6"
            onClick={() => setChannelDialog(true)}
            data-testid="new-channel-trigger"
            aria-label="Create channel"
            title="New channel"
          >
            <Hash className="h-3.5 w-3.5" />
          </Button>
          <Button
            size="icon"
            variant="ghost"
            className="h-6 w-6"
            onClick={() => setDmDialog(true)}
            data-testid="new-dm-trigger"
            aria-label="Start direct message"
            title="New direct message"
          >
            <Plus className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto py-2">
        {isLoading ? (
          <div className="px-4 py-2 text-xs text-muted-foreground">Loading…</div>
        ) : (
          <>
            {recent.length > 0 && (
              <Section label="Recent">
                {recent.map((c) => (
                  <ConversationRow key={c.id} conv={c} pathname={pathname} />
                ))}
              </Section>
            )}

            <Section label="Channels">
              {channels.length === 0 ? (
                <div className="px-4 py-2 text-xs text-muted-foreground">
                  No channels yet.
                </div>
              ) : (
                channels.map((c) => (
                  <ConversationRow key={c.id} conv={c} pathname={pathname} />
                ))
              )}
            </Section>

            {dms.length > 0 && (
              <Section label="Direct messages">
                {dms.map((c) => (
                  <ConversationRow key={c.id} conv={c} pathname={pathname} />
                ))}
              </Section>
            )}
          </>
        )}
      </div>

      <NewChannelDialog
        open={channelDialog}
        onOpenChange={setChannelDialog}
        onCreate={async (slug, name) => {
          const out = await createChannel.mutateAsync({ slug, name });
          setChannelDialog(false);
          if (typeof window !== "undefined") {
            window.location.assign(`/dashboard/communication/${out.id}`);
          }
        }}
        pending={createChannel.isPending}
      />

      <NewDMDialog
        open={dmDialog}
        onOpenChange={setDmDialog}
        onPick={async (userId) => {
          const out = await createDM.mutateAsync(userId);
          setDmDialog(false);
          if (typeof window !== "undefined") {
            window.location.assign(`/dashboard/communication/${out.id}`);
          }
        }}
        pending={createDM.isPending}
      />
    </aside>
  );
}

// ── helpers ───────────────────────────────────────────────────────────────

function Section({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="mb-2">
      <div className="px-4 mt-1 mb-1 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
        {label}
      </div>
      <nav className="px-2 space-y-0.5" aria-label={label}>
        {children}
      </nav>
    </div>
  );
}

function ConversationRow({
  conv,
  pathname,
}: {
  conv: ConversationListItem;
  pathname: string;
}) {
  const href = `/dashboard/communication/${conv.id}`;
  const active = pathname === href;
  const unread = conv.unreadCount ?? 0;

  return (
    <Link
      href={href}
      data-testid={`conv-link-${conv.id}`}
      data-conv-type={conv.type}
      data-unread={unread}
      className={cn(
        "flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent",
        active && "bg-accent text-accent-foreground font-medium",
        unread > 0 && !active && "font-medium",
      )}
    >
      <span className="shrink-0">
        {conv.type === "channel" ? (
          conv.isPrivate ? (
            <Lock className="h-3.5 w-3.5 text-muted-foreground" />
          ) : (
            <Hash className="h-3.5 w-3.5 text-muted-foreground" />
          )
        ) : (
          <MessageSquare className="h-3.5 w-3.5 text-muted-foreground" />
        )}
      </span>
      <span className="truncate flex-1 min-w-0">
        {conv.type === "channel" ? conv.slug : conv.name ?? "DM"}
      </span>
      {unread > 0 && (
        <Badge
          variant="default"
          className="h-4 min-w-4 px-1 text-[10px] leading-none"
          data-testid="unread-badge"
        >
          {unread > 99 ? "99+" : unread}
        </Badge>
      )}
    </Link>
  );
}
