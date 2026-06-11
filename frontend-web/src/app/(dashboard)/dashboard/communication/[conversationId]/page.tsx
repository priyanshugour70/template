"use client";

import { use } from "react";

import { ConversationPane } from "../_components/ConversationPane";

// In Next.js 16, dynamic route params are exposed as a Promise that the
// client component unwraps via React's `use` hook. Server-only props would
// be awaited; for a client page we use `use()` to keep the call concise.
export default function ConversationPage({
  params,
}: {
  params: Promise<{ conversationId: string }>;
}) {
  const { conversationId } = use(params);
  return <ConversationPane conversationId={conversationId} />;
}
