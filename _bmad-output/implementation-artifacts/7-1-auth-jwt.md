---
status: done
type: backend
story_key: 7-1-auth-jwt
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 7.1: Sistema de autenticación JWT con login y logout

## Story

**As a** usuario administrador del sistema
**I want** autenticarme con username/password y recibir un JWT
**So that** pueda acceder a los endpoints protegidos del sistema

## Acceptance Criteria

**Given** un usuario administrador del sistema
**When** envía solicitud POST a /api/auth/login con username y password
**Then** el sistema valida credenciales contra tabla admin_users
**And** si son válidas, genera JWT con claims: usuario_id, username, rol, empresa_id, empresa_nombre
**And** retorna token con expiry configurable (default 24h)

**Given** un usuario autenticado
**When** envía solicitud POST a /api/auth/logout
**Then** el token se agrega a blacklist en DB
**And** retorna confirmación de logout exitoso

**Given** un JWT válido
**When** el token está por expirar (menos de 1 hora)
**Then** el endpoint /api/auth/refresh retorna nuevo token
**And** el token anterior se invalida

## Tasks/Subtasks

- [x] 1. Crear config/jwt.go con configuración de JWT (secret, expiry, issuer)
- [x] 2. Crear middleware/auth.go con funciones de validación JWT
- [x] 3. Crear handler auth.go con endpoints login, logout, refresh
- [x] 4. Agregar ruta /api/auth al router
- [x] 5. Agregar endpoint /api/auth/login
- [x] 6. Agregar endpoint /api/auth/logout
- [x] 7. Agregar endpoint /api/auth/refresh
- [x] 8. Agregar endpoint /api/auth/me (info del usuario actual)
- [x] 9. Crear storage/token_blacklist.go para blacklist de tokens
- [x] 10. Verificar que build compile y tests pasen

## Dev Notes

### Architecture Requirements

- JWT con HS256
- Secret key configurable por entorno
- Expiry default 24h, configurable
- Claims incluyen: user_id, username, rol, empresa_id, empresa_ruc, empresa_nombre
- Blacklist de tokens en tabla (para logout)

### Technical Specifications

- Usar library golang-jwt/jwt/v5
- Password verification con bcrypt
- Token blacklist en tabla mysql (tabla: token_blacklist)

## File List

- internal/config/jwt.go
- internal/http/middleware/auth.go
- internal/http/handlers/auth.go
- internal/http/router.go (modificar)
- internal/storage/token_blacklist.go
- internal/storage/migrations/008_create_token_blacklist_table.up.sql
- internal/domain/admin_user.go (agregado WithTokenClaims context)

## Change Log

- (2026-04-15) Story creada para implementar autenticación JWT
- (2026-04-15) Implementados: config JWT, middleware, handlers, rutas /api/auth/\*

### Review Findings

- [x] [Review][Patch] JTI no es aleatorio — colisión en logins rápidos del mismo usuario [internal/http/handlers/auth.go — generateToken()]
- [x] [Review][Defer] `UpdateLastLogin` error ignorado — non-critical path, pre-existing — deferred, pre-existing
- [x] [Review][Patch] Missing 008_create_token_blacklist_table.down.sql — sin rollback para migración 008 [internal/storage/migrations/]
