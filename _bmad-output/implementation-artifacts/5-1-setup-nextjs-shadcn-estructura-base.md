# Story 5.1: Setup Next.js + shadcn + estructura base

Status: ready-for-dev

## Story

As a developer,
I want configurar la estructura base del frontend con Next.js, shadcn/ui y Zustand,
So that podamos empezar a desarrollar las páginas del panel administrativo.

## Acceptance Criteria

1. **Given** la carpeta `frontend/` no existe, **When** se ejecuta `npx create-next-app@latest frontend`, **Then** se crea proyecto Next.js 14+ con App Router y TypeScript.

2. **Given** shadcn/ui se inicializa, **When** se ejecuta `npx shadcn@latest init`, **Then** configurar theme y crear carpeta `components/ui/` con componentes base (Button, Card, Table, Badge, Input, Select, Dialog, Tabs, Switch, DropdownMenu).

3. **Given** el tema claro/oscuro, **When** el usuario cambia en settings, **Then** se guarda en localStorage y persiste al recargar página. El sistema debe detectar preferencia del sistema (`prefers-color-scheme`) como默认值.

4. **Given** Zustand se configura, **When** se crea la store, **Then** gestiona: tema (light/dark/system), densidad UI (compact/default/spacious), y navegación activa.

5. **Given** el build de Go se ejecuta, **When** `make build` o `go build`, **Then** también ejecuta `npm run build` del frontend y copia los archivos a `internal/http/static/`. El backend sirve estos archivos en `/`.

6. **Given** la estructura de carpetas, **When** se organiza el código, **Then** sigue el patrón: `app/`, `components/`, `lib/`, `stores/`, `hooks/`, `types/`.

## Tasks / Subtasks

- [ ] **Setup Next.js**
  - [ ] Ejecutar `npx create-next-app@latest frontend` con: App Router, TypeScript, Tailwind, ESLint
  - [ ] Limpiar código default innecesario

- [ ] **Instalar shadcn/ui**
  - [ ] Ejecutar `npx shadcn@latest init` con theme personalizado
  - [ ] Agregar componentes: Button, Card, Table, Badge, Input, Select, Dialog, Tabs, Switch, DropdownMenu, Sheet, Skeleton
  - [ ] Configurar Tailwind con tokens shadcn

- [ ] **Configurar Theme (Light/Dark)**
  - [ ] Crear `components/theme-provider.tsx` con next-themes
  - [ ] Crear toggle theme en el layout
  - [ ] Persistir preferencia en localStorage

- [ ] **Configurar Zustand**
  - [ ] Instalar `zustand` (`npm install zustand`)
  - [ ] Crear `stores/useThemeStore.ts`: tema, densidad UI, persistencia
  - [ ] Crear `stores/useAppStore.ts`: navegación, estados globales

- [ ] **Crear estructura de carpetas**
  - [ ] `app/` - Next.js App Router pages
  - [ ] `components/ui/` - shadcn primitives
  - [ ] `components/layout/` - Sidebar, Header
  - [ ] `lib/` - api.ts, utils.ts
  - [ ] `stores/` - Zustand stores
  - [ ] `hooks/` - custom hooks
  - [ ] `types/` - TypeScript types

- [ ] **Layout base con navegación**
  - [ ] Crear sidebar con links: Dashboard, Empresas, Mensajes, Sesiones, Broadcasts, Settings
  - [ ] Crear header con theme toggle y breadcrumb
  - [ ] Responsive: sidebarcollapsible en mobile

- [ ] **Configurar Makefile para build integrado**
  - [ ] Crear `Makefile` en raíz del proyecto
  - [ ] Target `build`: ejecuta `cd frontend && npm run build` luego `go build`
  - [ ] Copiar output de Next.js (`frontend/out/*`) a `internal/http/static/`

- [ ] **Actualizar backend para servir static**
  - [ ] Modificar `internal/http/router.go` para servir archivos estáticos de `static/`
  - [ ] Agregar handler para `/` que sirve `static/index.html` o redirige a `/dashboard`

- [ ] **Verificar integración**
  - [ ] Ejecutar `make build` y verificar que todo compila
  - [ ] Ejecutar `./wsapi` y verificar que sirve el frontend

## Dev Notes

### Dependencias necesarias

```json
{
  "next": "^14.2.0",
  "react": "^18.2.0",
  "react-dom": "^18.2.0",
  "zustand": "^4.5.0",
  "@shadcn/ui": "latest",
  "lucide-react": "^0.300.0",
  "clsx": "^2.1.0",
  "tailwind-merge": "^2.2.0",
  "next-themes": "^0.2.0"
}
```

### Estructura de archivos

```
frontend/
├── app/
│   ├── layout.tsx        # Root con ThemeProvider
│   ├── page.tsx       # Redirect a /dashboard
│   └── dashboard/      # Dashboard (próxima story)
├── components/
│   ├── ui/             # shadcn
│   └── layout/         # Sidebar, Header
├── lib/
│   ├── api.ts         # Fetch helpers
│   └── utils.ts       # clsx, etc
├── stores/
│   ├── useThemeStore.ts
│   └── useAppStore.ts
└── types/
    └── index.ts       # Types del backend
```

### Referencias técnicas

- shadcn installation: https://ui.shadcn.com/docs/installation
- next-themes: https://github.com/pacocoursey/next-themes
- Zustand: https://zustand-demo.pmnd.rs/

## Senior Developer Review (AI)

**Fecha:** 2026-04-15
**Resultado:** Pending Implementation

### Implementation Checklist

- [ ] Next.js 14 configurado
- [ ] shadcn/ui instalado
- [ ] Theme toggle (light/dark) funcionando
- [ ] Zustand stores configuradas
- [ ] Estructura de carpetas creada
- [ ] Layout base con sidebar
- [ ] Makefile para build integrado
- [ ] Backend sirve static files
- [ ] Verificación de build completo

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List