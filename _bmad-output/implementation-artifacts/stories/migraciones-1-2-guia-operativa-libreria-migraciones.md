---
status: done
type: docs
story_key: migraciones-1-2-guia-operativa-libreria-migraciones
created: 2026-04-18
last_updated: 2026-04-18
---

# Story 1.2: Guía operativa de la librería de migraciones

## Story

**As a** desarrollador
**I want** una guía operativa clara de la librería de migraciones
**So that** pueda ejecutar, mantener y diagnosticar las migraciones sin depender de conocimiento oral

## Acceptance Criteria

1. **Given** el estado actual del proyecto, **When** se documenta la librería de migraciones, **Then** queda claro que se usa `github.com/golang-migrate/migrate/v4` con driver MySQL y source file.

2. **Given** el comando de migraciones del proyecto, **When** se lee la guía, **Then** queda explicado cómo ejecutar `go run . migrate status`, `go run . migrate up` y `go run . migrate down`.

3. **Given** la ruta real del esquema, **When** se documenta el flujo de migraciones, **Then** queda claro que los archivos viven en `internal/storage/migrations`.

4. **Given** la implementación actual del runner, **When** se documenta el proceso, **Then** se explica que el runner limpia una tabla legacy `schema_migrations` con esquema antiguo antes de aplicar migraciones nuevas.

5. **Given** una migración fallida o una base de datos en estado dirty, **When** se revisa la guía, **Then** se explica cómo identificar el problema y qué comando usar para inspeccionar o revertir el estado.

6. **Given** una nueva migración SQL, **When** se sigue la guía, **Then** queda claro el formato de nombre, el orden esperado y la convención de uso para `up` y `down`.

## Tasks / Subtasks

- [ ] Documentar el flujo real de ejecución del runner
  - [ ] Explicar el entrypoint CLI en `main.go`
  - [ ] Explicar el uso de `MigrationRunner` en `internal/storage/migration.go`

- [ ] Documentar comandos operativos
  - [ ] `go run . migrate status`
  - [ ] `go run . migrate up`
  - [ ] `go run . migrate down`
  - [ ] Indicar prerequisitos de DB antes de ejecutar

- [ ] Documentar convenciones de archivos y orden
  - [ ] Ubicación de migraciones
  - [ ] Convención `*.up.sql` / `*.down.sql`
  - [ ] Secuencia de versionado y control de cambios

- [ ] Documentar troubleshooting basico
  - [ ] Estado dirty
  - [ ] Migracion faltante o duplicada
  - [ ] Inspeccion de version actual y migraciones aplicadas

## Dev Notes

### Contexto tecnico clave

- El CLI vive en `main.go` y expone `migrate status`, `migrate up` y `migrate down`.
- El runner vive en `internal/storage/migration.go`.
- La libreria instalada es `github.com/golang-migrate/migrate/v4`.
- El driver de base de datos es MySQL.
- La ruta de migraciones actual es `internal/storage/migrations`.

### Comportamiento actual a documentar

- `status` lista version actual y migraciones aplicadas.
- `up` aplica migraciones pendientes.
- `down` revierte la ultima migracion.
- Antes de aplicar migraciones nuevas, el runner intenta limpiar la tabla legacy `schema_migrations` si tiene el esquema antiguo.

### Regla de documentacion

- La guia debe ser operativa, no teorica.
- Debe decirle al desarrollador exactamente que comando usar, en que orden y con que expectativa.
- Debe mencionar como reconocer un estado dirty y que significa.

### References

- [Source: _bmad-output/planning-artifacts/epics-migraciones-telefonos.md#Story-12-Guia-operativa-de-la-libreria-de-migraciones]
- [Source: main.go]
- [Source: internal/storage/migration.go]
- [Source: go.mod]

## Dev Agent Record

### Agent Model Used

gpt-5.4-mini

### Debug Log References

### Completion Notes List

- Se creó la guía operativa en `_bmad-output/implementation-artifacts/guia-operativa-migraciones.md`.
- Se documentaron comandos `status`, `up` y `down`, junto con el flujo real del runner.
- Se explicó el manejo del estado dirty y la ruta `internal/storage/migrations`.

### File List

- `_bmad-output/implementation-artifacts/guia-operativa-migraciones.md`
