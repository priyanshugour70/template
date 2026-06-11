"use client";

import { useState } from "react";

import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from "@/components/ui";

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;
  onCreate(slug: string, name: string): Promise<void> | void;
  pending: boolean;
}

// Slug + name only — topic/description/private are Phase 4 polish.
export function NewChannelDialog({ open, onOpenChange, onCreate, pending }: Props) {
  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");

  function reset() {
    setSlug("");
    setName("");
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        onOpenChange(o);
        if (!o) reset();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create a channel</DialogTitle>
          <DialogDescription>
            Channels group conversations around a topic. Slug must be unique inside your org.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (slug.trim() && name.trim()) {
              void onCreate(slug.trim(), name.trim());
            }
          }}
        >
          <div className="space-y-1">
            <Label htmlFor="channel-slug">Slug</Label>
            <Input
              id="channel-slug"
              data-testid="channel-slug-input"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="general"
              minLength={2}
              maxLength={64}
              autoFocus
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="channel-name">Display name</Label>
            <Input
              id="channel-name"
              data-testid="channel-name-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="General"
              maxLength={200}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              data-testid="channel-create-submit"
              disabled={pending || !slug.trim() || !name.trim()}
            >
              {pending ? "Creating…" : "Create channel"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
