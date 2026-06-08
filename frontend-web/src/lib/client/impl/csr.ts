/**
 * Browser implementation of ApiClient. Hits the same-origin proxy at
 * `/api/v1/*` which adds the bearer token from HttpOnly cookies. Client code
 * never sees tokens; this file therefore never reads cookies.
 */

import type { ApiClient, ApiResponse, RequestOptions } from "../types";
import { buildQueryString, buildResponseFallback } from "../types";

const CSR_BASE = "/api/v1";

async function csrRequest<T>(path: string, opts: RequestOptions = {}): Promise<ApiResponse<T>> {
  const { method = "GET", body, query, headers = {}, signal, basePath } = opts;
  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };
  if (body !== undefined) finalHeaders["Content-Type"] = "application/json";

  const url = `${basePath ?? CSR_BASE}${path}${buildQueryString(query)}`;

  try {
    const res = await fetch(url, {
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
      // Empty/non-JSON response (e.g. 204) — keep the ok fallback.
    }
    json.success = json.success ?? res.ok;
    return json;
  } catch (err) {
    return buildResponseFallback<T>(false, err instanceof Error ? err.message : undefined);
  }
}

export const csrApi: ApiClient = {
  request: csrRequest,
  get: (path, opts) => csrRequest(path, { ...opts, method: "GET" }),
  post: (path, body, opts) => csrRequest(path, { ...opts, method: "POST", body }),
  put: (path, body, opts) => csrRequest(path, { ...opts, method: "PUT", body }),
  patch: (path, body, opts) => csrRequest(path, { ...opts, method: "PATCH", body }),
  delete: (path, opts) => csrRequest(path, { ...opts, method: "DELETE" }),
};
