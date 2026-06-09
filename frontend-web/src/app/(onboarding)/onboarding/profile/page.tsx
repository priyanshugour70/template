"use client";

import { ArrowLeft, ArrowRight } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useSetOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { toast } from "@/hooks/use-toast";
import { useUpdateUser, useUser } from "@/hooks/user/useUserQueries";
import { useAuth } from "@/providers";

export default function ProfileStep() {
  const router = useRouter();
  const { user, refreshUser } = useAuth();
  const profileQ = useUser(user?.id);
  const update = useUpdateUser(user?.id ?? "");
  const setState = useSetOnboardingState();

  const profile = profileQ.data;
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [phone, setPhone] = useState("");
  const [timezone, setTimezone] = useState("Asia/Kolkata");
  const [locale, setLocale] = useState("en-IN");
  const [avatarUrl, setAvatarUrl] = useState("");
  const [hydrated, setHydrated] = useState(false);

  if (profile && !hydrated) {
    setFirstName(profile.firstName ?? user?.firstName ?? "");
    setLastName(profile.lastName ?? user?.lastName ?? "");
    setPhone(profile.phone ?? "");
    setTimezone(profile.timezone ?? "Asia/Kolkata");
    setLocale(profile.locale ?? "en-IN");
    setAvatarUrl(profile.avatarUrl ?? "");
    setHydrated(true);
  }

  if (profileQ.isLoading) {
    return <Skeleton className="h-64 w-full" />;
  }

  const initials =
    (firstName?.[0] ?? user?.email?.[0] ?? "?").toUpperCase() +
    (lastName?.[0] ?? "").toUpperCase();

  const next = async () => {
    try {
      await update.mutateAsync({
        firstName,
        lastName,
        phone,
        timezone,
        locale,
        avatarUrl: avatarUrl || undefined,
      });
      await setState.mutateAsync({ patch: { step: "workspace" } });
      await refreshUser();
      router.push("/onboarding/workspace");
    } catch (e: unknown) {
      toast.error("Couldn't save", e instanceof Error ? e.message : undefined);
    }
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 2 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">About you</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          A few details for your profile. You can change these later.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-4 p-6">
          <div className="flex items-center gap-4">
            <Avatar className="h-14 w-14">
              {avatarUrl ? <AvatarImage src={avatarUrl} alt={user?.email ?? ""} /> : null}
              <AvatarFallback className="text-base">{initials}</AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-1">
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">
                Avatar URL
              </Label>
              <Input
                value={avatarUrl}
                onChange={(e) => setAvatarUrl(e.target.value)}
                placeholder="https://…/me.jpg (optional)"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Field label="First name">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
            <Field label="Last name">
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </Field>
          </div>

          <Field label="Phone">
            <Input
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+91 98765 43210"
            />
          </Field>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Field label="Timezone" hint="IANA name, e.g. Asia/Kolkata">
              <Input value={timezone} onChange={(e) => setTimezone(e.target.value)} />
            </Field>
            <Field label="Locale" hint="BCP 47, e.g. en-IN, en-US">
              <Input value={locale} onChange={(e) => setLocale(e.target.value)} />
            </Field>
          </div>
        </CardContent>
      </Card>

      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={() => router.push("/onboarding")}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <Button onClick={next} disabled={update.isPending || setState.isPending}>
          {update.isPending || setState.isPending ? "Saving…" : "Continue"}
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function Field({
  label,
  children,
  hint,
}: {
  label: string;
  children: React.ReactNode;
  hint?: string;
}) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs uppercase tracking-wider text-muted-foreground">{label}</Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}
