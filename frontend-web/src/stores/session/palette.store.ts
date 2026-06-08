"use client";

import { create } from "zustand";

import { readPalette, writePalette } from "@/lib/cookies";
import { DEFAULT_PALETTE, isValidPaletteId, type PaletteId } from "@/theme/palettes";

interface PaletteState {
  palette: PaletteId;
  setPalette: (id: PaletteId) => void;
}

function initialPalette(): PaletteId {
  if (typeof window === "undefined") return DEFAULT_PALETTE;
  const fromCookie = readPalette();
  return fromCookie && isValidPaletteId(fromCookie) ? fromCookie : DEFAULT_PALETTE;
}

export const usePaletteStore = create<PaletteState>((set) => ({
  palette: initialPalette(),
  setPalette: (id) => {
    if (!isValidPaletteId(id)) return;
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("data-palette", id);
    }
    writePalette(id);
    set({ palette: id });
  },
}));
