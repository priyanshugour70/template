import type { ReactNode } from "react";

import { PublicLayout } from "@/components/layouts/public/PublicLayout";

export default function PublicSegmentLayout({ children }: { children: ReactNode }) {
  return <PublicLayout>{children}</PublicLayout>;
}
