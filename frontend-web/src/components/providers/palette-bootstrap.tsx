import { DEFAULT_PALETTE } from "@/theme/palettes";

/**
 * Inline script that sets `data-palette` on <html> before the page renders.
 * Reads from the `app_palette` cookie (set by the theme picker) so SSR and
 * the first paint stay consistent.
 */
export function PaletteBootstrap() {
  const script = `
(function(){try{
  var valid=${JSON.stringify(["sunset-horizon","forest-trail","tropical-paradise","mountain-mist","desert-dunes"])};
  var match = document.cookie.match(/(?:^|; )app_palette=([^;]+)/);
  var stored = match ? decodeURIComponent(match[1]) : null;
  var palette = (stored && valid.indexOf(stored) !== -1) ? stored : ${JSON.stringify(DEFAULT_PALETTE)};
  document.documentElement.setAttribute('data-palette', palette);
}catch(e){}})();
`.trim();
  return <script dangerouslySetInnerHTML={{ __html: script }} />;
}
