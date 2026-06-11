"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import { createCommSocket, type CommSocket } from "@/lib/communication/ws-client";
import type { ServerFrame } from "@/types/communication";

// useCommSocket — one socket per tab. React Strict-Mode double-mount safe:
// we keep a module-level ref so a second mount in the same tab reuses the
// open socket.

let sharedSocket: CommSocket | null = null;

function getOrCreate(): CommSocket {
  if (!sharedSocket) sharedSocket = createCommSocket();
  return sharedSocket;
}

export function useCommSocket(): { socket: CommSocket; status: ReturnType<CommSocket["status"]> } {
  const socket = useMemo(getOrCreate, []);
  const [status, setStatus] = useState(socket.status());

  useEffect(() => {
    const unsub = socket.onFrame(() => setStatus(socket.status()));
    // Poll status briefly during connect — onFrame fires only after open.
    const id = setInterval(() => setStatus(socket.status()), 500);
    return () => {
      unsub();
      clearInterval(id);
    };
  }, [socket]);

  return { socket, status };
}

/**
 * Subscribe to a single conversation and run a callback for every server
 * frame scoped to it. Returns nothing — listener lifetime matches the
 * component.
 */
export function useConversationStream(
  conversationId: string | undefined,
  onFrame: (frame: ServerFrame) => void,
) {
  const { socket } = useCommSocket();
  const cbRef = useRef(onFrame);
  cbRef.current = onFrame;

  useEffect(() => {
    if (!conversationId) return;
    socket.subscribe(conversationId);
    const unsub = socket.onFrame((frame) => {
      // Some frames (presence) are global; let the consumer filter.
      cbRef.current(frame);
    });
    return () => {
      unsub();
      socket.unsubscribe(conversationId);
    };
  }, [socket, conversationId]);
}
