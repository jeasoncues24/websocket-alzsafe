---
status: done
type: backend
story_key: 7-2-schema-db
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 7.2: Schema de base de datos para usuarios y API keys

## Story

**As a** sistema backend
**I want** crear tablas de usuarios admin, empresas y API keys
**So that** el sistema tenga la base de datos necesaria para autenticación y gestión multiempresa

## Acceptance Criteria

**Given** el sistema necesita almacenar usuarios admin
**When** se ejecutan migraciones
**Then** crea tabla `admin_users`:

- id (INT, PK, AUTO_INCREMENT)
- username (VARCHAR(50), UNIQUE, NOT NULL)
- password_hash (VARCHAR(255), NOT NULL)
- empresa_id (INT, FK, nullable para super_admin)
- rol (ENUM: 'super_admin', 'admin', 'operador')
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

**Given** el sistema necesita almacenar API keys
**When** se ejecutan migraciones
**Then** crea tabla `api_keys`:

- id (INT, PK, AUTO_INCREMENT)
- empresa_id (INT, FK, NOT NULL)
- key_hash (VARCHAR(255), NOT NULL)
- nombre (VARCHAR(100))
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- expires_at (TIMESTAMP, nullable)

**Given** el sistema necesita almacenar empresas
**When** se ejecutan migraciones
**Then** crea tabla `empresas`:

- id (INT, PK, AUTO_INCREMENT)
- ruc (VARCHAR(11), UNIQUE, NOT NULL)
- nombre (VARCHAR(255), NOT NULL)
- nombre_comercial (VARCHAR(255))
- telefono (VARCHAR(20))
- direccion (TEXT)
- activo (BOOLEAN, DEFAULT TRUE)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

## Tasks/Subtasks

- [x] 1. Crear migración 005_create_empresas_table.up.sql
- [x] 2. Crear migración 005_create_empresas_table.down.sql
- [x] 3. Crear migración 006_update_admin_users_table.up.sql (agrega empresa_id y super_admin)
- [x] 4. Crear migración 006_update_admin_users_table.down.sql
- [x] 5. Crear migración 007_create_api_keys_table.up.sql
- [x] 6. Crear migración 007_create_api_keys_table.down.sql
- [x] 7. Crear modelo Go para Empresa en internal/domain/empresa.go
- [x] 8. Crear modelo Go para AdminUser en internal/domain/admin_user.go
- [x] 9. Crear modelo Go para ApiKey en internal/domain/api_key.go
- [x] 10. Crear storage/empresa.go con métodos CRUD
- [x] 11. Crear storage/admin_user.go con métodos CRUD
- [x] 12. Crear storage/api_key.go con métodos CRUD
- [x] 13. Ejecutar migraciones y verificar estructura
- [x] 14. Crear seed para usuario admin inicial (admin/admin123) en cmd/seed/main.go

## Dev Notes

### Architecture Requirements

- Tablas con foreign keys y índices para rendimiento
- Passwords almacenados como hash bcrypt (cost 12)
- API keys almacenadas como hash SHA-256
- Soft delete para empresas (campo activo)

### Technical Specifications

- Usar sistema de migraciones existente en internal/storage/migration.go
- Convenciones de nomenclatura: singular para tablas, PascalCase para modelos
- Timestamps con timezone UTC

## File List

- internal/storage/migrations/005_create_empresas_table.up.sql
- internal/storage/migrations/005_create_empresas_table.down.sql
- internal/storage/migrations/006_update_admin_users_table.up.sql
- internal/storage/migrations/006_update_admin_users_table.down.sql
- internal/storage/migrations/007_create_api_keys_table.up.sql
- internal/storage/migrations/007_create_api_keys_table.down.sql
- internal/domain/empresa.go
- internal/domain/admin_user.go
- internal/domain/api_key.go
- internal/storage/empresa.go
- internal/storage/admin_user.go
- internal/storage/api_key.go
- cmd/seed/main.go

## Change Log

- (2026-04-15) Story creada desde planning Epic 7
- (2026-04-15) Implementadas migraciones, modelos y storage para empresas, usuarios y API keys
