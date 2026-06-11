"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";

import { communicationService } from "@/services/communication";
import type {
  ChannelHook,
  Conversation,
  ConversationView,
  CreateChannelRequest,
  CreateHookRequest,
  CreateHookResponse,
  MessageView,
  SendMessageRequest,
  ServerFrame,
} from "@/types/communication";

import { useConversationStream } from "./useCommSocket";

const ROOT = "communication" as const;
const KEY = {
  conversations: (type?: "channel" | "dm") => [ROOT, "conversations", type ?? "_all"] as const,
  conversation: (id: string) => [ROOT, "conversation", id] as const,
  messages: (id: string) => [ROOT, "messages", id] as const,
  hooks: (id: string) => [ROOT, "hooks", id] as const,
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

export type { ChannelHook, Conversation, ConversationView, MessageView };
