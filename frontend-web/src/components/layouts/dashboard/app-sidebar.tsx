"use client";

import { ChevronDown, GalleryVerticalEnd } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useMemo, useState, type ReactNode } from "react";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/cn";
import { usePermissions, useTenant } from "@/providers";

interface NavItem {
  href: string;
  label: string;
  icon: string;
  permission?: string;
  /** If set, user needs any one of these permissions (overrides `permission`). */
  anyPermission?: string[];
}

interface NavSection {
  id: string;
  label: string;
  items: NavItem[];
  collapsible: boolean;
}

function activeHrefInSection(pathname: string, items: NavItem[]): string | undefined {
  const matches = items
    .map((i) => i.href)
    .filter((h) => pathname === h || pathname.startsWith(`${h}/`));
  if (matches.length === 0) return undefined;
  return matches.reduce((a, b) => (a.length >= b.length ? a : b));
}

const sections: NavSection[] = [
  {
    id: "main",
    label: "",
    collapsible: false,
    items: [{ href: "/dashboard", label: "Home", icon: "grid" }],
  },
  {
    id: "administrative",
    label: "Administrative",
    collapsible: true,
    items: [
      { href: "/dashboard/administrative/users", label: "Users", icon: "users", permission: "user.list" },
      { href: "/dashboard/administrative/roles", label: "Roles & Permissions", icon: "key", permission: "role.list" },
      { href: "/dashboard/administrative/departments", label: "Departments", icon: "git-branch", permission: "department.list" },
      { href: "/dashboard/administrative/groups", label: "Groups", icon: "users-round", permission: "group.list" },
      { href: "/dashboard/administrative/organizations", label: "Organizations", icon: "building", permission: "org.list" },
      { href: "/dashboard/administrative/audit", label: "Audit log", icon: "file-text", permission: "audit.read" },
    ],
  },
  {
    id: "billing",
    label: "Billing",
    collapsible: true,
    items: [
      { href: "/dashboard/subscription", label: "Subscription", icon: "credit-card", permission: "subscription.read" },
    ],
  },
  {
    id: "settings",
    label: "System",
    collapsible: true,
    items: [{ href: "/dashboard/settings", label: "Settings", icon: "settings" }],
  },
];

function NavIcon({ name }: { name: string }) {
  const paths: Record<string, ReactNode> = {
    grid: <path d="M4 4h6v6H4V4zm10 0h6v6h-6V4zM4 14h6v6H4v-6zm10 0h6v6h-6v-6z" />,
    users: <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" />,
    key: <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />,
    building: <path d="M3 21h18M6 21V10l6-3 6 3v11M9 21v-4h6v4M10 14h4M10 10h4" />,
    "file-text": <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6zM14 2v6h6M16 13H8M16 17H8M10 9H8" />,
    "credit-card": <path d="M21 4H3a2 2 0 00-2 2v12a2 2 0 002 2h18a2 2 0 002-2V6a2 2 0 00-2-2zM1 10h22" />,
    settings: <path d="M12 15a3 3 0 100-6 3 3 0 000 6zM19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 008 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H2a2 2 0 010-4h.09A1.65 1.65 0 004.6 8a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 112.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V2a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 112.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H22a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z" />,
    "git-branch": (
      <>
        <line x1="6" y1="3" x2="6" y2="15" />
        <circle cx="18" cy="6" r="3" />
        <circle cx="6" cy="18" r="3" />
        <path d="M18 9a9 9 0 01-9 9" />
      </>
    ),
    "users-round": <path d="M18 21a8 8 0 00-16 0M10 14a5 5 0 100-10 5 5 0 000 10zM22 21a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" />,
  };

  return (
    <svg
      className="h-4 w-4 shrink-0"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {paths[name] ?? paths.grid}
    </svg>
  );
}

export function AppSidebar() {
  const pathname = usePathname();
  const { has, hasAny } = usePermissions();
  const { tenant, activeOrganization } = useTenant();
  const { isMobile, setOpenMobile } = useSidebar();

  /** Collapsible groups: open by default; user toggles persist within session. */
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({});

  const closeMobile = useCallback(() => {
    if (isMobile) setOpenMobile(false);
  }, [isMobile, setOpenMobile]);

  const filteredSections = useMemo(
    () =>
      sections
        .map((s) => ({
          ...s,
          items: s.items.filter((i) => {
            if (i.anyPermission?.length) return hasAny(i.anyPermission);
            return !i.permission || has(i.permission);
          }),
        }))
        .filter((s) => s.items.length > 0),
    [has, hasAny],
  );

  return (
    <Sidebar collapsible="offcanvas">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link href="/dashboard" onClick={closeMobile}>
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  {tenant?.logoUrl ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={tenant.logoUrl}
                      alt={tenant.name}
                      className="size-8 rounded-lg object-cover"
                    />
                  ) : (
                    <GalleryVerticalEnd className="size-4" aria-hidden />
                  )}
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-medium">{tenant?.name ?? "Workspace"}</span>
                  <span className="text-xs text-sidebar-foreground/70">
                    {activeOrganization?.name ?? "—"}
                  </span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarMenu>
            {filteredSections.flatMap((section) => {
              const activeInSection = activeHrefInSection(pathname, section.items);

              if (!section.collapsible) {
                return section.items.map((item) => (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      asChild
                      isActive={activeInSection === item.href}
                      tooltip={item.label}
                    >
                      <Link href={item.href} onClick={closeMobile}>
                        <NavIcon name={item.icon} />
                        <span>{item.label}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ));
              }

              const isOpen = openSections[section.id] ?? true;
              return [
                <SidebarMenuItem key={section.id}>
                  <Collapsible
                    open={isOpen}
                    onOpenChange={(open) =>
                      setOpenSections((p) => ({ ...p, [section.id]: open }))
                    }
                  >
                    <CollapsibleTrigger asChild>
                      <SidebarMenuButton
                        type="button"
                        className="w-full justify-between gap-1 pr-1.5 h-auto min-h-8 py-1.5"
                        tooltip={section.label}
                      >
                        <span className="font-medium text-left flex-1 truncate">
                          {section.label}
                        </span>
                        <ChevronDown
                          className={cn(
                            "size-4 shrink-0 text-sidebar-foreground/70 transition-transform duration-200",
                            isOpen && "rotate-180",
                          )}
                          aria-hidden
                        />
                      </SidebarMenuButton>
                    </CollapsibleTrigger>
                    <CollapsibleContent className="group-data-[collapsible=icon]:hidden">
                      <SidebarMenuSub>
                        {section.items.map((item) => (
                          <SidebarMenuSubItem key={item.href}>
                            <SidebarMenuSubButton
                              asChild
                              isActive={activeInSection === item.href}
                            >
                              <Link href={item.href} onClick={closeMobile}>
                                <NavIcon name={item.icon} />
                                <span>{item.label}</span>
                              </Link>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </Collapsible>
                </SidebarMenuItem>,
              ];
            })}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
