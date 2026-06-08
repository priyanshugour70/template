/** Build a parameterized path: replaceParams("/items/:id", { id: 7 }) => "/items/7". */
export function replaceParams(template: string, params: Record<string, string | number>): string {
  let out = template;
  for (const [k, v] of Object.entries(params)) {
    out = out.replace(`:${k}`, encodeURIComponent(String(v)));
  }
  return out;
}

/** Append a `redirect` query param to /auth/login. */
export function loginWithRedirect(currentPath: string): string {
  const q = currentPath ? `?redirect=${encodeURIComponent(currentPath)}` : "";
  return `/auth/login${q}`;
}
