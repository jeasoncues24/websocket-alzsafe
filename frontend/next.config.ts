import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
      {
        source: "/admin/:path*",
        destination: "http://localhost:8080/admin/:path*",
      },
      {
        source: "/metrics",
        destination: "http://localhost:8080/metrics",
      },
      {
        source: "/ws",
        destination: "http://localhost:8080/ws",
      },
    ];
  },
};

export default nextConfig;
