"use client";

import {
  Activity,
  Building2,
  CreditCard,
  GitBranch,
  Home,
  Lock,
  LogOut,
  Moon,
  Settings as SettingsIcon,
  Sun,
  Users,
  UsersRound,
} from "lucide-react";
import { useTheme } from "next-themes";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import { useAuth, usePermissions } from "@/providers";

interface NavCmd {
  label: string;
  href: string;
  icon: typeof Home;
  permission?: string;
  shortcut?: string;
}

const NAV: NavCmd[] = [
  { label: "Home", href: "/dashboard", icon: Home, shortcut: "H" },
  { label: "Users", href: "/dashboard/administrative/users", icon: Users, permission: "user.list" },
  { label: "Roles & permissions", href: "/dashboard/administrative/roles", icon: Lock, permission: "role.list" },
  { label: "Departments", href: "/dashboard/administrative/departments", icon: GitBranch, permission: "department.list" },
  { label: "Groups", href: "/dashboard/administrative/groups", icon: UsersRound, permission: "group.list" },
  { label: "Organizations", href: "/dashboard/administrative/organizations", icon: Building2, permission: "org.list" },
  { label: "Audit log", href: "/dashboard/administrative/audit", icon: Activity, permission: "audit.read" },
  { label: "Subscription", href: "/dashboard/subscription", icon: CreditCard, permission: "subscription.read" },
  { label: "Settings", href: "/dashboard/settings", icon: SettingsIcon },
];

export function CommandPalette() {
  const router = useRouter();
  const { logout } = useAuth();
  const { has } = usePermissions();
  const { setTheme, resolvedTheme } = useTheme();
  const [open, setOpen] = useState(false);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  const run = (fn: () => void) => {
    setOpen(false);
    fn();
  };

  const items = NAV.filter((c) => !c.permission || has(c.permission));

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Type a command or search…" />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Navigate">
          {items.map((c) => {
            const Icon = c.icon;
            return (
              <CommandItem
                key={c.href}
                value={`${c.label} ${c.href}`}
                onSelect={() => run(() => router.push(c.href))}
              >
                <Icon className="text-muted-foreground" />
                <span>{c.label}</span>
                {c.shortcut && <CommandShortcut>{c.shortcut}</CommandShortcut>}
              </CommandItem>
            );
          })}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Actions">
          <CommandItem
            value="toggle theme"
            onSelect={() => run(() => setTheme(resolvedTheme === "dark" ? "light" : "dark"))}
          >
            {resolvedTheme === "dark" ? (
              <Sun className="text-muted-foreground" />
            ) : (
              <Moon className="text-muted-foreground" />
            )}
            <span>Toggle theme</span>
          </CommandItem>
          <CommandItem value="sign out" onSelect={() => run(() => void logout())}>
            <LogOut className="text-muted-foreground" />
            <span>Sign out</span>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}
