"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { userService } from "@/services/user";
import { useAuth } from "@/providers";
import type { JSONObject } from "@/types/common";

export const ONBOARDING_STEPS = [
  "welcome",
  "profile",
  "workspace",
  "invites",
  "plan",
  "done",
] as const;

export type OnboardingStep = (typeof ONBOARDING_STEPS)[number];

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
