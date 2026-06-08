import { api } from "@/lib/client";
import type { SessionResponse } from "@/types/auth";

export interface RegisterPayload {
  email: string;
  password: string;
  firstName: string;
  lastName?: string;
  organizationName: string;
  organizationSlug: string;
}

export const registerService = {
  register: (req: RegisterPayload) =>
    api.post<SessionResponse>("/register", req, { basePath: "/api/auth", skipAuth: true }),
};
