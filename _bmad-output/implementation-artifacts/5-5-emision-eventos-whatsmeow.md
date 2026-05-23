---
story_id: "5.5"
epic: "epic-5"
title: "Emisión de eventos desde el core de whatsmeow"
status: review
estimated_days: 2
priority: high
branch: "feature/integracion-loyo"
skills: ["golang-security", "bmad-code-review"]
affects:
  - backend/internal/whatsapp/service.go
  - backend/internal/whatsapp/send.go
  - backend/internal/whatsapp/webhook_events.go
  - backend/internal/whatsapp/webhook_events_test.go
  - backend/internal/storage/webhook_store.go
  - backend/internal/storage/telefono.go
  - backend/internal/storage/messages.go
  - _bmad-output/implementation-artifacts/sprint-status.yaml
---

# Story 5.5: Emisión de eventos desde el core de whatsmeow

Status: in-progress

## Story

Como integrador B2B que registró webhooks en wsapi,
quiero que los eventos reales de sesiones y mensajes de WhatsApp se encolen automáticamente,
para recibir notificaciones en tiempo real sin depender de WebSocket persistente.

## Contexto técnico

La Story 5.4 ya dejó operativo el worker que consume `webhooks_outbound_queue` y entrega payloads firmados al webhook destino. Esta story completa el circuito: detectar eventos reales en el runtime de whatsmeow, convertirlos a envelopes `{event_type,data}` y encolarlos para todos los webhooks activos de la empresa interesada.

Hoy `backend/internal/whatsapp/service.go` ya centraliza el ciclo de vida de sesión y registra un `AddEventHandler(...)` sobre el cliente de whatsmeow, pero solo usa esos eventos para estados internos (`markConnected`, `markDisconnected`, QR, bootstrap). No existe todavía una capa de emisión hacia `webhooks_outbound_queue`.

## Acceptance Criteria

**AC1 — Evento `message.received` se encola desde whatsmeow:**
**Dado** un `*events.Message` entrante emitido por whatsmeow,
**Cuando** el mensaje pertenece a un teléfono asociado a una empresa con webhooks activos suscritos a `message.received`,
**Entonces** se inserta un item en `webhooks_outbound_queue` por cada webhook activo suscrito
**Y** el payload encolado usa el envelope definido en 5.4:
```json
{
  "event_type": "message.received",
  "data": {
    "telefono_id": <telefono_id>,
    "from": "<numero_origen>",
    "message_id": "<whatsapp_message_id>",
    "content": "<texto_o_resumen>",
    "type": "<tipo_mensaje>",
    "timestamp": "<RFC3339>"
  }
}
```

**AC2 — Evento `session.connected` se encola al conectar sesión:**
**Dado** que el runtime de sesión entra a estado conectado,
**Cuando** `service.go` marca la sesión como activa,
**Entonces** además encola el evento `session.connected`
**Y** el payload contiene al menos `{telefono_id, phone, timestamp}`.

**AC3 — Evento `session.disconnected` se encola al desconectar o logout:**
**Dado** que el runtime recibe `*events.Disconnected`, `*events.LoggedOut`, `*events.StreamReplaced`, `*events.TemporaryBan` o `*events.ConnectFailure`,
**Cuando** la sesión termina en estado desconectado,
**Entonces** además encola `session.disconnected`
**Y** el payload contiene al menos `{telefono_id, phone, reason, timestamp}`
**Y** el reason conserva el valor operativo ya usado por `markDisconnected(...)` cuando exista.

**AC4 — `message.status_update` usa la mejor fuente real disponible sin romper el sistema:**
**Dado** que whatsmeow emite `*events.Receipt` para mensajes salientes y el código actual de `send.go` solo loggea el `message_id` de proveedor sin persistir relación directa con `reference_id`,
**Cuando** se implementa el evento `message.status_update`,
**Entonces** la story deja explícita en código la estrategia elegida y testeada:
- preferido: correlacionar `events.Receipt.MessageIDs` con un `reference_id` de wsapi si ya es resoluble con el modelo actual
- fallback aceptable MVP: emitir el update con `message_id` de proveedor y `reference_id` solo cuando pueda resolverse de forma confiable
**Y** en ningún caso se inventa una correlación no determinista.

