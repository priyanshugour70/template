import { ALL_PALETTES_CSS } from "@/theme/palettes";

/**
 * Emits all palette × mode CSS-variable blocks as one <style> tag in <head>.
 * Server-rendered so the first paint already has correct tokens.
 */
export function ThemeVars() {
  return <style dangerouslySetInnerHTML={{ __html: ALL_PALETTES_CSS }} />;
}
