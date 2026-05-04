# Story 1.7: Normalizar migraciones — una tabla por migración

Status: review

## Story

As a desarrollador,
I want que cada migración represente una tabla completa y correcta,
so that un fresh install ejecute `./wsapi migrate up` sin errores de columna duplicada ni ALTERs redundantes.

## Contexto y Problema

El binario falla en fresh install con:

```
Error 1060 (42S21): Duplicate column name 'is_root'
```

**Causa raíz:** Las migraciones CREATE TABLE fueron actualizadas para incluir el esquema final completo (columnas como `is_root`, `created_by`, `updated_by`), pero las migraciones ALTER TABLE que originalmente añadían esas columnas nunca fueron eliminadas. Resultado: conflictos al correr en una base de datos limpia.

### Conflictos identificados

| Migración ALTER | Conflicto |
|---|---|
| `020_add_is_root_to_roles` | `007` ya crea `is_root` → Error 1060 |
| `022_add_audit_columns_to_empresas` | `005` ya tiene `created_by`, `updated_by` → Error 1060 |
| `023_add_audit_columns_to_telefonos` | `006` ya tiene `created_by`, `updated_by` → Error 1060 |
| `024_add_audit_columns_to_roles` | `007` ya tiene `created_by`, `updated_by` → Error 1060 |

### Columnas que aún faltan en los CREATE TABLE

| Tabla | Columnas faltantes | Migración origen |
|---|---|---|
| `messages` | `adjuntos_json`, `error_reason`, `retry_count`, `last_attempt_at`, `timestamp_created/sent/confirmed` | 017 + 018 |
| `modules` | `slug` | 015 |
| `api_keys` | `updated_by` | 025 |
| `empresas` | columna se llama `telefono` (debe ser `telefono_contacto`) | 016 |

## Solución adoptada

Reescribir las migraciones siguiendo el principio: **una migración = una tabla completa y final**. Los seeds (INSERT INTO) se consolidan en una única migración final.

No se usan migraciones ALTER para columnas que debieron existir desde el inicio.

## Nueva estructura de migraciones

| # | Archivo | Acción |
|---|---|---|
| 001 | `messages` — esquema completo (absorbe 017 + 018) | Actualizar |
| 002 | `broadcasts` | Sin cambios |
| 003 | `broadcast_results` | Sin cambios |
| 004 | `admin_users` — mover INSERT a 016 | Actualizar |
| 005 | `empresas` — `telefono_contacto`, absorbe 016 + 022 | Actualizar |
| 006 | `telefonos` — sin cambios (ya completa) | Sin cambios |
| 007 | `roles` — sin cambios (ya completa) | Sin cambios |
| 008 | `modules` — agregar `slug`, absorbe 015 | Actualizar |
| 009 | `user_modules` — mover INSERT a 016 | Actualizar |
| 010 | `token_blacklist` | Sin cambios |
| 011 | `api_keys` — agregar `updated_by`, absorbe 025 | Actualizar |
| 012 | `api_key_usage_events` | Sin cambios |
| 013 | `api_key_usage_daily` | Sin cambios |
| 014 | `api_key_audit_events` | Sin cambios |
| 015 | `audit_log` — renombrado desde 019 | Renombrar |
| 016 | **Seeds** — todos los INSERT consolidados | Nuevo |

### Archivos a eliminar

```
015_add_missing_columns.up/down.sql
016_rename_empresa_telefono_contacto.up/down.sql
017_align_messages_schema_with_repository.up/down.sql
018_add_retry_fields_to_messages.up/down.sql
019_create_audit_log_table.up/down.sql       ← renombrado a 015
020_add_is_root_to_roles.up/down.sql
022_add_audit_columns_to_empresas.up/down.sql
023_add_audit_columns_to_telefonos.up/down.sql
024_add_audit_columns_to_roles.up/down.sql
025_add_audit_columns_to_api_keys.up/down.sql
```

### Seeds consolidados en 016 (referencia)

Provienen de las migraciones actuales:
- `roles`: 4 filas (super_admin, admin, operador, viewer) — actualmente en 007
- `empresas`: 1 fila demo — actualmente en 005
- `admin_users`: 1 fila (admin_usqay) — actualmente en 004
- `modules`: 8 filas — actualmente en 008
- `user_modules`: todos los módulos para user_id=1 — actualmente en 009

## Acceptance Criteria

