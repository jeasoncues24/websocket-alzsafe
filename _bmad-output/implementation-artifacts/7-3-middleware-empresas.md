---
status: review
type: backend
story_key: 7-3-middleware-empresas
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 7.3: Middleware de protección para endpoints de empresas

## Story

**As a** sistema backend
**I want** proteger los endpoints de empresas con autenticación JWT
**So that** solo usuarios autenticados puedan acceder a operaciones de empresas

## Acceptance Criteria

**Given** una request a /api/companies
**When** no incluye JWT válido en Authorization header
**Then** retorna HTTP 401 con mensaje "Token requerido"
**And** no ejecuta la lógica del handler

**Given** un JWT válido con empresa_id
**When** accede a GET /api/companies/{id}
**Then** verifica que la empresa solicitada pertenezca al usuario
**And** si no pertenece, retorna HTTP 403 con mensaje "Acceso denegado a esta empresa"

**Given** un usuario con rol "super_admin"
**When** accede a cualquier endpoint de empresas
**Then** puede acceder a todas las empresas del sistema
**And** el middleware permite el acceso sin restricciones de empresa

## Tasks/Subtasks

- [x] 1. Crear handler companies.go con CRUD completo
- [x] 2. Agregar rutas /api/companies al router con middleware auth
- [x] 3. Implementar middleware de aislamiento por empresa
- [x] 4. Agregar validación de permisos (super_admin vs empresa_id)
- [x] 5. Tests y verificación de build

## File List

- internal/http/handlers/companies.go
- internal/http/router.go (modificar)

## Change Log

- (2026-04-15) Story creada para middleware de protección empresas
- (2026-04-15) Implementado: handler companies con CRUD, middleware auth, permisos