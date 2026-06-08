import { getTokens } from "@/lib/cookies";

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: { code: string; message: string; details?: Record<string, unknown> };
  message?: string;
  timestamp?: string;
}

export interface RequestOptions {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  query?: Record<string, string | number | boolean | undefined>;
  headers?: Record<string, string>;
  /** Set true to skip the Authorization header. */
  skipAuth?: boolean;
  signal?: AbortSignal;
}

function buildUrl(path: string, query?: RequestOptions["query"]) {
  const url = new URL(path, typeof window === "undefined" ? "http://localhost" : window.location.origin);
  if (query) {
    for (const [k, v] of Object.entries(query)) {
      if (v !== undefined) url.searchParams.set(k, String(v));
    }
  }
  return url.pathname + url.search;
}

export async function request<T>(path: string, opts: RequestOptions = {}): Promise<ApiResponse<T>> {
  const { method = "GET", body, query, headers = {}, skipAuth, signal } = opts;
  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };

  if (body !== undefined) finalHeaders["Content-Type"] = "application/json";

  if (!skipAuth) {
    const { accessToken } = getTokens();
    if (accessToken) finalHeaders.Authorization = `Bearer ${accessToken}`;
  }

  const res = await fetch(buildUrl(`/api/v1${path}`, query), {
    method,
    headers: finalHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    credentials: "include",
    signal,
  });

  let json: ApiResponse<T> = { success: res.ok };
  try {
    json = (await res.json()) as ApiResponse<T>;
  } catch {
    // No JSON body — fall through.
  }
  json.success = json.success ?? res.ok;
  return json;
}

export const api = {
  get: <T,>(path: string, opts?: Omit<RequestOptions, "method" | "body">) =>
    request<T>(path, { ...opts, method: "GET" }),
  post: <T,>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    request<T>(path, { ...opts, method: "POST", body }),
  put: <T,>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    request<T>(path, { ...opts, method: "PUT", body }),
  patch: <T,>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    request<T>(path, { ...opts, method: "PATCH", body }),
  delete: <T,>(path: string, opts?: Omit<RequestOptions, "method" | "body">) =>
    request<T>(path, { ...opts, method: "DELETE" }),
};