**AC5 — Emisión no bloquea handlers de whatsmeow:**
**Dado** cualquier evento del runtime,
**Cuando** ocurre la lógica de emisión a webhooks,
**Entonces** el handler principal no se bloquea esperando network I/O ni delivery HTTP
**Y** la operación de emisión se limita a lookups/encolado rápido en DB
**Y** si la DB falla, se registra error estructurado pero el runtime de sesión NO crashea.

**AC6 — Filtrado por empresa y evento suscrito:**
**Dado** que la empresa tiene múltiples webhooks configurados con diferentes `eventos`,
**Cuando** se intenta emitir un evento concreto,
**Entonces** solo reciben cola los webhooks `activo = 1` de esa empresa cuyo JSON `eventos` contiene el `event_type` solicitado.

**AC7 — Reutiliza el contrato de cola de 5.4:**
**Dado** que 5.4 ya definió el envelope `{event_type,data}` y el worker espera exactamente ese formato,
**Cuando** se encolan eventos desde 5.5,
**Entonces** todos los items insertados cumplen ese contrato
**Y** no se cambia el schema de `webhooks_outbound_queue`
**Y** no se crean migraciones nuevas.

**AC8 — Seguridad y logging:**
**Dado** que los payloads pueden incluir contenido de mensajes y metadatos operativos,
**Cuando** se registran logs de emisión,
**Entonces** los logs usan `zerolog` estructurado
**Y** no exponen `secret`, firmas HMAC, bodies completos de mensajes ni payloads sensibles completos
**Y** incluyen solo contexto mínimo como `empresa_id`, `telefono_id`, `event_type`, `webhooks_match`, `resultado`.

**AC9 — Test de integración de encolado:**
**Dado** `backend/internal/whatsapp/webhook_events_test.go`,
**Cuando** se simula al menos un evento `*events.Message` y un evento de sesión conectada/desconectada,
**Entonces** los tests verifican que:
- se insertan items en `webhooks_outbound_queue`
- el envelope contiene `event_type` correcto
- no se encolan webhooks de otra empresa
- no se encolan webhooks no suscritos al evento
- el flujo no falla aunque no existan webhooks activos

**AC10 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `cd backend && go build ./...` y `cd backend && go test ./...`,
**Entonces** ambos comandos terminan sin errores ni regresiones.

## Tasks / Subtasks

- [x] **T1 — Capa de emisión a cola** (AC1, AC5, AC6, AC7, AC8)
  - [x] Crear `backend/internal/whatsapp/webhook_events.go`
  - [x] Definir una pequeña capa/struct de emisión (ej. `WebhookEmitter`) con dependencias explícitas (`WebhookStore`, `TelefonoStore`, logger)
  - [x] Implementar `emitWebhookEvent(empresaID, eventType, payload)` reutilizando `webhooks_outbound_queue`
  - [x] Filtrar solo webhooks activos y suscritos al evento
  - [x] Encolar envelope `{event_type,data}` compatible con 5.4

- [x] **T2 — Integración con `service.go` para sesión y mensajes entrantes** (AC1-AC3)
  - [x] Reusar el `AddEventHandler(...)` existente en `service.go`
  - [x] Encolar `message.received` desde `*events.Message`
  - [x] Encolar `session.connected` al marcar conexión exitosa
  - [x] Encolar `session.disconnected` al marcar desconexión/logout/ban/failure
  - [x] Mantener intactas las transiciones actuales de `sessionStore` y `telefonoStore`

- [x] **T3 — `message.status_update` con estrategia explícita** (AC4)
  - [x] Revisar `*events.Receipt` de whatsmeow como fuente principal
  - [x] Revisar si el modelo actual permite correlacionar `MessageIDs` con `reference_id`
  - [x] Implementar la mejor estrategia viable sin introducir correlación falsa
  - [x] Dejar documentado en Dev Notes / Completion Notes qué forma final tomó el payload

