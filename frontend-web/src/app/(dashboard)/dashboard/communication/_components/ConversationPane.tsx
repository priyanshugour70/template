"use client";

import {
  Hash,
  Lock,
  MessageSquare,
  Pencil,
  Settings,
  Users,
  Webhook,
} from "lucide-react";
import { useState } from "react";

import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui";
import {
  useConversation,
  useLiveMessages,
  useTypingUsers,
} from "@/hooks/communication/useCommunication";

import { Composer } from "./Composer";
import { EditChannelDialog } from "./EditChannelDialog";
import { HooksDrawer } from "./HooksDrawer";
import { MembersDrawer } from "./MembersDrawer";
import { MessageList } from "./MessageList";
import { TypingIndicator } from "./TypingIndicator";

interface Props {
  conversationId: string;
}

function titleFor(conv: { type: string; name?: string; slug?: string } | undefined) {
  if (!conv) return "";
  if (conv.type === "channel") return conv.name ?? conv.slug ?? "channel";
  return conv.name ?? "Direct message";
}

export function ConversationPane({ conversationId }: Props) {
  const { data: conv, isLoading: convLoading } = useConversation(conversationId);
  const { data: messages = [], isLoading: messagesLoading } = useLiveMessages(conversationId);
  const typingUserIds = useTypingUsers(conversationId);
  const [hooksOpen, setHooksOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [membersOpen, setMembersOpen] = useState(false);

  const canManage =
    conv?.type === "channel" &&
    (conv.myMembership?.role === "owner" || conv.myMembership?.role === "admin");

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-background" data-testid="conversation-pane">
      <header
        className="border-b border-border px-6 py-3 flex items-center justify-between gap-3"
        data-testid="conversation-header"
      >
        <div className="flex items-center gap-2 min-w-0">
          {conv?.type === "dm" ? (
            <MessageSquare className="h-4 w-4 text-muted-foreground shrink-0" />
          ) : conv?.isPrivate ? (
            <Lock className="h-4 w-4 text-muted-foreground shrink-0" />
          ) : (
            <Hash className="h-4 w-4 text-muted-foreground shrink-0" />
          )}
          <h1
            className="text-base font-semibold truncate"
            data-testid="conversation-title"
          >
            {convLoading && !conv ? "Loading…" : titleFor(conv)}
          </h1>
          {conv?.topic && (
            <span className="text-xs text-muted-foreground truncate ml-2">
              — {conv.topic}
            </span>
          )}
        </div>

        <div className="flex items-center gap-1 shrink-0">
          {conv && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setMembersOpen(true)}
              data-testid="members-trigger"
              title={conv.type === "channel" ? "Manage members" : "View participants"}
            >
              <Users className="h-4 w-4 mr-1" />
              {conv.members?.length ?? 0}
            </Button>
          )}

          {conv?.type === "channel" && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setHooksOpen(true)}
              data-testid="hooks-trigger"
              title="Inbound webhooks"
            >
              <Webhook className="h-4 w-4 mr-1" />
              Hooks
            </Button>
          )}

          {conv?.type === "channel" && canManage && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="icon"
                  variant="ghost"
                  data-testid="channel-settings-trigger"
                  aria-label="Channel settings"
                >
                  <Settings className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  onSelect={() => setEditOpen(true)}
                  data-testid="channel-settings-edit"
                >
                  <Pencil className="h-3.5 w-3.5 mr-2" />
                  Edit channel
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onSelect={() => setMembersOpen(true)}
                  data-testid="channel-settings-members"
                >
                  <Users className="h-3.5 w-3.5 mr-2" />
                  Manage members
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </header>

      <MessageList
        conversationId={conversationId}
        messages={messages}
        loading={messagesLoading}
      />
      <TypingIndicator typingUserIds={typingUserIds} />
      <Composer conversationId={conversationId} />

      {conv?.type === "channel" && (
        <>
          <HooksDrawer
            open={hooksOpen}
            onOpenChange={setHooksOpen}
            conversationId={conversationId}
          />
          <EditChannelDialog
            open={editOpen}
            onOpenChange={setEditOpen}
            conversation={conv}
          />
        </>
      )}

      {conv && (
        <MembersDrawer
          open={membersOpen}
          onOpenChange={setMembersOpen}
          conversation={conv}
          canManage={Boolean(canManage)}
        />
      )}
    </main>
  );
}
