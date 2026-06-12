"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";

import { communicationService } from "@/services/communication";
import type {
  AddMembersRequest,
  ChannelHook,
  Conversation,
  ConversationListItem,
  ConversationMemberView,
  ConversationView,
  CreateChannelRequest,
  CreateHookRequest,
  CreateHookResponse,
  MessageView,
  SendMessageRequest,
  ServerFrame,
  UpdateChannelRequest,
} from "@/types/communication";

import { useConversationStream } from "./useCommSocket";

const ROOT = "communication" as const;
const KEY = {
  conversations: (type?: "channel" | "dm") => [ROOT, "conversations", type ?? "_all"] as const,
  conversation: (id: string) => [ROOT, "conversation", id] as const,
  messages: (id: string) => [ROOT, "messages", id] as const,
  hooks: (id: string) => [ROOT, "hooks", id] as const,
  members: (id: string) => [ROOT, "members", id] as const,
};

// ── queries ────────────────────────────────────────────────────────────────

export function useConversations(type?: "channel" | "dm") {
  return useQuery({
    queryKey: KEY.conversations(type),
    queryFn: async () => {
      const res = await communicationService.listConversations({ type, limit: 100 });
      if (!res.success) throw new Error(res.error?.message ?? "list conversations failed");
      return res.data ?? [];
    },
    staleTime: 30_000,
  });
}

export function useConversation(id: string | undefined) {
  return useQuery({
    queryKey: id ? KEY.conversation(id) : [ROOT, "conversation", "_none"],
    enabled: Boolean(id),
    queryFn: async () => {
      const res = await communicationService.getConversation(id!);
      if (!res.success) throw new Error(res.error?.message ?? "load conversation failed");
      return res.data!;
    },
  });
}

export function useMessages(conversationId: string | undefined) {
  return useQuery({
    queryKey: conversationId ? KEY.messages(conversationId) : [ROOT, "messages", "_none"],
    enabled: Boolean(conversationId),
    queryFn: async () => {
      const res = await communicationService.listMessages(conversationId!, { limit: 50 });
      if (!res.success) throw new Error(res.error?.message ?? "list messages failed");
      // Backend returns newest-first; UI wants oldest-first chronological.
      return (res.data ?? []).slice().reverse();
    },
    staleTime: 5_000,
  });
}

export function useChannelHooks(conversationId: string | undefined) {
  return useQuery({
    queryKey: conversationId ? KEY.hooks(conversationId) : [ROOT, "hooks", "_none"],
    enabled: Boolean(conversationId),
    queryFn: async () => {
      const res = await communicationService.listHooks(conversationId!);
      if (!res.success) throw new Error(res.error?.message ?? "list hooks failed");
      return res.data ?? [];
    },
    staleTime: 60_000,
  });
}

// ── mutations ─────────────────────────────────────────────────────────────

export function useCreateChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateChannelRequest) => {
      const res = await communicationService.createChannel(req);
      if (!res.success) throw new Error(res.error?.message ?? "create channel failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT, "conversations"] });
    },
  });
}

export function useUpdateChannel(conversationId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: UpdateChannelRequest) => {
      const res = await communicationService.updateChannel(conversationId, req);
      if (!res.success) throw new Error(res.error?.message ?? "update channel failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.conversation(conversationId) });
      void qc.invalidateQueries({ queryKey: [ROOT, "conversations"] });
    },
  });
}

export function useArchiveChannel(conversationId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await communicationService.archiveChannel(conversationId);
      if (!res.success) throw new Error(res.error?.message ?? "archive failed");
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT, "conversations"] });
    },
  });
}

export function useMembers(conversationId: string | undefined) {
  return useQuery<ConversationMemberView[]>({
    queryKey: conversationId ? KEY.members(conversationId) : [ROOT, "members", "_none"],
    enabled: Boolean(conversationId),
    queryFn: async () => {
      const res = await communicationService.listMembers(conversationId!);
      if (!res.success) throw new Error(res.error?.message ?? "load members failed");
      return res.data ?? [];
    },
  });
}

export function useAddMembers(conversationId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: AddMembersRequest) => {
      const res = await communicationService.addMembers(conversationId, req);
      if (!res.success) throw new Error(res.error?.message ?? "add members failed");
      return res.data ?? [];
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.members(conversationId) });
      void qc.invalidateQueries({ queryKey: KEY.conversation(conversationId) });
    },
  });
}

export function useRemoveMember(conversationId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: string) => {
      const res = await communicationService.removeMember(conversationId, userId);
      if (!res.success) throw new Error(res.error?.message ?? "remove member failed");
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.members(conversationId) });
      void qc.invalidateQueries({ queryKey: KEY.conversation(conversationId) });
    },
  });
}

export function useCreateDM() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (recipientUserId: string) => {
      const res = await communicationService.createOrGetDM(recipientUserId);
      if (!res.success) throw new Error(res.error?.message ?? "create DM failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT, "conversations"] });
    },
  });
}

export function useAddReaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: { messageId: string; emoji: string }) => {
      const res = await communicationService.addReaction(input.messageId, input.emoji);
      if (!res.success) throw new Error(res.error?.message ?? "add reaction failed");
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT, "messages"] });
    },
  });
}

export function useRemoveReaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: { messageId: string; emoji: string }) => {
      const res = await communicationService.removeReaction(input.messageId, input.emoji);
      if (!res.success) throw new Error(res.error?.message ?? "remove reaction failed");
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT, "messages"] });
    },
  });
}

export function useMarkRead(conversationId: string | undefined) {
  return useMutation({
    mutationFn: async (lastReadMessageId: string) => {
      if (!conversationId) return;
      const res = await communicationService.markRead(conversationId, { lastReadMessageId });
      if (!res.success) throw new Error(res.error?.message ?? "mark read failed");
      return res.data;
    },
  });
}

