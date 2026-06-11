import { api } from "@/lib/client";
import type {
  ChannelHook,
  Conversation,
  ConversationView,
  CreateChannelRequest,
  CreateHookRequest,
  CreateHookResponse,
  MarkReadRequest,
  MessageView,
  SendMessageRequest,
  WSTicketResponse,
} from "@/types/communication";

// Every comm endpoint lives under /api/v1/comm/*. Cursor-pagination is used
// for the message list (before/after = message id); offset for everything else.

export const communicationService = {
  // conversations
  listConversations: (params?: { type?: "channel" | "dm"; includeArchived?: boolean; limit?: number }) =>
    api.get<Conversation[]>("/comm/conversations", { query: { limit: 50, ...params } }),
  getConversation: (id: string) => api.get<ConversationView>(`/comm/conversations/${id}`),
  createChannel: (req: CreateChannelRequest) =>
    api.post<Conversation>("/comm/conversations/channels", req),
  createOrGetDM: (recipientUserId: string) =>
    api.post<Conversation>("/comm/conversations/dms", { recipientUserId }),

  // messages
  listMessages: (
    conversationId: string,
    params?: { before?: string; after?: string; limit?: number },
  ) =>
    api.get<MessageView[]>(`/comm/conversations/${conversationId}/messages`, {
      query: { limit: 50, ...params },
    }),
  sendMessage: (conversationId: string, req: SendMessageRequest) =>
    api.post<MessageView>(`/comm/conversations/${conversationId}/messages`, req),
  markRead: (conversationId: string, req: MarkReadRequest) =>
    api.post<unknown>(`/comm/conversations/${conversationId}/read`, req),

  // inbound hooks
  listHooks: (conversationId: string) =>
    api.get<ChannelHook[]>(`/comm/conversations/${conversationId}/hooks`),
  createHook: (conversationId: string, req: CreateHookRequest) =>
    api.post<CreateHookResponse>(`/comm/conversations/${conversationId}/hooks`, req),
  revokeHook: (hookId: string) => api.delete<unknown>(`/comm/hooks/${hookId}`),

  // ws ticket — single-use, 60s TTL
  issueWSTicket: () => api.post<WSTicketResponse>("/comm/ws/ticket"),
};
