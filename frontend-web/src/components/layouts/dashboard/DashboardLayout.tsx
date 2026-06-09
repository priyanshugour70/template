"use client";

import type { ReactNode } from "react";

import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/toaster";

import { AppSidebar } from "./app-sidebar";
import { CommandPalette } from "./command-palette";
import { Header } from "./header";

export function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <SidebarProvider defaultOpen>
      <AppSidebar />
      <SidebarInset>
        <Header />
        <main className="flex-1 px-4 py-6 sm:px-6 sm:py-8">
          <div className="mx-auto w-full max-w-7xl">{children}</div>
        </main>
      </SidebarInset>
      <CommandPalette />
      <Toaster />
    </SidebarProvider>
  );
}