1. `./wsapi migrate up` en base de datos limpia completa sin errores.
2. Cada migración 001–015 contiene exactamente un `CREATE TABLE IF NOT EXISTS` con el esquema final correcto — sin `ALTER TABLE` para columnas que ya existen en el CREATE.
3. La migración 016 contiene todos los INSERT de datos iniciales.
4. No existen migraciones ALTER que dupliquen columnas ya presentes en el CREATE TABLE correspondiente.
5. Los archivos `.down.sql` hacen `DROP TABLE IF EXISTS` correctamente para cada tabla nueva.
6. El binario compilado pasa `go build ./...` sin errores.

## Tasks / Subtasks

- [x] Actualizar `001`: agregar `adjuntos_json`, `error_reason`, `retry_count`, `last_attempt_at`, `timestamp_created`, `timestamp_sent`, `timestamp_confirmed`; actualizar `.down.sql` (AC: 2, 5)
- [x] Actualizar `004`: eliminar INSERT de admin_users (se moverá a 016) (AC: 3)
- [x] Actualizar `005`: renombrar columna `telefono` → `telefono_contacto` en el CREATE TABLE; eliminar INSERT (AC: 2, 3)
- [x] Actualizar `008`: agregar columna `slug VARCHAR(50) UNIQUE` después de `name`; eliminar INSERT (AC: 2)
- [x] Actualizar `009`: eliminar INSERT de user_modules (se moverá a 016) (AC: 3)
- [x] Actualizar `011`: agregar `updated_by BIGINT NULL` con su INDEX; actualizar `.down.sql` (AC: 2, 5)
- [x] Renombrar `019_create_audit_log_table` → `015_create_audit_log_table` (up y down) (AC: 2)
- [x] Crear `016_seeds.up.sql` con todos los INSERT consolidados (AC: 3)
- [x] Crear `016_seeds.down.sql` que hace `DELETE FROM` en orden inverso de FK (AC: 5)
- [x] Eliminar archivos: 015, 016(viejo), 017, 018, 020, 022, 023, 024, 025 (up y down) (AC: 4)
- [x] Verificar compilación limpia con `go build ./...` (AC: 6)
- [x] Verificar `./wsapi migrate up` en base de datos limpia (AC: 1)

## Notas técnicas

- `golang-migrate` rastrea versiones por número. Al renumerar, la DB existente necesita ser recreada (o limpiar `schema_migrations`). Esto es aceptable en desarrollo.
- El gap de numeración (faltaba `021`) queda resuelto con la nueva numeración continua.
- Los archivos `.down.sql` deben hacer `DROP TABLE IF EXISTS` — sin intentar revertir ALTERs individuales que ya no existen.
- El orden de los seeds en 016 debe respetar las FKs: primero `roles`, luego `empresas`, luego `admin_users`, luego `modules`, luego `user_modules`.

## Dev Agent Record

### Implementation Plan
- Añadir pruebas automatizadas para validar el layout final de migraciones embebidas.
- Consolidar columnas finales dentro de los `CREATE TABLE` base y eliminar seeds intermedios.
- Renumerar `audit_log` a 015, mover todos los inserts iniciales a 016 y borrar migraciones ALTER legacy.
- Validar con `go test ./...`, `go build ./...` y `migrate up` sobre una base temporal limpia.

### Debug Log
- `go test ./internal/storage -run 'TestEmbeddedMigrationsMatchNormalizedLayout|TestNormalizedCreateTableMigrationsHaveFinalSchema|TestSeedsMigrationContainsAllInitialDataAndReversibleDeletes'` ❌ falló inicialmente, confirmando que la estructura legacy seguía activa.
- `go test ./internal/storage -run 'TestEmbeddedMigrationsMatchNormalizedLayout|TestNormalizedCreateTableMigrationsHaveFinalSchema|TestSeedsMigrationContainsAllInitialDataAndReversibleDeletes'` ✅ pasó tras normalizar las migraciones.
- `go test ./...` ✅
- `go build ./...` ✅
- `./wsapi migrate up` sobre base temporal limpia `wsapi_story17_verify` ✅ versión `16`; verificado además: 15 tablas creadas, 7 columnas esperadas en `messages`, `slug` en `modules`, `telefono_contacto` en `empresas`, `updated_by` en `api_keys` y seeds `4,1,1,8,8`.
- `go test ./...` ✅ (re-ejecución final tras ajustar comentarios explícitos en `001` y `011` down migrations).

