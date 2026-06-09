"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { authService } from "@/services/auth";
import type {
  AcceptInviteRequest,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  LoginRequest,
  ResetPasswordRequest,
  SessionResponse,
  SwitchOrgRequest,
} from "@/types/auth";

function ok<T>(p: Promise<{ success: boolean; data?: T; error?: { message?: string } }>) {
  return p.then((r) => {
    if (!r.success || !r.data) throw new Error(r.error?.message ?? "request failed");
    return r.data;
  });
}

export function useDiscoverMutation() {
  return useMutation({
    mutationFn: (email: string) => ok(authService.discover(email)),
  });
}

export function useLoginMutation() {
  return useMutation<SessionResponse, Error, LoginRequest>({
    mutationFn: (req) => ok(authService.login(req)),
  });
}

export function useSwitchOrgMutation() {
  return useMutation<SessionResponse, Error, SwitchOrgRequest>({
    mutationFn: (req) => ok(authService.switchOrg(req)),
  });
}

export function useAcceptInviteMutation() {
  return useMutation<SessionResponse, Error, AcceptInviteRequest>({
    mutationFn: (req) => ok(authService.acceptInvite(req)),
  });
}

export function useForgotPasswordMutation() {
  return useMutation<unknown, Error, ForgotPasswordRequest>({
    mutationFn: (req) => ok(authService.forgotPassword(req)),
  });
}

export function useResetPasswordMutation() {
  return useMutation<unknown, Error, ResetPasswordRequest>({
    mutationFn: (req) => ok(authService.resetPassword(req)),
  });
}

export function useChangePasswordMutation() {
  return useMutation<unknown, Error, ChangePasswordRequest>({
    mutationFn: (req) => ok(authService.changePassword(req)),
  });
}

export function useSessions(opts: { page?: number; limit?: number } = {}) {
  return useQuery({
    queryKey: ["auth", "sessions", opts] as const,
    queryFn: async () => {
      const res = await authService.listSessions(opts);
      if (!res.success) throw new Error(res.error?.message ?? "sessions failed");
      return {
        items: res.data ?? [],
        total: res.pagination?.total ?? (res.data?.length ?? 0),
        page: res.pagination?.page ?? 1,
        limit: res.pagination?.limit ?? (opts.limit ?? 200),
      };
    },
    placeholderData: (prev) => prev,
  });
}

export function useRevokeSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (jti: string) => {
      const res = await authService.revokeSession(jti);
      if (!res.success) throw new Error(res.error?.message ?? "revoke failed");
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["auth", "sessions"] }),
  });
}
