/**
 * Lss Goo design tokens.
 *
 * Single source of truth for **all** color / radius tokens used by the UI.
 *
 * ─────────────────────────────────────────────────────────────────────────
 *   PALETTE REGISTRY  →  5 curated themes, user-pickable at runtime.
 * ─────────────────────────────────────────────────────────────────────────
 *   The active palette is governed at runtime via:
 *     • `data-palette="…"` on `<html>` (set on first paint by an inline
 *       script; mutated reactively by `usePaletteStore`).
 *     • `class="dark"` on `<html>` (managed by `next-themes`).
 *
 *   `ThemeVars` emits one big `<style>` block that defines the CSS custom
 *   properties for **every** palette × mode combination, scoped by
 *   `[data-palette="…"]` and `.dark[data-palette="…"]`. Tailwind utilities
 *   (`bg-primary`, `text-foreground`, …) read those custom properties via
 *   `@theme inline` in `globals.css`, so the entire UI re-tints with a
 *   single attribute change — no JS re-render, no flash.
 *
 *   To add a palette: append a new `PaletteDef` entry to `PALETTES` below.
 *   Every palette **must** define both `light` and `dark` token sets with
 *   identical keys (the `ThemeTokens` shape is enforced by TypeScript).
 */

/* ---------- Token shape (light + dark have identical keys) ----------- */

export type ThemeTokens = {
  /* Surfaces */
  background: string;
  card: string;
  popover: string;
  muted: string;
  border: string;
  input: string;
  ring: string;

  /* Text */
  foreground: string;
  cardForeground: string;
  popoverForeground: string;
  mutedForeground: string;

  /* Brand roles */
  primary: string;
  primaryForeground: string;
  secondary: string;
  secondaryForeground: string;
  accent: string;
  accentForeground: string;
  success: string;
  successForeground: string;
  destructive: string;
  destructiveForeground: string;
};

export type PaletteId =
  | "sunset-horizon"
  | "forest-trail"
  | "tropical-paradise"
  | "mountain-mist"
  | "desert-dunes";

export type PaletteDef = {
  id: PaletteId;
  label: string;
  blurb: string;
  /** Three swatches used in the picker preview row (light-mode hues). */
  swatches: { primary: string; secondary: string; accent: string };
  light: ThemeTokens;
  dark: ThemeTokens;
};

/* ---------- Individual palettes -------------------------------------- */

const SUNSET_HORIZON: PaletteDef = {
  id: "sunset-horizon",
  label: "Sunset Horizon",
  blurb: "Golden hour · romantic destinations",
  swatches: { primary: "#E85D3C", secondary: "#0F2A4A", accent: "#F4B860" },
  light: {
    background: "#FFF8F1",
    card: "#FFFFFF",
    popover: "#FFFFFF",
    muted: "#FBEEDF",
    border: "#F1DEC4",
    input: "#F1DEC4",
    ring: "#E85D3C",
    foreground: "#1F1B16",
    cardForeground: "#1F1B16",
    popoverForeground: "#1F1B16",
    mutedForeground: "#7A6F5F",
    primary: "#E85D3C",
    primaryForeground: "#FFF8F1",
    secondary: "#0F2A4A",
    secondaryForeground: "#FFF8F1",
    accent: "#F4B860",
    accentForeground: "#1F1B16",
    success: "#2D9B5C",
    successForeground: "#FFF8F1",
    destructive: "#C8372F",
    destructiveForeground: "#FFF8F1",
  },
  dark: {
    background: "#1A0F08",
    card: "#2A1B11",
    popover: "#2A1B11",
    muted: "#2A1B11",
    border: "#3D2818",
    input: "#3D2818",
    ring: "#FF7B5A",
    foreground: "#FFF8F1",
    cardForeground: "#FFF8F1",
    popoverForeground: "#FFF8F1",
    mutedForeground: "#B5A593",
    primary: "#FF7B5A",
    primaryForeground: "#1A0F08",
    secondary: "#5B8DEF",
    secondaryForeground: "#1A0F08",
    accent: "#FFD089",
    accentForeground: "#1A0F08",
    success: "#4FBF7E",
    successForeground: "#1A0F08",
    destructive: "#FF6F5C",
    destructiveForeground: "#1A0F08",
  },
};

