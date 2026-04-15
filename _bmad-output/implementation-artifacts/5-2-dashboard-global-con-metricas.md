# Story 5.2: Dashboard global con métricas

Status: ready-for-dev

## Story

As a operations manager,
I want ver métricas globales del sistema en el dashboard,
So that pueda identificar problemas y medir rendimiento al instante.

## Acceptance Criteria

1. **Given** el usuario accede a `/dashboard`, **When** carga la página, **Then** muestra 4 métricas principales: empresas activas, mensajes hoy, broadcasts hoy, tasa de éxito.

2. **Given** las métricas, **When** los datos se cargan, **Then** muestran skeleton loading mientras cargan desde `/metrics`.

3. **Given** el dashboard, **When** hay alertas (sesiones caídas, errores recientes), **Then** se muestran en sección "Alertas" con color distintivo (rojo).

4. **Given** una métrica, **When** el usuario hace click, **Then** navega a la vista detallada (click en "mensajes hoy" → `/messages`).

5. **Given** los datos, **When** se actualizan, **Then** el dashboard muestra "última actualización" y tiene botón de refresh manual.

## Tasks

- [ ] Crear `frontend/app/dashboard/page.tsx`
- [ ] Crear componentes MetricCard
- [ ] Consumir `GET /metrics` del backend
- [ ] Implementar skeleton loading
- [ ] Crear sección de alertas
- [ ] Agregar navegación a vistas detalladas
- [ ] Mostrar timestamp de última actualización

## Dev Notes

- Endpoint: `GET /metrics` ya existe
- Usar Zustand para cache de métricas
- Skeleton usar componente `Skeleton` de shadcn