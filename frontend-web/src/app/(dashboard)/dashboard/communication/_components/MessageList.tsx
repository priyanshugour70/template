"use client";

import { useEffect, useRef } from "react";

import { Avatar, AvatarFallback } from "@/components/ui";
import { useMarkRead } from "@/hooks/communication/useCommunication";
import type { MessageView } from "@/types/communication";

import { ReactionPanel } from "./ReactionPanel";

interface Props {
  conversationId: string;
  messages: MessageView[];
  loading?: boolean;
}

function initials(name?: string): string {
  if (!name) return "?";
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((s) => s[0]?.toUpperCase() ?? "")
    .join("");
}

function fmtTime(iso?: string): string {
  if (!iso) return "";
  try {
    return new Date(iso).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

export function MessageList({ conversationId, messages, loading }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const markRead = useMarkRead(conversationId);
  const lastMarkedRef = useRef<string | null>(null);

  // Scroll-to-bottom on append. Cheap enough for Phase 4; virtualised list
  // lands in Phase 5 when we tackle long histories.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages.length]);

  // Auto mark-read: every time we render with a new last-message and the
  // bottom sentinel is in the viewport, POST /read. We dedupe by message id
  // so a re-render doesn't double-call.
  useEffect(() => {
    if (messages.length === 0 || !conversationId) return;
    const last = messages[messages.length - 1];
    if (!last || last.id === lastMarkedRef.current) return;
    const el = bottomRef.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            lastMarkedRef.current = last.id;
            markRead.mutate(last.id);
            obs.disconnect();
          }
        }
      },
      { root: null, threshold: 0.1 },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [messages, conversationId, markRead]);

  if (loading) {
    return <div className="flex-1 p-6 text-sm text-muted-foreground">Loading messages…</div>;
  }
  if (messages.length === 0) {
    return (
      <div className="flex-1 p-6 text-sm text-muted-foreground">
        No messages yet. Be the first to say hello.
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3" data-testid="message-list">
      {messages.map((m) => {
        const name =
          m.senderDisplayName ?? (m.senderType === "webhook" ? "Webhook" : "User");
        return (
          <article
            key={m.id}
            className="flex gap-3 group/message"
            data-testid="message-row"
            data-message-id={m.id}
          >
            <Avatar className="h-8 w-8 mt-0.5">
              <AvatarFallback className="text-[10px]">{initials(name)}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-baseline gap-2">
                <span className="text-sm font-medium">{name}</span>
                <span className="text-[11px] text-muted-foreground">{fmtTime(m.createdAt)}</span>
                {m.editedAt && (
                  <span className="text-[11px] text-muted-foreground italic">(edited)</span>
                )}
                {m.senderType === "webhook" && (
                  <span className="text-[10px] text-muted-foreground bg-muted px-1 rounded">
                    APP
                  </span>
                )}
              </div>
              <div className="text-sm whitespace-pre-wrap break-words">{m.body}</div>
              <div className="mt-1">
                <ReactionPanel message={m} />
              </div>
            </div>
          </article>
        );
      })}
      <div ref={bottomRef} data-testid="message-list-bottom" />
    </div>
  );
}