const FOREST_TRAIL: PaletteDef = {
  id: "forest-trail",
  label: "Forest Trail",
  blurb: "Mountain hikes · eco · slow travel",
  swatches: { primary: "#2F855A", secondary: "#C05621", accent: "#84A98C" },
  light: {
    background: "#F7F8F4",
    card: "#FFFFFF",
    popover: "#FFFFFF",
    muted: "#ECEFE5",
    border: "#DCE2D2",
    input: "#DCE2D2",
    ring: "#2F855A",
    foreground: "#1B2018",
    cardForeground: "#1B2018",
    popoverForeground: "#1B2018",
    mutedForeground: "#6B7561",
    primary: "#2F855A",
    primaryForeground: "#F7F8F4",
    secondary: "#C05621",
    secondaryForeground: "#F7F8F4",
    accent: "#84A98C",
    accentForeground: "#1B2018",
    success: "#38A169",
    successForeground: "#F7F8F4",
    destructive: "#C53030",
    destructiveForeground: "#F7F8F4",
  },
  dark: {
    background: "#0E1410",
    card: "#1A2218",
    popover: "#1A2218",
    muted: "#1A2218",
    border: "#283325",
    input: "#283325",
    ring: "#48BB78",
    foreground: "#F1F4EE",
    cardForeground: "#F1F4EE",
    popoverForeground: "#F1F4EE",
    mutedForeground: "#A3B19A",
    primary: "#48BB78",
    primaryForeground: "#0E1410",
    secondary: "#ED8936",
    secondaryForeground: "#0E1410",
    accent: "#A7C4AE",
    accentForeground: "#0E1410",
    success: "#68D391",
    successForeground: "#0E1410",
    destructive: "#FC8181",
    destructiveForeground: "#0E1410",
  },
};

const TROPICAL_PARADISE: PaletteDef = {
  id: "tropical-paradise",
  label: "Tropical Paradise",
  blurb: "Beach holidays · islands · vacation mode",
  swatches: { primary: "#0D9488", secondary: "#EC4899", accent: "#FACC15" },
  light: {
    background: "#F0FDFC",
    card: "#FFFFFF",
    popover: "#FFFFFF",
    muted: "#DEFAF7",
    border: "#C7F0EA",
    input: "#C7F0EA",
    ring: "#0D9488",
    foreground: "#0B1F1E",
    cardForeground: "#0B1F1E",
    popoverForeground: "#0B1F1E",
    mutedForeground: "#5C7574",
    primary: "#0D9488",
    primaryForeground: "#F0FDFC",
    secondary: "#EC4899",
    secondaryForeground: "#FFFFFF",
    accent: "#FACC15",
    accentForeground: "#0B1F1E",
    success: "#10B981",
    successForeground: "#FFFFFF",
    destructive: "#E11D48",
    destructiveForeground: "#FFFFFF",
  },
  dark: {
    background: "#0B1F1E",
    card: "#143231",
    popover: "#143231",
    muted: "#143231",
    border: "#1F4543",
    input: "#1F4543",
    ring: "#2DD4BF",
    foreground: "#ECFEFC",
    cardForeground: "#ECFEFC",
    popoverForeground: "#ECFEFC",
    mutedForeground: "#8FB4B1",
    primary: "#2DD4BF",
    primaryForeground: "#0B1F1E",
    secondary: "#F472B6",
    secondaryForeground: "#0B1F1E",
    accent: "#FDE047",
    accentForeground: "#0B1F1E",
    success: "#34D399",
    successForeground: "#0B1F1E",
    destructive: "#FB7185",
    destructiveForeground: "#0B1F1E",
  },
};

const MOUNTAIN_MIST: PaletteDef = {
  id: "mountain-mist",
  label: "Mountain Mist",
  blurb: "Premium alpine · editorial · curated stays",
  swatches: { primary: "#475569", secondary: "#B8553D", accent: "#84CC8C" },
  light: {
    background: "#F8FAFB",
    card: "#FFFFFF",
    popover: "#FFFFFF",
    muted: "#EEF1F4",
    border: "#DEE3E8",
    input: "#DEE3E8",
    ring: "#475569",
    foreground: "#1A2233",
    cardForeground: "#1A2233",
    popoverForeground: "#1A2233",
    mutedForeground: "#6B7587",
    primary: "#475569",
    primaryForeground: "#F8FAFB",
    secondary: "#B8553D",
    secondaryForeground: "#F8FAFB",
    accent: "#84CC8C",
    accentForeground: "#1A2233",
    success: "#16A34A",
    successForeground: "#F8FAFB",
    destructive: "#B91C1C",
    destructiveForeground: "#F8FAFB",
  },
  dark: {
    background: "#0D1117",
    card: "#161D29",
    popover: "#161D29",
    muted: "#161D29",
    border: "#232C3D",
    input: "#232C3D",
    ring: "#94A3B8",
    foreground: "#F1F5F9",
    cardForeground: "#F1F5F9",
    popoverForeground: "#F1F5F9",
    mutedForeground: "#94A3B8",
    primary: "#94A3B8",
    primaryForeground: "#0D1117",
    secondary: "#E07A5F",
    secondaryForeground: "#0D1117",
    accent: "#A7E3AC",
    accentForeground: "#0D1117",
    success: "#4ADE80",
    successForeground: "#0D1117",
    destructive: "#F87171",
    destructiveForeground: "#0D1117",
  },
};

