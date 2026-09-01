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
  LucideIcon,
} from "lucide-react";

export interface NavItem {
  id: string;
  label: string;
  icon: LucideIcon;
  href: string;
  category: "general" | "messaging" | "admin" | "system";
  badge?: string;
}

export const navSections = [
  { id: "general", label: "GENERAL" },
  { id: "messaging", label: "HERRAMIENTAS" },
  { id: "admin", label: "ADMINISTRACIÓN" },
  { id: "system", label: "SOPORTE" },
] as const;

export const navItems: NavItem[] = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, href: "/dashboard", category: "general" },
  { id: "companies", label: "Empresas", icon: Building2, href: "/empresas", category: "general" },
  { id: "sessions", label: "Sesiones", icon: Wifi, href: "/sessions", category: "general" },
  
  { id: "messages", label: "Mensajes", icon: MessageSquare, href: "/messages", category: "messaging" },
  { id: "broadcasts", label: "Broadcasts", icon: Send, href: "/broadcasts", category: "messaging" },

  { id: "users", label: "Usuarios", icon: Users, href: "/usuario_admin", category: "admin" },
  { id: "roles", label: "Roles", icon: Shield, href: "/roles", category: "admin" },
  { id: "modules", label: "Módulos", icon: LayoutGrid, href: "/modules", category: "admin" },

  { id: "settings", label: "Settings", icon: Settings, href: "/settings", category: "system" },
];

export function getActiveNavItem(pathname: string) {
  return navItems.find((item) => pathname.startsWith(item.href)) ?? null;
}

