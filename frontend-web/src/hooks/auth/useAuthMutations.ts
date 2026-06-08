"use client";

import { useMutation } from "@tanstack/react-query";

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
