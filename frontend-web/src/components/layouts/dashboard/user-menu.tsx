"use client";

import { LogOut, User as UserIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/providers";

export function UserMenu() {
  const { user, logout } = useAuth();
  const [open, setOpen] = useState(false);

  if (!user) return null;

  const initials =
    (user.firstName?.[0] ?? user.email?.[0] ?? "?").toUpperCase() +
    (user.lastName?.[0] ?? "").toUpperCase();

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2 rounded-full hover:bg-accent p-1 transition-colors"
      >
        <Avatar>
          {user.avatarUrl ? (
            <AvatarImage src={user.avatarUrl} alt={user.displayName ?? user.email} />
          ) : null}
          <AvatarFallback>{initials}</AvatarFallback>
        </Avatar>
      </button>
      {open && (
        <div className="absolute right-0 mt-2 w-56 rounded-md border bg-popover shadow-md z-50">
          <div className="p-3 border-b">
            <div className="text-sm font-medium truncate">{user.displayName ?? user.email}</div>
            <div className="text-xs text-muted-foreground truncate">{user.email}</div>
          </div>
          <div className="p-1">
            <Link
              href="/dashboard/settings"
              className="flex items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent"
              onClick={() => setOpen(false)}
            >
              <UserIcon className="h-4 w-4" />
              Profile & settings
            </Link>
            <Button
              variant="ghost"
              className="w-full justify-start gap-2 px-2 py-1.5 text-sm"
              onClick={() => {
                setOpen(false);
                void logout();
              }}
            >
              <LogOut className="h-4 w-4" />
              Sign out
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
