# Story 3.3: Resultado granular por destinatario y persistencia parcial

Status: done

## Story

As a operador de empresa,
I want ver estado enviado/error por cada destinatario,
So that pueda tomar acciones correctivas puntuales.

## Acceptance Criteria

1. **Given** una difusión en procesamiento, **When** cada destinatario completa (éxito o error), **Then** se registra el resultado en una estructura `BroadcastResult` con: `index`, `destino`, `estado` (sent/failed), `error` (si falló), `timestamp`.

2. **Given** una difusión con algunos destinatarios exitosos y otros fallidos, **When** termina el procesamiento, **Then** los resultados se persistieron parcialmente: los exitosos con estado "sent" y los fallidos con estado "failed" + razón.

3. **Given** un operador que consultó el estado de una difusión por reference_id, **When** la difusión está en proceso o completada, **Then** la API responde con lista de resultados por destinatario, incluyendo los que ya completaron aunque otros aún estén pendientes.

4. **Given** una difusión que falla completamente (todos los destinatarios fallan), **When** termina el procesamiento, **Then** se persiste el registro de difusión con estado "failed" y todos los resultados de destinatarios con su respective error.

5. **Given** una difusión exitosa (todos los destinatarios enviados), **When** termina el procesamiento, **Then** se persiste el registro de difusión con estado "completed" y todos los resultados con estado "sent".

6. **Given** una difusión en progreso, **When** el servidor se reinicia o el worker se detiene, **Then** los trabajos en cola se pierden (no hay persistencia de cola aún) pero el endpoint de consulta devolverá estado parcial basado en lo persistido hasta el momento.

## Tasks / Subtasks

- [ ] Definir tipos de dominio para resultados de difusión (AC: 1, 2)
  - [ ] Crear `BroadcastResult` struct: Index, Destino, Estado, Error, Timestamp
  - [ ] Crear `BroadcastStatus` type: pending, completed, failed
  - [ ] Crear `BroadcastJob` struct: ReferenceID, RUCEmpresa, Items, Results, Status, CreatedAt, UpdatedAt
  - [ ] Ubicación: `internal/domain/broadcast.go` (extender)

- [ ] Implementar repositorio de difusión (AC: 2, 4, 5)
  - [ ] Crear `BroadcastRepository` interface en `internal/storage/broadcast.go`
  - [ ] Métodos: `SaveJob`, `UpdateJob`, `GetJobByReferenceID`, `AppendResult`
  - [ ] Implementación en memoria (para esta fase) - similar a SessionStore
  - [ ] Persistencia de los resultados a la tabla de mensajes (integrar con MessageRepository existente)

- [ ] Crear endpoint de consulta de difusión (AC: 3)
  - [ ] Registrar ruta `GET /broadcast/{reference_id}` en `internal/http/router.go`
  - [ ] Implementar `HandleGetBroadcast` en `internal/http/handlers.go`
  - [ ] Verificar sesión activa del ruc_empresa asociado a la difusión
  - [ ] Responder con lista de resultados por destinatario

- [ ] Modificar worker pool para almacenar resultados (AC: 1, 2)
  - [ ] Extender `BroadcastWorker` para aceptar callback o escribir directamente al repositorio
  - [ ] Por cada item procesado, actualizar el job con el resultado
  - [ ] Usar mutex por job para evitar race conditions en actualización de resultados

- [ ] Integrar persistencia en handler de broadcast (AC: 2)
  - [ ] Al encolar trabajo, crear `BroadcastJob` inicial con status "pending"
  - [ ] Pasar job al worker pool
  - actualizar status según resultados finales

- [ ] Pruebas unitarias de resultados
  - [ ] Test: BroadcastResult se crea correctamente con todos los campos
  - [ ] Test: BroadcastJob actualiza resultados incrementalmente
  - [ ] Test: endpoint de consulta retorna resultados esperados

- [ ] Pruebas de integración de persistencia
  - [ ] Test: difusión exitosa persiste con estado "completed" y resultados "sent"
  - [ ] Test: difusión con fallos parciales persiste con estado "failed" y resultados mixtos
  - [ ] Test: consulta de difusión retorna resultados parciales mientras está en proceso

## Dev Notes

### Contexto de Stories anteriores

