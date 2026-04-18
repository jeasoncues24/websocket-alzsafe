---
title: "S-6.1: Schema DB — empresas (token_version, permissions) + CREATE telefonos"
type: "feature"
created: "2026-04-16"
status: "draft"
context:
  - "_bmad-output/implementation-artifacts/epic-6-context.md"
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** La tabla `empresas` no tiene los campos necesarios para JWT empresa (`token_version`, `permissions`), y no existe tabla `telefonos` que modele los números WhatsApp por empresa.

**Approach:** Agregar dos migraciones SQL (014 ALTER empresas, 015 CREATE telefonos) y actualizar/crear los structs Go correspondientes en `internal/domain/`.

## Boundaries & Constraints

**Always:** Seguir el patrón de migraciones existente (`NNN_name.up.sql` / `.down.sql`). Mantener backward compatibility — columnas nuevas en empresas usan DEFAULT. `numero_completo` como columna virtual generada. El down.sql debe ser el inverso exacto del up.sql.

**Ask First:** Cambiar el tipo de `permissions` a algo distinto de JSON. Agregar campos distintos a los listados en el I/O Matrix.

**Never:** Modificar tablas existentes fuera de `empresas`. Crear lógica de negocio en este story (solo schema + structs). Tocar el sistema admin.

## I/O & Edge-Case Matrix

| Scenario           | Input / State                         | Expected Output / Behavior                                      | Error Handling                                                                            |
| ------------------ | ------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| Migración 014 up   | DB con tabla empresas existente       | ALTER agrega token_version INT DEFAULT 1, permissions JSON NULL | Error si columna ya existe (idempotency no garantizada — migraciones corren una sola vez) |
| Migración 014 down | DB con columnas nuevas                | DROP COLUMN token_version, DROP COLUMN permissions              | -                                                                                         |
| Migración 015 up   | DB limpia de telefonos                | CREATE TABLE telefonos con todos los campos + índices           | Error si tabla ya existe                                                                  |
| Migración 015 down | DB con tabla telefonos                | DROP TABLE telefonos                                            | Falla si hay FK desde otras tablas                                                        |
| numero_completo    | codigo_pais="+51", numero="999888777" | numero_completo="+51999888777" (virtual generado)               | No se puede insertar directamente                                                         |

</frozen-after-approval>

## Code Map

- `internal/storage/migrations/` -- directorio con todos los archivos .up.sql / .down.sql
- `internal/domain/empresa.go` -- struct Empresa — agregar TokenVersion, Permissions
- `internal/domain/telefono.go` -- NUEVO — struct Telefono, TelefonoStatus

## Tasks & Acceptance

**Execution:**

- [ ] `internal/storage/migrations/014_alter_empresas_add_jwt_fields.up.sql` -- ALTER TABLE empresas: ADD COLUMN token_version INT NOT NULL DEFAULT 1, ADD COLUMN permissions JSON NULL -- campos requeridos para JWT empresa y control de permisos
- [ ] `internal/storage/migrations/014_alter_empresas_add_jwt_fields.down.sql` -- ALTER TABLE empresas: DROP COLUMN token_version, DROP COLUMN permissions -- rollback limpio
- [ ] `internal/storage/migrations/015_create_telefonos.up.sql` -- CREATE TABLE telefonos con todos los campos del modelo (incluyendo numero_completo como columna virtual generada) + índices -- tabla central para números WhatsApp por empresa
- [ ] `internal/storage/migrations/015_create_telefonos.down.sql` -- DROP TABLE IF EXISTS telefonos -- rollback
- [ ] `internal/domain/empresa.go` -- agregar campos TokenVersion int y Permissions []string (con json serialization) a struct Empresa -- struct debe reflejar el schema actualizado
- [ ] `internal/domain/telefono.go` -- crear struct Telefono, TelefonoStatus (active/qr_pending/disconnected), TelefonoResponse -- base del dominio multi-teléfono

**Acceptance Criteria:**

- Given la migración 014 se ejecuta en una DB con la tabla empresas existente, when se consulta DESCRIBE empresas, then aparecen las columnas token_version (INT, NOT NULL, DEFAULT 1) y permissions (JSON, NULL)
- Given la migración 015 se ejecuta, when se consulta SHOW CREATE TABLE telefonos, then la tabla tiene: empresa_id FK a empresas, numero_completo como columna generada de CONCAT(codigo_pais, numero), session_data LONGBLOB, status con valores permitidos, e índices en empresa_id y numero_completo
- Given el down de 014 y 015 se ejecutan, when se consulta el schema, then las columnas/tabla desaparecen sin error
- Given el struct Empresa en Go, when se hace json.Marshal de una instancia con Permissions=["send","broadcast"], then el JSON contiene "permissions":["send","broadcast"]
- Given el struct Telefono en Go, when se crea uno con Status=TelefonoStatusActive, then Status.String() == "active"

## Design Notes

**Columna virtual `numero_completo` en MariaDB:**

```sql
numero_completo VARCHAR(25) AS (CONCAT(codigo_pais, numero)) VIRTUAL NOT NULL
```

Las columnas virtuales en MariaDB son calculadas en tiempo de query, no almacenadas. Son indexables con un índice normal.

**`permissions` en Go:** usar `[]string` con un custom JSON marshaler no es necesario — `json:",omitempty"` y `pq.StringArray` no aplica aquí. Usar un type alias o simplemente `[]string` con un campo de tipo `json.RawMessage` en el storage scan es suficiente. En el struct de dominio, mantener `[]string`.

**Índices en telefonos:**

- `(empresa_id)` — para listar teléfonos de una empresa
- `(empresa_id, status)` — para filtrar por estado
- `(numero_completo)` — para lookup por número completo (evitar duplicados entre empresas)

## Verification

**Commands:**

- `go build ./...` -- expected: sin errores de compilación
- `go vet ./internal/domain/...` -- expected: sin warnings
