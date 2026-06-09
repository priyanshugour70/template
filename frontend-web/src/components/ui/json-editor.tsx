"use client";

import { AlertCircle, Check } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/cn";

export interface JsonEditorProps {
  /** Current parsed value. */
  value: Record<string, unknown> | unknown[] | null | undefined;
  /** Called with the parsed object on every valid edit. */
  onChange: (next: Record<string, unknown>) => void;
  disabled?: boolean;
  className?: string;
  rows?: number;
  placeholder?: string;
}

/**
 * JsonEditor — textarea-based JSON editor with parse validation. Renders the
 * incoming `value` as pretty-printed JSON on mount; emits the parsed object
 * via onChange after every successful parse. Shows an inline error when the
 * current draft doesn't parse.
 *
 * Stays uncontrolled internally to avoid clobbering the user's edits while
 * they're mid-typing.
 */
export function JsonEditor({
  value,
  onChange,
  disabled,
  className,
  rows = 10,
  placeholder = "{}",
}: JsonEditorProps) {
  const initial = React.useMemo(() => formatValue(value), []);
  const [draft, setDraft] = React.useState<string>(initial);
  const [error, setError] = React.useState<string | null>(null);

  // When the upstream value changes (e.g. after a successful save), refresh
  // the draft — but only when the draft hasn't been edited away from the
  // current upstream value (i.e. it parses to the same shape).
  React.useEffect(() => {
    const incoming = formatValue(value);
    try {
      const draftParsed = JSON.parse(draft);
      const incomingParsed = JSON.parse(incoming);
      if (deepEqual(draftParsed, incomingParsed)) return;
    } catch {
      // current draft is invalid — leave it alone so the user can fix
      return;
    }
    setDraft(incoming);
    setError(null);
  }, [value]); // eslint-disable-line react-hooks/exhaustive-deps

  const onChangeRaw = (raw: string) => {
    setDraft(raw);
    if (raw.trim() === "") {
      setError(null);
      onChange({});
      return;
    }
    try {
      const parsed = JSON.parse(raw);
      if (typeof parsed !== "object" || parsed === null) {
        setError("Must be a JSON object or array");
        return;
      }
      setError(null);
      onChange(parsed as Record<string, unknown>);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Invalid JSON");
    }
  };

  return (
    <div className={cn("space-y-1.5", className)}>
      <textarea
        value={draft}
        onChange={(e) => onChangeRaw(e.target.value)}
        disabled={disabled}
        rows={rows}
        spellCheck={false}
        placeholder={placeholder}
        className={cn(
          "w-full rounded-md border bg-background px-3 py-2 font-mono text-xs shadow-sm",
          "focus:outline-none focus:ring-2 focus:ring-ring",
          "disabled:cursor-not-allowed disabled:opacity-50",
          error ? "border-destructive" : "border-input",
        )}
      />
      {error ? (
        <p className="flex items-center gap-1 text-xs text-destructive">
          <AlertCircle className="h-3 w-3" />
          {error}
        </p>
      ) : (
        <p className="flex items-center gap-1 text-xs text-muted-foreground">
          <Check className="h-3 w-3 text-success" />
          Valid JSON
        </p>
      )}
    </div>
  );
}

function formatValue(v: unknown): string {
  if (v == null) return "{}";
  try {
    return JSON.stringify(v, null, 2);
  } catch {
    return "{}";
  }
}

function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (typeof a !== typeof b) return false;
  if (typeof a !== "object" || a === null || b === null) return false;
  const aKeys = Object.keys(a as object);
  const bKeys = Object.keys(b as object);
  if (aKeys.length !== bKeys.length) return false;
  for (const k of aKeys) {
    if (!deepEqual((a as Record<string, unknown>)[k], (b as Record<string, unknown>)[k])) {
      return false;
    }
  }
  return true;
}
