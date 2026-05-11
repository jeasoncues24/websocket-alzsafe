---
story_id: "3.4"
epic: "epic-3"
title: "Fix Correctness — MetricCard Anidado, Hydration Mismatches y Array Keys"
status: backlog
estimated_days: 2
priority: high
skills: ["bmad-code-review"]
affects:
  - frontend/app/dashboard/page.tsx
  - frontend/components/companies/empresa-detail-modal.tsx
  - frontend/components/api-key-metrics.tsx
---

# Story 3.4: Fix Correctness — MetricCard, Hydration Mismatches y Array Keys

## Contexto

React Doctor detectó 3 categorías de bugs de correctness con ubicaciones exactas:

1. **`MetricCard` definido dentro de `DashboardPage`** (`app/dashboard/page.tsx:58`): React crea una nueva instancia del componente en cada render del padre, destruyendo el estado interno y forzando re-mounts innecesarios.

2. **`new Date()` en JSX** (`components/companies/empresa-detail-modal.tsx:207` ×4): El servidor renderiza un timestamp, el cliente hidrata con un timestamp diferente → mismatch, warning de hidratación, y parpadeo visual.

3. **Array index como key** (`components/api-key-metrics.tsx:78` ×6): React usa el índice para reconciliar listas; cuando la lista se reordena o filtra, los componentes reciben las props del elemento incorrecto.

## User Story

Como desarrollador del equipo,
quiero que el código del dashboard y modales esté libre de bugs de correctness detectados por React Doctor,
para que los componentes mantengan estado estable y no haya diferencias entre renders server y client.

## Scope

### Incluido
- Mover `MetricCard` fuera de `DashboardPage` (a nivel de módulo en el mismo archivo o a `components/`)
- Corregir los 4 usos de `new Date()` en JSX en `empresa-detail-modal.tsx`
- Reemplazar índices de array por keys estables en `api-key-metrics.tsx`

### Excluido
- Cambios en la lógica de negocio de métricas, empresas o modales
- Cambios en el diseño visual de los componentes

## Acceptance Criteria

**AC1 — MetricCard movido fuera del padre:**
**Dado** que `MetricCard` está en `app/dashboard/page.tsx:58` dentro del cuerpo de `DashboardPage`
**Cuando** se mueve a nivel de módulo (antes de la declaración de `DashboardPage`) o a un archivo separado
**Entonces** React no crea nueva instancia de `MetricCard` en cada render de `DashboardPage`
**Y** React Doctor no reporta "Nested component definition" en `app/dashboard/page.tsx`

**AC2 — Hydration mismatches eliminados:**
**Dado** que `empresa-detail-modal.tsx:207` usa `new Date()` directamente en JSX (×4)
**Cuando** se envuelven en `useEffect + useState`:
```tsx
const [fecha, setFecha] = useState<string | null>(null);
useEffect(() => { setFecha(new Date(value).toLocaleDateString('es')); }, [value]);
return <span>{fecha ?? '—'}</span>;
```
**O** se agrega `suppressHydrationWarning` al elemento padre si el valor es puramente cosmético
**Entonces** no hay warnings de hidratación en la consola del browser
**Y** los timestamps se muestran correctamente después de la hidratación

**AC3 — Keys estables en listas:**
**Dado** que `api-key-metrics.tsx:78` usa `i` (índice) como key en ×6 mapeos
**Cuando** se identifican los campos únicos de cada item (`item.id`, `item.day`, `item.bucket`, `point.bucket`, etc.)
**Entonces** cada elemento de lista usa un identificador estable como key
**Y** React Doctor no reporta "Array index as key" en `api-key-metrics.tsx`

**AC4 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores de TypeScript ni ESLint

## Notas de Implementación

- Para `MetricCard`: si el componente es pequeño, moverlo al principio del mismo archivo (`app/dashboard/page.tsx`) antes de `DashboardPage` es suficiente. Si crece, moverlo a `components/dashboard/metric-card.tsx`.
- Para las keys en `api-key-metrics.tsx`: revisar la forma de los datos de la API (TelemetryTimeSeriesPoint tiene `bucket: string` — usarlo como key).
- Ejecutar `npx react-doctor@latest . --verbose` al final para confirmar que los 11 issues de Correctness quedan resueltos.
