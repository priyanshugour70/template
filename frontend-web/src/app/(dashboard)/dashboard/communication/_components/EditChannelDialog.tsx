"use client";

import { useEffect, useState } from "react";

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
import {
  useArchiveChannel,
  useUpdateChannel,
} from "@/hooks/communication/useCommunication";
import type { ConversationView } from "@/types/communication";

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;
  conversation: ConversationView;
}

// EditChannelDialog edits name, topic, description, and visibility. Slug is
// immutable (would break existing links) — Phase 5 can add slug rename with
// HTTP 301-style redirect support.
export function EditChannelDialog({ open, onOpenChange, conversation }: Props) {
  const update = useUpdateChannel(conversation.id);
  const archive = useArchiveChannel(conversation.id);

  const [name, setName] = useState(conversation.name ?? "");
  const [topic, setTopic] = useState(conversation.topic ?? "");
  const [description, setDescription] = useState(conversation.description ?? "");
  const [isPrivate, setIsPrivate] = useState<boolean>(conversation.isPrivate);

  useEffect(() => {
    if (!open) return;
    setName(conversation.name ?? "");
    setTopic(conversation.topic ?? "");
    setDescription(conversation.description ?? "");
    setIsPrivate(conversation.isPrivate);
  }, [open, conversation]);

  async function submit() {
    await update.mutateAsync({
      name: name.trim() || undefined,
      topic,
      description,
      isPrivate,
    });
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit channel</DialogTitle>
          <DialogDescription>
            Update the channel name and metadata. The slug stays the same.
          </DialogDescription>
        </DialogHeader>

        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            void submit();
          }}
        >
          <div className="space-y-1">
            <Label htmlFor="edit-channel-name">Name</Label>
            <Input
              id="edit-channel-name"
              data-testid="edit-channel-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={200}
              required
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="edit-channel-topic">Topic</Label>
            <Input
              id="edit-channel-topic"
              data-testid="edit-channel-topic"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              maxLength={500}
              placeholder="What is this channel about?"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="edit-channel-desc">Description</Label>
            <textarea
              id="edit-channel-desc"
              data-testid="edit-channel-desc"
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={2000}
            />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={isPrivate}
              onChange={(e) => setIsPrivate(e.target.checked)}
              data-testid="edit-channel-private"
            />
            Private channel
          </label>

          <DialogFooter className="flex sm:justify-between sm:flex-row-reverse">
            <div className="flex gap-2">
              <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button
                type="submit"
                data-testid="edit-channel-submit"
                disabled={update.isPending || !name.trim()}
              >
                {update.isPending ? "Saving…" : "Save changes"}
              </Button>
            </div>
            <Button
              type="button"
              variant="ghost"
              className="text-destructive hover:text-destructive"
              data-testid="edit-channel-archive"
              onClick={async () => {
                if (!confirm("Archive this channel? It will disappear from sidebars.")) return;
                await archive.mutateAsync();
                onOpenChange(false);
                if (typeof window !== "undefined") {
                  window.location.assign("/dashboard/communication");
                }
              }}
              disabled={archive.isPending}
            >
              {archive.isPending ? "Archiving…" : "Archive"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
