"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useDiscoverMutation, useLoginMutation } from "@/hooks/auth/useAuthMutations";
import { useAuth } from "@/providers";
import type { DiscoveredTenant } from "@/types/auth";

type Step = "email" | "tenant" | "password";

/** Reject redirect targets that would bounce the user back into the auth tree
 * or off-site. Anything that's not a local non-auth path falls back to /dashboard. */
function safeRedirect(raw: string | null): string {
  if (!raw) return "/dashboard";
  if (!raw.startsWith("/") || raw.startsWith("//")) return "/dashboard";
  if (raw === "/auth" || raw === "/auth/" || raw.startsWith("/auth/")) {
    // /auth/accept-invite is valid post-login but the invite flow handles that
    // explicitly; the safe default here is /dashboard.
    return "/dashboard";
  }
  return raw;
}

export default function LoginPage() {
  const router = useRouter();
  const search = useSearchParams();
  const redirect = safeRedirect(search.get("redirect"));
  const { isAuthenticated, loading } = useAuth();

  // Bounce already-signed-in users (covers browser bfcache, where the proxy
  // doesn't run on back-navigation).
  useEffect(() => {
    if (!loading && isAuthenticated) router.replace(redirect);
  }, [loading, isAuthenticated, redirect, router]);

  const [step, setStep] = useState<Step>("email");
  const [email, setEmail] = useState("");
  const [tenants, setTenants] = useState<DiscoveredTenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState<DiscoveredTenant | null>(null);
  const [password, setPassword] = useState("");

  const discover = useDiscoverMutation();
  const login = useLoginMutation();

  async function onDiscover(e: React.FormEvent) {
    e.preventDefault();
    try {
      const res = await discover.mutateAsync(email.trim());
      const list = (res as { tenants: DiscoveredTenant[] }).tenants ?? [];
      if (list.length === 0) {
        setStep("password");
        return;
      }
      if (list.length === 1) {
        setSelectedTenant(list[0]);
        setStep("password");
        return;
      }
      setTenants(list);
      setStep("tenant");
    } catch {
      setStep("password");
    }
  }

  async function onLogin(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedTenant) return;
    try {
      await login.mutateAsync({
        email: email.trim(),
        password,
        tenantId: selectedTenant.id,
      });
      // Hard navigate so server components re-read the new cookies.
      window.location.assign(redirect);
    } catch {
      // login.error is surfaced inline below.
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Welcome back</CardTitle>
          <CardDescription>
            {step === "email" && "Enter your work email to continue."}
            {step === "tenant" && "Pick the workspace you want to sign in to."}
            {step === "password" && selectedTenant
              ? `Signing in to ${selectedTenant.name}.`
              : step === "password"
              ? "Enter your password."
              : ""}
          </CardDescription>
        </CardHeader>

        {step === "email" && (
          <form onSubmit={onDiscover}>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Work email</Label>
                <Input
                  id="email"
                  type="email"
                  autoFocus
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@company.com"
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-2 items-stretch">
              <Button type="submit" disabled={discover.isPending || !email}>
                {discover.isPending ? "Looking up…" : "Continue"}
              </Button>
            </CardFooter>
          </form>
        )}

        {step === "tenant" && (
          <CardContent className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Found {tenants.length} workspaces for this email.
            </p>
            {tenants.map((t) => (
              <button
                key={t.id}
                onClick={() => {
                  setSelectedTenant(t);
                  setStep("password");
                }}
                className="w-full flex items-center gap-3 rounded-md border p-3 text-left hover:bg-accent transition-colors"
              >
                {t.logoUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={t.logoUrl} alt={t.name} className="h-10 w-10 rounded-md object-cover" />
                ) : (
                  <div className="h-10 w-10 rounded-md bg-primary/10" />
                )}
                <div className="flex-1">
                  <div className="font-medium">{t.name}</div>
                  <div className="text-xs text-muted-foreground">/{t.slug}</div>
                </div>
              </button>
            ))}
            <Button variant="ghost" type="button" onClick={() => setStep("email")}>
              Use a different email
            </Button>
          </CardContent>
        )}

        {step === "password" && (
          <form onSubmit={onLogin}>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3 rounded-md bg-muted/40 p-3">
                {selectedTenant?.logoUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={selectedTenant.logoUrl} alt={selectedTenant.name} className="h-8 w-8 rounded-md object-cover" />
                ) : (
                  <div className="h-8 w-8 rounded-md bg-primary/10" />
                )}
                <div className="flex-1 text-sm">
                  <div className="font-medium">{selectedTenant?.name ?? "—"}</div>
                  <div className="text-xs text-muted-foreground">{email}</div>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  autoFocus
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
              {login.isError && (
                <div className="text-sm text-destructive">
                  {login.error instanceof Error ? login.error.message : "Sign in failed"}
                </div>
              )}
              <div className="text-sm">
                <a href="/auth/forgot-password" className="text-primary hover:underline">
                  Forgot password?
                </a>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-2 items-stretch">
              <Button type="submit" disabled={login.isPending || !password || !selectedTenant}>
                {login.isPending ? "Signing in…" : "Sign in"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={() => {
                  setStep("email");
                  setPassword("");
                }}
              >
                Back
              </Button>
            </CardFooter>
          </form>
        )}
      </Card>
    </div>
  );
}
