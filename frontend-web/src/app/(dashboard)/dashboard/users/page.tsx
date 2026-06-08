"use client";

import { useState } from "react";
import { UserPlus } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useUsers } from "@/hooks/user/useUserQueries";
import { usePermissions } from "@/providers";

export default function UsersPage() {
  const [search, setSearch] = useState("");
  const usersQ = useUsers({ q: search || undefined, limit: 50 });
  const { has } = usePermissions();

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
          <p className="text-muted-foreground mt-1">
            People who have access to this organization.
          </p>
        </div>
        {has("user.invite") && (
          <Button>
            <UserPlus className="h-4 w-4" /> Invite user
          </Button>
        )}
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">All users</CardTitle>
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name or email…"
            className="max-w-xs"
          />
        </CardHeader>
        <CardContent>
          {usersQ.isLoading ? (
            <div className="text-sm text-muted-foreground py-8 text-center">Loading…</div>
          ) : usersQ.isError ? (
            <div className="text-sm text-destructive py-8 text-center">
              Failed to load users.
            </div>
          ) : !usersQ.data?.length ? (
            <div className="text-sm text-muted-foreground py-8 text-center">No users found.</div>
          ) : (
            <div className="rounded-md border">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/40">
                  <tr>
                    <th className="text-left p-3 font-medium">Name</th>
                    <th className="text-left p-3 font-medium">Email</th>
                    <th className="text-left p-3 font-medium">Status</th>
                    <th className="text-left p-3 font-medium">Last login</th>
                  </tr>
                </thead>
                <tbody>
                  {usersQ.data.map((u) => (
                    <tr key={u.id} className="border-b last:border-0">
                      <td className="p-3">
                        {u.displayName || `${u.firstName ?? ""} ${u.lastName ?? ""}`.trim() || "—"}
                      </td>
                      <td className="p-3 text-muted-foreground">{u.email}</td>
                      <td className="p-3">
                        <Badge variant={u.status === "active" ? "success" : u.status === "suspended" ? "danger" : "muted"}>
                          {u.status}
                        </Badge>
                      </td>
                      <td className="p-3 text-muted-foreground">
                        {u.lastLoginAt ? new Date(u.lastLoginAt).toLocaleString() : "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
