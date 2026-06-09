"use client";

import { Bell, Building2, ChevronRight, Code2, Lock, Monitor, User } from "lucide-react";
import Link from "next/link";

import { Card, CardContent, CardHeader } from "@/components/ui";
import { usePermissions } from "@/providers";

interface Section {
  href: string;
  title: string;
  description: string;
  icon: typeof User;
  anyPermission?: string[];
}

const SECTIONS: Section[] = [
  {
    href: "/dashboard/settings/profile",
    title: "Profile",
    description: "Your name, avatar, contact and bio.",
    icon: User,
  },
  {
    href: "/dashboard/settings/security",
    title: "Security",
    description: "Email, password, and multi-factor authentication.",
    icon: Lock,
  },
  {
    href: "/dashboard/settings/sessions",
    title: "Sessions",
    description: "Devices currently signed in to your account.",
    icon: Monitor,
  },
  {
    href: "/dashboard/settings/notifications",
    title: "Notifications",
    description: "Email digests, mentions, billing, and the in-app bell.",
    icon: Bell,
  },
  {
    href: "/dashboard/settings/developer",
    title: "Developer",
    description: "API keys and webhooks for integrations.",
    icon: Code2,
    anyPermission: ["api_key.list", "webhook.list"],
  },
  {
    href: "/dashboard/settings/tenant",
    title: "Tenant",
    description: "Workspace-wide branding and contact details.",
    icon: Building2,
    anyPermission: ["tenant.update"],
  },
];

export default function SettingsOverviewPage() {
  const { hasAny } = usePermissions();
  const visible = SECTIONS.filter((s) => !s.anyPermission || hasAny(s.anyPermission));

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Pick a section above, or jump in below.
      </p>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {visible.map(({ href, title, description, icon: Icon }) => (
          <Link key={href} href={href} className="group block">
            <Card className="h-full transition-colors hover:bg-muted/30">
              <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2">
                <div className="flex items-center gap-2">
                  <Icon
                    className="h-4 w-4 text-muted-foreground"
                    strokeWidth={1.5}
                    aria-hidden
                  />
                  <span className="text-sm font-medium group-hover:underline">{title}</span>
                </div>
                <ChevronRight
                  className="h-4 w-4 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
                  aria-hidden
                />
              </CardHeader>
              <CardContent className="pt-0">
                <p className="text-xs text-muted-foreground">{description}</p>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
