"use client";

import { ArrowLeft, ArrowRight, Plus, Trash2 } from "lucide-react";
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
import { useRoles } from "@/hooks/rbac/useRBACQueries";
import { useInviteUser } from "@/hooks/user/useUserQueries";
import { toast } from "@/hooks/use-toast";
import { useTenant } from "@/providers";

interface InviteRow {
  id: string;
  email: string;
  roleKey: string;
}

let rowSeq = 0;
const nextRowId = () => `r${++rowSeq}`;

export default function InvitesStep() {
  const router = useRouter();
  const { activeOrganization } = useTenant();
  const rolesQ = useRoles();
  const invite = useInviteUser();
  const setState = useSetOnboardingState();

  const defaultRole = "member";
  const [rows, setRows] = useState<InviteRow[]>([
    { id: nextRowId(), email: "", roleKey: defaultRole },
    { id: nextRowId(), email: "", roleKey: defaultRole },
  ]);

  const addRow = () =>
    setRows((r) => [...r, { id: nextRowId(), email: "", roleKey: defaultRole }]);
  const removeRow = (id: string) => setRows((r) => r.filter((x) => x.id !== id));
  const setEmail = (id: string, email: string) =>
    setRows((r) => r.map((x) => (x.id === id ? { ...x, email } : x)));
  const setRole = (id: string, roleKey: string) =>
    setRows((r) => r.map((x) => (x.id === id ? { ...x, roleKey } : x)));

  const validRows = rows.filter((r) => r.email.trim().includes("@"));

  const submitAndNext = async () => {
    if (!activeOrganization?.id) return;
    if (validRows.length > 0) {
      let sent = 0;
      let failed = 0;
      for (const r of validRows) {
        try {
          await invite.mutateAsync({
            email: r.email.trim(),
            organizationId: activeOrganization.id,
            roleKeys: [r.roleKey],
          });
          sent++;
        } catch {
          failed++;
        }
      }
      if (sent > 0)
        toast.success(`Sent ${sent} invite${sent === 1 ? "" : "s"}`, failed > 0 ? `${failed} failed` : undefined);
      else if (failed > 0) toast.error(`${failed} invite(s) failed`);
    }
    await setState.mutateAsync({ patch: { step: "plan" } });
    router.push("/onboarding/plan");
  };

  const skip = async () => {
    await setState.mutateAsync({ patch: { step: "plan" } });
    router.push("/onboarding/plan");
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">Step 4 of 6</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Invite your team</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          We&apos;ll email each person a link to set their password. You can skip this and invite later.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-3 p-6">
          {rows.map((row, idx) => (
            <div key={row.id} className="grid grid-cols-[1fr_180px_auto] gap-2">
              <Field label={idx === 0 ? "Email" : undefined}>
                <Input
                  type="email"
                  value={row.email}
                  onChange={(e) => setEmail(row.id, e.target.value)}
                  placeholder="teammate@company.com"
                />
              </Field>
              <Field label={idx === 0 ? "Role" : undefined}>
                <Select value={row.roleKey} onValueChange={(v) => setRole(row.id, v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {(rolesQ.data?.items ?? []).map((r) => (
                      <SelectItem key={r.key} value={r.key}>
                        {r.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>
              <div className={idx === 0 ? "pt-[26px]" : ""}>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removeRow(row.id)}
                  disabled={rows.length <= 1}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}

          <Button variant="outline" size="sm" onClick={addRow}>
            <Plus className="h-4 w-4" />
            Add another
          </Button>
        </CardContent>
      </Card>

      <div className="flex items-center justify-between gap-2">
        <Button variant="ghost" onClick={() => router.push("/onboarding/workspace")}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={skip} disabled={setState.isPending}>
            Skip
          </Button>
          <Button
            onClick={submitAndNext}
            disabled={invite.isPending || setState.isPending}
          >
            {validRows.length > 0
              ? `Send ${validRows.length} invite${validRows.length === 1 ? "" : "s"} & continue`
              : "Continue"}
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}

function Field({ label, children }: { label?: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-1.5">
      {label && (
        <Label className="text-xs uppercase tracking-wider text-muted-foreground">
          {label}
        </Label>
      )}
      {children}
    </div>
  );
}
