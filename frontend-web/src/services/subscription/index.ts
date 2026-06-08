import { api } from "@/lib/client";
import type {
  CancelSubscriptionRequest,
  ChangePlanRequest,
  FeatureSet,
  Plan,
  Subscription,
  UsageCounter,
} from "@/types/subscription";

export const subscriptionService = {
  listPlans: () => api.get<Plan[]>("/subscription-plans"),
  getActive: () => api.get<Subscription>("/subscriptions/active"),
  changePlan: (req: ChangePlanRequest) => api.post<Subscription>("/subscriptions/change", req),
  cancel: (req: CancelSubscriptionRequest) => api.post<unknown>("/subscriptions/cancel", req),
  features: () => api.get<FeatureSet>("/subscriptions/features"),
  listUsage: () => api.get<UsageCounter[]>("/subscriptions/usage"),
};