- [x] **T4 — Store helpers mínimos** (AC6, AC7)
  - [x] Extender `backend/internal/storage/webhook_store.go` con helpers de lookup por `empresa_id + event_type`
  - [x] Evitar duplicar lógica SQL de filtrado por JSON en varios lugares
  - [x] Reutilizar `EnqueueEvent(...)` cuando sea suficiente; extender solo si hace falta

- [x] **T5 — Tests de integración** (AC9)
  - [x] Crear `backend/internal/whatsapp/webhook_events_test.go`
  - [x] Cubrir `message.received`
  - [x] Cubrir `session.connected`
  - [x] Cubrir `session.disconnected`
  - [x] Cubrir filtrado por empresa y por evento suscrito
  - [x] Cubrir caso sin webhooks activos ni suscritos

- [x] **T6 — Verificación final** (AC10)
  - [x] `cd backend && go build ./...` sin errores
  - [x] `cd backend && go test ./...` sin regresiones

## Dev Notes

### Hallazgo crítico: hoy no existe correlación persistida entre `reference_id` y `message_id` de WhatsApp

`backend/internal/whatsapp/send.go` actualmente hace:
- `client.SendMessage(...)`
- loggea `resp.ID` (`message_id` del proveedor)
- retorna `error` o éxito

Pero **no persiste** `resp.ID` en `messages`, y `backend/internal/storage/messages.go` tampoco tiene una columna dedicada a ese ID de proveedor.

**Consecuencia:** para `message.status_update`, correlacionar un `events.Receipt.MessageIDs[]` con `domain.Message.ReferenceID` no es trivial con el modelo actual.

El dev agent debe evitar dos errores:
1. **no inventar correlación** por timestamp/destino de manera heurística
2. **no romper el modelo de mensajes existente** solo para “hacer que funcione” sin criterio

Si la correlación real no es viable sin migración, la implementación MVP debe dejarlo explícito y emitir `message_id` de proveedor como identificador fuerte, incluyendo `reference_id` solo cuando pueda resolverse de forma determinista.

### Punto de anclaje real para sesión

En `backend/internal/whatsapp/service.go` ya existen estos puntos estables:
- `markConnected(accountID)`
- `markDisconnected(accountID, reason)`
- `syncTelefonoConnected(...)`
- `syncTelefonoDisconnected(...)`
- handler `AddEventHandler(...)` sobre el cliente

**Recomendación:** usar esos puntos como lugares de emisión de `session.connected` y `session.disconnected`, porque ya representan la transición operativa consolidada. No dupliques la lógica de estado en otro lugar.

### Punto de anclaje real para mensajes entrantes

El mismo `AddEventHandler(...)` de `service.go` hoy solo procesa eventos de desconexión. La story 5.5 debe ampliarlo para escuchar al menos:
- `*events.Message`
- `*events.Receipt` (si se implementa `message.status_update` por receipts)

Referencia whatsmeow detectada en módulo local:
- `types/events.Message`
- `types/events.Receipt`
- `types/events.Connected`
- `types/events.Disconnected`
- `types/events.LoggedOut`

### Cómo resolver `empresa_id` y `telefono_id`

El runtime de `service.go` trabaja con `accountID = numero_completo`. Para emitir webhooks se necesita aterrizar a datos del dominio local.

Patrón más seguro:
1. usar `telefonoStore.GetByNumeroCompletoNormalized(accountID)`
2. obtener de ahí `telefono.ID` y `telefono.EmpresaID`
3. encolar por `empresa_id`

No derivar empresa desde otros stores ni desde memoria si ya existe el lookup canon en `TelefonoStore`.

### Contrato de payload recomendado

#### `message.received`
```json
{
  "event_type": "message.received",
  "data": {
    "telefono_id": 12,
    "from": "51911122233",
    "message_id": "wamid-xxx",
    "content": "hola",
    "type": "text",
    "timestamp": "2026-05-22T14:20:00Z"
  }
}
```

