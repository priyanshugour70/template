import { api } from "@/lib/client";
import type { User } from "@/types/auth";

export const authService = {
  me: () => api.get<User>("/auth/me"),
  login: (email: string, password: string) =>
    api.post<{ accessToken: string; refreshToken: string; user: User }>(
      "/auth/login",
      { email, password },
      { skipAuth: true },
    ),
  logout: (refreshToken: string) => api.post<{ ok: true }>("/auth/logout", { refreshToken }),
};
