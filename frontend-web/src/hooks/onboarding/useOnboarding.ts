"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { userService } from "@/services/user";
import { useAuth } from "@/providers";
import type { JSONObject } from "@/types/common";

// Full step set — used by tenant owners (founders) who need to brand the
// workspace, invite teammates, and pick a plan.
export const OWNER_ONBOARDING_STEPS = [
  "welcome",
  "profile",
  "workspace",
  "invites",
  "plan",
  "done",
] as const;

// Minimal step set for invited members. The workspace, plan, and team are
// already configured by the founder — invited users just confirm their own
// profile and jump into the dashboard.
export const MEMBER_ONBOARDING_STEPS = ["welcome", "profile", "done"] as const;

// Backwards-compat union: still the same set of step identifiers, so callers
// that import OnboardingStep don't break. The ordering used at runtime now
// depends on isOwner.
export const ONBOARDING_STEPS = OWNER_ONBOARDING_STEPS;

export type OnboardingStep = (typeof OWNER_ONBOARDING_STEPS)[number];

/** Returns the active step ordering for the current user. Owners get the
 *  full 6-step flow; invited members get a 3-step (welcome / profile / done)
 *  flow. */
export function useOnboardingSteps(): readonly OnboardingStep[] {
  const { user } = useAuth();
  const isOwner = user?.isOwner === true;
  return isOwner ? OWNER_ONBOARDING_STEPS : MEMBER_ONBOARDING_STEPS;
}

export interface OnboardingState {
  step?: OnboardingStep;
  completed?: boolean;
  completedAt?: string;
  role?: string;
  goals?: string[];
}

/**
 * Pulls the onboarding state out of the (cookie-hydrated) auth user's
 * preferences blob. The user lives in the AuthProvider — no extra fetch.
 */
export function useOnboardingState(): OnboardingState {
  const { user } = useAuth();
  const prefs = (user as unknown as { preferences?: JSONObject })?.preferences;
  const ob = (prefs?.onboarding ?? {}) as OnboardingState;
  return ob;
}

interface SetStateArgs {
  patch: Partial<OnboardingState>;
}

/**
 * Mutation that merges new fields into preferences.onboarding and PATCHes
 * /users/me. After success, the auth user is refreshed so other components
 * see the new step.
 */
export function useSetOnboardingState() {
  const qc = useQueryClient();
  const { user, refreshUser } = useAuth();

  return useMutation({
    mutationFn: async ({ patch }: SetStateArgs) => {
      const current =
        ((user as unknown as { preferences?: JSONObject })?.preferences ??
          ({} as JSONObject)) as JSONObject;
      const merged = {
        ...current,
        onboarding: {
          ...(current.onboarding as object | undefined),
          ...patch,
        },
      } as JSONObject;
      const res = await userService.updateMe({ preferences: merged });
      if (!res.success) throw new Error(res.error?.message ?? "update failed");
      return res.data;
    },
    onSuccess: async () => {
      await refreshUser();
      void qc.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
