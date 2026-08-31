"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import {
  Search,
  Bell,
  Sun,
  Moon,
  LogOut,
  User,
  ShieldCheck,
  Building,
  CheckCircle2,
  AlertTriangle,
  ExternalLink,
  Command,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { useAppStore } from "@/stores/useAppStore";
import { useTheme } from "next-themes";
import { navItems } from "@/components/layout/nav-items";
import { getMetrics, type Alert } from "@/lib/api";

export function TopBar() {
  const router = useRouter();
  const { user } = useAppStore();
  const { setTheme, resolvedTheme } = useTheme();

  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [alertsOpen, setAlertsOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const [alerts, setAlerts] = useState<Alert[]>([]);

  const alertsRef = useRef<HTMLDivElement>(null);
  const profileRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    getMetrics()
      .then((m) => {
        if (m?.alerts) {
          setAlerts(m.alerts);
        }
      })
      .catch(() => {});
  }, []);

  // Keyboard shortcut ⌘K / Ctrl+K
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setSearchOpen((prev) => !prev);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  // Click outside listener for dropdowns
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (alertsRef.current && !alertsRef.current.contains(e.target as Node)) {
        setAlertsOpen(false);
      }
      if (profileRef.current && !profileRef.current.contains(e.target as Node)) {
        setProfileOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const filteredNavResults = navItems.filter((item) =>
    item.label.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleSelectNav = (href: string) => {
    setSearchOpen(false);
    setSearchQuery("");
    router.push(href);
  };

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    window.location.href = "/login";
  };

  return (
    <>
      <header className="sticky top-0 z-30 flex h-16 w-full items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur md:px-6">
        {/* Barra de Búsqueda Rápida Trigger */}
        <div className="relative flex max-w-md flex-1 items-center">
          <button
            type="button"
            onClick={() => setSearchOpen(true)}
            className="flex h-10 w-full max-w-md items-center justify-between rounded-xl border border-input bg-muted/40 px-3.5 text-xs text-muted-foreground transition hover:bg-muted/70 focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <div className="flex items-center gap-2">
              <Search className="h-4 w-4 text-muted-foreground" />
              <span className="truncate">Buscar módulo, empresa o función...</span>
            </div>
            <kbd className="pointer-events-none hidden rounded border border-border bg-background px-1.5 py-0.5 font-mono text-[10px] font-medium text-muted-foreground sm:inline-block">
              ⌘ K
            </kbd>
          </button>
        </div>

        {/* Sección Derecha: Notificaciones, Tema y Perfil */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Toggle de Tema */}
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-9 rounded-xl text-muted-foreground hover:text-foreground"
            onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
            title="Cambiar tema"
          >
            {resolvedTheme === "dark" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </Button>

          {/* Notificaciones / Alertas */}
          <div className="relative" ref={alertsRef}>
            <Button
              variant="ghost"
              size="icon"
              className="relative h-9 w-9 rounded-xl text-muted-foreground hover:text-foreground"
              onClick={() => setAlertsOpen(!alertsOpen)}
              title="Alertas y Notificaciones"
            >
              <Bell className="h-4 w-4" />
              {alerts.length > 0 && (
                <span className="absolute top-1.5 right-1.5 flex h-2 w-2">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-destructive opacity-75"></span>
                  <span className="relative inline-flex h-2 w-2 rounded-full bg-destructive"></span>
                </span>
              )}
            </Button>

            {alertsOpen && (
              <div className="absolute right-0 mt-2 w-80 rounded-2xl border border-border bg-card p-4 shadow-xl z-50 motion-enter-up">
                <div className="flex items-center justify-between border-b border-border pb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold">Notificaciones</span>
                    {alerts.length > 0 && (
                      <Badge variant="destructive" className="h-5 px-1.5 text-[10px]">
                        {alerts.length}
                      </Badge>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-xs"
                    onClick={() => setAlertsOpen(false)}
                  >
                    Cerrar
                  </Button>
                </div>

                <div className="mt-3 flex max-h-60 flex-col gap-2 overflow-y-auto">
                  {alerts.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-6 text-center text-muted-foreground">
                      <CheckCircle2 className="mb-2 h-6 w-6 text-primary" />
                      <span className="text-xs">Sin alertas pendientes</span>
                    </div>
                  ) : (
                    alerts.map((al, idx) => (
                      <div
                        key={idx}
                        className="flex items-start gap-2.5 rounded-lg border border-border/50 bg-muted/40 p-2.5 text-xs"
                      >
                        <AlertTriangle className="mt-0.5 h-3.5 w-3.5 flex-shrink-0 text-amber-500" />
                        <div className="flex-1">
                          <p className="font-medium text-foreground">{al.message}</p>
                          <span className="text-[10px] text-muted-foreground uppercase">
                            {al.type} • {al.level}
                          </span>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Separador vertical */}
          <div className="hidden h-6 w-[1px] bg-border sm:block" />

          {/* Perfil de Usuario */}
          <div className="relative" ref={profileRef}>
            <button
              type="button"
              onClick={() => setProfileOpen(!profileOpen)}
              className="flex items-center gap-2.5 rounded-xl p-1.5 transition hover:bg-muted/60 focus:outline-none"
            >
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/15 text-xs font-semibold text-primary">
                {user?.username ? user.username.slice(0, 2).toUpperCase() : <User className="h-4 w-4" />}
              </div>
              <div className="hidden flex-col text-left text-xs md:flex">
                <span className="font-medium text-foreground leading-none">
                  {user?.username ?? "Administrador"}
                </span>
                <span className="mt-0.5 text-[11px] text-muted-foreground">
                  {user?.is_root ? "Superadmin" : "Operador"}
                </span>
              </div>
            </button>

            {profileOpen && (
              <div className="absolute right-0 mt-2 w-64 rounded-2xl border border-border bg-card p-3 shadow-xl z-50 motion-enter-up">
                <div className="border-b border-border pb-3 px-2">
                  <p className="text-sm font-semibold">{user?.username ?? "Usuario"}</p>
                  <p className="text-xs text-muted-foreground truncate">{user?.email || "admin@wsapi.local"}</p>
                  <div className="mt-2 flex items-center gap-1.5">
                    <Badge variant={user?.is_root ? "default" : "secondary"} className="text-[10px] py-0">
                      <ShieldCheck className="mr-1 h-3 w-3" />
                      {user?.is_root ? "Acceso Total" : "Módulos Asignados"}
                    </Badge>
                  </div>
                </div>

                <div className="mt-2 flex flex-col gap-1">
                  <button
                    onClick={() => {
                      setProfileOpen(false);
                      router.push("/settings");
                    }}
                    className="flex items-center gap-2 rounded-lg px-2.5 py-2 text-xs text-foreground transition hover:bg-accent"
                  >
                    <Building className="h-4 w-4 text-muted-foreground" />
                    Configuración del Sistema
                  </button>
                  <button
                    onClick={handleLogout}
                    className="flex items-center gap-2 rounded-lg px-2.5 py-2 text-xs text-destructive transition hover:bg-destructive/10"
                  >
                    <LogOut className="h-4 w-4" />
                    Cerrar Sesión
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Modal de Búsqueda Renderizado en el Portal Raíz con Overlay Completo */}
      <Dialog open={searchOpen} onOpenChange={setSearchOpen}>
        <DialogContent className="max-w-lg p-0 gap-0 overflow-hidden border-border bg-card shadow-2xl rounded-2xl">
          <DialogHeader className="sr-only">
            <DialogTitle>Búsqueda Rápida</DialogTitle>
            <DialogDescription>Buscar módulos y páginas</DialogDescription>
          </DialogHeader>

          {/* Search Input Header */}
          <div className="flex items-center gap-3 border-b border-border px-4 py-3 bg-muted/20">
            <Search className="h-4 w-4 text-muted-foreground flex-shrink-0" />
            <input
              autoFocus
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Buscar páginas, empresas, mensajes, sesiones..."
              className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground text-foreground"
            />
            {searchQuery && (
              <Button
                variant="ghost"
                size="sm"
                className="h-6 px-1.5 text-xs text-muted-foreground"
                onClick={() => setSearchQuery("")}
              >
                Limpiar
              </Button>
            )}
          </div>

          {/* Search Results */}
          <div className="max-h-80 overflow-y-auto p-2">
            <div className="px-2 py-1 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              Navegación del Sistema
            </div>
            <div className="flex flex-col gap-1 mt-1">
              {filteredNavResults.length === 0 ? (
                <div className="py-8 text-center text-xs text-muted-foreground">
                  No se encontraron resultados para &ldquo;{searchQuery}&rdquo;
                </div>
              ) : (
                filteredNavResults.map((item) => {
                  const Icon = item.icon;
                  return (
                    <button
                      key={item.id}
                      onClick={() => handleSelectNav(item.href)}
                      className="group flex w-full items-center justify-between rounded-xl px-3 py-2.5 text-left text-xs font-medium transition hover:bg-primary/10 hover:text-primary"
                    >
                      <div className="flex items-center gap-3">
                        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted group-hover:bg-primary/15 transition">
                          <Icon className="h-4 w-4 text-muted-foreground group-hover:text-primary transition" />
                        </div>
                        <span className="text-foreground group-hover:text-primary font-medium">
                          {item.label}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                        <span className="font-mono">{item.href}</span>
                        <ExternalLink className="h-3 w-3 opacity-60" />
                      </div>
                    </button>
                  );
                })
              )}
            </div>
          </div>

          {/* Footer Shortcuts */}
          <div className="flex items-center justify-between border-t border-border bg-muted/30 px-4 py-2 text-[10px] text-muted-foreground">
            <span className="flex items-center gap-1">
              <Command className="h-3 w-3" /> Escribe para filtrar en tiempo real
            </span>
            <span className="font-mono">ESC para cerrar</span>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
