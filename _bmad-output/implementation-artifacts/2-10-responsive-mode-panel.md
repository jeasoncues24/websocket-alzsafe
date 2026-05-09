---
title: 'Story 2.10 — Responsive mode del panel administrativo'
type: 'feature'
created: '2026-05-08'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '3c41157'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador que accede al panel desde un móvil o tablet,
quiero que la interfaz se adapte a pantallas pequeñas mostrando un menú hamburguesa en lugar del sidebar fijo,
para poder navegar sin que el sidebar tape el contenido.

## Acceptance Criteria

**AC1 — Sidebar oculto en móvil:**
En pantallas `< md` (< 768px) el sidebar lateral no es visible. El área de contenido ocupa el 100% del ancho.

**AC2 — Topbar móvil con hamburguesa:**
En pantallas `< md` aparece una barra superior fija con: nombre de la app ("WhatsApp API") a la izquierda y un botón hamburguesa (`Menu` icon de lucide-react) a la derecha.

**AC3 — Sheet de navegación móvil:**
Al hacer click en el hamburguesa se abre un `Sheet` desde la izquierda con todos los items de navegación (los mismos 9 items del sidebar desktop). Al hacer click en un item el Sheet se cierra y navega a la ruta correspondiente. El item activo mantiene el mismo estilo de highlight que el sidebar desktop (`bg-primary/10 text-primary`).

**AC4 — Comportamiento desktop intacto:**
En pantallas `≥ md` el sidebar lateral funciona exactamente igual que antes (visible, colapsable con `sidebarOpen`/`setSidebarOpen`, transición `w-64`/`w-16`). El topbar móvil NO aparece en desktop.

**AC5 — Padding responsivo en main:**
El `<main>` usa `p-4 md:p-6` (en móvil padding más compacto, en desktop el p-6 actual).

**AC6 — Lint:**
`cd frontend && npm run lint` pasa sin nuevos errores (los 3 pre-existentes en `api.ts` son aceptables).

## Tasks / Subtasks

- [x] **Tarea 1: Modificar `sidebar.tsx` para ocultarse en móvil** (AC: 1, 4)
  - [x] Cambiar `"flex flex-col h-screen ..."` por `"hidden md:flex flex-col h-screen ..."` en el div raíz

- [x] **Tarea 2: Crear componente `MobileNav`** (AC: 2, 3)
  - [x] Crear `frontend/components/layout/mobile-nav.tsx`
  - [x] Topbar: `<div className="flex md:hidden items-center justify-between px-4 h-14 border-b bg-background">`
  - [x] Hamburguesa con `SheetTrigger asChild` (sin useState — shadcn maneja el estado)
  - [x] `SheetContent side="left"` con `SheetHeader` + `SheetTitle` (accesibilidad Radix)
  - [x] Items del nav envueltos en `SheetClose asChild` — se cierra al navegar
  - [x] Items: mismo estilo que sidebar desktop (`h-11`, `justify-start px-3`, highlight activo)

- [x] **Tarea 3: Modificar `admin-auth-check.tsx`** (AC: 2, 5)
  - [x] Envolver `<Sidebar />` + `<main>` en un layout de columna para acomodar el topbar móvil
  - [x] Incluir `<MobileNav />` antes del `<main>`
  - [x] Cambiar `p-6` → `p-4 md:p-6` en `<main>`

- [x] **Tarea 4: Lint** (AC: 6)
  - [x] `cd frontend && npm run lint`

## Dev Notes

### 📐 Estado actual de los archivos a modificar

**`frontend/components/admin-auth-check.tsx` (MODIFICAR):**
```tsx
return (
  <div className="flex h-screen">
    <Sidebar />
    <main className="flex-1 overflow-auto bg-background p-6">{children}</main>
  </div>
);
```

**`frontend/components/layout/sidebar.tsx` (MODIFICAR — solo 1 línea):**
```tsx
// Línea 54 — cambiar:
"flex flex-col h-screen bg-background border-r transition-all duration-300",
// Por:
"hidden md:flex flex-col h-screen bg-background border-r transition-all duration-300",
```
El resto del componente sidebar no cambia nada.

### 📐 Resultado esperado en `admin-auth-check.tsx`

```tsx
import { MobileNav } from "@/components/layout/mobile-nav";

// En el return del bloque autenticado:
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
```

### 📐 Código exacto de `MobileNav`

El patrón shadcn correcto usa `SheetTrigger` + `SheetClose` — sin `useState`. El Sheet de este proyecto:
- Ya renderiza botón X de cierre (`showCloseButton=true` por defecto)
- Requiere `SheetTitle` para accesibilidad (es `DialogPrimitive.Title` internamente)
- `SheetClose asChild` en cada nav item lo cierra al navegar

