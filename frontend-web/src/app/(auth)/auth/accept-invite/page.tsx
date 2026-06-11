"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";

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
import { useAcceptInviteMutation } from "@/hooks/auth/useAuthMutations";

/**
 * Consumes an invite token from the email link. Backend signature:
 *   POST /auth/accept-invite { token, firstName, lastName?, password }
 * Returns a full session — we persist it (HttpOnly cookies, scoped to the
 * current tenant subdomain) and land the user on /dashboard.
 */
export default function AcceptInvitePage() {
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const emailHint = params.get("email") ?? "";

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [mismatch, setMismatch] = useState(false);
  const accept = useAcceptInviteMutation();

  // Dead-end the user immediately if the link is malformed. The backend would
  // 404 anyway but it's a worse UX to show a form that can't possibly succeed.
  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-2xl">Invalid invite link</CardTitle>
            <CardDescription>
              This link is missing its invitation token. Ask the person who invited you to send
              a fresh invite.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button asChild className="w-full">
              <Link href="/auth/login">Back to sign in</Link>
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setMismatch(true);
      return;
    }
    setMismatch(false);
    try {
      await accept.mutateAsync({
        token,
        firstName: firstName.trim(),
        lastName: lastName.trim() || undefined,
        password,
      });
      // Hard navigate so server components re-read the newly-set cookies.
      window.location.assign("/dashboard");
    } catch {
      // accept.error is surfaced inline below — typical causes: expired
      // token, already-used token, password too short.
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Accept your invitation</CardTitle>
          <CardDescription>
            {emailHint
              ? `Finish setting up your account for ${emailHint}.`
              : "Finish setting up your account."}
          </CardDescription>
        </CardHeader>
        <form onSubmit={onSubmit}>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label htmlFor="firstName">First name</Label>
                <Input
                  id="firstName"
                  type="text"
                  autoFocus
                  required
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName">Last name</Label>
                <Input
                  id="lastName"
                  type="text"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                required
                minLength={8}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm">Confirm password</Label>
              <Input
                id="confirm"
                type="password"
                required
                minLength={8}
                value={confirm}
                onChange={(e) => {
                  setConfirm(e.target.value);
                  if (mismatch) setMismatch(false);
                }}
              />
            </div>
            {mismatch && (
              <div className="text-sm text-destructive">
                Passwords don&apos;t match. Re-enter the same password in both fields.
              </div>
            )}
            {accept.isError && !mismatch && (
              <div className="text-sm text-destructive">
                {accept.error instanceof Error
                  ? accept.error.message
                  : "Couldn't accept this invite — the link may have expired or already been used."}
              </div>
            )}
          </CardContent>
          <CardFooter className="flex flex-col gap-2 items-stretch">
            <Button
              type="submit"
              disabled={accept.isPending || !firstName || !password || !confirm}
            >
              {accept.isPending ? "Setting up…" : "Accept invite"}
            </Button>
            <Link
              href="/auth/login"
              className="text-sm text-center text-muted-foreground hover:text-foreground"
            >
              Already have an account? Sign in
            </Link>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
