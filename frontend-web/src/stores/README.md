# src/stores/

Zustand stores, grouped by feature. Use stores for **cross-component, client-side ephemeral state** (current cart, current store-tag, palette preference).

For server state (data fetched from APIs), prefer **TanStack Query** — see `src/providers/query-provider.tsx`. Stores duplicating server state quickly drift out of sync.

```
stores/
├── session/                 # palette.store.ts, session.store.ts
├── cart/                    # cart.store.ts
└── …
```

Each store ships with selectors and is wrapped in `persist` middleware when the state should survive a reload.
