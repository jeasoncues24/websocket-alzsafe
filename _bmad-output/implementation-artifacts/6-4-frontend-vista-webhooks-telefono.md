---
story_id: "6.4"
epic: "epic-6"
title: "Frontend — Vista de webhooks por teléfono (solo lectura)"
status: review
estimated_days: 2
priority: high
branch: "feature/panel-telefonos-webhooks"
skills: ["ui-ux-pro-max", "bmad-code-review"]
affects:
  - frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/webhooks/page.tsx
  - frontend/lib/api.ts
---

# Story 6.4: Frontend — Vista de webhooks por teléfono (solo lectura)

## Story

Como admin del panel wsapi,
quiero ver todos los webhooks registrados para un teléfono con información de diagnóstico detallada,
para apoyar a los integradores B2B cuando reportan que no les llegan eventos.

## Contexto técnico

Nueva página en el App Router de Next.js. El endpoint backend es `GET /api/admin/telefonos/{id}/webhooks` (story 6.3). La página es completamente read-only — el admin solo visualiza, no puede crear ni eliminar webhooks.

Datos disponibles por webhook: `id`, `empresa_id`, `telefono_id`, `api_key_id`, `url`, `eventos[]`, `activo`, `failure_count`, `last_error`, `last_success_at`, `created_at`, `updated_at`.

Stack: Next.js 15 App Router, `"use client"`, shadcn/ui, Tailwind CSS 4, lucide-react.

## Acceptance Criteria

**AC1 — Ruta y navegación:**
**Dado** que el admin hace clic en `[Webhooks]` en la card de un teléfono,
**Cuando** navega a `/empresas/[empresaId]/telefonos/[telefonoId]/webhooks`,
**Entonces** la página carga correctamente con el número del teléfono en el header.

**AC2 — Fila bento de resumen (3 mini-cards):**
**Dado** que hay webhooks cargados,
**Cuando** se renderiza la página,
**Entonces** se muestran 3 cards de resumen:
- `N webhooks activos`
- `N webhooks inactivos`
- `N con fallos` (failure_count > 0 y activo = true) — con badge destructivo si N > 0

**AC3 — Tabla de webhooks:**
**Dado** que hay webhooks cargados,
**Cuando** se renderiza la tabla,
**Entonces** cada fila muestra:
- URL truncada (max 40 chars + `...`) con tooltip de URL completa
- Badges de eventos suscritos en español
- Badge de estado en español
- Conteo de fallos (resaltado en rojo si > 0)
- Último éxito (fecha relativa: "hace 2 min", "hace 3 días", o "Nunca")
- Fecha de registro

**AC4 — Accordion de detalle por fila:**
**Dado** una fila en la tabla,
**Cuando** el admin hace clic en ella o en un botón expandir,
**Entonces** se muestra el detalle completo:
- URL completa (con botón de copiar)
- Prefijo de la API key que lo registró (`api_key_id` — mostrar como "API key #42" por ahora)
- Lista completa de eventos como badges
- `last_error` si existe (en rojo)

**AC5 — Labels en español:**

Eventos:
| Evento técnico | Label |
|---------------|-------|
| `message.received` | `Mensaje recibido` |
| `message.status_update` | `Estado de mensaje` |
| `session.connected` | `Sesión conectada` |
| `session.disconnected` | `Sesión desconectada` |

Estados del webhook:
| Condición | Label | Badge variant |
|-----------|-------|--------------|
| `activo = true`, `failure_count = 0` | `Activo` | `default` |
| `activo = true`, `failure_count > 0` | `Con fallos` | `destructive` |
| `activo = false` | `Inactivo` | `secondary` |

**AC6 — Empty state:**
**Dado** que el teléfono no tiene webhooks registrados,
**Cuando** se renderiza la página,
**Entonces** se muestra: "Este teléfono no tiene webhooks registrados. Los webhooks son creados por los integradores a través de la API."

**AC7 — Loading state:**
**Dado** que los datos están cargando,
**Cuando** se renderiza la página,
**Entonces** se muestran skeletons en la fila de resumen y en las filas de la tabla.

**AC8 — Breadcrumb y navegación:**
**Dado** que el admin está en la página de webhooks,
**Cuando** hace clic en volver,
**Entonces** regresa a `/empresas/[empresaId]/telefonos` (lista de teléfonos de la empresa).

## Tasks / Subtasks

- [x] **T1 — Interface `AdminWebhook` en `api.ts`** (AC3)

- [x] **T2 — Función `getAdminTelefonoWebhooks` en `api.ts`** (AC1)
  - [x] `GET /api/admin/telefonos/{id}/webhooks`
  - [x] Retorna `{ ok: boolean; webhooks: AdminWebhook[]; total: number; error?: string }`

- [x] **T3 — Crear página** `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/webhooks/page.tsx` (AC1-AC8)
  - [x] `"use client"` + `useParams` para `empresaId` y `telefonoId`
  - [x] `useEffect` para cargar teléfono (número) y webhooks en paralelo
  - [x] Header con número de teléfono y botón volver
  - [x] Fila bento de 3 mini-cards de resumen (activos, inactivos, con fallos)
  - [x] Tabla con shadcn `Table` y filas expandibles
  - [x] Accordion por fila con `Set<number>` local

