/**
 * The HTTP client interface. Both impl/csr (browser) and impl/ssr (server)
 * satisfy this shape — consumer code (services) never needs to know which
 * one is wired in.
 */

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
  message?: string;
  timestamp?: string;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export interface RequestOptions {
  method?: HttpMethod;
  body?: unknown;
  /**
   * Query-string params. Accepts any object — `object` rather than `Record<string, unknown>`
   * so callers can pass typed interfaces without an explicit index signature.
   */
  query?: object;
  headers?: Record<string, string>;
  /** Skip the auth header attach. Used by login/discover/refresh. */
  skipAuth?: boolean;
  signal?: AbortSignal;
  /** Override base path (default: /api/v1 for CSR, API_URL/api/v1 for SSR). */
  basePath?: string;
}

export interface ApiClient {
  request<T>(path: string, opts?: RequestOptions): Promise<ApiResponse<T>>;
  get<T>(path: string, opts?: Omit<RequestOptions, "method" | "body">): Promise<ApiResponse<T>>;
  post<T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">): Promise<ApiResponse<T>>;
  put<T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">): Promise<ApiResponse<T>>;
  patch<T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">): Promise<ApiResponse<T>>;
  delete<T>(path: string, opts?: Omit<RequestOptions, "method" | "body">): Promise<ApiResponse<T>>;
}

export function buildQueryString(query?: RequestOptions["query"]): string {
  if (!query) return "";
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(query as Record<string, unknown>)) {
    if (v === undefined || v === null) continue;
    if (Array.isArray(v)) {
      for (const item of v) params.append(k, String(item));
      continue;
    }
    if (typeof v === "object") {
      params.set(k, JSON.stringify(v));
      continue;
    }
    params.set(k, String(v));
  }
  const s = params.toString();
  return s ? `?${s}` : "";
}

export function buildResponseFallback<T>(ok: boolean, message?: string): ApiResponse<T> {
  return { success: ok, error: ok ? undefined : { code: "NETWORK_ERROR", message: message ?? "request failed" } };
}
