"use client";

import { create } from "zustand";

import type { ToastVariant } from "@/components/ui/toast";

export interface ToastItem {
  id: string;
  title?: string;
  description?: string;
  variant?: ToastVariant;
  duration?: number;
}

interface ToastStore {
  toasts: ToastItem[];
  push: (t: Omit<ToastItem, "id">) => string;
  dismiss: (id: string) => void;
}

let counter = 0;
const nextId = () => `t-${++counter}`;

export const useToastStore = create<ToastStore>((set) => ({
  toasts: [],
  push: (t) => {
    const id = nextId();
    set((s) => ({ toasts: [...s.toasts, { id, ...t }] }));
    return id;
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((x) => x.id !== id) })),
}));

/** Imperative helpers — usable outside React. */
export const toast = {
  show: (t: Omit<ToastItem, "id">) => useToastStore.getState().push(t),
  success: (title: string, description?: string) =>
    useToastStore.getState().push({ title, description, variant: "success" }),
  error: (title: string, description?: string) =>
    useToastStore.getState().push({ title, description, variant: "destructive" }),
  info: (title: string, description?: string) =>
    useToastStore.getState().push({ title, description }),
};

export function useToast() {
  return useToastStore();
}
