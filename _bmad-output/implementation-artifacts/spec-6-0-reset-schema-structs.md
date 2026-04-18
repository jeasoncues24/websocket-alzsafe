---
title: "S-6.0: Reset schema DB + actualización structs Go (base multi-tenant)"
type: "feature"
created: "2026-04-16"
status: "ready-for-dev"
context:
  - "_bmad-output/implementation-artifacts/epic-6-context.md"
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El schema actual tiene 13 migraciones incrementales con inconsistencias: `messages` y `broadcasts` usan `ruc_empresa VARCHAR` en lugar de FKs enteras, `empresas` no tiene `token_version` ni `permissions`, y no existe tabla `telefonos`. Esta deuda de schema causará problemas en todas las stories del Epic 6.

**Approach:** Una migración "reset" (014) que hace DROP de todas las tablas en orden FK-safe y las recrea con el schema correcto definitivo. Paralelamente, actualizar todos los structs Go de dominio y las queries de storage que referencian el schema viejo, para que el código compile contra el nuevo schema.

## Boundaries & Constraints

**Always:** El schema nuevo debe ser la fuente de verdad — no al revés. Mantener todas las tablas del sistema admin (`admin_users`, `roles`, `modules`, `user_modules`, `token_blacklist`) con su lógica intacta. El down.sql debe hacer DROP de todas las tablas en orden inverso (FK-safe). Eliminar `api_keys` del schema nuevo (reemplazada por JWT empresa).

**Ask First:** Cambiar el tipo de alguna columna no listada en el I/O Matrix. Añadir lógica de negocio en este story (scope: solo schema + structs + storage queries).

**Never:** Reescribir los HTTP handlers en este story (eso es S-6.6). Tocar el frontend. Introducir lógica de autenticación JWT empresa (eso es S-6.4).

## I/O & Edge-Case Matrix

| Scenario                                       | Input / State                                 | Expected Output / Behavior                                                 | Error Handling                                                               |
| ---------------------------------------------- | --------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Migration 014 up sobre DB vacía                | DB sin tablas                                 | Todas las tablas creadas en orden correcto, trigger de protección recreado | Error si dependencia FK falta — orden de CREATE importa                      |
| Migration 014 up sobre DB con datos existentes | DB con tablas de epics anteriores             | DROP en orden FK-safe + CREATE limpio — **se pierden todos los datos**     | Si hay FKs circulares, el DROP falla — resolver con SET FOREIGN_KEY_CHECKS=0 |
| Migration 014 down                             | DB con schema nuevo                           | DROP de todas las tablas + trigger                                         | Falla si alguna tabla ya fue dropeada manualmente                            |
| Message con empresa_id/telefono_id             | `EmpresaID=1, TelefonoID=3`                   | Storage insert usa empresa_id y telefono_id en SQL                         | -                                                                            |
| BroadcastRequest con TelefonoID                | `{"telefono_id": 2, "lista_difusion": [...]}` | BroadcastJob almacena empresa_id y telefono_id                             | -                                                                            |
| `go build ./...` post-cambios                  | Todos los archivos actualizados               | Sin errores de compilación                                                 | -                                                                            |

</frozen-after-approval>

## Code Map

- `internal/storage/migrations/` -- directorio de migraciones — crear 014 up/down
- `internal/domain/empresa.go` -- struct Empresa — añadir TokenVersion, Permissions
- `internal/domain/telefono.go` -- NUEVO — struct Telefono, TelefonoStatus enum
- `internal/domain/message.go` -- struct Message, MessageRequest — reemplazar RUCEmpresa con EmpresaID/TelefonoID
- `internal/domain/broadcast.go` -- BroadcastRequest, BroadcastJob, BroadcastResult, BroadcastDetailResponse — reemplazar RUCEmpresa con EmpresaID/TelefonoID
- `internal/storage/messages.go` -- queries SQL — reemplazar ruc_empresa con empresa_id/telefono_id
- `internal/storage/broadcast.go` -- queries SQL — reemplazar ruc_empresa con empresa_id/telefono_id
- `internal/http/handlers.go` -- referencias a RUCEmpresa en structs locales y llamadas a storage — mínimo para que compile
- `internal/http/validator.go` -- ValidateMessageRequest, ValidateBroadcastRequest — reemplazar ruc_empresa por telefono_id

## Tasks & Acceptance

**Execution:**

