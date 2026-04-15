import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  output: isProd ? "export" : undefined,
  distDir: isProd ? "out" : undefined,
  images: {
    unoptimized: true,
  },
  async rewrites() {
    if (isProd) return [];
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
        source: "/companies",
        destination: "http://localhost:8080/companies",
      },
      {
        source: "/ws",
        destination: "http://localhost:8080/ws",
      },
    ];
  },
};

export default nextConfig;