export default function DashboardPage() {
  return (
    <main className="px-6 py-10">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Place your authenticated dashboard here. The route group{" "}
        <code className="rounded bg-muted px-1 py-0.5">(dashboard)</code> can host a shared layout
        in <code className="rounded bg-muted px-1 py-0.5">(dashboard)/layout.tsx</code>.
      </p>
    </main>
  );
}
