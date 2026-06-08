"use client";

import { Check, ChevronDown, Building2 } from "lucide-react";
import { useState } from "react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { authService } from "@/services/auth";
import { useTenant } from "@/providers";

export function OrgSwitcher() {
  const { activeOrganization, organizations } = useTenant();
  const [switching, setSwitching] = useState<string | null>(null);

  async function switchTo(orgId: string) {
    if (!orgId || orgId === activeOrganization?.id) return;
    setSwitching(orgId);
    try {
      const res = await authService.switchOrg({ organizationId: orgId });
      if (res.success) {
        window.location.assign("/dashboard");
      }
    } finally {
      setSwitching(null);
    }
  }

  if (!activeOrganization) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="flex items-center gap-2 rounded-md border border-border bg-background px-3 h-9 text-sm hover:bg-accent transition-colors">
          {activeOrganization.logoUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={activeOrganization.logoUrl} alt="" className="h-5 w-5 rounded-sm" />
          ) : (
            <Building2 className="h-4 w-4 text-muted-foreground" />
          )}
          <span className="hidden sm:inline max-w-[180px] truncate">
            {activeOrganization.name}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[240px]">
        <DropdownMenuLabel>Switch organization</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {organizations.length === 0 ? (
          <div className="px-2 py-3 text-sm text-muted-foreground text-center">
            No other organizations
          </div>
        ) : (
          organizations.map((o) => {
            const isActive = o.id === activeOrganization.id;
            return (
              <DropdownMenuItem
                key={o.id}
                onSelect={() => !isActive && void switchTo(o.id)}
                disabled={switching === o.id}
                className="gap-3"
              >
                {o.logoUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={o.logoUrl} alt="" className="h-6 w-6 rounded-sm" />
                ) : (
                  <div className="h-6 w-6 rounded-sm bg-primary/10 flex items-center justify-center">
                    <Building2 className="h-3 w-3 text-primary" />
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <div className="truncate font-medium">{o.name}</div>
                  <div className="text-xs text-muted-foreground truncate">/{o.slug}</div>
                </div>
                {isActive && <Check className="h-4 w-4 text-primary" />}
              </DropdownMenuItem>
            );
          })
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
