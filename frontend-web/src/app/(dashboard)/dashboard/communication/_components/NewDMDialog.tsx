"use client";

import { Loader2, MessageSquarePlus, Search } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  Input,
} from "@/components/ui";
import { useAuth } from "@/providers";
import { userService } from "@/services/user";
import type { UserProfile } from "@/types/user";

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;
  onPick(userId: string): Promise<void> | void;
  pending: boolean;
}

// NewDMDialog — type to filter, click a user, fire onPick. Excludes self.
// Search is client-side over the first 100 users; for orgs that outgrow
// that, swap in a debounced /users?q= call. Kept simple here because Phase 1
// of comm targets small teams.
export function NewDMDialog({ open, onOpenChange, onPick, pending }: Props) {
  const { user } = useAuth();
  const [query, setQuery] = useState("");
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    userService
      .list({ limit: 100 })
      .then((res) => setUsers(res.success ? res.data ?? [] : []))
      .finally(() => setLoading(false));
  }, [open]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const myId = user?.id;
    return users
      .filter((u) => u.id !== myId)
      .filter((u) => {
        if (!q) return true;
        return (
          u.email.toLowerCase().includes(q) ||
          (u.displayName ?? "").toLowerCase().includes(q) ||
          (u.firstName ?? "").toLowerCase().includes(q) ||
          (u.lastName ?? "").toLowerCase().includes(q)
        );
      })
      .slice(0, 50);
  }, [users, query, user]);

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        onOpenChange(o);
        if (!o) setQuery("");
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <MessageSquarePlus className="h-4 w-4" />
            Start a direct message
          </DialogTitle>
          <DialogDescription>
            Pick a teammate to start a one-on-one conversation.
          </DialogDescription>
        </DialogHeader>

        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            data-testid="dm-search-input"
            className="pl-7"
            placeholder="Search by name or email"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
          />
        </div>

        <div
          className="max-h-72 overflow-y-auto -mx-2 px-2 space-y-0.5"
          data-testid="dm-user-list"
        >
          {loading ? (
            <div className="flex items-center gap-2 p-3 text-sm text-muted-foreground">
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              Loading…
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-3 text-sm text-muted-foreground">No teammates found.</div>
          ) : (
            filtered.map((u) => {
              const label =
                u.displayName ??
                [u.firstName, u.lastName].filter(Boolean).join(" ").trim() ??
                u.email;
              return (
                <Button
                  key={u.id}
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  data-testid={`dm-user-${u.email}`}
                  disabled={pending}
                  onClick={() => void onPick(u.id)}
                >
                  <div className="flex flex-col items-start min-w-0">
                    <span className="text-sm truncate">{label || u.email}</span>
                    <span className="text-[11px] text-muted-foreground truncate">
                      {u.email}
                    </span>
                  </div>
                </Button>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
