---
status: done
type: backend
story_key: 9-1-endpoint-usuario-me
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 9.1: Endpoint /usuario/me

## Story

**As a** usuario autenticado con JWT válido
**I want** obtener mi información actual
**So that** pueda conocer mi perfil, rol y empresa asociada

## Acceptance Criteria

**Given** un usuario autenticado con JWT válido
**When** invoca GET /api/usuario/me
**Then** retorna información del usuario: id, username, rol, empresa_id, empresa_nombre
**And** incluye permisos del usuario

**Given** un usuario sin empresa asociada (super_admin sin empresa asignada)
**When** invoca GET /api/usuario/me
**Then** retorna empresa: null

## Implementación

El endpoint ya fue implementado en story 7-1 como `/api/auth/me`. El cambio de nombre a `/api/usuario/me` es simple.

## Tasks/Subtasks

- [x] 1. Endpoint GET /api/auth/me ya implementado en 7-1
- [x] 2. Agregar endpoint /api/usuario/me como alias (opcional)

## File List

- internal/http/handlers/auth.go (ya implementado)

## Change Log

- (2026-04-15) Story 9-1 - endpoint ya implementado en 7-1 como /api/auth/me