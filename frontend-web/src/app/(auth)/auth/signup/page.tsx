"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import Link from "next/link";

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
  const [step, setStep] = useState<1 | 2>(1);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [organizationName, setOrganizationName] = useState("");
  const [organizationSlug, setOrganizationSlug] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function next(e: React.FormEvent) {
    e.preventDefault();
    if (!firstName || !email || password.length < 8) {
      setError("Please fill the required fields and use an 8+ character password.");
      return;
    }
    setError(null);
    setStep(2);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (!organizationName) {
      setError("Workspace name is required.");
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
      if (!res.success) {
        throw new Error(res.error?.message ?? "Registration failed");
      }
      // New tenant has no subscription — jump to onboarding.
      window.location.assign("/onboarding/subscription");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Registration failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4 py-12">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Create your workspace</CardTitle>
          <CardDescription>
            {step === 1 ? "Step 1 of 2 — your account." : "Step 2 of 2 — your workspace."}
          </CardDescription>
        </CardHeader>

        {step === 1 && (
          <form onSubmit={next}>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="firstName">First name</Label>
                  <Input id="firstName" required value={firstName} onChange={(e) => setFirstName(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="lastName">Last name</Label>
                  <Input id="lastName" value={lastName} onChange={(e) => setLastName(e.target.value)} />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Work email</Label>
                <Input id="email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password (8+ characters)</Label>
                <Input id="password" type="password" required minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </CardContent>
            <CardFooter className="flex flex-col gap-2 items-stretch">
              <Button type="submit">Continue</Button>
              <p className="text-center text-sm text-muted-foreground">
                Have an account?{" "}
                <Link href="/auth/login" className="text-primary hover:underline">Sign in</Link>
              </p>
            </CardFooter>
          </form>
        )}

        {step === 2 && (
          <form onSubmit={submit}>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="orgName">Workspace name</Label>
                <Input
                  id="orgName"
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
                <Label htmlFor="orgSlug">Workspace URL</Label>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">app.example.com/</span>
                  <Input
                    id="orgSlug"
                    required
                    value={organizationSlug}
                    onChange={(e) => setOrganizationSlug(slugify(e.target.value))}
                    placeholder="acme"
                  />
                </div>
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </CardContent>
            <CardFooter className="flex flex-col gap-2 items-stretch">
              <Button type="submit" disabled={submitting}>
                {submitting ? "Creating workspace…" : "Create workspace"}
              </Button>
              <Button type="button" variant="ghost" onClick={() => setStep(1)}>Back</Button>
            </CardFooter>
          </form>
        )}
      </Card>
    </div>
  );
}
