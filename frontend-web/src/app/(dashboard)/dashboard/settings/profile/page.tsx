"use client";

import { ProfileSection, RegionalSection } from "../_components/sections";

export default function SettingsProfilePage() {
  return (
    <div className="space-y-4">
      <ProfileSection />
      <RegionalSection />
    </div>
  );
}
