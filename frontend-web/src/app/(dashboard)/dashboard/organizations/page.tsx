"use client";

import { Building2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useOrganizations } from "@/hooks/tenant/useTenantQueries";

export default function OrganizationsPage() {
  const orgsQ = useOrganizations();
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Organizations</h1>
        <p className="text-muted-foreground mt-1">
          Workspaces inside your tenant. Each has its own users, roles, and subscription.
        </p>
      </div>
      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {orgsQ.data?.map((o) => (
          <Card key={o.id}>
            <CardHeader>
              <div className="flex items-center gap-3">
                {o.logoUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={o.logoUrl} alt={o.name} className="h-10 w-10 rounded-md" />
                ) : (
                  <div className="h-10 w-10 rounded-md bg-primary/10 flex items-center justify-center">
                    <Building2 className="h-5 w-5 text-primary" />
                  </div>
                )}
                <div className="flex-1">
                  <CardTitle className="text-base">{o.name}</CardTitle>
                  <p className="text-xs text-muted-foreground">/{o.slug}</p>
                </div>
                {o.isDefault && <Badge>default</Badge>}
              </div>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              {o.description ?? "—"}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