- [x] **T4 — Helpers de formato** (AC3, AC5)
  - [x] `etiquetaEvento`, `etiquetaEstadoWebhook`, `tiempoRelativo`, `truncarUrl`

- [x] **T5 — Nota de solo lectura visible** (AC6)
  - [x] `Alert` con mensaje informativo visible en header y descripción de página

- [x] **T6 — Verificación visual** (AC1-AC8)
  - [x] TypeScript compila sin errores
  - [x] DataEmptyState para 0 webhooks
  - [x] Accordion con ChevronDown/Right indica estado
  - [x] Botón copiar URL con feedback "Copiado"

## Dev Notes

### Estructura de la página

```
Header
  ← Volver a teléfonos
  Webhooks de +51 999 888 777
  [badge: Conectado]
  Solo lectura — gestionado por integradores vía API

Fila bento resumen (grid-cols-3)
  Card: "2 activos"
  Card: "1 inactivo"
  Card: "1 con fallos" (destructivo si > 0)

Tabla
  Columnas: URL | Eventos | Estado | Fallos | Último éxito | Registrado
  Filas expandibles (Collapsible o estado local)
```

### Componentes a usar

| Elemento | Componente |
|----------|-----------|
| Resumen | `Card` en `grid-cols-3` |
| Tabla | `Table`, `TableHeader`, `TableRow`, `TableCell`, `TableHead` |
| Badges de evento | `Badge` variant `secondary` tamaño `sm` |
| Badge de estado | `Badge` con variant según estado |
| Expandir fila | `Collapsible` de shadcn o `useState` por fila |
| URL copiable | Button con `Copy` icon de lucide + `navigator.clipboard` |
| Loading | `Skeleton` en grid y en tabla |
| Empty state | `DataEmptyState` ya existente |
| Alerta info | `Alert` de shadcn |

### Patrón de accordion por fila (sin dependencia nueva)

```tsx
// estado local por webhook id
const [expanded, setExpanded] = useState<Set<number>>(new Set())

const toggle = (id: number) => {
  setExpanded(prev => {
    const next = new Set(prev)
    next.has(id) ? next.delete(id) : next.add(id)
    return next
  })
}

// en la fila:
<TableRow
  className="cursor-pointer hover:bg-muted/50"
  onClick={() => toggle(webhook.id)}
>
  ...
</TableRow>
{expanded.has(webhook.id) && (
  <TableRow>
    <TableCell colSpan={6} className="bg-muted/30 p-4">
      {/* detalle expandido */}
    </TableCell>
  </TableRow>
)}
```

### Tiempo relativo (helper simple)

```ts
function tiempoRelativo(fecha?: string | null): string {
  if (!fecha) return "Nunca"
  const diff = Date.now() - new Date(fecha).getTime()
  const min = Math.floor(diff / 60000)
  if (min < 1) return "hace un momento"
  if (min < 60) return `hace ${min} min`
  const hrs = Math.floor(min / 60)
  if (hrs < 24) return `hace ${hrs}h`
  return `hace ${Math.floor(hrs / 24)} días`
}
```

### Cómo obtener el número de teléfono para el header

Reutilizar `getAdminEmpresaTelefonos(empresaId)` y filtrar por `telefonoId`, o añadir una función `getAdminTelefono(telefonoId)` si ya existe el endpoint `GET /api/admin/telefonos/{id}` (existe en routes_admin.go línea 42).

### References

- Endpoint backend: `GET /api/admin/telefonos/{id}/webhooks` (story 6.3)
- Patrón de página similar: `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/api-keys/page.tsx`
- Componentes feedback: `frontend/components/feedback/`
- API client: `frontend/lib/api.ts`
- shadcn Table: `frontend/components/ui/table.tsx`
- shadcn Collapsible: verificar si está instalado con `ls frontend/components/ui/collapsible.tsx`

## Dev Agent Record

### Agent Model Used
claude-sonnet-4-6

### Debug Log References
- No existe `getAdminTelefonoById` en api.ts; teléfono obtenido con `getAdminEmpresaTelefonos` + find (mismo patrón de api-keys/page.tsx)
- Accordion implementado con `Set<number>` local (sin Collapsible — no está instalado en ui/)
- `fetchWithAuth` está definida en api.ts línea 754 (disponible para la nueva función)

### Completion Notes List
- T1: Interface `AdminWebhook` + `AdminWebhooksResponse` añadidos a api.ts
- T2: `getAdminTelefonoWebhooks(telefonoId)` → `GET /api/admin/telefonos/{id}/webhooks`
- T3-T5: Página creada con header, fila bento 3-cols, tabla expandible, helpers de formato, Alert de solo lectura, DataEmptyState
- T6: TypeScript limpio

### File List
- frontend/lib/api.ts
- frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/webhooks/page.tsx
