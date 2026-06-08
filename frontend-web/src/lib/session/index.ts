import { clearTokens, getTokens, setTokens } from "@/lib/cookies";

/** Convenience helpers around token lifecycle. */
export const session = {
  read: getTokens,
  write: setTokens,
  clear: clearTokens,
  isAuthenticated: () => !!getTokens().accessToken,
};