### Completion Notes
- Se absorbieron los cambios de `017` y `018` dentro de `001_create_messages_table.up.sql` para dejar el esquema final de `messages` en una sola migración base.
- Se removieron inserts de `004`, `005`, `007`, `008` y `009`, y se centralizaron en `016_seeds.up.sql` con su reversa segura en `016_seeds.down.sql`.
- Se incorporó `telefono_contacto` en `005`, `slug` en `008`, `updated_by` en `011` y se renumeró `audit_log` a `015_create_audit_log_table`.
- Se eliminaron las migraciones ALTER y renames legacy (`015`, `016`, `017`, `018`, `020`, `022`, `023`, `024`, `025` antiguos) para evitar duplicados en instalaciones limpias.
- Se agregó una prueba automatizada (`backend/internal/storage/migrations_layout_test.go`) para impedir regresiones en el layout normalizado de migraciones.

## File List
- `_bmad-output/implementation-artifacts/1-7-normalizar-migraciones.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `backend/internal/storage/migrations/001_create_messages_table.up.sql`
- `backend/internal/storage/migrations/001_create_messages_table.down.sql`
- `backend/internal/storage/migrations/004_create_admin_users_table.up.sql`
- `backend/internal/storage/migrations/005_create_empresas_table.up.sql`
- `backend/internal/storage/migrations/007_create_roles_table.up.sql`
- `backend/internal/storage/migrations/008_create_modules_table.up.sql`
- `backend/internal/storage/migrations/009_create_user_modules_table.up.sql`
- `backend/internal/storage/migrations/011_create_api_keys_table.up.sql`
- `backend/internal/storage/migrations/011_create_api_keys_table.down.sql`
- `backend/internal/storage/migrations/015_create_audit_log_table.up.sql`
- `backend/internal/storage/migrations/015_create_audit_log_table.down.sql`
- `backend/internal/storage/migrations/016_seeds.up.sql`
- `backend/internal/storage/migrations/016_seeds.down.sql`
- `backend/internal/storage/migrations/015_add_missing_columns.up.sql` (deleted)
- `backend/internal/storage/migrations/015_add_missing_columns.down.sql` (deleted)
- `backend/internal/storage/migrations/016_rename_empresa_telefono_contacto.up.sql` (deleted)
- `backend/internal/storage/migrations/016_rename_empresa_telefono_contacto.down.sql` (deleted)
- `backend/internal/storage/migrations/017_align_messages_schema_with_repository.up.sql` (deleted)
- `backend/internal/storage/migrations/017_align_messages_schema_with_repository.down.sql` (deleted)
- `backend/internal/storage/migrations/018_add_retry_fields_to_messages.up.sql` (deleted)
- `backend/internal/storage/migrations/018_add_retry_fields_to_messages.down.sql` (deleted)
- `backend/internal/storage/migrations/019_create_audit_log_table.up.sql` (deleted)
- `backend/internal/storage/migrations/019_create_audit_log_table.down.sql` (deleted)
- `backend/internal/storage/migrations/020_add_is_root_to_roles.up.sql` (deleted)
- `backend/internal/storage/migrations/020_add_is_root_to_roles.down.sql` (deleted)
- `backend/internal/storage/migrations/022_add_audit_columns_to_empresas.up.sql` (deleted)
- `backend/internal/storage/migrations/022_add_audit_columns_to_empresas.down.sql` (deleted)
- `backend/internal/storage/migrations/023_add_audit_columns_to_telefonos.up.sql` (deleted)
- `backend/internal/storage/migrations/023_add_audit_columns_to_telefonos.down.sql` (deleted)
- `backend/internal/storage/migrations/024_add_audit_columns_to_roles.up.sql` (deleted)
- `backend/internal/storage/migrations/024_add_audit_columns_to_roles.down.sql` (deleted)
- `backend/internal/storage/migrations/025_add_audit_columns_to_api_keys.up.sql` (deleted)
- `backend/internal/storage/migrations/025_add_audit_columns_to_api_keys.down.sql` (deleted)
- `backend/internal/storage/migrations_layout_test.go`

## Change Log
- 2026-05-04: Normalizadas las migraciones base 001–015, consolidado el seed inicial en 016 y eliminadas migraciones ALTER/rename legacy para soportar `migrate up` limpio en instalaciones frescas.
