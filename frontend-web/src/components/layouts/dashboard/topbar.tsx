"use client";

import { ThemePicker } from "./theme-picker";
import { UserMenu } from "./user-menu";

export function Topbar() {
  return (
    <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b bg-background/80 px-6 backdrop-blur-sm md:ml-64">
      <div className="flex-1" />
      <div className="flex items-center gap-3">
        <ThemePicker />
        <UserMenu />
      </div>
    </header>
  );
}
