# src/components/

| Folder | Purpose |
|--------|---------|
| `ui/` | **shadcn/ui primitives.** Add with `pnpm dlx shadcn add <component>` (see `components.json`). Owned by you — modify freely. |
| `layouts/` | Per-route-group app shells: `dashboard/`, `public/`, `back-office/`. Each exports a layout component used in `(group)/layout.tsx`. |
| `shared/` | Cross-feature presentation components (PageHeader, EmptyState, Pagination, …). |
| `features/` | Feature-specific composites (`features/sample/sample-table.tsx`). Co-locate with their domain. |
| `providers/` | React-context providers small enough not to warrant a top-level `providers/` entry (e.g. `theme-vars.tsx`, `palette-bootstrap.tsx`). |

## Naming

- Files: `kebab-case.tsx`, exporting `PascalCase` named exports.
- Server components by default. Add `"use client"` only at the leaf component that needs hooks/events.
