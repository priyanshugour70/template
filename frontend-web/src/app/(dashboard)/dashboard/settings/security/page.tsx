"use client";

import { AccountSection, MFASection } from "../_components/sections";

export default function SettingsSecurityPage() {
  return (
    <div className="space-y-4">
      <AccountSection />
      <MFASection />
    </div>
  );
}
