# src/theme/

Design tokens and the dynamic palette registry.

| File | Purpose |
|------|---------|
| `palettes.ts` | The 5 runtime-switchable palettes (Lss Goo design tokens). Single source of truth for **all** color / radius tokens. Exports `PALETTES`, `PALETTE_LIST`, `DEFAULT_PALETTE`, `ALL_PALETTES_CSS`, `RADIUS`, `LIGHT_THEME`, `DARK_THEME`. |
| `fonts.ts` | `next/font` declarations (Inter by default). Mount the CSS variable on `<html className={inter.variable}>`. |
| `index.ts` | Public re-exports. |

The palette is selected at runtime via `data-palette="…"` on `<html>`. The `<ThemeVars/>` component (in `components/providers/theme-vars.tsx`) emits the full CSS block, and `usePaletteStore` mutates the attribute on user selection — no JS re-render, no flash.

To add a palette: append a new `PaletteDef` entry to `PALETTES`. The shape (`ThemeTokens`) is enforced by TypeScript; both `light` and `dark` must define every key.
