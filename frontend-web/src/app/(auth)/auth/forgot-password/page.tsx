"use client";

import Link from "next/link";
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
import { useForgotPasswordMutation } from "@/hooks/auth/useAuthMutations";

/**
 * Forgot-password page. POSTs the email to the backend which always 202s —
 * we never reveal whether the address exists. The success state shows a
 * neutral "if it exists, check your inbox" message regardless of outcome.
 */
export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const forgot = useForgotPasswordMutation();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await forgot.mutateAsync({ email: email.trim() });
    } catch {
      // intentional: backend is silent about existence; surface a neutral
      // success state to the user regardless.
    }
    setSubmitted(true);
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">
            {submitted ? "Check your inbox" : "Reset your password"}
          </CardTitle>
          <CardDescription>
            {submitted
              ? "If an account exists for that email, we just sent a reset link. The link expires in 1 hour."
              : "Enter your work email. We'll send you a link to set a new password."}
          </CardDescription>
        </CardHeader>

        {!submitted && (
          <form onSubmit={onSubmit}>
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
              <Button type="submit" disabled={forgot.isPending || !email}>
                {forgot.isPending ? "Sending…" : "Send reset link"}
              </Button>
              <Link
                href="/auth/login"
                className="text-sm text-center text-muted-foreground hover:text-foreground"
              >
                Back to sign in
              </Link>
            </CardFooter>
          </form>
        )}

        {submitted && (
          <CardFooter className="flex flex-col gap-2 items-stretch">
            <Button asChild>
              <Link href="/auth/login">Back to sign in</Link>
            </Button>
            <button
              type="button"
              className="text-sm text-muted-foreground hover:text-foreground"
              onClick={() => {
                setSubmitted(false);
                setEmail("");
              }}
            >
              Use a different email
            </button>
          </CardFooter>
        )}
      </Card>
    </div>
  );
}
