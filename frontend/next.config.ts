import type { NextConfig } from "next";

const backendUrl = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");

if (!backendUrl) {
  throw new Error("NEXT_PUBLIC_API_URL is required in frontend/.env.local");
}

const nextConfig: NextConfig = {
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${backendUrl}/api/:path*`,
      },
      {
        source: "/admin/:path*",
        destination: `${backendUrl}/admin/:path*`,
      },
      {
        source: "/metrics",
        destination: `${backendUrl}/metrics`,
      },
      {
        source: "/ws",
        destination: `${backendUrl}/ws`,
      },
    ];
  },
};

export default nextConfig;
