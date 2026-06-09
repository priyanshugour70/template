/** Top-level Suspense fallback. Shown briefly during initial app shell loads. */
export default function Loading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-primary" />
        <span>Loading…</span>
      </div>
    </div>
  );
}
