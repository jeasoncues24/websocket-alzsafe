# Story 2.3: Persistencia de mensajes directos y trazabilidad mínima

Status: done

## Story

As a administrador de plataforma,
I want registrar cada envío directo,
So that tenga trazabilidad y auditoría operativa.

## Acceptance Criteria

1. Dado un envío directo exitoso validado, cuando se guarda en DB, entonces se persiste registro con ruc_empresa, destino, timestamp, referenceId, contenido, estado y adjuntosInfo.

2. Dado múltiples envíos por la misma empresa, cuando se consultan por empresa, entonces se retorna lista ordenada por timestamp descendente con paginación.

3. Dado un envío con estado "pending", cuando el proveedor confirma entrega, entonces se actualiza estado a "sent" y se registra timestamp de confirmación.

4. Dado una consulta de auditoría, cuando se filtra por ruc_empresa y rango de fecha, entonces se retorna solo los envíos dentro del rango con todos los detalles.

5. Dado un envío fallido, cuando se persiste, entonces se guarda con estado "failed" y se registra razón del error para diagnóstico.

## Tasks / Subtasks

- [ ] Crear schema de tabla messages en DB (AC: 1, 2, 3, 4, 5)
  - [ ] Tabla: messages (id, reference_id, ruc_empresa, destino, contenido, adjuntos_json, estado, timestamp_create, timestamp_sent, timestamp_confirmed, error_reason)
  - [ ] Indices: (ruc_empresa, timestamp_create), (reference_id), (estado)
  - [ ] Definir tipos: id BIGINT PK, reference_id VARCHAR(36) UNIQUE, estado ENUM, timestamps DATETIME
- [ ] Crear migration file para schema (AC: 1)
  - [ ] Usar framework de migraciones de Epic 4.1 (o crear basis ahora)
  - [ ] Migration UP: CREATE TABLE + INDICES
  - [ ] Migration DOWN: DROP TABLE
  - [ ] Nombre: 001_create_messages_table.sql
- [ ] Extender domain/message.go con fields de DB (AC: 1, 3)
  - [ ] Agregar: ID, TimestampCreated, TimestampSent, TimestampConfirmed, ErrorReason
  - [ ] Mantener backward compatibility con MessageResponse
- [ ] Crear storage/messages.go repository (AC: 1, 2, 3, 4, 5)
  - [ ] MessagesRepository interface: Create, Update, GetByReferenceID, GetByEmpresa, GetByEstado
  - [ ] Implementación MariaDB con prepared statements
  - [ ] Connection pooling via existing DB conn
  - [ ] JSON marshaling para adjuntos
- [ ] Integración con handler HandlePostMessage (AC: 1)
  - [ ] Post-validación: crear Message en DB con estado "pending"
  - [ ] Capturar ID generado y retornar en MessageResponse
  - [ ] Manejar error de DB: retornar 500 con detalle
  - [ ] Usar transacciones si aplica
- [ ] Crear endpoint GET /messages para auditoría (AC: 2, 4)
  - [ ] Query params: ruc_empresa (required), start_date, end_date, estado, page, limit
  - [ ] Retorna: lista de Message con paginación {total, page, limit, messages}
  - [ ] Validar que ruc_empresa de request coincida con sesión activa
  - [ ] Default limit 50, max 500
- [ ] Agregar pruebas unitarias/integración (AC: 1, 2, 3, 4, 5)
  - [ ] Test: create message → DB record created con estado pending
  - [ ] Test: get by empresa → retorna ordenado por timestamp DESC
  - [ ] Test: update estado → cambio persistido
  - [ ] Test: query by fecha range → solo dentro del rango
  - [ ] Test: error reason logged → available para diagnostico
  - [ ] Test: adjuntos JSON → parsed correctamente
  - [ ] Integration: HandlePostMessage → message persisted

## Dev Notes

- No implementar proveedor real en esta story; foco es persistencia y auditoría.
- Adjuntos se serializan como JSON array en campo adjuntos_json
- Estado es ENUM: pending, sent, delivered, failed, rejected
- Timestamp_confirmed es NULL hasta que proveedor confirme (future story)
- Para auditoría: ruc_empresa es la clave de particionamiento lógico

### Technical Requirements

- Usar database/sql con prepared statements (inyección SQL safe)
- Conexión MariaDB existente desde internal/config/config.go
- Nuevo archivo: internal/storage/messages.go
- Nuevo bundle: internal/storage/migrations/ (crear carpeta)
- JSON marshaling: encoding/json con custom MarshalJSON si necesario
- Timestamps: time.Time con formato RFC3339
- Error handling: log errors, return 500 si DB unavailable

### Suggested File Targets

- [internal/storage/messages.go](internal/storage/messages.go) - NEW (repository)
- [internal/domain/message.go](internal/domain/message.go) - Extend con DB fields
- [internal/storage/migrations/001_create_messages_table.sql](internal/storage/migrations/001_create_messages_table.sql) - NEW
- [internal/http/handlers.go](internal/http/handlers.go) - Extend HandlePostMessage + add HandleGetMessages
- [internal/http/router.go](internal/http/router.go) - Add GET /messages route
- [internal/http/handlers_test.go](internal/http/handlers_test.go) - Add integration tests

### Testing Requirements

- Unit tests para MessagesRepository (CRUD operations)
- Integration tests para HandlePostMessage (message persisted)
- Mock DB para unit tests, real DB para integration (si está disponible)
- Casos de cobertura:
  - Create message success → record in DB
  - Create with adjuntos → JSON serialized correctly
  - Get by empresa paginado → correct ordering + pagination
  - Update estado pending→sent → persisted
  - Query by date range → only within range
  - Concurrent creates → no race conditions, unique reference_id

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L155)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md)
- Story 2.1: [\_bmad-output/implementation-artifacts/2-1-endpoint-de-envio-directo-con-validacion-de-payload.md](_bmad-output/implementation-artifacts/2-1-endpoint-de-envio-directo-con-validacion-de-payload.md)
- Story 2.2: [\_bmad-output/implementation-artifacts/2-2-soporte-de-adjuntos-en-envio-directo-con-politicas-de-seguridad.md](_bmad-output/implementation-artifacts/2-2-soporte-de-adjuntos-en-envio-directo-con-politicas-de-seguridad.md)
- DB setup: internal/storage/mariadb.go

## Implementation Priority

**Must Have** (AC 1, 3, 5):

- Create messages table with migration
- Repository Create + Update methods
- Integration in HandlePostMessage

**Should Have** (AC 2, 4):

- GetByEmpresa with pagination
- GET /messages endpoint for auditoría

**Nice to Have**:

- Advanced query filtering
- Export messages to CSV

## File List

- \_bmad-output/implementation-artifacts/2-3-persistencia-de-mensajes-directos-y-trazabilidad-minima.md
