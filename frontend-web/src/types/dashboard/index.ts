// Wire format for GET /api/v1/dashboard/summary — mirror of Go's
// dashboard.Summary. Keep this in lockstep with model.go on the backend.

export interface DashboardKPIs {
  mrrCents: number;
  mrrDeltaPct: number;
  invoicedThisMonthCents: number;
  invoicedDeltaPct: number;
  activeUsers7d: number;
  activeUsersDeltaPct: number;
  outstandingDueCents: number;
  openInvoiceCount: number;
}

export interface RevenueBucket {
  month: string; // ISO timestamp
  issuedCents: number;
  paidCents: number;
}

export interface RequestBucket {
  day: string;
  requests: number;
  errors: number;
}

export interface StatusSlice {
  group: "2xx" | "3xx" | "4xx" | "5xx" | "other";
  count: number;
}

export interface EndpointBucket {
  route: string;
  method: string;
  count: number;
  avgLatencyMs: number;
}

export interface AgingBucket {
  bucket: "current" | "1-30" | "31-60" | "61-90" | "90+";
  count: number;
  totalDueCents: number;
}

export interface ActivityEntry {
  occurredAt: string;
  action: string;
  userEmail?: string;
  method?: string;
  path?: string;
  statusCode: number;
  targetType?: string;
}

export interface DashboardSummary {
  kpis: DashboardKPIs;
  revenueByMonth: RevenueBucket[];
  requestsByDay: RequestBucket[];
  statusBreakdown: StatusSlice[];
  topEndpoints: EndpointBucket[];
  invoiceAging: AgingBucket[];
  recentActivity: ActivityEntry[];
  generatedAt: string;
}