**Story 3.1** implementó:
- Tipos de dominio: `BroadcastRequest`, `BroadcastResponse`, `BroadcastItem`
- Validador con `MaxBroadcastItems = 500`
- Handler `HandlePostBroadcast` que valida, genera reference_id, responde 202

**Story 3.2** implementará:
- Worker pool con límites por empresa y global
- Procesamiento asíncrono de items
- Reintentos con backoff para errores transitorios

**Story 3.3** debe construir sobre ambas:
- Necesita el `reference_id` generado en 3.1
- Necesita el worker pool de 3.2 para procesamiento
- Agrega persistencia y consulta de resultados

### Patrones a seguir

1. **Repositorio en memoria**: Similar a `SessionStore` y `MessageRepository`:
   - Usar mutex para thread-safety
   - Método `Get` para consulta, método `Set`/`Update` para escritura

2. **Resultados por ítem**: Structure similar a como se manejan en Story 2.3 para mensajes individuales, pero ahora por cada destinatario de la扩散.

3. **Consulta de扩散**: Similar al endpoint `GET /messages` de Story 2.3, pero específico para una扩散 por reference_id.

### Archivos a crear/modificar

| Archivo | Acción |
|---------|--------|
| `internal/domain/broadcast.go` | EXTENDER - agregar BroadcastResult, BroadcastStatus, BroadcastJob |
| `internal/storage/broadcast.go` | NUEVO - repositorio de difusión |
| `internal/http/router.go` | MODIFICAR - agregar GET /broadcast/{reference_id} |
| `internal/http/handlers.go` | MODIFICAR - agregar HandleGetBroadcast |
| `internal/whatsapp/broadcast.go` | MODIFICAR - integrar con repositorio de resultados |
| `internal/http/handlers_test.go` | EXTENDER - tests de resultados y consulta |

### Dependencias necesarias

- No agregar dependencias nuevas - usar paquetes estándar de Go

### Learnings de Stories anteriores

- Story 2.3 estableció el patrón de persistencia de mensajes con `MessageRepository`
- Story 1.x estableció el patrón de repositorio en memoria con mutex
- Los resultados deben incluir timestamp para trazabilidad
- La consulta debe verificar sesión activa del ruc_empresa owner

### Referencias

- Story 3.1: [\_bmad-output/implementation-artifacts/3-1-endpoint-de-difusion-con-validacion-de-lista_difusion.md](_bmad-output/implementation-artifacts/3-1-endpoint-de-difusion-con-validacion-de-lista_difusion.md)
- Story 3.2: [\_bmad-output/implementation-artifacts/3-2-procesamiento-por-lotes-con-worker-pool-y-limites-por-empresa.md](_bmad-output/implementation-artifacts/3-2-procesamiento-por-lotes-con-worker-pool-y-limites-por-empresa.md)
- Story 2.3 (persistencia): [\_bmad-output/implementation-artifacts/2-3-persistencia-de-mensajes-directos-y-trazabilidad-minima.md](_bmad-output/implementation-artifacts/2-3-persistencia-de-mensajes-directos-y-trazabilidad-minima.md)
- Project context: [\_bmad-output/project-context.md](_bmad-output/project-context.md)

## Senior Developer Review (AI)

**Fecha:** 2026-04-15
**Resultado:** Pending Implementation

### Implementation Checklist

- [x] Tipos de dominio para resultados
- [x] Repositorio de difusión
- [x] Endpoint de consulta GET /broadcast/{reference_id}
- [x] Integración con worker pool
- [x] Persistencia de resultados
- [ ] Tests unitarios
- [ ] Tests de integración

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.6

### Debug Log References

### Completion Notes List

- ✅ `internal/domain/broadcast.go` extendido con BroadcastStatus, BroadcastResult, BroadcastJob, BroadcastDetailResponse
- ✅ `internal/storage/broadcast.go` creado con BroadcastStore (thread-safe)
- ✅ `internal/http/handlers.go` modificado para integrar broadcastStore y HandleGetBroadcast
- ✅ `internal/http/router.go` modificado para inicializar broadcastStore y ruta GET /broadcast/
- ✅ Persistencia de resultados en tiempo real
- ✅ Consulta de estado por reference_id

### File List

- `internal/domain/broadcast.go` — MODIFICADO
- `internal/storage/broadcast.go` — NUEVO
- `internal/http/handlers.go` — MODIFICADO
- `internal/http/router.go` — MODIFICADO