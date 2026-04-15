---
status: ready-for-dev
type: frontend
story_key: 8-3-panel-empresas-frontend
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 8.3: Panel admin para gestión de empresas (Frontend)

## Story

**As a** usuario autenticado en el panel
**I want** gestionar empresas desde la interfaz
**So that** pueda crear, editar y ver empresas sin usar la API directamente

## Acceptance Criteria

**Given** un usuario autenticado en el panel
**When** accede a /companies
**Then** muestra lista de empresas con columnas: RUC, Nombre, Estado, Sesión WhatsApp
**And** incluye búsqueda por nombre o RUC
**And** incluye filtros por estado (activo/inactivo)

**Given** un usuario autenticado
**When** hace clic en "Nueva Empresa"
**Then** muestra modal/formulario con campos: RUC, Nombre, Teléfono, Dirección
**And** valida que RUC no exista previamente

**Given** un usuario autenticado
**When** hace clic en Editar empresa
**Then** muestra formulario con datos actuales
**And** permite modificar nombre, teléfono, dirección, estado

**Given** un usuario autenticado
**When** hace clic en Ver detalle de empresa
**Then** muestra panel con: datos de empresa, sesión WhatsApp actual, últimos mensajes

## Tasks/Subtasks

- [ ] 1. Crear componente de lista de empresas con búsqueda y filtros
- [ ] 2. Crear modal/formulario para nueva empresa
- [ ] 3. Crear formulario de edición de empresa
- [ ] 4. Crear página de detalle de empresa
- [ ] 5. Conectar con API /api/companies
- [ ] 6. Tests y verificación

## File List

- frontend/app/companies/page.tsx (actualizar)
- frontend/components/companies/ (nuevos componentes)

## Change Log

- (2026-04-15) Story creada para panel frontend de empresas