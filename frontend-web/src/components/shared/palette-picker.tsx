"use client";

import { Check } from "lucide-react";

import { usePaletteStore } from "@/stores/session/palette.store";
import { PALETTE_LIST } from "@/theme/palettes";

/** A simple palette picker. Drop into a settings page or a header menu. */
export function PalettePicker() {
  const palette = usePaletteStore((s) => s.palette);
  const setPalette = usePaletteStore((s) => s.setPalette);

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {PALETTE_LIST.map((p) => {
        const active = p.id === palette;
        return (
          <button
            key={p.id}
            type="button"
            onClick={() => setPalette(p.id)}
            className="group flex items-center justify-between rounded-lg border border-border bg-card p-3 text-left transition hover:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            aria-pressed={active}
          >
            <div className="flex items-center gap-3">
              <span className="flex h-8 items-center gap-1 rounded-md border border-border bg-background px-2">
                <span
                  className="h-4 w-4 rounded-full"
                  style={{ background: p.swatches.primary }}
                  aria-hidden
                />
                <span
                  className="h-4 w-4 rounded-full"
                  style={{ background: p.swatches.secondary }}
                  aria-hidden
                />
                <span
                  className="h-4 w-4 rounded-full"
                  style={{ background: p.swatches.accent }}
                  aria-hidden
                />
              </span>
              <span>
                <span className="block text-sm font-medium text-foreground">{p.label}</span>
                <span className="block text-xs text-muted-foreground">{p.blurb}</span>
              </span>
            </div>
            {active ? <Check className="h-4 w-4 text-primary" aria-hidden /> : null}
          </button>
        );
      })}
    </div>
  );
}
