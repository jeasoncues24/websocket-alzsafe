"use client";

import { usePathname, useRouter } from "next/navigation";
import {
  Menu,
  LayoutDashboard,
  Building2,
  MessageSquare,
  Wifi,
  Send,
  Settings,
  Users,
  Shield,
  LayoutGrid,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

const navItems = [
  { id: "dashboard",  label: "Dashboard",     icon: LayoutDashboard, href: "/dashboard" },
  { id: "companies",  label: "Empresas",       icon: Building2,       href: "/empresas" },
  { id: "messages",   label: "Mensajes",       icon: MessageSquare,   href: "/messages" },
  { id: "sessions",   label: "Sesiones",       icon: Wifi,            href: "/sessions" },
  { id: "broadcasts", label: "Broadcasts",     icon: Send,            href: "/broadcasts" },
  { id: "users",      label: "Usuario Admin",  icon: Users,           href: "/usuario_admin" },
  { id: "roles",      label: "Roles",          icon: Shield,          href: "/roles" },
  { id: "modules",    label: "Módulos",        icon: LayoutGrid,      href: "/modules" },
  { id: "settings",   label: "Settings",       icon: Settings,        href: "/settings" },
];

export function MobileNav() {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <div className="flex md:hidden items-center justify-between px-4 h-14 border-b bg-background flex-shrink-0">
      <span className="font-semibold text-lg">WhatsApp API</span>
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon">
            <Menu className="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="p-0">
          <SheetHeader className="px-4 pt-6 pb-2">
            <SheetTitle>Navegación</SheetTitle>
          </SheetHeader>
          <nav className="flex flex-col p-2 space-y-1">
            {navItems.map((item) => {
              const isActive = pathname.startsWith(item.href);
              return (
                <SheetClose asChild key={item.id}>
                  <Button
                    variant="ghost"
                    onClick={() => router.push(item.href)}
                    className={cn(
                      "w-full h-11 justify-start px-3 transition-colors",
                      isActive &&
                        "bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary",
                    )}
                  >
                    <item.icon className="h-5 w-5 flex-shrink-0" />
                    <span className="ml-3">{item.label}</span>
                  </Button>
                </SheetClose>
              );
            })}
          </nav>
        </SheetContent>
      </Sheet>
    </div>
  );
}
