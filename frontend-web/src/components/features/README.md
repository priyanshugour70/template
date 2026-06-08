# src/components/features/

Feature-specific composite components. One folder per domain feature.

Example: `features/sample/sample-table.tsx`, `features/sample/sample-form.tsx`.

A component here is allowed to import from:

- `@/components/ui/*` (primitives)
- `@/components/shared/*` (cross-feature)
- `@/hooks/<same-feature>`, `@/services/<same-feature>`, `@/types/<same-feature>`
- `@/stores/<same-feature>`

A component here MUST NOT import from `@/components/features/<other-feature>/*` — surface that need via `shared/` or a service interface instead.
