# Epic 5: Panel Administrativo Frontend

## Overview

Implementar un panel administrativo interno para gestionar el backend de WhatsApp API. El panel será una aplicación web Next.js con shadcn/ui, integrada en el mismo proyecto Go.

## Requirements

### Functional Requirements

- FR-01 Dashboard global con métricas agregadas de todas las empresas
- FR-02 Lista de empresas con búsqueda y filtrado
- FR-03 Lista de mensajes con filtros por empresa, fecha, estado
- FR-04 Gestión de sesiones WhatsApp (ver estado, iniciar/desconectar)
- FR-05 Crear y ejecutar broadcasts globales
- FR-06 Ver resultados de broadcasts por destinatario
- FR-07 Configuración de apariencia (tema claro/oscuro)
- FR-08 Integración de build con Go

### Non-Functional Requirements

- NFR-01 UI responsiva con shadcn/ui
- NFR-02 Theme toggle (light/dark) persistido
- NFR-03 Estado global con Zustand
- NFR-04 Build conjunto: `go build` también build el frontend
- NFR-05 Navegación fluida con Next.js App Router

---

## Stories

### Story 5.1: Setup Next.js + shadcn + estructura base

**Objective:** Crear estructura base del frontend con Next.js, shadcn/ui, y configuración de tema.

**Acceptance Criteria:**

1. **Given** la carpeta `frontend/` no existe, **When** se crea el proyecto, **Then** Next.js 14+ está instalado con App Router.

2. **Given** shadcn/ui se inicializa, **When** se ejecuta `npx shadcn@latest init`, **Then** componentes base están disponibles en `components/ui/`.

3. **Given** el tema claro/oscuro, **When** el usuario cambia en settings, **Then** se guarda en localStorage y persiste al recargar.

4. **Given** Zustand se configura, **When** se crea la store, **Then** gestiona tema y navegación global.

5. **Given** el build de Go se ejecuta, **When** `make build` o `go build`, **Then** también ejecuta `npm run build` del frontend y copia los archivos a `static/`.

**Tasks:**

- [ ] Crear `frontend/` con Next.js (App Router)
- [ ] Instalar shadcn/ui y componentes base (Button, Card, Table, etc.)
- [ ] Configurar Tailwind con theme shadcn
- [ ] Implementar theme provider (light/dark) con Tailwind
- [ ] Crear Zustand store para tema y app state
- [ ] Crear layout base con sidebar navigation
- [ ] Configurar Makefile para build integrado
- [ ] Actualizar Go para servir archivos estáticos del frontend

---

### Story 5.2: Dashboard global con métricas

**Objective:** Crear página principal con métricas agregadas del sistema.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/`, **When** carga el dashboard, **Then** muestra: empresas activas, mensajes hoy, broadcasts hoy, tasa de éxito.

2. **Given** el dashboard, **When** hay alertas (sesiones caídas, errores), **Then** se muestran en sección de alertas con color distintivo.

3. **Given** el dashboard, **When** se hace click en una métrica, **Then** navega a la vista detallada correspondiente.

4. **Given** las métricas, **When** los datos se cargan, **Then** muestran skeleton loading mientras cargan.

**Tasks:**

- [ ] Crear page `app/page.tsx` (dashboard)
- [ ] Crear componentes de metric cards (empresas activas, mensajes hoy, broadcasts, tasa éxito)
- [ ] Consumir endpoint `GET /metrics` del backend
- [ ] Crear componente de alertas/notificaciones
- [ ] Implementar loading states con skeleton
- [ ] Crear navegación desde métricas a vistas detalladas

---

### Story 5.3: Lista de empresas con búsqueda

**Objective:** Mostrar listado de todas las empresas con búsqueda y filtrado.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/companies`, **When** carga la página, **Then** muestra tabla con: RUC, nombre, estado sesión, último actividad.

2. **Given** la lista de empresas, **When** el usuario busca por RUC o nombre, **Then** filtra en tiempo real.

3. **Given** cada fila de empresa, **When** se muestra el estado de sesión, **Then** usa indicador visual (verde=activa, rojo=inactiva, amarillo=conectando).

4. **Given** el usuario hace click en una empresa, **When** navega a `/companies/[id]`, **Then** muestra detalle con historial de mensajes y broadcasts.

**Tasks:**

- [ ] Crear page `app/companies/page.tsx`
- [ ] Crear endpoint backend `GET /admin/companies` (listar todas)
- [ ] Implementar tabla con shadcn Table
- [ ] Agregar input de búsqueda con debounce
- [ ] Crear indicador de estado (badge con color)
- [ ] Crear page de detalle `/companies/[id]`
- [ ] Mostrar historial de mensajes y broadcasts de esa empresa

---

### Story 5.4: Lista de mensajes con filtros

**Objective:** Mostrar todos los mensajes del sistema con filtros poderosos.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/messages`, **When** carga la página, **Then** muestra tabla con: referencia, empresa, destino, mensaje, estado, fecha.

2. **Given** los filtros, **When** el usuario filtra por empresa, estado, o rango de fecha, **Then** la tabla se actualiza con los resultados.

3. **Given** un mensaje, **When** se muestra el detalle, **Then** incluye: mensaje completo, metadata, intentos de envío, errores si falló.

4. **Given** la paginación, **When** hay muchos mensajes, **Then** se cargan de 20 en 20 con infinite scroll o páginao.

**Tasks:**

- [ ] Crear page `app/messages/page.tsx`
- [ ] Crear endpoint backend `GET /admin/messages` con filtros
- [ ] Implementar filtros: empresa (select), estado (select), fecha (date picker)
- [ ] Crear tabla con shadcn y paginación
- [ ] Crear modal o page de detalle de mensaje
- [ ] Implementar paginación o infinite scroll

---

### Story 5.5: Gestión de sesiones WhatsApp

**Objective:** Administrar sesiones de WhatsApp por empresa.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/sessions`, **When** carga la página, **Then** muestra lista de empresas con estado de sesión WhatsApp.

