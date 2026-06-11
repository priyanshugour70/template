import type { BaseEntity, ID, ISODate, JSONObject } from "@/types/common";

// ── Conversations ─────────────────────────────────────────────────────────

export type ConversationType = "dm" | "channel";

export interface Conversation extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  type: ConversationType;
  slug?: string;
  name?: string;
  topic?: string;
  description?: string;
  isPrivate: boolean;
  archivedAt?: ISODate | null;
  lastMessageAt?: ISODate | null;
  messageCount: number;
}

export interface ConversationMemberView {
  id: ID;
  userId: ID;
  role: string;
  joinedAt: ISODate;
  lastReadMessageId?: ID | null;
  unreadCount: number;
  notificationPref: string;
  userEmail?: string;
  userDisplayName?: string;
  userAvatarUrl?: string;
}

export interface ConversationView extends Conversation {
  members?: ConversationMemberView[];
  myMembership?: ConversationMemberView | null;
}

/** Row shape returned by GET /comm/conversations — Conversation + the
 * caller's unread state, so the sidebar can render badges without a second
 * round trip. */
export interface ConversationListItem extends Conversation {
  unreadCount: number;
  lastReadMessageId?: ID | null;
}

// ── Messages ──────────────────────────────────────────────────────────────

export type MessageSenderType = "user" | "system" | "webhook";

export interface MessageMention {
  id: ID;
  messageId: ID;
  mentionType: "user" | "here" | "channel" | "everyone";
  targetUserId?: ID | null;
  indexInBody: number;
  createdAt: ISODate;
}

export interface MessageReaction {
  id: ID;
  messageId: ID;
  userId: ID;
  emoji: string;
  createdAt: ISODate;
}

export interface MessageView extends BaseEntity {
  conversationId: ID;
  tenantId: ID;
  organizationId: ID;
  parentMessageId?: ID | null;
  senderType: MessageSenderType;
  senderUserId?: ID | null;
  senderWebhookId?: ID | null;
  body: string;
  bodyFormat: "markdown" | "plain";
  attachments?: JSONObject[];
  metadata?: JSONObject;
  editedAt?: ISODate | null;
  mentions?: MessageMention[];
  reactions?: MessageReaction[];
  senderDisplayName?: string;
  senderAvatarUrl?: string;
}

// ── Requests ──────────────────────────────────────────────────────────────

export interface CreateChannelRequest {
  slug: string;
  name: string;
  topic?: string;
  description?: string;
  isPrivate?: boolean;
  memberIds?: ID[];
}

export interface SendMessageRequest {
  body: string;
  bodyFormat?: "markdown" | "plain";
  parentMessageId?: ID;
  attachments?: JSONObject[];
}

export interface MarkReadRequest {
  lastReadMessageId: ID;
}

// ── Hooks (channel inbound webhooks) ──────────────────────────────────────

export interface ChannelHook extends BaseEntity {
  tenantId: ID;
  organizationId: ID;
  conversationId: ID;
  name: string;
  iconUrl?: string;
  displayName?: string;
  isActive: boolean;
  lastUsedAt?: ISODate | null;
  useCount: number;
}

export interface CreateHookRequest {
  name: string;
  displayName?: string;
  iconUrl?: string;
}

export interface CreateHookResponse {
  hook: ChannelHook;
  token: string;
  url: string;
}

// ── WS protocol (mirrors backend internal/modules/comm/ws/protocol.go) ────

export type ServerFrameType =
  | "hello"
  | "pong"
  | "error"
  | "message.created"
  | "message.updated"
  | "message.deleted"
  | "reaction.added"
  | "reaction.removed"
  | "typing"
  | "presence"
  | "read"
  | "member.added"
  | "member.removed"
  | "conversation.updated";

export interface ServerFrame {
  type: ServerFrameType;
  conversationId?: ID;
  userId?: ID;
  messageId?: ID;
  message?: MessageView;
  conversation?: Conversation;
  member?: ConversationMemberView;
  emoji?: string;
  untilMs?: number;
  status?: "online" | "away" | "offline";
  reason?: string;
  error?: string;
}

export interface ClientFrame {
  type: "subscribe" | "unsubscribe" | "typing" | "ping";
  conversationId?: ID;
}

export interface WSTicketResponse {
  ticket: string;
  expiresAt: ISODate;
}
