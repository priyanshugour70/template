/**
 * Server entry for the HTTP client. Use this from Server Components, Route
 * Handlers, and Server Actions. It hits the backend directly (no proxy hop)
 * and reads the access token from HttpOnly cookies via next/headers.
 *
 * For client components / hooks, import from "@/lib/client" instead.
 */

import "server-only";

export type {
  ApiClient,
  ApiResponse,
  ApiError,
  HttpMethod,
  RequestOptions,
} from "./types";

import { ssrApi } from "./impl/ssr";

export const api = ssrApi;
