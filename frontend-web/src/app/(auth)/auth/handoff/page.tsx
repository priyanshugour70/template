"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * Consumes a single-use SSO handoff token in the URL, persists the real
 * session (HttpOnly cookies, scoped to this subdomain), and redirects to the
 * dashboard. Renders a brief loading state in case the network is slow.
 */
export default function HandoffPage() {
  const router = useRouter();
  const search = useSearchParams();
  const token = search.get("token");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setError("Missing handoff token");
      return;
    }
    (async () => {
      try {
        const res = await fetch("/api/auth/handoff/consume", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
          credentials: "include",
        });
        const body = (await res.json()) as { success: boolean; error?: { message?: string } };
        if (!body.success) {
          setError(body.error?.message ?? "Sign-in failed. Please try again.");
          return;
        }
        // Hard navigate so server components re-read the new cookies.
        window.location.assign("/dashboard");
      } catch {
        setError("Network error. Please try again.");
      }
    })();
  }, [token, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">{error ? "Sign-in failed" : "Signing you in…"}</CardTitle>
          <CardDescription>
            {error ? error : "Finishing your workspace handoff. This only takes a second."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error ? (
            <a href="/auth/login" className="text-sm text-primary hover:underline">
              Try signing in again
            </a>
          ) : (
            <div className="flex justify-center py-6">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-r-transparent" />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
