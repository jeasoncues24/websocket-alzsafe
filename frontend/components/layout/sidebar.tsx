"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  ChevronLeft,
  ChevronRight,
  MessageCircle,
  Activity,
  Layers,
  Sparkles,
} from "lucide-react";
import { useAppStore } from "@/stores/useAppStore";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { getActiveNavItem, navItems, navSections, type NavItem } from "@/components/layout/nav-items";

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

  const handleNavClick = (item: NavItem) => {
    setActiveNav(item.id);
    router.push(item.href);
  };

  const filteredNavItems = navItems.filter((item) => {
    if (user?.is_root) return true;
    if (item.id === "dashboard" || item.id === "settings") return true;
    return allowedModules.includes(item.id);
  });

  return (
    <aside
      className={cn(
        "hidden md:flex flex-col h-screen bg-card border-r border-border motion-fade-in transition-all duration-[var(--motion-duration-slow)] ease-[var(--motion-ease-emphasized)] select-none",
        sidebarOpen ? "w-64" : "w-20",
      )}
    >
      {/* Brand Header */}
      <div className="flex h-16 items-center justify-between border-b border-border px-4">
        {sidebarOpen ? (
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <MessageCircle className="h-5 w-5 fill-current" />
            </div>
            <div className="flex flex-col">
              <span className="font-bold text-sm tracking-tight text-foreground flex items-center gap-1.5">
                WhatsApp API
              </span>
              <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">
                Enterprise Hub
              </span>
            </div>
          </div>
        ) : (
          <div className="flex w-full justify-center">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <MessageCircle className="h-5 w-5 fill-current" />
            </div>
          </div>
        )}

        {sidebarOpen && (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-muted-foreground hover:text-foreground"
            onClick={() => setSidebarOpen(false)}
            title="Colapsar menú"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
        )}
      </div>

      {!sidebarOpen && (
        <div className="flex justify-center pt-2">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:text-foreground"
            onClick={() => setSidebarOpen(true)}
            title="Expandir menú"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Navigation Sections */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-5">
        {navSections.map((section) => {
          const sectionItems = filteredNavItems.filter(
            (item) => item.category === section.id
          );
          if (sectionItems.length === 0) return null;

          return (
            <div key={section.id} className="space-y-1">
              {sidebarOpen && (
                <div className="px-3 pb-1 text-[10px] font-bold tracking-wider text-muted-foreground/80 uppercase">
                  {section.label}
                </div>
              )}
              <div className="space-y-1">
                {sectionItems.map((item) => {
                  const isActive = pathname.startsWith(item.href);
                  const Icon = item.icon;

                  return (
                    <button
                      key={item.id}
                      onClick={() => handleNavClick(item)}
                      aria-current={isActive ? "page" : undefined}
                      className={cn(
                        "group relative flex w-full items-center rounded-xl font-medium text-xs transition-all duration-150",
                        sidebarOpen ? "px-3 py-2.5" : "h-11 justify-center px-0",
                        isActive
                          ? "bg-primary text-primary-foreground font-semibold shadow-xs"
                          : "text-muted-foreground hover:bg-accent hover:text-foreground",
                      )}
                      title={!sidebarOpen ? item.label : undefined}
                    >
                      <Icon
                        className={cn(
                          "h-4 w-4 flex-shrink-0 transition-transform group-hover:scale-105",
                          isActive ? "text-primary-foreground" : "text-muted-foreground group-hover:text-foreground",
                          sidebarOpen && "mr-3",
                        )}
                      />
                      {sidebarOpen && (
                        <div className="flex flex-1 items-center justify-between overflow-hidden">
                          <span className="truncate">{item.label}</span>
                          {item.id === "broadcasts" && (
                            <Badge
                              variant={isActive ? "secondary" : "outline"}
                              className={cn(
                                "h-4 px-1 text-[9px] font-semibold tracking-wide",
                                isActive ? "bg-primary-foreground/20 text-primary-foreground border-none" : ""
                              )}
                            >
                              PRO
                            </Badge>
                          )}
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}
      </nav>

      {/* Footer / Account / Workspace Card */}
      {sidebarOpen ? (
        <div className="p-3 border-t border-border bg-muted/20">
          <div className="rounded-2xl border border-border/80 bg-card p-3 shadow-2xs">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <Layers className="h-5 w-5" />
              </div>
              <div className="flex flex-1 flex-col overflow-hidden">
                <div className="flex items-center justify-between">
                  <span className="truncate text-xs font-semibold text-foreground">
                    Instancia Principal
                  </span>
                </div>
                <span className="text-[10px] text-muted-foreground flex items-center gap-1">
                  <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
                  Cluster Operativo
                </span>
              </div>
            </div>

            <div className="mt-2.5 flex items-center justify-between border-t border-border/50 pt-2 text-[10px] text-muted-foreground">
              <span className="font-mono">v2026.1</span>
              <span className="flex items-center gap-1 text-primary font-medium">
                <Activity className="h-3 w-3" />
                99.9% Uptime
              </span>
            </div>
          </div>
        </div>
      ) : (
        <div className="p-2 border-t border-border flex justify-center">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary" title="Cluster Operativo">
            <Sparkles className="h-4 w-4" />
          </div>
        </div>
      )}
    </aside>
  );
}
