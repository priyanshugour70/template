import { api } from "@/lib/client";
import type {
  CancelSubscriptionRequest,
  ChangePlanRequest,
  ChangePlanResponse,
  FeatureSet,
  Invoice,
  PauseRequest,
  Plan,
  PreviewChangeRequest,
  PreviewChangeResponse,
  Subscription,
  UpdateBillingRequest,
  UsageCounter,
  ValidateCouponRequest,
  ValidateCouponResponse,
} from "@/types/subscription";

// Backend list endpoints are paginated (default 25, max 200). Defaulting to
// limit=200 keeps existing dashboards rendering full lists in typical tenants.
export const subscriptionService = {
  listPlans: (params?: { page?: number; limit?: number }) =>
    api.get<Plan[]>("/subscription-plans", { query: { limit: 200, ...params } }),
  getActive: () => api.get<Subscription>("/subscriptions/active"),
  features: () => api.get<FeatureSet>("/subscriptions/features"),
  listUsage: (params?: { page?: number; limit?: number }) =>
    api.get<UsageCounter[]>("/subscriptions/usage", { query: { limit: 200, ...params } }),

  // lifecycle
  changePlan: (req: ChangePlanRequest) =>
    api.post<ChangePlanResponse>("/subscriptions/change", req),
  previewChange: (req: PreviewChangeRequest) =>
    api.post<PreviewChangeResponse>("/subscriptions/preview-change", req),
  cancel: (req: CancelSubscriptionRequest) =>
    api.post<unknown>("/subscriptions/cancel", req),
  reactivate: () => api.post<Subscription>("/subscriptions/reactivate"),
  pause: (req: PauseRequest = {}) => api.post<Subscription>("/subscriptions/pause", req),
  resume: () => api.post<Subscription>("/subscriptions/resume"),
  updateBilling: (req: UpdateBillingRequest) =>
    api.patch<Subscription>("/subscriptions/billing", req),

  // invoices
  listInvoices: (limit = 50) =>
    api.get<Invoice[]>("/subscriptions/invoices", { query: { limit } }),
  getInvoice: (id: string) => api.get<Invoice>(`/subscriptions/invoices/${id}`),

  // coupons
  validateCoupon: (req: ValidateCouponRequest) =>
    api.post<ValidateCouponResponse>("/subscriptions/coupons/validate", req),
};