#### `session.connected`
```json
{
  "event_type": "session.connected",
  "data": {
    "telefono_id": 12,
    "phone": "51999888777",
    "timestamp": "2026-05-22T14:20:00Z"
  }
}
```

#### `session.disconnected`
```json
{
  "event_type": "session.disconnected",
  "data": {
    "telefono_id": 12,
    "phone": "51999888777",
    "reason": "logged_out",
    "timestamp": "2026-05-22T14:20:00Z"
  }
}
```

#### `message.status_update`
Usar el mejor contrato que el modelo actual soporte de forma honesta. Si no hay `reference_id` determinista, preferir:
```json
{
  "event_type": "message.status_update",
  "data": {
    "telefono_id": 12,
    "message_id": "wamid-xxx",
    "reference_id": "ref-123", 
    "status": "delivered",
    "timestamp": "2026-05-22T14:20:00Z"
  }
}
```
`reference_id` solo debe incluirse cuando esté resuelto de forma confiable.

### Filtrado por evento en DB

El epic pide:
```sql
webhooks_outbound WHERE empresa_id = ? AND activo = 1 AND JSON_CONTAINS(eventos, ?)
```

No repitas esta query inline por todo el código. Centralízala en `WebhookStore` con un helper claro (por ejemplo `ListActiveByEmpresaAndEvent`).

### No bloquear el runtime

La emisión NO debe hacer HTTP; solo insertar cola.

Aun así, el handler de whatsmeow no debe quedar cargado con operaciones lentas o frágiles. Si el código de emisión crece, encapsúlalo en un emisor propio y mantén el `AddEventHandler` lo más delgado posible.

### Seguridad aplicada (golang-security)

- no loggear bodies de mensajes completos
- no exponer `secret` ni detalles de webhooks
- tratar `content` como dato de usuario: serializar limpio, sin interpolaciones inseguras
- manejar errores de DB explícitamente, pero sin panics
- si el mensaje entrante no puede parsearse a texto, usar un resumen seguro del tipo (`image`, `audio`, `document`, etc.)

### Testing strategy

Para tests no necesitas una sesión real de WhatsApp completa.

Preferir:
- testear una capa de emisión desacoplada (`WebhookEmitter`) con SQLite en memoria
- simular inputs equivalentes a `events.Message` / `events.Receipt` con payload mínimo necesario
- verificar inserts en `webhooks_outbound_queue`
- verificar que webhooks de otras empresas no reciben eventos

### Learnings de stories previas

**De 5.3:**
- los webhooks válidos ya vienen filtrados por HTTPS y eventos permitidos; no revalidar eso aquí
- `Webhook.Secret` está protegido con `json:"-"`; mantener la misma disciplina

**De 5.4:**
- el worker ya exige envelope `{event_type,data}`
- no tocar schema de cola
- no introducir payloads incompatibles con `decodeWebhookDeliveryEnvelope(...)`

### Project Structure Notes

- Runtime y event handlers de WhatsApp: `backend/internal/whatsapp/service.go`
- Envío saliente: `backend/internal/whatsapp/send.go`
- Worker de delivery ya existente: `backend/internal/whatsapp/webhook_delivery_worker.go`
- Persistencia webhook/cola: `backend/internal/storage/webhook_store.go`
- Lookup canónico de teléfono/empresa: `backend/internal/storage/telefono.go`
- Modelo de mensajes actual: `backend/internal/storage/messages.go`

### References

