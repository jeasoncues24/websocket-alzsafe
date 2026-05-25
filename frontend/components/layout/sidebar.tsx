"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight, Moon, Sun, LogOut } from "lucide-react";
import { useAppStore } from "@/stores/useAppStore";
import { Button } from "@/components/ui/button";
import { useTheme } from "next-themes";
import { getActiveNavItem, navItems } from "@/components/layout/nav-items";

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { sidebarOpen, setSidebarOpen, setActiveNav, allowedModules, user } = useAppStore();
  const activeItem = getActiveNavItem(pathname);

  useEffect(() => {
    if (activeItem) {
      setActiveNav(activeItem.id);
    }
  }, [activeItem, setActiveNav]);
  const { setTheme, resolvedTheme } = useTheme();

  const handleNavClick = (item: (typeof navItems)[number]) => {
    setActiveNav(item.id);
    router.push(item.href);
  };

  const filteredNavItems = navItems.filter((item) => {
    if (user?.is_root) return true;
    if (item.id === "dashboard" || item.id === "settings") return true;
    return allowedModules.includes(item.id);
  });

  return (
    <div
      className={cn(
        "hidden md:flex flex-col h-screen bg-background border-r motion-fade-in transition-[width] duration-[var(--motion-duration-slow)] ease-[var(--motion-ease-emphasized)]",
        sidebarOpen ? "w-64" : "w-16",
      )}
    >
      <div className="flex items-center justify-between border-b p-4">
        {sidebarOpen && (
          <div className="motion-enter-up flex flex-col gap-0.5">
            <span className="text-lg font-semibold">WhatsApp API</span>
            <span className="text-xs text-muted-foreground">
              {activeItem?.label ?? "Panel administrativo"}
            </span>
          </div>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setSidebarOpen(!sidebarOpen)}
        >
          {sidebarOpen ? (
            <ChevronLeft className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </Button>
      </div>

      <nav className="flex flex-1 flex-col gap-1 p-2">
        {filteredNavItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Button
              key={item.id}
              variant="ghost"
              onClick={() => handleNavClick(item)}
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "motion-transition relative h-11 w-full overflow-hidden rounded-lg",
                sidebarOpen ? "justify-start px-3" : "justify-center px-0",
                isActive
                  ? "bg-primary/10 text-primary shadow-sm hover:bg-primary/15 hover:text-primary"
                  : "hover:bg-muted/60",
              )}
            >
              <span
                aria-hidden="true"
                className={cn(
                  "absolute inset-y-2 left-1 w-1 rounded-full bg-primary motion-transform",
                  isActive ? "opacity-100" : "opacity-0",
                )}
              />
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {sidebarOpen && <span className="ml-3 truncate">{item.label}</span>}
            </Button>
          );
        })}
      </nav>

      <div className="p-4 border-t flex justify-between">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
        >
          {resolvedTheme === "dark" ? (
            <Sun className="h-4 w-4" />
          ) : (
            <Moon className="h-4 w-4" />
          )}
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => {
            localStorage.removeItem("admin_token");
            window.location.href = "/login";
          }}
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
