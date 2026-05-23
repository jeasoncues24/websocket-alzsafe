---
story_id: "6.2"
epic: "epic-6"
title: "Frontend — Rediseño bento grid de lista de teléfonos"
status: review
estimated_days: 2
priority: high
branch: "feature/panel-telefonos-webhooks"
skills: ["ui-ux-pro-max", "bmad-code-review"]
affects:
  - frontend/app/empresas/[empresaId]/telefonos/page.tsx
  - frontend/lib/api.ts
---

# Story 6.2: Frontend — Rediseño bento grid de lista de teléfonos

## Story

Como admin del panel wsapi,
quiero ver la lista de teléfonos de una empresa como un bento grid con información relevante en cada card,
para diagnosticar el estado operativo de un número sin tener que hacer clics adicionales.

## Contexto técnico

Vista actual: `frontend/app/empresas/[empresaId]/telefonos/page.tsx` — lista vertical de divs con flex. Cada "card" es un `div` con clase `rounded-xl border bg-card p-4`. No usa el componente `Card` de shadcn correctamente, no hay grid, y no muestra conteos de API keys ni webhooks.

La story 6.1 añade `api_key_count` y `webhook_count` a la respuesta del backend. Esta story los consume.

Stack: Next.js App Router, React 19, shadcn/ui, Tailwind CSS 4, lucide-react.

## Acceptance Criteria

**AC1 — Layout bento grid responsivo:**
**Dado** que hay teléfonos cargados,
**Cuando** se renderiza la vista,
**Entonces** el grid es: `1 columna` en mobile, `2 columnas` en tablet (md), `3 columnas` en desktop (lg).

**AC2 — Card de teléfono con shadcn Card:**
**Dado** un teléfono en el grid,
**Cuando** se renderiza su card,
**Entonces** usa el componente `Card` de shadcn con `CardHeader` y `CardContent`
**Y** muestra: número completo, badge de estado, contador de claves activas, contador de webhooks.

**AC3 — Card grande para teléfonos activos con webhooks:**
**Dado** un teléfono con estado `active` (runtime conectado) y `webhook_count > 0`,
**Cuando** se renderiza en desktop,
**Entonces** la card ocupa `col-span-2` (dos columnas).

**AC4 — Botón Webhooks en cada card:**
**Dado** cualquier teléfono,
**Cuando** se renderiza su card,
**Entonces** existe un botón `[Webhooks]` variant `outline` que navega a `/empresas/[id]/telefonos/[id]/webhooks`.

**AC5 — Labels de estado en español:**

| Estado técnico | Label UI | Variante badge |
|---------------|----------|---------------|
| `active` + runtime_connected=true | `Conectado` | `default` (verde) |
| `disconnected` o runtime_connected=false | `Desconectado` | `destructive` |
| mismatch=true | `Desajuste` | amarillo/warning |
| cualquier otro | `En espera` | `secondary` |

**AC6 — Contadores mini:**
**Dado** un teléfono con `api_key_count = 3` y `webhook_count = 2`,
**Cuando** se renderiza la card,
**Entonces** muestra `3 claves activas` (icono KeyRound) y `2 webhooks` (icono Webhook o Link).

**AC7 — Alerta visual para webhooks con fallos:**
**Dado** que el endpoint incluye conteo de webhooks (ver nota en Dev Notes),
**Cuando** `webhook_count > 0`,
**Entonces** el contador de webhooks es visible (no oculto si es 0).
*Nota: la alerta de fallos es nice-to-have para esta story; el conteo básico es suficiente.*

**AC8 — Loading y empty states:**
**Dado** que los datos están cargando,
**Cuando** se muestra el skeleton,
**Entonces** el skeleton respeta el grid bento (3 cards skeleton en grid).
**Dado** que no hay teléfonos,
**Entonces** el empty state dice: "No hay teléfonos registrados para esta empresa."

**AC9 — Sin regresión en acciones existentes:**
**Dado** los cambios aplicados,
**Cuando** se interactúa con Editar, Eliminar, API Keys, Conectar/Ver QR,
**Entonces** todas las acciones siguen funcionando igual que antes.

## Tasks / Subtasks

- [x] **T1 — Actualizar interface `AdminTelefono` en `api.ts`** (AC6)
  - [x] Añadir `api_key_count?: number` y `webhook_count?: number` al interface

- [x] **T2 — Reescribir layout de la página** (AC1, AC2, AC3)
  - [x] Reemplazar el `div` contenedor por `<div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">`
  - [x] Envolver cada teléfono en `<Card className={cn("...", isLarge && "md:col-span-2")}>`
  - [x] Lógica `isLarge`: `telefono.status === "active" && telefono.runtime_connected && (telefono.webhook_count ?? 0) > 0`

- [x] **T3 — Contenido de la card** (AC2, AC4, AC5, AC6)
  - [x] `CardHeader`: número completo + `<Badge>` de estado en español
  - [x] `CardContent`: mini-stats (claves, webhooks) + botones de acción
  - [x] Botón `[Webhooks]` con icono `Link` de lucide