```tsx
"use client";

import { usePathname, useRouter } from "next/navigation";
import {
  Menu, LayoutDashboard, Building2, MessageSquare,
  Wifi, Send, Settings, Users, Shield, LayoutGrid,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet, SheetContent, SheetHeader, SheetTitle,
  SheetTrigger, SheetClose,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

const navItems = [
  { id: "dashboard", label: "Dashboard",      icon: LayoutDashboard, href: "/dashboard" },
  { id: "companies", label: "Empresas",        icon: Building2,       href: "/empresas" },
  { id: "messages",  label: "Mensajes",        icon: MessageSquare,   href: "/messages" },
  { id: "sessions",  label: "Sesiones",        icon: Wifi,            href: "/sessions" },
  { id: "broadcasts",label: "Broadcasts",      icon: Send,            href: "/broadcasts" },
  { id: "users",     label: "Usuario Admin",   icon: Users,           href: "/usuario_admin" },
  { id: "roles",     label: "Roles",           icon: Shield,          href: "/roles" },
  { id: "modules",   label: "Módulos",         icon: LayoutGrid,      href: "/modules" },
  { id: "settings",  label: "Settings",        icon: Settings,        href: "/settings" },
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
                      isActive && "bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary",
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
```

### ⚠️ Sheet — por qué SheetTrigger y SheetClose (no useState)

El `Sheet` de shadcn/Radix maneja su propio estado open/closed. Usar `useState` + `open={open}` es innecesario y rompe el patrón. El flujo correcto:
- `SheetTrigger asChild` — delega el evento click al Button hijo, abre el Sheet
- `SheetClose asChild` — delega el evento click al Button hijo, cierra el Sheet
- El overlay y la tecla Escape también cierran el Sheet automáticamente (Radix lo maneja)

### ⚠️ `navItems` duplicado — es intencional

`navItems` aparece tanto en `sidebar.tsx` como en `mobile-nav.tsx`. **No refactorizar** para compartirlos en un archivo común — la story es de baja prioridad y la deuda de DRY es menor. Si se centraliza en el futuro, hacerlo en una story dedicada.

### ⚠️ `flex-shrink-0` en el topbar

El topbar móvil necesita `flex-shrink-0` para que no se encoja cuando el contenido crece. El `<main>` debe tener `flex-1 overflow-auto` para que sea el que scrollea.

### ⚠️ Tailwind v4 — sin `tailwind.config.js`

Este proyecto usa Tailwind CSS v4 con configuración implícita (solo PostCSS). Los breakpoints `md:` funcionan igual que en v3 (768px). No hay que configurar nada extra.

### ⚠️ `overflow-hidden` en el contenedor flex-col

El div `<div className="flex flex-col flex-1 overflow-hidden">` necesita `overflow-hidden` para que el scroll quede contenido en el `<main>` y no en el contenedor padre.

### Learnings de stories anteriores

- Package de frontend: Next.js 16.x, React 19, Tailwind v4, componentes shadcn copiados en `components/ui/`
- `usePathname()` para detectar ruta activa, `useRouter()` para navegar — mismos hooks que usa `sidebar.tsx`
- `useAppStore` con Zustand para `sidebarOpen` — el topbar móvil NO necesita tocar este store (tiene su propio `useState` local para el Sheet)
- `npm run lint` con ESLint — los 3 errores `no-explicit-any` en `api.ts` son pre-existentes y aceptables

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- `sidebar.tsx`: `"flex"` → `"hidden md:flex"` en el div raíz. Una sola línea. El sidebar desktop sigue funcionando igual (colapsable, Zustand).
- `mobile-nav.tsx`: nuevo componente. Topbar `flex md:hidden` con `SheetTrigger asChild` para el hamburguesa. `SheetContent side="left"` con `SheetHeader` + `SheetTitle` ("Navegación") para accesibilidad Radix. 9 navItems con `SheetClose asChild` — al hacer click navegan y cierran el Sheet sin useState.
- `admin-auth-check.tsx`: el bloque autenticado ahora envuelve `Sidebar` + `<div className="flex flex-col flex-1 overflow-hidden">` que contiene `<MobileNav />` y `<main className="flex-1 overflow-auto bg-background p-4 md:p-6">`. `overflow-hidden` en el wrapper evita scroll doble.
- Lint: solo 3 errores pre-existentes en `api.ts` (aceptables per AC6).

### File List

- frontend/components/layout/sidebar.tsx
- frontend/components/layout/mobile-nav.tsx
- frontend/components/admin-auth-check.tsx

### Change Log

- 2026-05-08: Story creada — responsive mode: sidebar oculto en móvil, MobileNav con Sheet
- 2026-05-08: Implementación completa — sidebar hidden md:flex, MobileNav con SheetTrigger+SheetClose (shadcn correcto), padding responsivo p-4 md:p-6
