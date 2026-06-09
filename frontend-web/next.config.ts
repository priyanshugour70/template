import type { NextConfig } from "next";

const ADMIN_SEGMENTS = ["users", "roles", "organizations", "audit"] as const;

const nextConfig: NextConfig = {
  reactCompiler: true,
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
