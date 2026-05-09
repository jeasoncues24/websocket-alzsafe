"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Sidebar } from "@/components/layout/sidebar";
import { MobileNav } from "@/components/layout/mobile-nav";

export function AdminAuthCheck({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | undefined;

    if (pathname === "/login" || pathname.startsWith("/qr")) {
      timeoutId = setTimeout(() => setChecking(false), 0);
      return;
    }

    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
    } else {
      timeoutId = setTimeout(() => setChecking(false), 0);
    }

    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [router, pathname]);

  if (checking) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (pathname === "/login" || pathname.startsWith("/qr")) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <MobileNav />
        <main className="flex-1 overflow-auto bg-background p-4 md:p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
