"use client";

import { Building2, User } from "lucide-react";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

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
import { Separator } from "@/components/ui/separator";
import { useAuth } from "@/providers";
import { registerService } from "@/services/auth/register";

function slugify(s: string): string {
  return s
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export default function SignupPage() {
  const router = useRouter();
  const { isAuthenticated, loading } = useAuth();

  // Bounce already-signed-in users (covers browser bfcache).
  useEffect(() => {
    if (!loading && isAuthenticated) router.replace("/dashboard");
  }, [loading, isAuthenticated, router]);

  const [organizationName, setOrganizationName] = useState("");
  const [organizationSlug, setOrganizationSlug] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!organizationName || !firstName || !email || password.length < 8) {
      setError("Please fill all required fields. Password must be 8+ characters.");
      return;
    }
    const slug = organizationSlug || slugify(organizationName);
    setSubmitting(true);
    try {
      const res = await registerService.register({
        email: email.trim(),
        password,
        firstName: firstName.trim(),
        lastName: lastName.trim() || undefined,
        organizationName: organizationName.trim(),
        organizationSlug: slug,
      });
      if (!res.success) throw new Error(res.error?.message ?? "Registration failed");
      // New tenant has no subscription — jump to plan selection.
      window.location.assign("/onboarding/subscription");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Registration failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4 py-12">
      <Card className="w-full max-w-xl">
        <CardHeader>
          <CardTitle className="text-2xl">Create your workspace</CardTitle>
          <CardDescription>
            Set up your organization and your owner account in one step. You&apos;ll pick a plan next.
          </CardDescription>
        </CardHeader>

        <form onSubmit={submit}>
          <CardContent className="space-y-6">
            {/* Workspace */}
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-sm font-semibold">
                <Building2 className="h-4 w-4 text-primary" />
                Workspace
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="org-name">Workspace name</Label>
                  <Input
                    id="org-name"
                    required
                    value={organizationName}
                    onChange={(e) => {
                      const v = e.target.value;
                      setOrganizationName(v);
                      if (!organizationSlug) setOrganizationSlug(slugify(v));
                    }}
                    placeholder="Acme Corporation"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="org-slug">URL slug</Label>
                  <Input
                    id="org-slug"
                    required
                    value={organizationSlug}
                    onChange={(e) => setOrganizationSlug(slugify(e.target.value))}
                    placeholder="acme"
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* Owner account */}
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-sm font-semibold">
                <User className="h-4 w-4 text-primary" />
                Owner account
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="first-name">First name</Label>
                  <Input id="first-name" required value={firstName} onChange={(e) => setFirstName(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="last-name">Last name</Label>
                  <Input id="last-name" value={lastName} onChange={(e) => setLastName(e.target.value)} />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Work email</Label>
                <Input id="email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password (8+ characters)</Label>
                <Input
                  id="password"
                  type="password"
                  required
                  minLength={8}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            </div>

            {error && <p className="text-sm text-destructive">{error}</p>}
          </CardContent>

          <CardFooter className="flex flex-col gap-3 items-stretch">
            <Button type="submit" disabled={submitting}>
              {submitting ? "Creating workspace…" : "Continue to plan selection"}
            </Button>
            <p className="text-center text-sm text-muted-foreground">
              Already have an account?{" "}
              <Link href="/auth/login" className="text-primary hover:underline">
                Sign in
              </Link>
            </p>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
