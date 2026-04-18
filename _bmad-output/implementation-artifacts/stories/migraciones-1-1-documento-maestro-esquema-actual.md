---
status: done
type: docs
story_key: migraciones-1-1-documento-maestro-esquema-actual
created: 2026-04-18
last_updated: 2026-04-18
---

# Story 1.1: Documento maestro del esquema actual

## Story

**As a** equipo tecnico
**I want** un documento maestro del esquema actual
**So that** podamos borrar y recrear migraciones con una fuente de verdad clara y consistente

## Acceptance Criteria

1. **Given** el estado actual del codigo y las migraciones, **When** se crea el documento maestro, **Then** cada tabla relevante queda documentada con columnas, tipos, relaciones y proposito.

2. **Given** una columna de cualquier tabla, **When** se documenta su definicion, **Then** el documento indica explicitamente si es `NULL` o `NOT NULL` y cual es su `DEFAULT`.

3. **Given** una tabla con indices, uniques o llaves foraneas, **When** se documenta el esquema, **Then** se registran explicitamente los indices, PK, FK, uniques y restricciones relevantes.

4. **Given** el runner de migraciones actual, **When** se revisa la documentacion, **Then** se explica que usa `github.com/golang-migrate/migrate/v4`, MySQL y `internal/storage/migrations`.

5. **Given** la historia de migraciones actual, **When** se documenta el esquema maestro, **Then** se incluyen al menos las tablas `messages`, `broadcasts`, `broadcast_results`, `admin_users`, `empresas`, `telefonos`, `roles`, `modules`, `user_modules`, `token_blacklist`, `api_keys`, `api_key_usage_events`, `api_key_usage_daily`, `api_key_audit_events` y `schema_migrations`.

6. **Given** el documento maestro no fue aprobado, **When** alguien intenta borrar o regenerar migraciones, **Then** esa accion no debe considerarse lista para ejecucion.

7. **Given** una consulta critica del sistema, **When** se revisa el documento, **Then** se puede identificar que indice la soporta y por que existe.

## Tasks / Subtasks

- [ ] Levantar inventario completo del esquema actual desde migraciones y stores existentes
  - [ ] Revisar `internal/storage/migrations/`
  - [ ] Revisar stores y dominios relacionados con cada tabla

- [ ] Redactar el documento maestro de tablas
  - [ ] Definir columnas, tipos, nullability y defaults por tabla
  - [ ] Documentar PK, FK, uniques e indices
  - [ ] Incluir observaciones de relaciones y dependencias entre tablas

- [ ] Documentar el uso correcto de la libreria de migraciones
  - [ ] Explicar `go run . migrate status`
  - [ ] Explicar `go run . migrate up`
  - [ ] Explicar `go run . migrate down`
  - [ ] Explicar la ruta `internal/storage/migrations`
  - [ ] Explicar la limpieza legacy de `schema_migrations`

- [ ] Validar que el documento sirva como prerequisito para borrar migraciones
  - [ ] Verificar que no falten tablas criticas
  - [ ] Verificar que no falten defaults o nullability explicitos


## Dev Notes

### Contexto tecnico clave

- El proyecto ya usa `github.com/golang-migrate/migrate/v4` con MySQL.
- El runner vive en `internal/storage/migration.go`.
- El comando CLI de migraciones vive en `main.go`.
- La ruta real de migraciones es `internal/storage/migrations`.
- El runner elimina una tabla legacy `schema_migrations` con schema antiguo antes de aplicar las migraciones nuevas.

### Regla de documentacion obligatoria

- Cada columna debe quedar expresada con su estado exacto:
  - `nullable: true` o `nullable: false`
  - `default: ...` o `default: none`
- No asumir defaults implícitos.
- Si un campo es requerido, documentarlo como `NOT NULL` de forma explicita.

### Tablas a cubrir como minimo

- `messages`
- `broadcasts`
- `broadcast_results`
- `admin_users`
- `empresas`
- `telefonos`
- `roles`
- `modules`
- `user_modules`
- `token_blacklist`
- `api_keys`
- `api_key_usage_events`
- `api_key_usage_daily`
- `api_key_audit_events`
- `schema_migrations`

### References

- [Source: _bmad-output/planning-artifacts/epics-migraciones-telefonos.md#Story-11-Documento-maestro-del-esquema-actual]
- [Source: internal/storage/migration.go]
- [Source: main.go]
- [Source: go.mod]
- [Source: internal/storage/migrations/]

## Dev Agent Record

### Agent Model Used

gpt-5.4-mini

### Debug Log References

### Completion Notes List

- Se creó el documento maestro del esquema en `_bmad-output/implementation-artifacts/documento-maestro-esquema-migraciones.md`.
- Se documentaron columnas, nullability, defaults, indices y relaciones por tabla.
- Se agregaron reglas explícitas para `telefono_contacto`, relaciones lógicas y defaults `0` en columnas relacionales.

### File List

- `_bmad-output/implementation-artifacts/documento-maestro-esquema-migraciones.md`
