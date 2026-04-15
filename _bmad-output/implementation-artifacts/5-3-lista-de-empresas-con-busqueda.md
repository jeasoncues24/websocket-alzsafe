# Story 5.3: Lista de empresas con búsqueda

Status: ready-for-dev

## Story

As a operations manager,
I want ver todas las empresas registradas con su estado de sesión,
So that pueda monitorear y gestionar conexiones rápidamente.

## Acceptance Criteria

1. **Given** `/companies`, **When** carga, **Then** muestra tabla: RUC, nombre, estado sesión (badge), último mensaje, acciones.

2. **Given** búsqueda, **When** usuario escribe RUC o nombre, **Then** filtra en tiempo real (debounce 300ms).

3. **Given** estado de sesión, **When** es "active", **Then** badge verde; "connecting" → amarillo; "inactive" → rojo.

4. **Given** click en empresa, **When** navega a `/companies/[id]`, **Then** muestra: detalle empresa, historial mensajes, broadcasts recientes.

## Tasks

- [ ] Crear `GET /admin/companies` endpoint en backend
- [ ] Crear `app/companies/page.tsx`
- [ ] Implementar tabla con shadcn Table
- [ ] Agregar búsqueda con debounce
- [ ] Crear badge de estado
- [ ] Crear page detalle `/companies/[id]/page.tsx`
- [ ] Mostrar historial mensajes/broadcasts