- [x] **T4 — Función helper de estado** (AC5)
  - [x] `function estadoTelefono(t: AdminTelefono): { label: string; variant: ... }`
  - [x] Mapeado: `active`+runtime → `"Conectado"`; `disconnected` → `"Desconectado"`; mismatch → `"Desajuste"`; else → `"En espera"`

- [x] **T5 — Skeleton y empty state** (AC8)
  - [x] Skeleton: `Array(3).fill(0).map(...)` con `<Card><Skeleton ... /></Card>` en grid
  - [x] Empty state: componente `DataEmptyState` ya existente en `components/feedback/data-empty-state.tsx`

- [x] **T6 — Verificación visual** (AC1-AC9)
  - [x] TypeScript compila sin errores (`tsc --noEmit`)
  - [x] Botón Webhooks navega a `/empresas/[id]/telefonos/[id]/webhooks`
  - [x] Acciones Editar, Eliminar, API Keys, Conectar conservadas

## Dev Notes

### Estructura de la card propuesta

```tsx
<Card className={cn(
  "transition-shadow hover:shadow-md",
  isLarge && "md:col-span-2"
)}>
  <CardHeader className="pb-3">
    <div className="flex items-center justify-between gap-2">
      <div className="flex items-center gap-2">
        <Phone className="h-4 w-4 text-muted-foreground" />
        <CardTitle className="text-base font-semibold">
          {telefono.numero_completo}
        </CardTitle>
      </div>
      <Badge variant={estado.variant}>{estado.label}</Badge>
    </div>
  </CardHeader>
  <CardContent className="space-y-4">
    {/* Mini stats */}
    <div className="flex gap-4 text-sm text-muted-foreground">
      <span className="flex items-center gap-1">
        <KeyRound className="h-3.5 w-3.5" />
        {telefono.api_key_count ?? 0} claves activas
      </span>
      <span className="flex items-center gap-1">
        <Link className="h-3.5 w-3.5" />
        {telefono.webhook_count ?? 0} webhooks
      </span>
    </div>
    {/* Acciones */}
    <div className="flex flex-wrap gap-2">
      <Button size="sm" variant="outline" onClick={() => openEdit(telefono)}>
        <Pencil className="mr-1.5 h-3.5 w-3.5" /> Editar
      </Button>
      <Button size="sm" variant="outline" onClick={() => router.push(`.../api-keys`)}>
        <KeyRound className="mr-1.5 h-3.5 w-3.5" /> API Keys
      </Button>
      <Button size="sm" variant="outline" onClick={() => router.push(`.../webhooks`)}>
        <Link className="mr-1.5 h-3.5 w-3.5" /> Webhooks
      </Button>
      {telefono.status !== "active" && (
        <Button size="sm" variant="outline" onClick={...}>
          <QrCode className="mr-1.5 h-3.5 w-3.5" /> Conectar
        </Button>
      )}
      <Button size="sm" variant="destructive" onClick={() => handleDelete(telefono)}>
        <Trash2 className="mr-1.5 h-3.5 w-3.5" /> Eliminar
      </Button>
    </div>
  </CardContent>
</Card>
```

### Variantes de badge de estado

shadcn no tiene variante `warning` por defecto. Opciones:
- Usar `variant="outline"` con clase `text-amber-600 border-amber-400` para `Desajuste`
- O añadir variante custom en `components/ui/badge.tsx` (preferible: clase inline para no cambiar el sistema)

### Icono para webhooks

`lucide-react` tiene `Webhook` desde v0.400+. Verificar con:
```bash
grep -r "from 'lucide-react'" frontend/app/ | head -5
# Si no existe Webhook, usar Link2 o Globe
```

### References

- Página actual: `frontend/app/empresas/[empresaId]/telefonos/page.tsx`
- API client: `frontend/lib/api.ts` (interface AdminTelefono ~línea 98)
- Card shadcn: `frontend/components/ui/card.tsx`
- Badge shadcn: `frontend/components/ui/badge.tsx`
- DataEmptyState: `frontend/components/feedback/data-empty-state.tsx`
- SessionStatusBadge (referencia de lógica de estado): `frontend/components/session/session-status-badge.tsx`

## Dev Agent Record

### Agent Model Used
claude-sonnet-4-6

### Debug Log References
- `Webhook` icon existe en lucide-react (v0.400+), confirmado
- Badge no tiene variante `warning`; se usó `variant="outline"` + `className="text-amber-600 border-amber-400"` para Desajuste
- `SessionStatusBadge` reemplazado por función helper `estadoTelefono` directa (más simple para el nuevo layout)

### Completion Notes List
- T1: `api_key_count` y `webhook_count` añadidos al interface `AdminTelefono`
- T2-T5: Página reescrita con bento grid (1/2/3 col), cards shadcn, skeleton en grid, DataEmptyState
- `estadoTelefono()` helper: mapea status+runtime+mismatch a label/variant en español
- Cards grandes (`md:col-span-2`) para teléfonos activos con webhooks
- Botón Webhooks navega a `/empresas/{id}/telefonos/{id}/webhooks`

### File List
- frontend/lib/api.ts
- frontend/app/empresas/[empresaId]/telefonos/page.tsx
