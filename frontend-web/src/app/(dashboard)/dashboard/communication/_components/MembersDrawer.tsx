"use client";

import { LogOut, Search, UserMinus, UserPlus } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  Avatar,
  AvatarFallback,
  Button,
  Input,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui";
import {
  useAddMembers,
  useMembers,
  useRemoveMember,
} from "@/hooks/communication/useCommunication";
import { useAuth } from "@/providers";
import { userService } from "@/services/user";
import type { ConversationView } from "@/types/communication";
import type { UserProfile } from "@/types/user";

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;
  conversation: ConversationView;
  canManage: boolean;
}

function initials(name?: string, email?: string): string {
  const src = (name ?? email ?? "?").trim();
  return src
    .split(/[\s@.]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((s) => s[0]?.toUpperCase() ?? "")
    .join("");
}

// MembersDrawer — list members, add (channels only), remove (channels +
// manage permission), leave conversation. The "leave" action calls the
// same DELETE /members/:userId path but with the caller's own id.
export function MembersDrawer({ open, onOpenChange, conversation, canManage }: Props) {
  const { user } = useAuth();
  const { data: members = [], isLoading } = useMembers(conversation.id);
  const addMembers = useAddMembers(conversation.id);
  const removeMember = useRemoveMember(conversation.id);

  const [tab, setTab] = useState<"list" | "add">("list");
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [query, setQuery] = useState("");
  const [loadingUsers, setLoadingUsers] = useState(false);

  useEffect(() => {
    if (!open) return;
    setTab("list");
    setQuery("");
  }, [open, conversation.id]);

  useEffect(() => {
    if (!open || tab !== "add") return;
    setLoadingUsers(true);
    userService
      .list({ limit: 100 })
      .then((res) => setUsers(res.success ? res.data ?? [] : []))
      .finally(() => setLoadingUsers(false));
  }, [open, tab]);

  const existingIds = useMemo(() => new Set(members.map((m) => m.userId)), [members]);
  const addCandidates = useMemo(() => {
    const q = query.trim().toLowerCase();
    return users.filter((u) => {
      if (existingIds.has(u.id)) return false;
      if (!q) return true;
      return (
        u.email.toLowerCase().includes(q) ||
        (u.displayName ?? "").toLowerCase().includes(q)
      );
    });
  }, [users, query, existingIds]);

  const isChannel = conversation.type === "channel";
  const myId = user?.id;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-md">
        <div className="space-y-1.5">
          <SheetTitle>{isChannel ? "Channel members" : "Conversation"}</SheetTitle>
          <SheetDescription>
            {isChannel
              ? "Manage who has access to this channel."
              : "Direct message — both participants are listed below."}
          </SheetDescription>
        </div>

        {isChannel && canManage && (
          <div className="flex gap-1 mt-4 rounded-md bg-muted p-0.5" data-testid="members-tabs">
            <Button
              variant={tab === "list" ? "default" : "ghost"}
              size="sm"
              className="flex-1"
              onClick={() => setTab("list")}
              data-testid="members-tab-list"
            >
              Members ({members.length})
            </Button>
            <Button
              variant={tab === "add" ? "default" : "ghost"}
              size="sm"
              className="flex-1"
              onClick={() => setTab("add")}
              data-testid="members-tab-add"
            >
              <UserPlus className="h-3.5 w-3.5 mr-1" />
              Add
            </Button>
          </div>
        )}

        {tab === "list" ? (
          <div className="mt-4 space-y-1" data-testid="members-list">
            {isLoading ? (
              <div className="text-xs text-muted-foreground p-2">Loading…</div>
            ) : members.length === 0 ? (
              <div className="text-xs text-muted-foreground p-2">No members.</div>
            ) : (
              members.map((m) => {
                const isMe = m.userId === myId;
                const canRemove = isChannel && canManage && !isMe;
                return (
                  <div
                    key={m.id}
                    data-testid="member-row"
                    data-user-id={m.userId}
                    className="flex items-center gap-2 rounded border border-border px-2 py-1.5 text-sm"
                  >
                    <Avatar className="h-7 w-7">
                      <AvatarFallback className="text-[10px]">
                        {initials(m.userDisplayName, m.userEmail)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <div className="truncate">
                        {m.userDisplayName ?? m.userEmail ?? "user"}
                        {isMe && (
                          <span className="text-[10px] text-muted-foreground ml-1">(you)</span>
                        )}
                      </div>
                      <div className="text-[11px] text-muted-foreground truncate">
                        {m.role} · {m.userEmail}
                      </div>
                    </div>
                    {canRemove && (
                      <Button
                        size="icon"
                        variant="ghost"
                        onClick={() => void removeMember.mutate(m.userId)}
                        aria-label={`Remove ${m.userDisplayName ?? m.userEmail}`}
                        data-testid={`member-remove-${m.userId}`}
                      >
                        <UserMinus className="h-3.5 w-3.5" />
                      </Button>
                    )}
                  </div>
                );
              })
            )}
            {isChannel && myId && existingIds.has(myId) && (
              <Button
                size="sm"
                variant="ghost"
                className="w-full mt-2 text-destructive hover:text-destructive"
                onClick={() => {
                  void removeMember.mutateAsync(myId).then(() => onOpenChange(false));
                }}
                data-testid="member-leave"
              >
                <LogOut className="h-3.5 w-3.5 mr-1" />
                Leave channel
              </Button>
            )}
          </div>
        ) : (
          <div className="mt-4 space-y-2" data-testid="members-add-panel">
            <div className="relative">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search teammates"
                className="pl-7"
                data-testid="members-add-search"
              />
            </div>
            {loadingUsers ? (
              <div className="text-xs text-muted-foreground p-2">Loading…</div>
            ) : addCandidates.length === 0 ? (
              <div className="text-xs text-muted-foreground p-2">
                No more teammates to add.
              </div>
            ) : (
              addCandidates.slice(0, 30).map((u) => (
                <Button
                  key={u.id}
                  size="sm"
                  variant="ghost"
                  className="w-full justify-start"
                  data-testid={`members-add-${u.email}`}
                  onClick={() => void addMembers.mutate({ userIds: [u.id] })}
                  disabled={addMembers.isPending}
                >
                  <div className="flex flex-col items-start min-w-0">
                    <span className="text-sm truncate">
                      {u.displayName ??
                        (`${u.firstName ?? ""} ${u.lastName ?? ""}`.trim() || u.email)}
                    </span>
                    <span className="text-[11px] text-muted-foreground truncate">
                      {u.email}
                    </span>
                  </div>
                </Button>
              ))
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
