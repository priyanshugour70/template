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

export const subscriptionService = {
  listPlans: () => api.get<Plan[]>("/subscription-plans"),
  getActive: () => api.get<Subscription>("/subscriptions/active"),
  features: () => api.get<FeatureSet>("/subscriptions/features"),
  listUsage: () => api.get<UsageCounter[]>("/subscriptions/usage"),

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
