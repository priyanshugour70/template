"use client";

import { create } from "zustand";

import {
  readSidebarCollapsed,
  readSidebarSections,
  writeSidebarCollapsed,
  writeSidebarSections,
} from "@/lib/cookies";

interface SidebarState {
  collapsed: boolean;
  sections: Record<string, boolean>; // sectionId → collapsed
  toggleCollapsed: () => void;
  setCollapsed: (v: boolean) => void;
  toggleSection: (id: string) => void;
}

function initialCollapsed(): boolean {
  if (typeof window === "undefined") return false;
  return readSidebarCollapsed();
}
function initialSections(): Record<string, boolean> {
  if (typeof window === "undefined") return {};
  return readSidebarSections();
}

export const useSidebarStore = create<SidebarState>((set, get) => ({
  collapsed: initialCollapsed(),
  sections: initialSections(),
  toggleCollapsed: () => {
    const next = !get().collapsed;
    writeSidebarCollapsed(next);
    set({ collapsed: next });
  },
  setCollapsed: (v) => {
    writeSidebarCollapsed(v);
    set({ collapsed: v });
  },
  toggleSection: (id) => {
    const next = { ...get().sections, [id]: !get().sections[id] };
    writeSidebarSections(next);
    set({ sections: next });
  },
}));
