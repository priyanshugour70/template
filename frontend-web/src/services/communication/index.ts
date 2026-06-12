import { api } from "@/lib/client";
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
  MarkReadRequest,
  MessageView,
  SendMessageRequest,
  UpdateChannelRequest,
  WSTicketResponse,
} from "@/types/communication";

// Every comm endpoint lives under /api/v1/comm/*. Cursor-pagination is used
// for the message list (before/after = message id); offset for everything else.

export const communicationService = {
  // conversations
  listConversations: (params?: { type?: "channel" | "dm"; includeArchived?: boolean; limit?: number }) =>
    api.get<ConversationListItem[]>("/comm/conversations", { query: { limit: 50, ...params } }),
  getConversation: (id: string) => api.get<ConversationView>(`/comm/conversations/${id}`),
  createChannel: (req: CreateChannelRequest) =>
    api.post<Conversation>("/comm/conversations/channels", req),
  createOrGetDM: (recipientUserId: string) =>
    api.post<Conversation>("/comm/conversations/dms", { recipientUserId }),
  updateChannel: (conversationId: string, req: UpdateChannelRequest) =>
    api.patch<Conversation>(`/comm/conversations/${conversationId}`, req),
  archiveChannel: (conversationId: string) =>
    api.delete<unknown>(`/comm/conversations/${conversationId}`),

  // members
  listMembers: (conversationId: string) =>
    api.get<ConversationMemberView[]>(`/comm/conversations/${conversationId}/members`),
  addMembers: (conversationId: string, req: AddMembersRequest) =>
    api.post<ConversationMemberView[]>(
      `/comm/conversations/${conversationId}/members`,
      req,
    ),
  removeMember: (conversationId: string, userId: string) =>
    api.delete<unknown>(
      `/comm/conversations/${conversationId}/members/${userId}`,
    ),

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

  // reactions
  addReaction: (messageId: string, emoji: string) =>
    api.post<unknown>(`/comm/messages/${messageId}/reactions`, { emoji }),
  removeReaction: (messageId: string, emoji: string) =>
    api.delete<unknown>(`/comm/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`),

  // inbound hooks
  listHooks: (conversationId: string) =>
    api.get<ChannelHook[]>(`/comm/conversations/${conversationId}/hooks`),
  createHook: (conversationId: string, req: CreateHookRequest) =>
    api.post<CreateHookResponse>(`/comm/conversations/${conversationId}/hooks`, req),
  revokeHook: (hookId: string) => api.delete<unknown>(`/comm/hooks/${hookId}`),

  // ws ticket — single-use, 60s TTL
  issueWSTicket: () => api.post<WSTicketResponse>("/comm/ws/ticket"),
};