export function useSendMessage(conversationId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: SendMessageRequest) => {
      const res = await communicationService.sendMessage(conversationId, req);
      if (!res.success) throw new Error(res.error?.message ?? "send failed");
      return res.data!;
    },
    onSuccess: (msg) => {
      // Optimistically insert into cache. If WS arrives afterwards with the
      // same id, useLiveMessages dedupes. This makes the message appear
      // immediately even when WS is still negotiating its handshake.
      qc.setQueryData<MessageView[]>(KEY.messages(conversationId), (prev) => {
        const list = prev ?? [];
        if (list.some((m) => m.id === msg.id)) return list;
        return [...list, msg];
      });
    },
  });
}

export function useCreateHook(conversationId: string) {
  const qc = useQueryClient();
  return useMutation<CreateHookResponse, Error, CreateHookRequest>({
    mutationFn: async (req) => {
      const res = await communicationService.createHook(conversationId, req);
      if (!res.success) throw new Error(res.error?.message ?? "create hook failed");
      return res.data!;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.hooks(conversationId) });
    },
  });
}

export function useRevokeHook(conversationId: string) {
  const qc = useQueryClient();
  return useMutation<unknown, Error, string>({
    mutationFn: async (hookId) => {
      const res = await communicationService.revokeHook(hookId);
      if (!res.success) throw new Error(res.error?.message ?? "revoke hook failed");
      return res.data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY.hooks(conversationId) });
    },
  });
}

// ── live-event glue ───────────────────────────────────────────────────────

/**
 * Subscribe to a conversation's WS stream and mutate the React-Query cache
 * for every relevant frame. Returns the live message list so the consumer
 * can render directly. Designed for the conversation pane.
 */
export function useLiveMessages(conversationId: string | undefined) {
  const qc = useQueryClient();
  const query = useMessages(conversationId);

  useConversationStream(conversationId, (frame: ServerFrame) => {
    if (!conversationId || frame.conversationId !== conversationId) return;
    if (frame.type === "message.created" && frame.message) {
      qc.setQueryData<MessageView[]>(KEY.messages(conversationId), (prev) => {
        const list = prev ?? [];
        // Idempotent: REST send → WS may double-deliver.
        if (list.some((m) => m.id === frame.message!.id)) return list;
        return [...list, frame.message!];
      });
    } else if (frame.type === "message.updated" && frame.message) {
      qc.setQueryData<MessageView[]>(KEY.messages(conversationId), (prev) =>
        (prev ?? []).map((m) => (m.id === frame.message!.id ? frame.message! : m)),
      );
    } else if (frame.type === "message.deleted" && frame.messageId) {
      qc.setQueryData<MessageView[]>(KEY.messages(conversationId), (prev) =>
        (prev ?? []).filter((m) => m.id !== frame.messageId),
      );
    } else if (
      (frame.type === "reaction.added" || frame.type === "reaction.removed") &&
      frame.messageId &&
      frame.emoji
    ) {
      qc.setQueryData<MessageView[]>(KEY.messages(conversationId), (prev) =>
        (prev ?? []).map((m) => {
          if (m.id !== frame.messageId) return m;
          const reactions = m.reactions ?? [];
          if (frame.type === "reaction.added") {
            const exists = reactions.some(
              (r) => r.userId === frame.userId && r.emoji === frame.emoji,
            );
            if (exists) return m;
            return {
              ...m,
              reactions: [
                ...reactions,
                {
                  id: `${frame.messageId}-${frame.userId}-${frame.emoji}`,
                  messageId: frame.messageId!,
                  userId: frame.userId!,
                  emoji: frame.emoji!,
                  createdAt: new Date().toISOString(),
                },
              ],
            };
          }
          return {
            ...m,
            reactions: reactions.filter(
              (r) => !(r.userId === frame.userId && r.emoji === frame.emoji),
            ),
          };
        }),
      );
    }
  });

  return query;
}

/**
 * Subscribe to a conversation's WS stream just for `typing` events. Returns
 * a Set of currently-typing user ids; entries expire automatically.
 */
export function useTypingUsers(conversationId: string | undefined) {
  // Plain Set kept in a stable ref via state — re-render on add/expire.
  const qc = useQueryClient();
  const key = ["communication", "typing", conversationId ?? "_none"] as const;
  const { data } = useQuery<Record<string, number>>({
    queryKey: key,
    initialData: {},
    queryFn: () => ({}),
    staleTime: Infinity,
  });

  useConversationStream(conversationId, (frame: ServerFrame) => {
    if (!conversationId || frame.conversationId !== conversationId) return;
    if (frame.type !== "typing" || !frame.userId) return;
    qc.setQueryData<Record<string, number>>(key, (prev) => ({
      ...(prev ?? {}),
      [frame.userId!]: frame.untilMs ?? Date.now() + 5_000,
    }));
  });

  // Expire entries that are past untilMs.
  useEffect(() => {
    if (!conversationId) return;
    const id = setInterval(() => {
      const now = Date.now();
      qc.setQueryData<Record<string, number>>(key, (prev) => {
        if (!prev) return prev;
        let changed = false;
        const next: Record<string, number> = {};
        for (const [uid, until] of Object.entries(prev)) {
          if (until > now) next[uid] = until;
          else changed = true;
        }
        return changed ? next : prev;
      });
    }, 1_000);
    return () => clearInterval(id);
  }, [qc, conversationId, key]);

  return Object.keys(data ?? {});
}

export type {
  ChannelHook,
  Conversation,
  ConversationListItem,
  ConversationView,
  MessageView,
};
