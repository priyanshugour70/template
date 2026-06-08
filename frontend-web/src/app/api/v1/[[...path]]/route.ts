/**
 * Same-origin proxy: browser → /api/v1/* → backend at API_URL.
 *
 * Keeps cookies same-origin (no CORS, no SameSite=None headaches).
 * The backend is the source of truth for auth, validation, and response shape.
 */

const upstream = (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "");

type Ctx = { params: Promise<{ path?: string[] }> };

async function proxy(req: Request, { params }: Ctx) {
  const { path = [] } = await params;
  const url = new URL(req.url);
  const target = `${upstream}/api/v1/${path.join("/")}${url.search}`;

  const headers = new Headers(req.headers);
  headers.delete("host");
  // Forward only what the backend needs; let fetch set Content-Length.
  headers.delete("content-length");

  const init: RequestInit = {
    method: req.method,
    headers,
    redirect: "manual",
    body: ["GET", "HEAD"].includes(req.method) ? undefined : req.body,
    // Body streaming requires duplex: 'half' in Node 18+.
    // @ts-expect-error — duplex is a fetch RequestInit extension.
    duplex: "half",
    cache: "no-store",
  };

  const res = await fetch(target, init);
  return new Response(res.body, {
    status: res.status,
    statusText: res.statusText,
    headers: res.headers,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
export const OPTIONS = proxy;
