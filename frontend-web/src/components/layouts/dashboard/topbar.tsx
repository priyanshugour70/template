"use client";

import { Search } from "lucide-react";

import { Breadcrumbs } from "./breadcrumbs";
import { OrgSwitcher } from "./org-switcher";
import { ThemePicker } from "./theme-picker";
import { UserMenu } from "./user-menu";
import { cn } from "@/lib/cn";
import { useSidebarStore } from "@/stores/ui/sidebar.store";

export function Topbar() {
  const collapsed = useSidebarStore((s) => s.collapsed);
  return (
    <header
      className={cn(
        "sticky top-0 z-20 border-b bg-background/85 backdrop-blur-sm transition-[margin] duration-200",
        collapsed ? "md:ml-16" : "md:ml-64",
      )}
    >
      <div className="flex h-16 items-center justify-between gap-4 px-6">
        <div className="flex items-center gap-3 min-w-0">
          <OrgSwitcher />
          <div className="hidden md:block">
            <Breadcrumbs />
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="hidden md:flex items-center gap-2 rounded-md border border-border bg-background px-3 h-9 w-72 text-sm">
            <Search className="h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search…"
              className="flex-1 bg-transparent outline-none placeholder:text-muted-foreground"
            />
            <kbd className="hidden lg:inline rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
              ⌘K
            </kbd>
          </div>
          <ThemePicker />
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
