/**
 * lib/client — the HTTP client INTERFACE. Concrete impls live under ./impl/.
 *
 * Selection by environment:
 *   import { api } from "@/lib/client"          // CSR — browser / client components
 *   import { api } from "@/lib/client/server"   // SSR — RSC / route handlers / server actions
 *
 * Both impls satisfy the same `ApiClient` type, so service code reads
 * identically. The CSR impl hits the same-origin proxy (`/api/v1/*`) which
 * lets the proxy attach the bearer token from HttpOnly cookies. The SSR
 * impl hits the backend directly using API_URL and reads the token from
 * next/headers.
 */

export type {
  ApiClient,
  ApiResponse,
  ApiError,
  HttpMethod,
  RequestOptions,
} from "./types";

export { buildQueryString } from "./types";

import { csrApi } from "./impl/csr";

/** Default export is the CSR impl — safe to import from any client context. */
export const api = csrApi;