const DESERT_DUNES: PaletteDef = {
  id: "desert-dunes",
  label: "Desert Dunes",
  blurb: "Heritage routes · Rajasthan · road trips",
  swatches: { primary: "#B45309", secondary: "#0E7C7B", accent: "#D4A04C" },
  light: {
    background: "#FBF6EE",
    card: "#FFFFFF",
    popover: "#FFFFFF",
    muted: "#F3E9D6",
    border: "#E8D9BC",
    input: "#E8D9BC",
    ring: "#B45309",
    foreground: "#231A12",
    cardForeground: "#231A12",
    popoverForeground: "#231A12",
    mutedForeground: "#7A6849",
    primary: "#B45309",
    primaryForeground: "#FBF6EE",
    secondary: "#0E7C7B",
    secondaryForeground: "#FBF6EE",
    accent: "#D4A04C",
    accentForeground: "#231A12",
    success: "#15803D",
    successForeground: "#FBF6EE",
    destructive: "#B91C1C",
    destructiveForeground: "#FBF6EE",
  },
  dark: {
    background: "#1B130B",
    card: "#2A1F12",
    popover: "#2A1F12",
    muted: "#2A1F12",
    border: "#3D2D1A",
    input: "#3D2D1A",
    ring: "#F59E0B",
    foreground: "#FBF6EE",
    cardForeground: "#FBF6EE",
    popoverForeground: "#FBF6EE",
    mutedForeground: "#B5A484",
    primary: "#F59E0B",
    primaryForeground: "#1B130B",
    secondary: "#2DB5B3",
    secondaryForeground: "#1B130B",
    accent: "#F1C470",
    accentForeground: "#1B130B",
    success: "#4ADE80",
    successForeground: "#1B130B",
    destructive: "#F87171",
    destructiveForeground: "#1B130B",
  },
};

/* ---------- Public registry ----------------------------------------- */

export const PALETTES: Record<PaletteId, PaletteDef> = {
  "sunset-horizon": SUNSET_HORIZON,
  "forest-trail": FOREST_TRAIL,
  "tropical-paradise": TROPICAL_PARADISE,
  "mountain-mist": MOUNTAIN_MIST,
  "desert-dunes": DESERT_DUNES,
};

/** Stable iteration order for the picker UI. */
export const PALETTE_LIST: readonly PaletteDef[] = [
  SUNSET_HORIZON,
  FOREST_TRAIL,
  TROPICAL_PARADISE,
  MOUNTAIN_MIST,
  DESERT_DUNES,
];

/** Default palette applied when a user has no saved preference. */
export const DEFAULT_PALETTE: PaletteId = "forest-trail";

export function isValidPaletteId(value: unknown): value is PaletteId {
  return typeof value === "string" && value in PALETTES;
}

/* ---------- Non-color tokens ---------------------------------------- */

export const RADIUS = {
  base: "0.75rem",
} as const;

/* ---------- CSS serialization --------------------------------------- */

/** `cardForeground` → `--card-foreground` */
function tokenToCssVar(key: keyof ThemeTokens): string {
  return `--${key.replace(/[A-Z]/g, (m) => `-${m.toLowerCase()}`)}`;
}

const TOKEN_VARS = (Object.keys(SUNSET_HORIZON.light) as (keyof ThemeTokens)[])
  .map((key) => ({ key, cssVar: tokenToCssVar(key) }))
  .sort((a, b) => a.cssVar.localeCompare(b.cssVar));

function declarations(theme: ThemeTokens): string {
  return TOKEN_VARS.map(({ key, cssVar }) => `  ${cssVar}: ${theme[key]};`).join("\n");
}

/**
 * Emits all palette × mode combinations as a single CSS string.
 *
 *   :root[data-palette="forest-trail"] { …light tokens… }
 *   .dark[data-palette="forest-trail"] { …dark tokens… }
 *   …repeated for every palette
 *
 * Plus a `:root` block carrying the shared `--radius` token and a fallback
 * for the default palette so unbranded SSR HTML still has tokens to read.
 */
export const ALL_PALETTES_CSS = [
  `:root {\n  --radius: ${RADIUS.base};\n}`,
  `:root:not([data-palette]) {\n${declarations(PALETTES[DEFAULT_PALETTE].light)}\n}`,
  `.dark:not([data-palette]) {\n${declarations(PALETTES[DEFAULT_PALETTE].dark)}\n}`,
  ...PALETTE_LIST.flatMap((p) => [
    `:root[data-palette="${p.id}"] {\n${declarations(p.light)}\n}`,
    `.dark[data-palette="${p.id}"] {\n${declarations(p.dark)}\n}`,
  ]),
].join("\n");

/* ---------- Convenience exports for non-runtime consumers ----------- */

/** Light + dark of the *default* palette — used by `viewport.themeColor`. */
export const LIGHT_THEME = PALETTES[DEFAULT_PALETTE].light;
export const DARK_THEME = PALETTES[DEFAULT_PALETTE].dark;
