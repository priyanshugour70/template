"use client";

import type { ReactNode } from "react";

import { cn } from "@/lib/cn";
import { useSidebarStore } from "@/stores/ui/sidebar.store";

import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

export function DashboardLayout({ children }: { children: ReactNode }) {
  const collapsed = useSidebarStore((s) => s.collapsed);
  return (
    <div className="min-h-screen bg-muted/20">
      <Sidebar />
      <Topbar />
      <main
        className={cn(
          "px-6 py-8 transition-[margin] duration-200",
          collapsed ? "md:ml-16" : "md:ml-64",
        )}
      >
        <div className="mx-auto max-w-7xl">{children}</div>
      </main>
    </div>
  );
}
