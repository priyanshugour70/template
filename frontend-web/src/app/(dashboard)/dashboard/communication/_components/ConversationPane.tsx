"use client";

import { Hash, Lock, Webhook } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui";
import {
  useConversation,
  useLiveMessages,
  useTypingUsers,
} from "@/hooks/communication/useCommunication";

import { Composer } from "./Composer";
import { HooksDrawer } from "./HooksDrawer";
import { MessageList } from "./MessageList";
import { TypingIndicator } from "./TypingIndicator";

interface Props {
  conversationId: string;
}

export function ConversationPane({ conversationId }: Props) {
  const { data: conv } = useConversation(conversationId);
  const { data: messages = [], isLoading } = useLiveMessages(conversationId);
  const typingUserIds = useTypingUsers(conversationId);
  const [hooksOpen, setHooksOpen] = useState(false);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-background" data-testid="conversation-pane">
      <header
        className="border-b border-border px-6 py-3 flex items-center justify-between"
        data-testid="conversation-header"
      >
        <div className="flex items-center gap-2 min-w-0">
          {conv?.isPrivate ? (
            <Lock className="h-4 w-4 text-muted-foreground" />
          ) : (
            <Hash className="h-4 w-4 text-muted-foreground" />
          )}
          <h1 className="text-base font-semibold truncate" data-testid="conversation-title">
            {conv?.name ?? conv?.slug ?? "Channel"}
          </h1>
          {conv?.topic && (
            <span className="text-xs text-muted-foreground truncate ml-2">— {conv.topic}</span>
          )}
        </div>
        {conv?.type === "channel" && (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => setHooksOpen(true)}
            data-testid="hooks-trigger"
          >
            <Webhook className="h-4 w-4 mr-1" />
            Hooks
          </Button>
        )}
      </header>

      <MessageList messages={messages} loading={isLoading} />
      <TypingIndicator typingUserIds={typingUserIds} />
      <Composer conversationId={conversationId} />

      {conv?.type === "channel" && (
        <HooksDrawer
          open={hooksOpen}
          onOpenChange={setHooksOpen}
          conversationId={conversationId}
        />
      )}
    </main>
  );
}
