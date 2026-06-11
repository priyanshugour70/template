import { communicationService } from "@/services/communication";
import type { ClientFrame, ServerFrame } from "@/types/communication";

import { commWSBaseURL } from "./ws-url";

// CommSocket — a thin reconnecting WebSocket wrapper for the comm module.
// Owns ticket issuance, reconnect with exponential backoff, subscribe
// bookkeeping, and a single fan-out callback. The React hook layer adds
// per-component subscriptions on top.

type Status = "idle" | "connecting" | "open" | "closed";

export type SocketListener = (frame: ServerFrame) => void;

export interface CommSocket {
  status(): Status;
  onFrame(cb: SocketListener): () => void;
  subscribe(conversationId: string): void;
  unsubscribe(conversationId: string): void;
  typing(conversationId: string): void;
  close(): void;
}

interface InternalState {
  ws: WebSocket | null;
  status: Status;
  listeners: Set<SocketListener>;
  subscriptions: Set<string>;
  reconnectAttempts: number;
  closed: boolean;
}

/**
 * Create a CommSocket. Starts connecting immediately. Reconnects on close
 * with backoff (capped at 30s). Re-subscribes to every previously-subscribed
 * conversation after a successful reconnect.
 */
export function createCommSocket(): CommSocket {
  const state: InternalState = {
    ws: null,
    status: "idle",
    listeners: new Set(),
    subscriptions: new Set(),
    reconnectAttempts: 0,
    closed: false,
  };

  async function connect() {
    if (state.closed || state.status === "connecting" || state.status === "open") return;
    state.status = "connecting";

    const tRes = await communicationService.issueWSTicket();
    if (!tRes.success || !tRes.data?.ticket) {
      scheduleReconnect();
      return;
    }
    const url = `${commWSBaseURL()}/api/v1/comm/ws?ticket=${encodeURIComponent(tRes.data.ticket)}`;
    let ws: WebSocket;
    try {
      ws = new WebSocket(url);
    } catch {
      scheduleReconnect();
      return;
    }
    state.ws = ws;

    ws.onopen = () => {
      state.status = "open";
      state.reconnectAttempts = 0;
      // Re-establish every prior subscription on a fresh connection.
      for (const convId of state.subscriptions) {
        sendFrame(ws, { type: "subscribe", conversationId: convId });
      }
    };

    ws.onmessage = (ev) => {
      let frame: ServerFrame;
      try {
        frame = JSON.parse(typeof ev.data === "string" ? ev.data : "") as ServerFrame;
      } catch {
        return;
      }
      for (const cb of state.listeners) {
        try {
          cb(frame);
        } catch {
          // Listener errors are swallowed so a single buggy consumer can't
          // bring the whole stream down.
        }
      }
    };

    ws.onclose = () => {
      state.status = "closed";
      state.ws = null;
      if (!state.closed) scheduleReconnect();
    };

    ws.onerror = () => {
      // onclose fires right after; the reconnect path lives there.
    };
  }

  function scheduleReconnect() {
    if (state.closed) return;
    state.status = "closed";
    state.reconnectAttempts += 1;
    const delay = Math.min(1000 * 2 ** state.reconnectAttempts, 30_000);
    setTimeout(() => {
      void connect();
    }, delay);
  }

  function sendFrame(ws: WebSocket, frame: ClientFrame) {
    if (ws.readyState !== WebSocket.OPEN) return;
    try {
      ws.send(JSON.stringify(frame));
    } catch {
      // Ignored — close handler will reconnect.
    }
  }

  function sendIfOpen(frame: ClientFrame) {
    if (state.ws) sendFrame(state.ws, frame);
  }

  void connect();

  return {
    status: () => state.status,
    onFrame(cb) {
      state.listeners.add(cb);
      return () => state.listeners.delete(cb);
    },
    subscribe(conversationId) {
      state.subscriptions.add(conversationId);
      sendIfOpen({ type: "subscribe", conversationId });
    },
    unsubscribe(conversationId) {
      state.subscriptions.delete(conversationId);
      sendIfOpen({ type: "unsubscribe", conversationId });
    },
    typing(conversationId) {
      sendIfOpen({ type: "typing", conversationId });
    },
    close() {
      state.closed = true;
      if (state.ws) state.ws.close();
      state.ws = null;
      state.status = "closed";
    },
  };
}
