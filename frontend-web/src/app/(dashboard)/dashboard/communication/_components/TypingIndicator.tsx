"use client";

interface Props {
  typingUserIds: string[];
}

export function TypingIndicator({ typingUserIds }: Props) {
  if (typingUserIds.length === 0) return null;
  const label =
    typingUserIds.length === 1
      ? "Someone is typing…"
      : `${typingUserIds.length} people are typing…`;
  return (
    <div
      className="px-6 pb-1 text-xs text-muted-foreground italic h-5"
      data-testid="typing-indicator"
    >
      {label}
    </div>
  );
}
