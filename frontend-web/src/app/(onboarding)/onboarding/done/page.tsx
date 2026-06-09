"use client";

import { ArrowRight, PartyPopper } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useSetOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { toast } from "@/hooks/use-toast";
import { useAuth, useTenant } from "@/providers";

export default function DoneStep() {
  const router = useRouter();
  const { user, refreshUser } = useAuth();
  const { activeOrganization } = useTenant();
  const setState = useSetOnboardingState();
  const [marked, setMarked] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (marked) return;
    (async () => {
      try {
        await setState.mutateAsync({
          patch: {
            step: "done",
            completed: true,
            completedAt: new Date().toISOString(),
          },
        });
        await refreshUser();
        setMarked(true);
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Couldn't mark complete";
        setError(msg);
        toast.error("Hiccup finishing onboarding", msg);
      }
    })();
    // run-once on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const goToDashboard = () => router.push("/dashboard");

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 6 of 6</p>
        <div className="mt-2 inline-flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
          <PartyPopper className="h-7 w-7 text-primary" />
        </div>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight">
          You&apos;re all set{user?.firstName ? `, ${user.firstName}` : ""}!
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {activeOrganization?.name ?? "Your workspace"} is ready to go.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-3 p-6">
          <p className="text-sm font-medium">Useful next steps</p>
          <ul className="space-y-1.5 text-sm">
            <Bullet>Invite the rest of your team — Settings → Profile</Bullet>
            <Bullet>Configure departments + roles for fine-grained access</Bullet>
            <Bullet>Issue an API key under Settings → Developer</Bullet>
            <Bullet>Add webhooks to react to events in real time</Bullet>
          </ul>
        </CardContent>
      </Card>

      {error && (
        <p className="text-center text-sm text-destructive">{error}</p>
      )}

      <div className="flex justify-center">
        <Button
          size="lg"
          onClick={goToDashboard}
          disabled={setState.isPending && !marked && !error}
        >
          Go to dashboard
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function Bullet({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-start gap-2 text-muted-foreground">
      <span className="mt-2 h-1 w-1 shrink-0 rounded-full bg-primary" />
      <span>{children}</span>
    </li>
  );
}
