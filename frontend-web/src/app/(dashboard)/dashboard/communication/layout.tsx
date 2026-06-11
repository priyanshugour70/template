import type { ReactNode } from "react";

import { ChannelSidebar } from "./_components/ChannelSidebar";

// Layout owns the persistent left rail. Selecting a channel navigates to a
// child route that fills the right-hand pane.
export default function CommunicationLayout({ children }: { children: ReactNode }) {
  return (
    <div
      className="flex h-[calc(100vh-12rem)] rounded-md border border-border overflow-hidden"
      data-testid="comm-layout"
    >
      <ChannelSidebar />
      {children}
    </div>
  );
}
