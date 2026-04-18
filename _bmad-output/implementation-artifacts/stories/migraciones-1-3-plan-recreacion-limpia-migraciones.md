---
status: done
type: docs
story_key: migraciones-1-3-plan-recreacion-limpia-migraciones
created: 2026-04-18
last_updated: 2026-04-18
---

# Story 1.3: Plan de recreación limpia de migraciones

## Story

**As a** desarrollador
**I want** un plan claro para borrar y recrear migraciones de forma limpia
**So that** el esquema pueda reconstruirse sin duplicados, deuda técnica accidental ni pérdida de trazabilidad

## Acceptance Criteria

1. **Given** el documento maestro del esquema ya fue creado y revisado, **When** se redacta el plan de recreación, **Then** el plan indica que no se borra ninguna migración antes de aprobar la documentación base.

2. **Given** una base de datos existente, **When** se prepara el reset, **Then** el plan contempla backup/export del estado actual y validación previa del esquema documentado.

3. **Given** el directorio `internal/storage/migrations`, **When** se recrean las migraciones, **Then** el plan define una secuencia ordenada de archivos `*.up.sql` y `*.down.sql` sin duplicados ni versiones ambiguas.

4. **Given** el runner actual de migraciones, **When** se sigue el plan, **Then** se especifica cómo validar `status`, `up` y `down` después de recrear el esquema.

5. **Given** la estrategia de reset aprobada, **When** se ejecuta la recreación, **Then** el esquema nuevo debe coincidir con el documento maestro y no introducir defaults `0` en columnas relacionales nuevas.

6. **Given** un reset completo desde cero, **When** se aplican todas las migraciones, **Then** el sistema debe poder levantar el esquema sin errores y sin divergencias respecto al documento maestro.

## Tasks / Subtasks

- [ ] Definir la estrategia de reset antes de borrar archivos
  - [ ] Enumerar prerequisitos obligatorios
  - [ ] Definir backup y punto de restauración
  - [ ] Definir criterios de aprobacion para iniciar el borrado

- [ ] Definir la nueva secuencia de migraciones
  - [ ] Ordenar las tablas segun dependencias reales
  - [ ] Separar cambios de estructura y datos semilla
  - [ ] Mantener nombres claros y versionados

- [ ] Definir validaciones posteriores al reset
  - [ ] Verificar `go run . migrate status`
  - [ ] Verificar `go run . migrate up`
  - [ ] Verificar `go run . migrate down`
  - [ ] Verificar ausencia de indices duplicados y defaults problemáticos

- [ ] Documentar riesgos y puntos de control
  - [ ] Relaciones logicas sin FK fisica
  - [ ] Tablas que dependen de `telefono_id`
  - [ ] Campos relacionales con defaults historicos `0`

## Dev Notes

### Contexto tecnico clave

- El documento maestro de esquema ya existe y debe actuar como fuente de verdad.
- El runner de migraciones usa `github.com/golang-migrate/migrate/v4` con MySQL.
- Los archivos de migración viven en `internal/storage/migrations`.
- El CLI de migraciones vive en `main.go`.

### Reglas del plan

- No borrar ninguna migración antes de aprobar el documento maestro.
- No mezclar recreación de esquema con cambios funcionales nuevos.
- No introducir nuevos defaults `0` en relaciones.
- No asumir FK implícitas si no existen en el esquema final.

### Referencias

- [Source: _bmad-output/implementation-artifacts/documento-maestro-esquema-migraciones.md]
- [Source: _bmad-output/planning-artifacts/epics-migraciones-telefonos.md#Story-13-Plan-de-recreacion-limpia-de-migraciones]
- [Source: main.go]
- [Source: internal/storage/migration.go]
- [Source: internal/storage/migrations/]

## Dev Agent Record

### Agent Model Used

gpt-5.4-mini

### Debug Log References

### Completion Notes List

- Se creó el plan de recreación limpia en `_bmad-output/implementation-artifacts/plan-recreacion-limpia-migraciones.md`.
- Se dejaron prerrequisitos, reglas, orden de reconstrucción y validaciones posteriores.
- Se reforzó la regla de no borrar migraciones antes del documento maestro.

### File List

- `_bmad-output/implementation-artifacts/plan-recreacion-limpia-migraciones.md`
