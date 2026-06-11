"use client";

import { MessagesSquare } from "lucide-react";

// Landing pane when no channel is selected. The sidebar from the parent
// layout is still visible.
export default function CommunicationLanding() {
  return (
    <main
      className="flex-1 flex items-center justify-center bg-background"
      data-testid="comm-landing"
    >
      <div className="text-center max-w-sm">
        <MessagesSquare className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
        <h2 className="text-base font-semibold mb-1">Pick a channel</h2>
        <p className="text-sm text-muted-foreground">
          Select a channel from the left to start the conversation, or create a new one.
        </p>
      </div>
    </main>
  );
}
