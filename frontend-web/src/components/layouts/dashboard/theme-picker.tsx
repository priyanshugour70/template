"use client";

import { Moon, Palette, Sun } from "lucide-react";
import { useTheme } from "next-themes";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { writePalette } from "@/lib/cookies";
import { PALETTE_LIST, type PaletteId } from "@/theme/palettes";
import { usePaletteStore } from "@/stores/session/palette.store";

export function ThemePicker() {
  const { setTheme, resolvedTheme } = useTheme();
  const palette = usePaletteStore((s) => s.palette);
  const setPalette = usePaletteStore((s) => s.setPalette);

  function selectPalette(id: PaletteId) {
    setPalette(id);
    writePalette(id);
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

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Pick palette">
            <Palette className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-[220px]">
          <DropdownMenuLabel>Palette</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {PALETTE_LIST.map((p) => (
            <DropdownMenuItem
              key={p.id}
              onSelect={() => selectPalette(p.id)}
              className="gap-3"
            >
              <span
                className="h-4 w-4 rounded-full border ring-1 ring-inset ring-border/50"
                style={{ background: p.swatches.primary }}
              />
              <div className="flex-1">
                <div className="font-medium">{p.label}</div>
                <div className="text-xs text-muted-foreground">{p.blurb}</div>
              </div>
              {palette === p.id && <span className="h-2 w-2 rounded-full bg-primary" />}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
