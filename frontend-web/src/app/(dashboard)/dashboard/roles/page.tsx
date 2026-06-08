"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useRoles } from "@/hooks/rbac/useRBACQueries";

export default function RolesPage() {
  const rolesQ = useRoles();
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Roles & Permissions</h1>
        <p className="text-muted-foreground mt-1">
          Roles bundle permissions; assign roles to members from the user detail page.
        </p>
      </div>
      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {rolesQ.data?.map((r) => (
          <Card key={r.id}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">{r.name}</CardTitle>
                {r.isSystem && <Badge variant="muted">system</Badge>}
              </div>
              <p className="text-sm text-muted-foreground">{r.description}</p>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">priority {r.priority}</div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
