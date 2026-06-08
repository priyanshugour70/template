"use client";

import { Moon, Palette, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { writePalette } from "@/lib/cookies";
import { PALETTE_LIST, type PaletteId } from "@/theme/palettes";
import { usePaletteStore } from "@/stores/session/palette.store";

export function ThemePicker() {
  const { setTheme, resolvedTheme } = useTheme();
  const palette = usePaletteStore((s) => s.palette);
  const setPalette = usePaletteStore((s) => s.setPalette);
  const [open, setOpen] = useState(false);

  // Hydrate palette from cookie on first mount (server can read cookie too).
  useEffect(() => {
    // no-op — store hydrates itself.
  }, []);

  function selectPalette(id: PaletteId) {
    setPalette(id);
    writePalette(id); // persist to cookie so SSR has it next request.
    setOpen(false);
  }

  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon"
        aria-label="Toggle theme"
        onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
      >
        {resolvedTheme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
      </Button>

      <div className="relative">
        <Button
          variant="ghost"
          size="icon"
          aria-label="Pick palette"
          onClick={() => setOpen((v) => !v)}
        >
          <Palette className="h-4 w-4" />
        </Button>
        {open && (
          <div className="absolute right-0 mt-2 w-56 rounded-md border bg-popover p-2 shadow-md z-50">
            <div className="text-xs font-semibold text-muted-foreground uppercase px-2 py-1">
              Palette
            </div>
            {PALETTE_LIST.map((p) => (
              <button
                key={p.id}
                onClick={() => selectPalette(p.id)}
                className={`w-full text-left rounded-sm px-2 py-1.5 text-sm hover:bg-accent flex items-center gap-2 ${palette === p.id ? "bg-accent" : ""}`}
              >
                <span
                  className="h-3 w-3 rounded-full border"
                  style={{ background: p.swatches.primary }}
                />
                {p.label}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