2. **Given** una sesión inactiva, **When** el usuario clickea "Conectar", **Then** genera QR y muestra para escanear.

3. **Given** una sesión activa, **When** el usuario clickea "Desconectar", **Then** cierra la sesión y actualiza el estado.

4. **Given** errores de conexión, **When** ocurre un problema, **Then** muestra notificación con el error específico.

**Tasks:**

- [ ] Crear page `app/sessions/page.tsx`
- [ ] Crear endpoints backend: `GET /admin/sessions`, `POST /admin/sessions/{ruc}/connect`, `POST /admin/sessions/{ruc}/disconnect`
- [ ] Mostrar grid de empresas con estado de sesión
- [ ] Implementar modal de QR para nueva conexión
- [ ] Crear botón de desconexión con confirmación
- [ ] Mostrar notificaciones de errores/éxito

---

### Story 5.6: Broadcasts y resultados

**Objective:** Crear y gestionar difusiones masivas.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/broadcasts`, **When** carga la página, **Then** muestra lista de broadcasts con: referencia, empresa, total destinatarios, estado, fecha.

2. **Given** un broadcast, **When** el usuario hace click en ver detalle, **Then** muestra tabla de resultados por destinatario (enviado/fallido/error).

3. **Given** la creación de broadcast, **When** el usuario completa el formulario, **Then** selecciona empresa, ingresa destinatarios (manual o CSV), escribe mensaje, y ejecuta.

4. **Given** el broadcast se está ejecutando, **When** hay progreso, **Then** muestra progress bar en tiempo real.

**Tasks:**

- [ ] Crear page `app/broadcasts/page.tsx`
- [ ] Crear endpoint `GET /admin/broadcasts` (todas las difusiones)
- [ ] Implementar tabla de broadcasts con estado
- [ ] Crear page de detalle `/broadcasts/[id]` con resultados por destinatario
- [ ] Crear formulario de creación de broadcast
- [ ] Soporte para input manual de destinatarios o paste de lista
- [ ] Mostrar progress en tiempo real (polling o WebSocket)

---

### Story 5.7: Settings - Apariencia y Configuración

**Objective:** Panel de configuración con enfoque en apariencia.

**Acceptance Criteria:**

1. **Given** el usuario accede a `/settings`, **When** carga la página, **Then** muestra tabs: Apariencia, General, Acerca de.

2. **Given** la sección de apariencia, **When** el usuario cambia entre modo claro y oscuro, **Then** el cambio es instantáneo y se guarda en persistencia.

3. **Given** la apariencia, **When** el usuario puede configurar, **Then** tiene opciones: theme (light/dark/system), densidad de UI (compact/default/spacious).

4. **Given** la sección General, **When** el usuario ve configuración, **Then** muestra: notificaciones, auto-refresh intervals, items por página por defecto.

5. **Given** el theme se guarda, **When** el usuario cierra y abre el navegador, **Then** mantiene su preferencia.

**Tasks:**

- [ ] Crear page `app/settings/page.tsx`
- [ ] Crear tabs: Apariencia, General, Acerca de
- [ ] Implementar theme toggle (light/dark) con Zustand + localStorage
- [ ] Agregar opciones de densidad de UI
- [ ] Crear sección de notificaciones (toggle)
- [ ] Crear sección "Acerca de" con versión, links
- [ ] Persistir settings en localStorage via Zustand

---

## Technical Stack

- **Frontend:** Next.js 14 (App Router)
- **UI:** shadcn/ui + Tailwind CSS
- **State:** Zustand
- **Icons:** Lucide React
- **Build:** Makefile integrado con Go

## Project Structure

```
frontend/
├── app/
│   ├── layout.tsx          # Root layout con providers
│   ├── page.tsx           # Dashboard
│   ├── companies/
│   │   ├── page.tsx       # Lista empresas
│   │   └── [id]/page.tsx  # Detalle empresa
│   ├── messages/
│   │   ├── page.tsx       # Lista mensajes
│   │   └── [id]/page.tsx  # Detalle mensaje
│   ├── sessions/page.tsx  # Gestión sesiones
│   ├── broadcasts/
│   │   ├── page.tsx       # Lista broadcasts
│   │   └── [id]/page.tsx  # Detalle broadcast
│   └── settings/page.tsx  # Configuración
├── components/
│   ├── ui/                # Shadcn primitives
│   ├── layout/            # Sidebar, Header
│   └── ...                # Componentes específicos
├── lib/
│   ├── api.ts             # API calls al backend
│   └── utils.ts           # Helpers
├── hooks/                 # Custom hooks
├── stores/                # Zustand stores
│   └── useThemeStore.ts   # Theme + settings
│   └── useAppStore.ts     # Estado global
└── types/                 # TypeScript types
```

---

## Dependencies

- `next`: ^14.0.0
- `react`, `react-dom`: ^18.0.0
- `@shadcn/ui`: latest
- `zustand`: ^4.5.0
- `lucide-react`: ^0.300.0
- `tailwindcss`: ^3.4.0
- `clsx`, `tailwind-merge`: utilities

---

## Build Integration

```makefile
# Makefile en raíz del proyecto

build: frontend frontend/out/* 
	go build -o wsapi .

frontend:
	cd frontend && npm run build

.PHONY: frontend
```

El backend Go sirve la carpeta `static/` que contiene el output de Next.js (`out/`).