"use client";

import { useEffect } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { useCommSocket } from "./useCommSocket";

// usePresence — maintains a Map<userId, "online" | "offline"> in React-Query
// cache. The backend pushes a presence frame on every online/offline
// transition (filtered to org membership). We don't need a query — only
// listener setup — so the queryFn is a no-op and we mutate via setQueryData.

const PRESENCE_KEY = ["communication", "presence"] as const;

export function usePresenceTracker() {
  const { socket } = useCommSocket();
  const qc = useQueryClient();

  useEffect(() => {
    const unsub = socket.onFrame((frame) => {
      if (frame.type !== "presence" || !frame.userId || !frame.status) return;
      qc.setQueryData<Record<string, string>>(PRESENCE_KEY, (prev) => ({
        ...(prev ?? {}),
        [frame.userId!]: frame.status!,
      }));
    });
    return unsub;
  }, [socket, qc]);
}

export function useIsOnline(userId: string | undefined): boolean {
  const { data } = useQuery<Record<string, string>>({
    queryKey: PRESENCE_KEY,
    initialData: {},
    queryFn: () => ({}),
    staleTime: Infinity,
  });
  if (!userId) return false;
  return data?.[userId] === "online";
}
