"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

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
import { useSubdomain } from "@/hooks/useSubdomain";
import { useAuth } from "@/providers";
import type { DiscoveredTenant } from "@/types/auth";

type Step = "email" | "tenant" | "password";

interface TenantBySlugResponse {
  id: string;
  slug: string;
  name: string;
  logoUrl?: string;
  primaryColor?: string;
}

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
  const subdomain = useSubdomain();

  // Bounce already-signed-in users (covers browser bfcache, where the proxy
  // doesn't run on back-navigation).
  useEffect(() => {
    if (!loading && isAuthenticated) router.replace(redirect);
  }, [loading, isAuthenticated, redirect, router]);

  // On a tenant subdomain we skip discovery entirely: the tenant is fixed by
  // the URL. We resolve it once via the public /auth/tenant-by-slug endpoint
  // so the password form can show "Signing in to Acme" with the right brand.
  const [subdomainTenant, setSubdomainTenant] = useState<TenantBySlugResponse | null>(null);
  const [subdomainResolveError, setSubdomainResolveError] = useState<string | null>(null);

  useEffect(() => {
    if (!subdomain) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch(`/api/v1/auth/tenant-by-slug?slug=${encodeURIComponent(subdomain)}`, {
          credentials: "include",
        });
        const body = (await res.json()) as { success: boolean; data?: TenantBySlugResponse; error?: { message?: string } };
        if (cancelled) return;
        if (body.success && body.data) {
          setSubdomainTenant(body.data);
        } else {
          setSubdomainResolveError(body.error?.message ?? "Workspace not found");
        }
      } catch {
        if (!cancelled) setSubdomainResolveError("Could not reach the server");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [subdomain]);

  // Initial step depends on whether we're on the apex or a tenant subdomain.
  const initialStep: Step = useMemo(() => (subdomain ? "email" : "email"), [subdomain]);
  const [step, setStep] = useState<Step>(initialStep);
  const [email, setEmail] = useState("");
  const [tenants, setTenants] = useState<DiscoveredTenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState<DiscoveredTenant | null>(null);
  const [password, setPassword] = useState("");

  const discover = useDiscoverMutation();
  const login = useLoginMutation();

  // When the subdomain tenant resolves AND the user has typed an email, skip
  // the discovery step and go straight to password.
  useEffect(() => {
    if (subdomainTenant && !selectedTenant) {
      setSelectedTenant({
        id: subdomainTenant.id,
        name: subdomainTenant.name,
        slug: subdomainTenant.slug,
        logoUrl: subdomainTenant.logoUrl,
        primaryColor: subdomainTenant.primaryColor,
      });
    }
  }, [subdomainTenant, selectedTenant]);

  async function onDiscover(e: React.FormEvent) {
    e.preventDefault();
    // On a tenant subdomain we don't need /discover — the tenant is fixed and
    // we already filled it from /tenant-by-slug. Just advance to password.
    if (subdomain && subdomainTenant) {
      setStep("password");
      return;
    }
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
      const data = (await login.mutateAsync({
        email: email.trim(),
        password,
        tenantId: selectedTenant.id,
      })) as { mode?: string; redirect?: string };
      // On an apex login the server returns a handoff redirect URL pointing at
      // the tenant subdomain. On a same-subdomain login the cookies are
      // already set; just go to /dashboard.
      if (data?.mode === "handoff" && data.redirect) {
        window.location.assign(data.redirect);
        return;
      }
      window.location.assign(redirect);
    } catch {
      // login.error is surfaced inline below.
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">
            {subdomainTenant ? `Sign in to ${subdomainTenant.name}` : "Welcome back"}
          </CardTitle>
          <CardDescription>
            {subdomainResolveError && step === "email"
              ? subdomainResolveError
              : step === "email" && subdomain
              ? "Enter your work email to continue."
              : step === "email"
              ? "Enter your work email — we'll find your workspace."
              : step === "tenant"
              ? "Pick the workspace you want to sign in to."
              : step === "password" && selectedTenant
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
