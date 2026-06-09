import { Badge } from "@/components/ui/badge";

const VARIANTS: Record<string, "default" | "success" | "warning" | "danger" | "muted"> = {
  active: "success",
  trial: "warning",
  pending: "muted",
  paid: "success",
  open: "warning",
  draft: "muted",
  accepted: "success",
  expired: "danger",
  cancelled: "danger",
  past_due: "danger",
  void: "muted",
  refunded: "muted",
  recorded: "success",
  failed: "danger",
};

export function StatusBadge({ status }: { status: string }) {
  const variant = VARIANTS[status?.toLowerCase()] ?? "muted";
  return <Badge variant={variant}>{status}</Badge>;
}
