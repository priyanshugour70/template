"use client";

import { ChevronLeft, ChevronRight, MoreHorizontal } from "lucide-react";
import * as React from "react";

import { type ButtonProps } from "@/components/ui/button";
import { cn } from "@/lib/cn";

export const Pagination = ({ className, ...props }: React.ComponentProps<"nav">) => (
  <nav
    role="navigation"
    aria-label="pagination"
    className={cn("mx-auto flex w-full justify-center", className)}
    {...props}
  />
);
Pagination.displayName = "Pagination";

export const PaginationContent = React.forwardRef<
  HTMLUListElement,
  React.ComponentProps<"ul">
>(({ className, ...props }, ref) => (
  <ul ref={ref} className={cn("flex flex-row items-center gap-1", className)} {...props} />
));
PaginationContent.displayName = "PaginationContent";

export const PaginationItem = React.forwardRef<
  HTMLLIElement,
  React.ComponentProps<"li">
>(({ className, ...props }, ref) => <li ref={ref} className={cn("", className)} {...props} />);
PaginationItem.displayName = "PaginationItem";

type PaginationLinkProps = {
  isActive?: boolean;
} & Pick<ButtonProps, "size"> &
  React.ComponentProps<"button">;

export const PaginationLink = ({
  className,
  isActive,
  size = "icon",
  ...props
}: PaginationLinkProps) => (
  <button
    type="button"
    aria-current={isActive ? "page" : undefined}
    className={cn(
      "inline-flex items-center justify-center gap-1 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
      size === "icon" ? "h-9 w-9" : "h-9 px-3",
      isActive
        ? "border border-input bg-background hover:bg-muted"
        : "hover:bg-muted hover:text-foreground",
      className,
    )}
    {...props}
  />
);
PaginationLink.displayName = "PaginationLink";

export const PaginationPrevious = ({
  className,
  ...props
}: React.ComponentProps<typeof PaginationLink>) => (
  <PaginationLink
    aria-label="Go to previous page"
    size="default"
    className={cn("gap-1 pl-2.5", className)}
    {...props}
  >
    <ChevronLeft className="h-4 w-4" />
    <span>Previous</span>
  </PaginationLink>
);
PaginationPrevious.displayName = "PaginationPrevious";

export const PaginationNext = ({
  className,
  ...props
}: React.ComponentProps<typeof PaginationLink>) => (
  <PaginationLink
    aria-label="Go to next page"
    size="default"
    className={cn("gap-1 pr-2.5", className)}
    {...props}
  >
    <span>Next</span>
    <ChevronRight className="h-4 w-4" />
  </PaginationLink>
);
PaginationNext.displayName = "PaginationNext";

export const PaginationEllipsis = ({ className, ...props }: React.ComponentProps<"span">) => (
  <span
    aria-hidden
    className={cn("flex h-9 w-9 items-center justify-center", className)}
    {...props}
  >
    <MoreHorizontal className="h-4 w-4" />
    <span className="sr-only">More pages</span>
  </span>
);
PaginationEllipsis.displayName = "PaginationEllipsis";

/**
 * `usePager` — given total/page/limit, returns the prev/next/list-of-pages
 * to render. Handles edge-trimming with ellipses for long page lists.
 */
export function usePager(total: number, page: number, limit: number, edge = 1, around = 1) {
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const pages: (number | "ellipsis")[] = [];
  const showFrom = Math.max(1 + edge, page - around);
  const showTo = Math.min(totalPages - edge, page + around);
  for (let i = 1; i <= Math.min(edge, totalPages); i++) pages.push(i);
  if (showFrom > edge + 1) pages.push("ellipsis");
  for (let i = showFrom; i <= showTo; i++) {
    if (i > edge && i <= totalPages - edge) pages.push(i);
  }
  if (showTo < totalPages - edge) pages.push("ellipsis");
  for (let i = Math.max(totalPages - edge + 1, edge + 1); i <= totalPages; i++) {
    if (!pages.includes(i)) pages.push(i);
  }
  return { totalPages, pages, hasPrev: page > 1, hasNext: page < totalPages };
}
