import type { NextConfig } from "next";

const ADMIN_SEGMENTS = ["users", "roles", "organizations", "audit"] as const;

const nextConfig: NextConfig = {
  reactCompiler: true,
  // Multi-tenant dev hosts. Next.js 16 blocks dev-only assets (HMR WebSocket,
  // RSC fetches) from origins other than the bind host by default. We serve
  // the app at `lvh.me` and `<tenant>.lvh.me` in dev, so both must be allowed.
  // *.localhost is included for the Chrome/Firefox "*.localhost auto-resolves
  // to 127.0.0.1" convention. Has no effect outside `next dev`.
  allowedDevOrigins: ["lvh.me", "*.lvh.me", "*.localhost"],
  // Enable the experimental forbidden()/unauthorized() functions so server
  // code can call them to render app/forbidden.tsx and app/unauthorized.tsx.
  experimental: {
    authInterrupts: true,
  },
  async redirects() {
    return ADMIN_SEGMENTS.flatMap((seg) => [
      {
        source: `/dashboard/${seg}`,
        destination: `/dashboard/administrative/${seg}`,
        permanent: false,
      },
      {
        source: `/dashboard/${seg}/:path*`,
        destination: `/dashboard/administrative/${seg}/:path*`,
        permanent: false,
      },
    ]);
  },
};

export default nextConfig;
