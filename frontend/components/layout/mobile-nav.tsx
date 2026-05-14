"use client";

import { useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Menu } from "lucide-react";
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
import { getActiveNavItem, navItems } from "@/components/layout/nav-items";

export function MobileNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const activeItem = getActiveNavItem(pathname);

  const handleNavigate = (href: string) => {
    setOpen(false);
    router.push(href);
  };

  return (
    <div className="motion-fade-in flex h-14 flex-shrink-0 items-center justify-between border-b bg-background px-4 md:hidden">
      <div className="flex min-w-0 flex-col">
        <span className="truncate text-sm font-semibold">WhatsApp API</span>
        <span className="truncate text-xs text-muted-foreground">
          {activeItem?.label ?? "Panel administrativo"}
        </span>
      </div>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Abrir navegación">
            <Menu className="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="p-0">
          <SheetHeader className="border-b px-4 pt-6 pb-3">
            <SheetTitle>Navegación</SheetTitle>
            <SheetDescription>
              Accede rápido a los módulos del panel.
            </SheetDescription>
          </SheetHeader>
          <nav className="flex flex-col gap-1 p-2">
            {navItems.map((item) => {
              const isActive = pathname.startsWith(item.href);
              return (
                <Button
                  key={item.id}
                  variant="ghost"
                  onClick={() => handleNavigate(item.href)}
                  aria-current={isActive ? "page" : undefined}
                  className={cn(
                    "motion-transition relative h-11 w-full justify-start overflow-hidden rounded-lg px-3",
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
                  <span className="ml-3 truncate">{item.label}</span>
                </Button>
              );
            })}
          </nav>
        </SheetContent>
      </Sheet>
    </div>
  );
}
