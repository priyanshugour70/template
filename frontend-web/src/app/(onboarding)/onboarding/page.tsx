"use client";

import { ArrowRight, Briefcase, Code2, Headphones, LineChart, Sparkles } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useSetOnboardingState, useOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";
import { useAuth } from "@/providers";

const ROLES = [
  { value: "founder", label: "Founder / owner", icon: Sparkles, description: "Building the company" },
  { value: "admin", label: "Operations / IT", icon: Briefcase, description: "Managing access and policy" },
  { value: "developer", label: "Developer", icon: Code2, description: "Integrating the API" },
  { value: "support", label: "Customer support", icon: Headphones, description: "Helping users" },
  { value: "growth", label: "Marketing / sales", icon: LineChart, description: "Driving growth" },
];

const GOAL_OPTIONS = [
  "Replace a competing tool",
  "Manage team access (RBAC)",
  "Audit and compliance",
  "Build an integration",
  "Test it out before committing",
];

export default function WelcomeStep() {
  const router = useRouter();
  const { user } = useAuth();
  const state = useOnboardingState();
  const setState = useSetOnboardingState();

  const [role, setRole] = useState<string>(state.role ?? "");
  const [goals, setGoals] = useState<string[]>(state.goals ?? []);

  const next = async () => {
    if (!role) {
      toast.error("Pick a role to continue");
      return;
    }
    await setState.mutateAsync({
      patch: { step: "profile", role, goals },
    });
    router.push("/onboarding/profile");
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 1 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">
          Welcome{user?.firstName ? `, ${user.firstName}` : ""} 👋
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Help us tailor the setup. This takes about a minute.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-4 p-6">
          <div>
            <h3 className="text-sm font-semibold">What best describes your role?</h3>
            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              {ROLES.map((r) => {
                const Icon = r.icon;
                const active = role === r.value;
                return (
                  <button
                    key={r.value}
                    type="button"
                    onClick={() => setRole(r.value)}
                    className={cn(
                      "flex items-start gap-3 rounded-md border p-3 text-left transition-colors",
                      active
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-primary/40 hover:bg-muted/40",
                    )}
                  >
                    <Icon
                      className={cn("mt-0.5 h-4 w-4 shrink-0", active && "text-primary")}
                    />
                    <div className="min-w-0">
                      <div className="text-sm font-medium">{r.label}</div>
                      <div className="text-xs text-muted-foreground">{r.description}</div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <h3 className="text-sm font-semibold">
              What are you trying to do? <span className="text-muted-foreground">(optional)</span>
            </h3>
            <div className="mt-3 flex flex-wrap gap-2">
              {GOAL_OPTIONS.map((g) => {
                const active = goals.includes(g);
                return (
                  <button
                    key={g}
                    type="button"
                    onClick={() =>
                      setGoals((prev) =>
                        prev.includes(g) ? prev.filter((x) => x !== g) : [...prev, g],
                      )
                    }
                    className={cn(
                      "rounded-full border px-3 py-1 text-xs transition-colors",
                      active
                        ? "border-primary bg-primary/10 text-primary"
                        : "border-border text-muted-foreground hover:border-primary/40",
                    )}
                  >
                    {g}
                  </button>
                );
              })}
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={next} disabled={setState.isPending}>
          {setState.isPending ? "Saving…" : "Continue"}
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
