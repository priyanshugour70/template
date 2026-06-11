"use client";

import { SmilePlus } from "lucide-react";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui";
import {
  useAddReaction,
  useRemoveReaction,
} from "@/hooks/communication/useCommunication";
import { useAuth } from "@/providers";
import type { MessageView } from "@/types/communication";

interface Props {
  message: MessageView;
}

// Quick-reaction emoji set. A full emoji picker is Phase 5 — for now the
// fixed six covers most "yes/no/etc" use cases. Plus button is a placeholder
// for the future picker.
const QUICK = ["👍", "❤️", "😄", "🎉", "🤔", "👀"] as const;

export function ReactionPanel({ message }: Props) {
  const { user } = useAuth();
  const myId = user?.id;
  const add = useAddReaction();
  const remove = useRemoveReaction();
  const [open, setOpen] = useState(false);

  const summary = useMemo(() => {
    // Group reactions by emoji with a `mine` flag so the existing-reactions
    // strip can render distinct counts.
    const groups: Record<string, { count: number; mine: boolean }> = {};
    for (const r of message.reactions ?? []) {
      const g = groups[r.emoji] ?? { count: 0, mine: false };
      g.count += 1;
      if (r.userId === myId) g.mine = true;
      groups[r.emoji] = g;
    }
    return groups;
  }, [message.reactions, myId]);

  function toggle(emoji: string) {
    const g = summary[emoji];
    if (g?.mine) {
      void remove.mutate({ messageId: message.id, emoji });
    } else {
      void add.mutate({ messageId: message.id, emoji });
    }
  }

  const summaryKeys = Object.keys(summary);

  return (
    <div className="flex items-center gap-1 flex-wrap" data-testid="reaction-panel">
      {summaryKeys.map((emoji) => {
        const g = summary[emoji]!;
        return (
          <button
            type="button"
            key={emoji}
            onClick={() => toggle(emoji)}
            data-testid={`reaction-chip-${emoji}`}
            data-mine={g.mine}
            className={
              "inline-flex items-center gap-1 rounded-full border px-1.5 py-0.5 text-[11px] leading-none transition-colors " +
              (g.mine
                ? "border-primary/50 bg-primary/10 text-foreground"
                : "border-border bg-background hover:bg-accent text-muted-foreground")
            }
          >
            <span>{emoji}</span>
            <span>{g.count}</span>
          </button>
        );
      })}

      <div className="relative">
        <Button
          size="icon"
          variant="ghost"
          className="h-5 w-5 opacity-0 group-hover/message:opacity-100 transition-opacity"
          onClick={() => setOpen((o) => !o)}
          data-testid="reaction-add-trigger"
          aria-label="Add reaction"
        >
          <SmilePlus className="h-3.5 w-3.5" />
        </Button>
        {open && (
          <div
            className="absolute right-0 top-6 z-10 flex items-center gap-1 rounded border border-border bg-popover px-1.5 py-1 shadow"
            data-testid="reaction-quick"
            onMouseLeave={() => setOpen(false)}
          >
            {QUICK.map((emoji) => (
              <button
                key={emoji}
                type="button"
                onClick={() => {
                  toggle(emoji);
                  setOpen(false);
                }}
                data-testid={`reaction-pick-${emoji}`}
                className="text-sm leading-none px-1 hover:bg-accent rounded"
              >
                {emoji}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
