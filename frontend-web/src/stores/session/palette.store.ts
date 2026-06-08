"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

import { DEFAULT_PALETTE, isValidPaletteId, type PaletteId } from "@/theme/palettes";

interface PaletteState {
  palette: PaletteId;
  setPalette: (id: PaletteId) => void;
}

export const usePaletteStore = create<PaletteState>()(
  persist(
    (set) => ({
      palette: DEFAULT_PALETTE,
      setPalette: (id) => {
        if (!isValidPaletteId(id)) return;
        if (typeof document !== "undefined") {
          document.documentElement.setAttribute("data-palette", id);
        }
        set({ palette: id });
      },
    }),
    {
      name: "palette",
      // localStorage key matches the inline bootstrap script in PaletteBootstrap.
    },
  ),
);
