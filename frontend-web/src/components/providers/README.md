# src/components/providers/

Small context/wiring components mounted by app-root providers. Keep top-level concerns (Auth, Query, Theme) in `src/providers/`; this folder is for the helpers those providers compose.

- `theme-vars.tsx` — emits all palette × mode CSS-variable blocks as a single `<style>` tag.
- `palette-bootstrap.tsx` — inline script that sets `data-palette` on `<html>` before first paint to avoid flash.
