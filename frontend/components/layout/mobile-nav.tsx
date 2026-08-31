"use client";

import { useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Menu, MessageCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { getActiveNavItem, navItems, navSections } from "@/components/layout/nav-items";
import { useAppStore } from "@/stores/useAppStore";

export function MobileNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const activeItem = getActiveNavItem(pathname);
  const { allowedModules, user } = useAppStore();

  const handleNavigate = (href: string) => {
    setOpen(false);
    router.push(href);
  };

  const filteredNavItems = navItems.filter((item) => {
    if (user?.is_root) return true;
    if (item.id === "dashboard" || item.id === "settings") return true;
    return allowedModules.includes(item.id);
  });

  return (
    <div className="motion-fade-in flex h-14 flex-shrink-0 items-center justify-between border-b border-border bg-card px-4 md:hidden">
      <div className="flex items-center gap-2.5">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <MessageCircle className="h-4 w-4 fill-current" />
        </div>
        <div className="flex min-w-0 flex-col">
          <span className="truncate text-sm font-semibold">WhatsApp API</span>
          <span className="truncate text-[11px] text-muted-foreground">
            {activeItem?.label ?? "Panel administrativo"}
          </span>
        </div>
      </div>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Abrir navegación">
            <Menu className="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="p-0 flex flex-col justify-between">
          <div>
            <SheetHeader className="border-b border-border px-4 pt-6 pb-3">
              <SheetTitle className="text-base font-bold">WhatsApp API Hub</SheetTitle>
              <SheetDescription className="text-xs">
                Navegación y administración de servicios
              </SheetDescription>
            </SheetHeader>

            <nav className="flex flex-col gap-4 p-3 overflow-y-auto max-h-[calc(100vh-140px)]">
              {navSections.map((section) => {
                const sectionItems = filteredNavItems.filter(
                  (item) => item.category === section.id
                );
                if (sectionItems.length === 0) return null;

                return (
                  <div key={section.id} className="space-y-1">
                    <div className="px-2 pb-1 text-[10px] font-bold tracking-wider text-muted-foreground uppercase">
                      {section.label}
                    </div>
                    <div className="space-y-1">
                      {sectionItems.map((item) => {
                        const isActive = pathname.startsWith(item.href);
                        const Icon = item.icon;

                        return (
                          <Button
                            key={item.id}
                            variant="ghost"
                            onClick={() => handleNavigate(item.href)}
                            aria-current={isActive ? "page" : undefined}
                            className={cn(
                              "relative h-10 w-full justify-start rounded-xl px-3 text-xs font-medium transition",
                              isActive
                                ? "bg-primary text-primary-foreground font-semibold shadow-xs"
                                : "hover:bg-accent text-muted-foreground",
                            )}
                          >
                            <Icon className="mr-3 h-4 w-4 flex-shrink-0" />
                            <span className="truncate">{item.label}</span>
                          </Button>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </nav>
          </div>

          <div className="border-t border-border p-4 bg-muted/20 text-xs text-muted-foreground">
            <div className="flex items-center justify-between">
              <span>{user?.username ?? "Usuario"}</span>
              <span className="font-semibold text-primary">{user?.is_root ? "Root" : "Operador"}</span>
            </div>
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
