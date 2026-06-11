"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/providers";

/**
 * Defence-in-depth guard for onboarding steps that only make sense for the
 * tenant owner (workspace branding, plan picking, team invites). Non-owners
 * who land on these pages — typically by hand-editing the URL — get bounced
 * to /onboarding/profile, which IS in their flow.
 *
 * The sidebar already hides owner-only steps for members, so this guard
 * exists for direct-URL navigation only.
 */
export function useRequireOwner() {
  const router = useRouter();
  const { user, loading } = useAuth();

  useEffect(() => {
    if (loading) return;
    if (!user) return; // useRequireAuth in the layout handles unauthenticated
    if (user.isOwner !== true) {
      router.replace("/onboarding/profile");
    }
  }, [loading, user, router]);
}