- [ ] `internal/storage/migrations/014_reset_schema.up.sql` -- DROP todas las tablas en orden FK-safe (SET FOREIGN_KEY_CHECKS=0/1) + CREATE todas las tablas con schema correcto definitivo -- base limpia sin deuda técnica de schema
- [ ] `internal/storage/migrations/014_reset_schema.down.sql` -- DROP todas las tablas en orden FK-safe -- rollback completo
- [ ] `internal/domain/empresa.go` -- añadir `TokenVersion int` y `Permissions []string` al struct Empresa; actualizar NewEmpresa si aplica -- refleja schema nuevo
- [ ] `internal/domain/telefono.go` -- crear struct Telefono (ID, EmpresaID, CodigoPais, Numero, NumeroCompleto, Status, SessionData, QRString, LastConnected, CreatedAt, UpdatedAt), TelefonoStatus type y constantes (Active/QRPending/Disconnected), TelefonoResponse -- dominio base de teléfonos
- [ ] `internal/domain/message.go` -- en struct Message reemplazar `RUCEmpresa string` por `EmpresaID int64` y `TelefonoID int64`; en MessageRequest reemplazar `RUCEmpresa` por `TelefonoID int64`; actualizar NewMessage, MessageResponse, MessagesListResponse -- alineado con schema nuevo
- [ ] `internal/domain/broadcast.go` -- en BroadcastRequest reemplazar `RUCEmpresa` por `TelefonoID int64`; en BroadcastJob/BroadcastResult/BroadcastDetailResponse reemplazar `RUCEmpresa` por `EmpresaID int64, TelefonoID int64` -- alineado con schema nuevo
- [ ] `internal/storage/messages.go` -- actualizar todas las queries SQL: INSERT usa empresa_id/telefono_id, SELECT/WHERE usan empresa_id en lugar de ruc_empresa; actualizar Scan() para los nuevos campos; actualizar GetMetricsByRUC para usar empresa_id (o dejar stub hasta S-6.2) -- storage consistente con schema nuevo
- [ ] `internal/storage/broadcast.go` -- actualizar queries SQL que usan ruc_empresa para usar empresa_id/telefono_id; actualizar structs de scan -- storage consistente con schema nuevo
- [ ] `internal/http/handlers.go` -- actualizar structs locales que tenían `RUCEmpresa string` para usar `TelefonoID int64`; en los handlers que usaban ruc para lookup en whatsapp manager, usar string vacío o TODO comentado hasta S-6.6 -- compilación sin errores
- [ ] `internal/http/validator.go` -- reemplazar validación de ruc_empresa por validación de telefono_id (> 0) en ValidateMessageRequest y ValidateBroadcastRequest -- validadores alineados con nuevos structs

**Acceptance Criteria:**

- Given la migración 014 up se ejecuta en una DB limpia, when se ejecuta `SHOW TABLES`, then aparecen exactamente: roles, modules, admin_users, user_modules, token_blacklist, empresas, telefonos, messages, broadcasts, broadcast_results (10 tablas — sin api_keys)
- Given la migración 014 up se ejecuta, when se ejecuta `SHOW CREATE TABLE empresas`, then incluye columnas token_version y permissions
- Given la migración 014 up se ejecuta, when se ejecuta `SHOW CREATE TABLE messages`, then incluye empresa_id INT y telefono_id INT y NO incluye columna ruc_empresa
- Given la migración 014 up se ejecuta, when se ejecuta `SHOW CREATE TABLE telefonos`, then incluye empresa_id FK, numero_completo como columna virtual generada, session_data LONGBLOB
- Given `go build ./...` se ejecuta después de todos los cambios de código, when compila, then exit code 0 sin errores

## Design Notes

**Orden de CREATE en 014 up (respetar FK):**

```sql
SET FOREIGN_KEY_CHECKS = 0;
-- 1. roles
-- 2. modules
-- 3. empresas   ← sin FK externas
-- 4. admin_users  ← FK → empresas, roles
-- 5. user_modules  ← FK → admin_users, modules
-- 6. token_blacklist  ← sin FK
-- 7. telefonos  ← FK → empresas
-- 8. messages  ← FK → empresas, telefonos (o solo empresa_id/telefono_id sin FK explícita)
-- 9. broadcasts  ← FK → empresas, telefonos
-- 10. broadcast_results  ← FK → broadcasts
SET FOREIGN_KEY_CHECKS = 1;
```

**FK en messages/broadcasts:** Para mensajes y broadcasts, usar `empresa_id INT NOT NULL` y `telefono_id INT NOT NULL` como columnas indexadas pero **sin FK constraint** — facilita inserts cuando el teléfono aún no está en DB (seed, tests). La integridad se garantiza a nivel de aplicación.

**numero_completo virtual en telefonos:**

```sql
numero_completo VARCHAR(25) AS (CONCAT(codigo_pais, numero)) VIRTUAL NOT NULL
```

**Stub en handlers.go para whatsapp manager:** Los handlers que usaban `NormalizeAccountID(req.RUCEmpresa)` pueden temporalmente usar `""` o `fmt.Sprintf("%d", req.TelefonoID)` — serán reemplazados en S-6.6. Documentar con `// TODO(S-6.6): reemplazar con lookup real por telefono_id`.

**GetMetricsByRUC en storage/messages.go:** La función `GetMetricsByEmpresa` ahora recibe `empresaID int64` en lugar de `ruc string`. Los handlers que la llamen deben actualizarse también (buscar en handlers.go).

## Verification

**Commands:**

- `go build ./...` -- expected: exit 0, sin errores de compilación
- `go vet ./internal/...` -- expected: sin warnings relevantes
- `go test ./internal/domain/...` -- expected: tests pasan (si existen unit tests de dominio)
