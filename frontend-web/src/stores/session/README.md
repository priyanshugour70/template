# src/stores/session/

Session-scoped client state.

- `palette.store.ts` — current palette ID, persisted to `localStorage` under `"palette"`. The `PaletteBootstrap` inline script reads the same key on first paint to avoid flash.
- Add session-related stores here (e.g. last-selected store, recent searches).
