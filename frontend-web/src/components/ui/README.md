# src/components/ui/

shadcn/ui primitives. Each component is generated from the shadcn registry and **owned by your project** — modify freely.

```bash
# Add a new primitive
pnpm dlx shadcn add dropdown-menu
pnpm dlx shadcn add toast
```

All primitives use the semantic Tailwind utilities (`bg-primary`, `text-foreground`, …) which read from the active palette set by `<ThemeVars/>` + `usePaletteStore`. Re-tinting happens automatically when the user changes palette.