- Epic base: `_bmad-output/planning-artifacts/epic-integracion-loyo.md`
- Estado y orden autoritativo: `_bmad-output/implementation-artifacts/sprint-status.yaml`
- Runtime de sesión: `backend/internal/whatsapp/service.go`
- Send saliente: `backend/internal/whatsapp/send.go`
- Worker de delivery ya implementado: `backend/internal/whatsapp/webhook_delivery_worker.go`
- Store de webhook: `backend/internal/storage/webhook_store.go`
- Store de teléfono: `backend/internal/storage/telefono.go`
- Repo de mensajes: `backend/internal/storage/messages.go`
- Tipos whatsmeow locales: `/home/fulanito/go/pkg/mod/go.mau.fi/whatsmeow@v0.0.0-20260410162419-b95d92207080/types/events/events.go`
- Reglas del proyecto: `docs/bmad-project-rules.md`
- Contexto técnico global: `_bmad-output/project-context.md`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (bmad-create-story)

### Debug Log References

- Se creó `backend/internal/whatsapp/webhook_events.go` con `WebhookEmitter`, payloads tipados y encolado reutilizando el envelope `{event_type,data}` de 5.4.
- `service.go` ahora emite `message.received` y `message.status_update` desde `AddEventHandler(...)`, y reutiliza `markConnected`/`markDisconnected` para `session.connected`/`session.disconnected`.
- Se añadió en `Manager` un mapa en memoria `provider message_id -> reference_id` por `accountID`, alimentado desde `SendRichMessageWithReference(...)`, para correlación determinista de receipts sin tocar el schema de `messages`.
- La estrategia MVP para `message.status_update` quedó así: si existe correlación en memoria, se incluye `reference_id`; si no, el webhook igualmente sale con `message_id`, `status` y `timestamp`, sin heurísticas frágiles.
- Validaciones ejecutadas: `cd backend && go test ./internal/whatsapp/...`, `cd backend && go test ./...`, `cd backend && go build ./...`.

### Completion Notes List

- Se implementó emisión de webhooks para `message.received`, `session.connected`, `session.disconnected` y `message.status_update`.
- El filtrado por empresa/evento quedó centralizado en `WebhookStore.ListActiveByEmpresaAndEvent(...)`.
- La correlación de receipts evita migraciones: usa registro en memoria al enviar mensajes y solo agrega `reference_id` cuando está resuelto de forma confiable.
- Se añadieron tests de integración con SQLite para eventos de sesión, mensajes entrantes, receipts y caso sin suscriptores activos.
- La story quedó lista para `code-review` con build y suite de tests en verde.

### File List

- `_bmad-output/implementation-artifacts/5-5-emision-eventos-whatsmeow.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `backend/internal/http/container.go`
- `backend/internal/http/handlers/admin_messages.go`
- `backend/internal/http/handlers/v1_messages.go`
- `backend/internal/storage/webhook_store.go`
- `backend/internal/whatsapp/broadcast.go`
- `backend/internal/whatsapp/manager.go`
- `backend/internal/whatsapp/send.go`
- `backend/internal/whatsapp/service.go`
- `backend/internal/whatsapp/webhook_events.go`
- `backend/internal/whatsapp/webhook_events_test.go`

## Change Log

- 2026-05-22 — Implementada la Story 5.5: emisión de eventos de whatsmeow a `webhooks_outbound_queue`, correlación determinista de receipts sin migración y tests de integración.
- **2026-05-23** — Reabierta. Ajuste por cambio de ownership (ver 5-2 changelog):
  - `WebhookEmitter.emitWebhookEvent` ahora filtra suscriptores con `WebhookStore.ListActiveByTelefonoAndEvent(telefonoID, eventType)` en lugar de `ListActiveByEmpresaAndEvent`. El emitter ya resolvía `phone.ID` en cada lookup, así que el cambio se concentró en el método de filtrado.
  - **AC6** se reescribe: el filtrado ahora es por `telefono_id + activo + JSON_CONTAINS(eventos, ?)`. Una empresa con N teléfonos puede tener N webhooks distintos suscritos al mismo evento, cada uno recibe sólo los eventos de su teléfono.
  - Tests `webhook_events_test.go` ajustados: cada `seedEmitterWebhook(...)` ahora recibe `telefono_id` y `api_key_id` explícitos, y los webhooks de "other company" usan teléfonos distintos para verificar aislamiento por teléfono.