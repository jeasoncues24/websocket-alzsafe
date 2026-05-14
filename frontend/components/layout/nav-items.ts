import {
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

export const navItems = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, href: "/dashboard" },
  { id: "companies", label: "Empresas", icon: Building2, href: "/empresas" },
  { id: "messages", label: "Mensajes", icon: MessageSquare, href: "/messages" },
  { id: "sessions", label: "Sesiones", icon: Wifi, href: "/sessions" },
  { id: "broadcasts", label: "Broadcasts", icon: Send, href: "/broadcasts" },
  { id: "users", label: "Usuario Admin", icon: Users, href: "/usuario_admin" },
  { id: "roles", label: "Roles", icon: Shield, href: "/roles" },
  { id: "modules", label: "Módulos", icon: LayoutGrid, href: "/modules" },
  { id: "settings", label: "Settings", icon: Settings, href: "/settings" },
] as const;

export function getActiveNavItem(pathname: string) {
  return navItems.find((item) => pathname.startsWith(item.href)) ?? null;
}
