"use client";

import { ThemeProvider as NextThemesProvider } from "next-themes";
import type { ReactNode } from "react";

import { ThemeVars } from "@/components/providers/theme-vars";
import { PaletteBootstrap } from "@/components/providers/palette-bootstrap";

export function ThemeProvider({ children }: { children: ReactNode }) {
  return (
    <NextThemesProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <ThemeVars />
      <PaletteBootstrap />
      {children}
    </NextThemesProvider>
  );
}
