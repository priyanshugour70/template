import Link from "next/link";

import { PalettePicker } from "@/components/shared/palette-picker";

export default function Home() {
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col gap-10 px-6 py-16">
      <header className="space-y-3">
        <p className="text-sm uppercase tracking-widest text-muted-foreground">Template</p>
        <h1 className="text-3xl font-semibold text-foreground sm:text-4xl">Frontend Web</h1>
        <p className="text-base text-muted-foreground">
          Next.js 16 + Tailwind v4 + shadcn primitives with 5 runtime-switchable palettes.
        </p>
      </header>

      <section className="space-y-4 rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-medium text-foreground">Pick a palette</h2>
        <p className="text-sm text-muted-foreground">
          Each option mutates <code className="rounded bg-muted px-1 py-0.5">data-palette</code> on{" "}
          <code className="rounded bg-muted px-1 py-0.5">&lt;html&gt;</code>. Tailwind utilities
          re-tint instantly.
        </p>
        <PalettePicker />
      </section>

      <section className="flex flex-wrap gap-3 text-sm">
        <Link
          href="/dashboard"
          className="rounded-md bg-primary px-4 py-2 text-primary-foreground hover:opacity-90"
        >
          Go to Dashboard
        </Link>
        <Link
          href="/auth/login"
          className="rounded-md border border-border px-4 py-2 text-foreground hover:bg-muted"
        >
          Sign in
        </Link>
      </section>
    </main>
  );
}
