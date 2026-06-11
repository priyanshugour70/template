"use client";

import { Send } from "lucide-react";
import { useRef, useState } from "react";

import { Button } from "@/components/ui";
import { useCommSocket } from "@/hooks/communication/useCommSocket";
import { useSendMessage } from "@/hooks/communication/useCommunication";

interface Props {
  conversationId: string;
}

// Composer — single-line textarea, Enter sends, Shift+Enter newline. Emits
// a typing event throttled to once every 2s while the user is actively
// typing. The backend has its own throttle so this is just polite.
export function Composer({ conversationId }: Props) {
  const [body, setBody] = useState("");
  const send = useSendMessage(conversationId);
  const { socket } = useCommSocket();
  const lastTypingAt = useRef(0);

  function onChange(v: string) {
    setBody(v);
    const now = Date.now();
    if (now - lastTypingAt.current > 2_000) {
      lastTypingAt.current = now;
      socket.typing(conversationId);
    }
  }

  async function submit() {
    const trimmed = body.trim();
    if (!trimmed) return;
    try {
      await send.mutateAsync({ body: trimmed });
      setBody("");
    } catch {
      // The mutation already exposes error; toast wiring is Phase 4.
    }
  }

  return (
    <div className="border-t border-border p-3 flex items-end gap-2">
      <textarea
        data-testid="composer-input"
        className="flex-1 resize-none rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring max-h-32"
        value={body}
        rows={2}
        placeholder="Message…"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            void submit();
          }
        }}
      />
      <Button
        size="icon"
        data-testid="composer-send"
        onClick={() => void submit()}
        disabled={!body.trim() || send.isPending}
        aria-label="Send message"
      >
        <Send className="h-4 w-4" />
      </Button>
    </div>
  );
}
