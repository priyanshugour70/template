/**
 * Server implementation of ApiClient. Hits the backend directly using API_URL,
 * pulling the access token from the HttpOnly cookie via next/headers. Use
 * this from Server Components / Route Handlers / Server Actions.
 */

import "server-only";

import { readAccessToken } from "@/lib/cookies/server";

import type { ApiClient, ApiResponse, RequestOptions } from "../types";
import { buildQueryString, buildResponseFallback } from "../types";

const BACKEND_BASE = (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "");
const SSR_BASE = `${BACKEND_BASE}/api/v1`;

async function ssrRequest<T>(path: string, opts: RequestOptions = {}): Promise<ApiResponse<T>> {
  const { method = "GET", body, query, headers = {}, signal, basePath, skipAuth } = opts;
  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };
  if (body !== undefined) finalHeaders["Content-Type"] = "application/json";
  if (!skipAuth) {
    const token = await readAccessToken();
    if (token) finalHeaders.Authorization = `Bearer ${token}`;
  }

  const url = `${basePath ?? SSR_BASE}${path}${buildQueryString(query)}`;

  try {
    const res = await fetch(url, {
      method,
      headers: finalHeaders,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      cache: "no-store",
      signal,
    });
    let json: ApiResponse<T> = { success: res.ok };
    try {
      json = (await res.json()) as ApiResponse<T>;
    } catch {
      // Empty body — keep ok fallback.
    }
    json.success = json.success ?? res.ok;
    return json;
  } catch (err) {
    return buildResponseFallback<T>(false, err instanceof Error ? err.message : undefined);
  }
}

export const ssrApi: ApiClient = {
  request: ssrRequest,
  get: (path, opts) => ssrRequest(path, { ...opts, method: "GET" }),
  post: (path, body, opts) => ssrRequest(path, { ...opts, method: "POST", body }),
  put: (path, body, opts) => ssrRequest(path, { ...opts, method: "PUT", body }),
  patch: (path, body, opts) => ssrRequest(path, { ...opts, method: "PATCH", body }),
  delete: (path, opts) => ssrRequest(path, { ...opts, method: "DELETE" }),
};
