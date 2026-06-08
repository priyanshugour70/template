export default function PublicLanding() {
  return (
    <main className="mx-auto max-w-3xl px-6 py-16">
      <h1 className="text-3xl font-semibold">Public landing</h1>
      <p className="mt-3 text-muted-foreground">
        Pages under the <code className="rounded bg-muted px-1 py-0.5">(public)</code> route group
        are accessible without a session.
      </p>
    </main>
  );
}
