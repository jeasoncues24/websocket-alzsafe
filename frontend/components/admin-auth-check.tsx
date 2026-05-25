"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Sidebar } from "@/components/layout/sidebar";
import { MobileNav } from "@/components/layout/mobile-nav";
import { getAuthMe } from "@/lib/api";
import { useAppStore } from "@/stores/useAppStore";
import { getActiveNavItem } from "@/components/layout/nav-items";

export function AdminAuthCheck({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    if (pathname === "/login" || pathname.startsWith("/qr")) {
      setChecking(false);
      return;
    }

    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    const currentUser = useAppStore.getState().user;
    if (currentUser) {
      const activeItem = getActiveNavItem(pathname);
      if (
        activeItem &&
        activeItem.id !== "dashboard" &&
        activeItem.id !== "settings" &&
        !currentUser.is_root
      ) {
        const allowed = useAppStore.getState().allowedModules;
        if (!allowed.includes(activeItem.id)) {
          router.push("/dashboard");
          return;
        }
      }
      setChecking(false);
      return;
    }

    setChecking(true);
    getAuthMe()
      .then((res) => {
        if (res.ok && res.user) {
          useAppStore.getState().setAllowedModules(res.user.allowed_modules || []);
          useAppStore.getState().setUser(res.user);

          const activeItem = getActiveNavItem(pathname);
          if (
            activeItem &&
            activeItem.id !== "dashboard" &&
            activeItem.id !== "settings" &&
            !res.user.is_root
          ) {
            const allowed = res.user.allowed_modules || [];
            if (!allowed.includes(activeItem.id)) {
              router.push("/dashboard");
              return;
            }
          }
          setChecking(false);
        } else {
          localStorage.removeItem("admin_token");
          router.push("/login");
        }
      })
      .catch(() => {
        localStorage.removeItem("admin_token");
        router.push("/login");
      });
  }, [router, pathname]);

  if (checking) {
    return (
      <div className="motion-fade-in flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (pathname === "/login" || pathname.startsWith("/qr")) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <MobileNav />
        <main className="flex-1 overflow-auto bg-background p-4 md:p-6">
          <div key={pathname} className="motion-enter-up flex h-full flex-col">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
