"use client";

import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
  usePager,
} from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/cn";

interface PaginationBarProps {
  /** 1-indexed current page. */
  page: number;
  /** Page size. */
  limit: number;
  /** Total row count from the server (use 0 while loading). */
  total: number;
  /** Called when the user clicks Previous/Next/a numeric page. */
  onPageChange: (nextPage: number) => void;
  /** Called when the user picks a new page size. Optional — pass to enable the size picker. */
  onLimitChange?: (nextLimit: number) => void;
  /** Available page sizes for the picker. Default: 10, 25, 50, 100. */
  pageSizes?: number[];
  /** Hide the "Showing X–Y of Z" line. */
  hideSummary?: boolean;
  className?: string;
}

/** Shared pagination footer for every list page in the dashboard.
 *
 * Renders a "Showing 1–25 of 432" summary, a page-size picker, and the
 * Previous/Next/numeric controls. Hides itself entirely when there's nothing
 * to paginate (single page + nothing loading).
 */
export function PaginationBar({
  page,
  limit,
  total,
  onPageChange,
  onLimitChange,
  pageSizes = [10, 25, 50, 100],
  hideSummary,
  className,
}: PaginationBarProps) {
  const { totalPages, pages, hasPrev, hasNext } = usePager(total, page, limit);
  if (total <= 0 && totalPages <= 1) return null;

  const start = total === 0 ? 0 : (page - 1) * limit + 1;
  const end = Math.min(total, page * limit);

  return (
    <div
      className={cn(
        "flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between",
        className,
      )}
    >
      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        {!hideSummary && (
          <span>
            Showing <span className="font-medium text-foreground">{start}</span>–
            <span className="font-medium text-foreground">{end}</span> of{" "}
            <span className="font-medium text-foreground">{total.toLocaleString()}</span>
          </span>
        )}
        {onLimitChange && (
          <div className="flex items-center gap-2">
            <span className="hidden sm:inline">Rows per page</span>
            <Select
              value={String(limit)}
              onValueChange={(v) => {
                const next = Number(v);
                if (Number.isFinite(next) && next > 0) onLimitChange(next);
              }}
            >
              <SelectTrigger className="h-8 w-[80px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {pageSizes.map((s) => (
                  <SelectItem key={s} value={String(s)}>
                    {s}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </div>

      <Pagination className="m-0 w-auto justify-end">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              disabled={!hasPrev}
              onClick={() => hasPrev && onPageChange(page - 1)}
            />
          </PaginationItem>
          {pages.map((p, i) =>
            p === "ellipsis" ? (
              <PaginationItem key={`e${i}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={p}>
                <PaginationLink isActive={p === page} onClick={() => onPageChange(p)}>
                  {p}
                </PaginationLink>
              </PaginationItem>
            ),
          )}
          <PaginationItem>
            <PaginationNext
              disabled={!hasNext}
              onClick={() => hasNext && onPageChange(page + 1)}
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
