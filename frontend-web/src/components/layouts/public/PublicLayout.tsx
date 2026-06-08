"use client";

import Link from "next/link";
import type { ReactNode } from "react";

import { useTenant } from "@/providers";

export function PublicLayout({ children }: { children: ReactNode }) {
  const { tenant } = useTenant();
  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground">
      <header className="border-b">
        <div className="container mx-auto flex h-16 items-center justify-between px-6">
          <Link href="/" className="flex items-center gap-2">
            {tenant?.logoUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={tenant.logoUrl} alt={tenant.name} className="h-7 w-7 rounded-md" />
            ) : (
              <div className="h-7 w-7 rounded-md bg-primary/10" />
            )}
            <span className="font-semibold">{tenant?.name ?? "App"}</span>
          </Link>
          <nav className="hidden md:flex items-center gap-6 text-sm">
            <Link href="/pricing" className="hover:text-primary">Pricing</Link>
            <Link href="/auth/login" className="hover:text-primary">Sign in</Link>
          </nav>
        </div>
      </header>
      <div className="flex-1">{children}</div>
      <footer className="border-t">
        <div className="container mx-auto flex h-16 items-center justify-between px-6 text-sm text-muted-foreground">
          <div>&copy; {new Date().getFullYear()} {tenant?.name ?? "App"}</div>
          <div className="flex gap-4">
            <Link href="/terms" className="hover:text-foreground">Terms</Link>
            <Link href="/privacy" className="hover:text-foreground">Privacy</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
