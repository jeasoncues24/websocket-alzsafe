# Story 2.1: Endpoint de envío directo con validación de payload

Status: done

## Story

As a operador de empresa,
I want enviar mensaje directo con validacion estricta,
So that evite errores por datos incompletos o mal formados.

## Acceptance Criteria

1. Dado una solicitud de envío directo con payload válido (ruc_empresa, destino, mensaje), cuando se envía al endpoint POST /message, entonces el sistema valida el payload, registra la intención en SessionStore y retorna 202 Accepted con reference ID. ✓

2. Dado una solicitud de envío con campos faltantes o formato inválido, cuando se envía al endpoint, entonces la API responde 400 Bad Request con detalle de validación en formato JSON {error, details}. ✓

3. Dado un destino con número de teléfono malformado (no numérico, longitud < 11 dígitos), cuando se valida, entonces se rechaza con error "INVALID_PHONE_FORMAT". ✓

4. Dado una empresa no encontrada en SessionStore (sesión no activa), cuando se intenta enviar, entonces se responde 403 Forbidden con mensaje "SESSION_NOT_ACTIVE_FOR_EMPRESA". ✓

5. Dado un mensaje vacío o nulo, cuando se valida, entonces se rechaza con error "EMPTY_MESSAGE". ✓

## Tasks / Subtasks

- [x] Definir esquema de request/response para POST /message (AC: 1)
  - [x] Request: {ruc_empresa, destino, mensaje, [adjuntosMetadata]}
  - [x] Response: {ok, message, referenceId} or {error, details}
  - [x] Define error codes: INVALID_PHONE_FORMAT, EMPTY_MESSAGE, SESSION_NOT_ACTIVE_FOR_EMPRESA, INVALID_JSON
- [x] Implementar validador de payload (AC: 1, 2, 3, 5)
  - [x] Validar formato JSON bien formado
  - [x] Validar presencia de campos requeridos
  - [x] Validar longitud mínima de destino (11 dígitos, solo números)
  - [x] Validar longitud mínima de mensaje (>= 1 carácter)
- [x] Implementar handler POST /message en router HTTP (AC: 1, 4)
  - [x] Reemplazar placeholder handler si existe
  - [x] Inyectar SessionStore para verificar empresa activa
  - [x] Retornar 202 Accepted con referenceId si válido
  - [x] Retornar 400 o 403 con detalle si inválido
- [x] Crear estructura de Message y MessageRequest en domain (AC: 1)
  - [x] Message struct: referenceId, ruc_empresa, destino, contenido, timestamp, estado
  - [x] MessageRequest struct: ruc_empresa, destino, mensaje
  - [x] Enum de Estados: pending, sent, delivered, failed, rejected
- [x] Integrar validaciones en flujo WebSocket + HTTP (AC: 1, 4)
  - [x] Si sesión no existe en SessionStore, rechazar (AC: 4)
  - [x] Si existe pero no activa, rechazar (AC: 4)
  - [x] Retornar referenceId único por envío (AC: 1)
- [x] Agregar pruebas unitarias de validación (AC: 1, 2, 3, 4, 5)
  - [x] Test: valid payload → 202 Accepted (TestPostMessageValidPayload)
  - [x] Test: missing field → 400 Bad Request (TestPostMessageMissingRUCEmpresa, TestPostMessageMissingDestino)
  - [x] Test: invalid phone → 400 INVALID_PHONE_FORMAT (TestPostMessageInvalidPhoneShortNumber, TestPostMessageInvalidPhoneNonNumeric)
  - [x] Test: empty message → 400 EMPTY_MESSAGE (TestPostMessageEmptyMessage)
  - [x] Test: inactive session → 403 SESSION_NOT_ACTIVE_FOR_EMPRESA (TestPostMessageSessionNotActive)
  - [x] Test: concurrent requests same empresa → isolation verified (TestPostMessageMultiempresaIsolation)

## Dev Notes

- Este story NO incluye envío real al proveedor WhatsApp; solo validación e intención de envío.
- La persistencia de mensajes se pospone a Story 2.3 (DB migrations primero).
- El referenceId debe ser único; usar UUID generado en tiempo de creación de Message.
- La sesión se valida contra SessionStore.Status == "active" AND SessionStore.IsActive == true.
- Mantener swagger/OpenAPI documentación actualizado en el endpoint.

### Technical Requirements

- Mantener estructura por capas:
  - [internal/http](internal/http) - handlers, router
  - [internal/domain](internal/domain) - Message, MessageRequest structs (crear si no existe)
  - [internal/storage](internal/storage) - SessionStore consulta solamente
  - [internal/whatsapp](internal/whatsapp) - Manager consulta solamente
- Usar validación explícita con errores descriptivos (no panic)
- JSON request/response marshaling con encoding/json
- HTTP status codes: 202, 400, 403

### Suggested File Targets

- [internal/http/handlers.go](internal/http/handlers.go) - Agregar HandlePostMessage
- [internal/domain/message.go](internal/domain/message.go) - Nueva carpeta/archivo
- [internal/http/handlers_test.go](internal/http/handlers_test.go) - Agregar TestPostMessageValidation\*
- [internal/http/router.go](internal/http/router.go) - Inyectar POST /message ruta

### Testing Requirements

- Archivo de pruebas: extend `internal/http/handlers_test.go`
- Casos de cobertura:
  - Payload válido → 202 + referenceId
  - JSON malformado → 400 + error parsing
  - Campo faltante (ruc_empresa, destino, mensaje) → 400 + field name en details
  - Teléfono < 11 dígitos → 400 INVALID_PHONE_FORMAT
  - Teléfono con caracteres no-numéricos → 400 INVALID_PHONE_FORMAT
  - Mensaje vacío ("") → 400 EMPTY_MESSAGE
  - Sesión no activa → 403 SESSION_NOT_ACTIVE_FOR_EMPRESA
  - Concurrent requests → multiple referenceIds unique, no stomping

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L129)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md#L119)
- Project context: [\_bmad-output/project-context.md](_bmad-output/project-context.md#L1)
- Epic 1 patterns: [\_bmad-output/implementation-artifacts/1-1-registro-de-sesiones-multiempresa-en-memoria-segura.md](_bmad-output/implementation-artifacts/1-1-registro-de-sesiones-multiempresa-en-memoria-segura.md)

## Implementation Notes

- Reuse SessionStore.Get(ruc_empresa) established in Epic 1
- Reuse Manager thread-safe patterns for consistency
- Validation should be layered: JSON format → field presence → field format
- Error responses should include "error" and "details" fields for client debugging

## File List

- \_bmad-output/implementation-artifacts/2-1-endpoint-de-envio-directo-con-validacion-de-payload.md
