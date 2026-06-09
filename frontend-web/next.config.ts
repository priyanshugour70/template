import type { NextConfig } from "next";

const ADMIN_SEGMENTS = ["users", "roles", "organizations", "audit"] as const;

const nextConfig: NextConfig = {
  reactCompiler: true,
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
