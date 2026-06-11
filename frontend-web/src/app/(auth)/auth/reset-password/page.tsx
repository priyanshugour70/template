"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
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
import { useResetPasswordMutation } from "@/hooks/auth/useAuthMutations";

/**
 * Consumes a reset-password token from the URL (sent by /auth/forgot-password
 * email) and sets a new password. Backend tokens are 1-hour single-use; on
 * success we redirect to /auth/login so the user signs in with the new pw.
 */
export default function ResetPasswordPage() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") ?? "";

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [mismatch, setMismatch] = useState(false);
  const reset = useResetPasswordMutation();

  // Missing/empty token — render a dead-end message rather than letting the
  // user submit and get a backend error. They'd land here from a manual URL
  // edit or a bad email link.
  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-2xl">Invalid reset link</CardTitle>
            <CardDescription>
              This reset link is missing its token. Request a new one and try again.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex flex-col gap-2 items-stretch">
            <Button asChild>
              <Link href="/auth/forgot-password">Request new link</Link>
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
      await reset.mutateAsync({ token, newPassword: password });
      router.replace("/auth/login?reset=ok");
    } catch {
      // reset.error is surfaced inline below — likely token expired/used.
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Set a new password</CardTitle>
          <CardDescription>
            Pick a strong password — at least 8 characters. You&apos;ll use it next time you sign in.
          </CardDescription>
        </CardHeader>
        <form onSubmit={onSubmit}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="password">New password</Label>
              <Input
                id="password"
                type="password"
                autoFocus
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
            {reset.isError && !mismatch && (
              <div className="text-sm text-destructive">
                {reset.error instanceof Error
                  ? reset.error.message
                  : "Reset failed — the link may have expired."}
              </div>
            )}
          </CardContent>
          <CardFooter className="flex flex-col gap-2 items-stretch">
            <Button type="submit" disabled={reset.isPending || !password || !confirm}>
              {reset.isPending ? "Saving…" : "Set password"}
            </Button>
            <Link
              href="/auth/login"
              className="text-sm text-center text-muted-foreground hover:text-foreground"
            >
              Back to sign in
            </Link>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
