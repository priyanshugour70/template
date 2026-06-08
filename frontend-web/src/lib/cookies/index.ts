import { destroyCookie, parseCookies, setCookie } from "nookies";

export const COOKIE_ACCESS = "app_access_token";
export const COOKIE_REFRESH = "app_refresh_token";

export interface Tokens {
  accessToken?: string;
  refreshToken?: string;
}

const baseOptions = {
  path: "/",
  sameSite: "lax" as const,
  // Browser cookies inherit the page's protocol; only mark Secure in production.
  secure: process.env.NODE_ENV === "production",
};

export function getTokens(): Tokens {
  const all = parseCookies();
  return {
    accessToken: all[COOKIE_ACCESS],
    refreshToken: all[COOKIE_REFRESH],
  };
}

export function setTokens(tokens: Required<Tokens>, maxAgeSeconds = 60 * 60 * 24 * 7) {
  setCookie(null, COOKIE_ACCESS, tokens.accessToken, { ...baseOptions, maxAge: maxAgeSeconds });
  setCookie(null, COOKIE_REFRESH, tokens.refreshToken, {
    ...baseOptions,
    maxAge: maxAgeSeconds,
  });
}

export function clearTokens() {
  destroyCookie(null, COOKIE_ACCESS, baseOptions);
  destroyCookie(null, COOKIE_REFRESH, baseOptions);
}
