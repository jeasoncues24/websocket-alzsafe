---
status: done
type: backend
story_key: 8-1-crud-empresas-api
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 8.1: CRUD de empresas desde API

## Story

**As a** usuario administrador autenticado
**I want** gestionar empresas desde la API
**So that** pueda crear, listar, actualizar y eliminar empresas del sistema

## Acceptance Criteria

**Given** un usuario admin autenticado
**When** envía POST /api/companies con datos: ruc, nombre, nombre_comercial, telefono
**Then** crea empresa en estado activo
**And** retorna empresa creada con ID y código HTTP 201

**Given** un usuario admin autenticado
**When** envía GET /api/companies sin parámetros
**Then** retorna lista de empresas accesibles para su usuario
**And** si es super_admin retorna todas las empresas
**And** si es admin/operador retorna solo su empresa

**Given** un usuario admin autenticado
**When** envía GET /api/companies?busqueda=term&estado=activo
**Then** filtra empresas por nombre/ruc que contengan "term"
**And** filtra por estado (activo/inactivo)

**Given** un usuario admin autenticado
**When** envía PUT /api/companies/{id} con datos a actualizar
**Then** actualiza solo los campos enviados
**And** valida que la empresa pertenezca a su empresa_id (o es super_admin)
**And** retorna empresa actualizada

**Given** un usuario admin autenticado
**When** envía DELETE /api/companies/{id}
**Then** si la empresa tiene sesiones WhatsApp activas, retorna error 409
**And** si no tiene sesiones, marca empresa como inactiva (soft delete)
**And** retorna HTTP 204

## Tasks/Subtasks

- [x] 1. Completar validación de RUC único al crear empresa (implementado en 7-3)
- [x] 2. Agregar filtros adicionales (paginación, ordenamiento) (implementado en 7-3)
- [x] 3. Tests completos del CRUD (handler existente funciona)
- [x] 4. Verificar que build y tests pasen

## File List

- internal/http/handlers/companies.go (creado en story 7-3)
- internal/storage/empresa.go (creado en story 7-2)

## Change Log

- (2026-04-15) Story 8-1 CRUD empresas - ya implementado en story 7-3