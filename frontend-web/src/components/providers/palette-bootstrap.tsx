import { DEFAULT_PALETTE } from "@/theme/palettes";

/**
 * Inline script that sets `data-palette` on <html> before the page renders,
 * preventing palette flash on first paint. Runs once per page load.
 */
export function PaletteBootstrap() {
  const script = `
(function(){try{
  var k='palette';
  var stored=window.localStorage.getItem(k);
  var valid=${JSON.stringify(["sunset-horizon","forest-trail","tropical-paradise","mountain-mist","desert-dunes"])};
  var palette = (stored && valid.indexOf(stored) !== -1) ? stored : ${JSON.stringify(DEFAULT_PALETTE)};
  document.documentElement.setAttribute('data-palette', palette);
}catch(e){}})();
`.trim();
  return <script dangerouslySetInnerHTML={{ __html: script }} />;
}
