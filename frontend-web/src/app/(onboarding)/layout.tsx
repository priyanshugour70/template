"use client";

import { Check, CircleDashed } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { useRequireAuth } from "@/hooks/auth/useRequireAuth";
import {
  ONBOARDING_STEPS,
  useOnboardingState,
  type OnboardingStep,
} from "@/hooks/onboarding/useOnboarding";
import { cn } from "@/lib/cn";

interface StepDef {
  step: OnboardingStep;
  href: string;
  label: string;
  description: string;
}

const STEPS: StepDef[] = [
  { step: "welcome", href: "/onboarding", label: "Welcome", description: "Tell us about you" },
  { step: "profile", href: "/onboarding/profile", label: "Profile", description: "Personal details" },
  { step: "workspace", href: "/onboarding/workspace", label: "Workspace", description: "Brand your org" },
  { step: "invites", href: "/onboarding/invites", label: "Teammates", description: "Optional invites" },
  { step: "plan", href: "/onboarding/plan", label: "Plan", description: "Pick a subscription" },
  { step: "done", href: "/onboarding/done", label: "Finish", description: "Jump to the dashboard" },
];

function stepIndex(s?: OnboardingStep): number {
  if (!s) return 0;
  const i = ONBOARDING_STEPS.indexOf(s);
  return i === -1 ? 0 : i;
}

export default function OnboardingLayout({ children }: { children: ReactNode }) {
  const { loading } = useRequireAuth();
  const state = useOnboardingState();
  const pathname = usePathname();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-sm text-muted-foreground">Loading…</div>
      </div>
    );
  }

  const currentStepIdx = Math.max(0, STEPS.findIndex((s) => s.href === pathname));
  const furthestStepIdx = stepIndex(state.step);
  const completedIdx = state.completed ? STEPS.length : furthestStepIdx;
  const progressPct = ((currentStepIdx + 1) / STEPS.length) * 100;

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted/30">
      <div className="mx-auto grid min-h-screen max-w-6xl grid-cols-1 lg:grid-cols-[280px_1fr]">
        {/* Sidebar progress */}
        <aside className="hidden border-r border-border bg-background/60 p-6 lg:block">
          <div className="mb-6">
            <p className="text-xs uppercase tracking-wider text-muted-foreground">Setup</p>
            <h2 className="mt-1 text-lg font-semibold tracking-tight">Welcome aboard</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              A few quick steps and you&apos;re in.
            </p>
          </div>

          <ol className="space-y-1">
            {STEPS.map((s, i) => {
              const isCurrent = i === currentStepIdx;
              const isDone = i < completedIdx;
              const isReachable = i <= Math.max(furthestStepIdx, currentStepIdx);
              return (
                <li key={s.step}>
                  <Link
                    href={isReachable ? s.href : pathname}
                    className={cn(
                      "flex items-start gap-3 rounded-md px-2 py-2 transition-colors",
                      isCurrent && "bg-muted",
                      !isCurrent && isReachable && "hover:bg-muted/50",
                      !isReachable && "pointer-events-none opacity-50",
                    )}
                  >
                    <div
                      className={cn(
                        "flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[10px] font-semibold",
                        isDone
                          ? "border-success bg-success text-success-foreground"
                          : isCurrent
                            ? "border-primary bg-primary text-primary-foreground"
                            : "border-border text-muted-foreground",
                      )}
                    >
                      {isDone ? <Check className="h-3 w-3" /> : i + 1}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div
                        className={cn(
                          "text-sm",
                          isCurrent ? "font-semibold text-foreground" : "font-medium",
                        )}
                      >
                        {s.label}
                      </div>
                      <div className="truncate text-xs text-muted-foreground">
                        {s.description}
                      </div>
                    </div>
                  </Link>
                </li>
              );
            })}
          </ol>
        </aside>

        {/* Main */}
        <main className="flex min-h-screen flex-col">
          {/* Top progress bar (mobile-visible) */}
          <div className="border-b border-border bg-background/40 px-4 py-3 lg:hidden">
            <div className="mb-1 flex items-center justify-between text-xs">
              <span className="font-medium">
                Step {currentStepIdx + 1} of {STEPS.length}
              </span>
              <span className="text-muted-foreground">{STEPS[currentStepIdx]?.label}</span>
            </div>
            <div className="h-1.5 overflow-hidden rounded-full bg-muted">
              <div
                className="h-full bg-primary transition-all"
                style={{ width: `${progressPct}%` }}
              />
            </div>
          </div>

          <div className="flex flex-1 items-start justify-center p-6 sm:p-10">
            <div className="w-full max-w-2xl">
              {children}
              <div className="mt-6 flex items-center justify-center gap-1">
                {STEPS.map((s, i) => (
                  <CircleDashed
                    key={s.step}
                    className={cn(
                      "h-2.5 w-2.5",
                      i === currentStepIdx
                        ? "text-primary"
                        : i < completedIdx
                          ? "text-success"
                          : "text-muted-foreground/30",
                    )}
                    fill={i < completedIdx || i === currentStepIdx ? "currentColor" : "none"}
                  />
                ))}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
