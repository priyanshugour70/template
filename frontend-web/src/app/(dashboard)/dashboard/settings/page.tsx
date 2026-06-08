"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useChangePasswordMutation } from "@/hooks/auth/useAuthMutations";
import { useAuth } from "@/providers";

export default function SettingsPage() {
  const { user } = useAuth();
  const change = useChangePasswordMutation();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!current || !next) return;
    try {
      await change.mutateAsync({ currentPassword: current, newPassword: next });
      setCurrent("");
      setNext("");
    } catch {
      /* error surfaced via change.error */
    }
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        <p className="text-muted-foreground mt-1">Manage your profile and security.</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Profile</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div>Email: <span className="font-medium">{user?.email}</span></div>
          <div>Name: <span className="font-medium">{user?.displayName ?? "—"}</span></div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Change password</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={submit} className="space-y-4 max-w-md">
            <div className="space-y-2">
              <Label htmlFor="current">Current password</Label>
              <Input id="current" type="password" value={current} onChange={(e) => setCurrent(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="next">New password</Label>
              <Input id="next" type="password" value={next} onChange={(e) => setNext(e.target.value)} />
            </div>
            {change.isError && (
              <p className="text-sm text-destructive">
                {change.error instanceof Error ? change.error.message : "Failed"}
              </p>
            )}
            {change.isSuccess && (
              <p className="text-sm text-emerald-600">Password updated.</p>
            )}
            <Button type="submit" disabled={change.isPending || !current || !next}>
              {change.isPending ? "Updating…" : "Update password"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
