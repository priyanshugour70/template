"use client";

import { ArrowLeft, ArrowRight, Building2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSetOnboardingState } from "@/hooks/onboarding/useOnboarding";
import { useUpdateOrganization } from "@/hooks/tenant/useTenantQueries";
import { toast } from "@/hooks/use-toast";
import { cn } from "@/lib/cn";
import { useTenant } from "@/providers";

const SIZE_OPTIONS = [
  { value: "1-10", label: "1 – 10 people" },
  { value: "11-50", label: "11 – 50 people" },
  { value: "51-200", label: "51 – 200 people" },
  { value: "201-1000", label: "201 – 1 000 people" },
  { value: "1000+", label: "1 000+ people" },
];

const INDUSTRY_OPTIONS = [
  "SaaS",
  "E-commerce",
  "Healthcare",
  "Finance",
  "Education",
  "Manufacturing",
  "Media",
  "Government",
  "Non-profit",
  "Other",
];

export default function WorkspaceStep() {
  const router = useRouter();
  const { activeOrganization } = useTenant();
  const orgId = activeOrganization?.id ?? "";
  const update = useUpdateOrganization(orgId);
  const setState = useSetOnboardingState();

  const [name, setName] = useState(activeOrganization?.name ?? "");
  const [description, setDescription] = useState("");
  const [industry, setIndustry] = useState("");
  const [size, setSize] = useState("");
  const [primaryColor, setPrimaryColor] = useState("");

  if (!activeOrganization) {
    return (
      <div className="text-sm text-muted-foreground">No active organization.</div>
    );
  }

  const next = async () => {
    if (!name) {
      toast.error("Give your workspace a name");
      return;
    }
    try {
      await update.mutateAsync({
        name,
        description: description || undefined,
        industry: industry || undefined,
        size: size || undefined,
        primaryColor: primaryColor || undefined,
      });
      await setState.mutateAsync({ patch: { step: "invites" } });
      router.push("/onboarding/invites");
    } catch (e: unknown) {
      toast.error("Couldn't save", e instanceof Error ? e.message : undefined);
    }
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 3 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Brand your workspace</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Teammates will see this name, colour, and logo throughout the product.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-4 p-6">
          <div className="flex items-center gap-4">
            <div
              className="flex h-14 w-14 items-center justify-center rounded-lg"
              style={{
                background: primaryColor ? `${primaryColor}1a` : undefined,
              }}
            >
              <Building2
                className="h-6 w-6"
                style={{ color: primaryColor || undefined }}
              />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium">Live preview</p>
              <p className="truncate text-lg font-semibold">{name || "Your workspace"}</p>
            </div>
          </div>

          <Field label="Workspace name">
            <Input value={name} onChange={(e) => setName(e.target.value)} required />
          </Field>
          <Field label="What does this workspace do? (optional)">
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="A short tagline"
            />
          </Field>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Field label="Industry">
              <Select value={industry || "_none"} onValueChange={(v) => setIndustry(v === "_none" ? "" : v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none">—</SelectItem>
                  {INDUSTRY_OPTIONS.map((i) => (
                    <SelectItem key={i} value={i}>
                      {i}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field label="Team size">
              <Select value={size || "_none"} onValueChange={(v) => setSize(v === "_none" ? "" : v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none">—</SelectItem>
                  {SIZE_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </div>

          <Field label="Primary colour">
            <div className="flex items-center gap-2">
              <input
                type="color"
                value={primaryColor || "#2563eb"}
                onChange={(e) => setPrimaryColor(e.target.value)}
                className="h-9 w-12 cursor-pointer rounded border border-input bg-background"
              />
              <Input
                value={primaryColor}
                onChange={(e) => setPrimaryColor(e.target.value)}
                placeholder="#2563eb"
                className={cn("font-mono")}
              />
            </div>
          </Field>
        </CardContent>
      </Card>

      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={() => router.push("/onboarding/profile")}>
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

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1.5">
      <Label className="text-xs uppercase tracking-wider text-muted-foreground">{label}</Label>
      {children}
    </div>
  );
}
