---
story_id: "5.4"
epic: "epic-5"
title: "Worker de envío con retry exponencial"
status: done
estimated_days: 2
priority: high
branch: "feature/integracion-loyo"
skills: ["golang-security", "bmad-code-review"]
affects:
  - backend/internal/whatsapp/webhook_delivery_worker.go
  - backend/internal/whatsapp/webhook_delivery_worker_test.go
  - backend/internal/storage/webhook_store.go
  - backend/internal/domain/webhook.go
  - backend/internal/http/container.go
  - backend/internal/http/router.go
  - _bmad-output/implementation-artifacts/sprint-status.yaml
---

# Story 5.4: Worker de envío con retry exponencial

Status: review

## Contexto

La Story 5.2 dejó listas las tablas `webhooks_outbound` y `webhooks_outbound_queue`, y la Story 5.3 ya permite a integradores crear/listar/eliminar webhooks. Falta el componente crítico que consume la cola y entrega eventos HTTP de forma confiable.

Sin este worker, los webhooks quedan persistidos pero nunca salen de la cola. Esta story habilita la entrega real con timeout, firma HMAC, reintentos exponenciales, logs estructurados y apagado limpio del proceso.

## User Story

Como integrador B2B que registró webhooks en wsapi,
quiero que los eventos encolados se entreguen automáticamente con reintentos controlados,
para recibir notificaciones confiables incluso ante fallos temporales del endpoint receptor.

## Objetivo

Implementar un worker background que:

- haga polling cada 5s de `webhooks_outbound_queue`
- reclame items con lock optimista
- envíe POST firmado al webhook destino
- reprograme fallos transitorios con backoff fijo por tabla
- marque fallos terminales y desactive webhooks problemáticos
- se integre al arranque actual sin romper el bootstrap de sesiones WhatsApp

## Scope

### Incluido
- Worker background en `backend/internal/whatsapp/` siguiendo el patrón de workers ya existente en el proyecto.
- Integración con `StartupTasks` para que arranque junto al servidor y se detenga vía `context.Context`.
- Cliente HTTP con timeout de 10s por request.
- Firma `HMAC-SHA256` en header `X-Wsapi-Signature`.
- Header `X-Wsapi-Event` a partir del `event_type` del payload en cola.
- Header `X-Wsapi-Delivery` con `queue.id` como token de idempotencia.
- Retry schedule fijo: `1m`, `5m`, `30m`, `2h`, `6h`, `24h`.
- Tests con `httptest` cubriendo `2xx`, `5xx`, timeout y shutdown limpio.

### Excluido
- Emisión de eventos desde whatsmeow (`5.5`).
- Documentación pública de integración (`5.6`).
- Reaper de items atascados en estado `sending` tras crash del proceso. La tabla actual no tiene `sending_at`; no ampliar esquema en esta story.
- Nuevas migraciones. La solución debe apoyarse en el esquema existente creado en `017` y `018`.

## Acceptance Criteria

**AC1 — Worker arranca con el servidor y se detiene con context:**
**Dado** que el servidor HTTP inicia y `StartupTasks` ya existe para bootstrap de sesiones,
**Cuando** el proceso arranca con DB disponible,
**Entonces** el worker de webhooks se registra dentro del mismo flujo de startup
**Y** se ejecuta en background sin bloquear el arranque del servidor
**Y** cuando el `context.Context` de startup se cancela, el worker deja de hacer polling y aborta requests en curso limpiamente.

**AC2 — Polling de cola con lock optimista:**
**Dado** que existen items en `webhooks_outbound_queue` con `estado='pending'` y `proximo_intento_at <= NOW()`,
**Cuando** corre el loop del worker,
**Entonces** consulta la cola cada `5s`
**Y** intenta reclamar cada item con `UPDATE ... WHERE estado='pending'`
**Y** si otro worker/proceso ya tomó el item, el actual lo omite sin tratarlo como error fatal.

**AC3 — Entrega HTTP firmada con timeout:**
**Dado** un item válido de cola,
**Cuando** el worker hace el POST al webhook,
**Entonces** usa `http.Client{Timeout: 10s}`
**Y** envía headers:
- `Content-Type: application/json`
- `X-Wsapi-Signature: sha256=<hex_hmac>`
- `X-Wsapi-Event: <event_type>`
- `X-Wsapi-Delivery: <queue.id>`
**Y** el body enviado es el JSON del evento real (`data`), no secretos ni metadatos internos.

**AC4 — Contrato del payload en cola sin cambiar esquema:**
**Dado** que la tabla `webhooks_outbound_queue` no tiene columna `event_type`,
**Cuando** el worker lee `payload`,
**Entonces** espera un envelope JSON con forma:
```json
{
  "event_type": "message.received",
  "data": { ...payload real... }
}
```
**Y** usa `event_type` para el header `X-Wsapi-Event`
**Y** usa `data` como body del POST
**Y** si el envelope está mal formado o no trae `event_type`, el item falla de forma terminal con log estructurado, sin retry infinito.

**AC5 — Éxito resetea estado del webhook:**
**Dado** que el receptor responde `2xx`,
**Cuando** el worker completa la entrega,
**Entonces** el item de cola pasa a `done`
**Y** `intentos` se incrementa correctamente
**Y** el webhook asociado resetea `failure_count` a `0`
**Y** `last_error` se limpia
**Y** `last_success_at` se actualiza.

**AC6 — Fallo temporal reprograme con backoff:**
**Dado** que el receptor responde `5xx`, error de red o timeout,
**Cuando** el intento actual todavía no agotó la política de 6 entregas,
**Entonces** el item vuelve a `pending`
**Y** `intentos` se incrementa en 1
**Y** `last_error` guarda una versión acotada del error
**Y** `proximo_intento_at` se recalcula con esta tabla:
- intento 1 fallido → `+1m`
- intento 2 fallido → `+5m`
- intento 3 fallido → `+30m`
- intento 4 fallido → `+2h`
- intento 5 fallido → `+6h`
- intento 6 fallido → `+24h` solo si se decide un último reintento; si no, pasar a fallo terminal según implementación elegida
**Y** la implementación deja explícito en código cuál es el criterio exacto para considerar el intento terminal.

**AC7 — Fallo terminal actualiza failure_count y puede desactivar webhook:**
**Dado** que el item agotó los intentos permitidos o tiene payload inválido no recuperable,
**Cuando** el worker lo da por perdido,
**Entonces** el item queda en `failed`
**Y** el webhook asociado incrementa `failure_count`
**Y** actualiza `last_error`
**Y** si el nuevo `failure_count >= 20`, el webhook se desactiva automáticamente (`activo = false`).

**AC8 — Logs estructurados sin filtrar secretos:**
**Dado** cualquier entrega o error del worker,
**Cuando** se registran logs,
**Entonces** se usan logs estructurados con `zerolog`
**Y** incluyen como mínimo `queue_id`, `webhook_id`, `event_type`, `status_code` (si aplica), `latency_ms`, `attempt`, `resultado`
**Y** NO incluyen `secret`, firma completa, body completo ni payload sensible.

**AC9 — Tests de worker cubren flujos críticos:**
**Dado** `backend/internal/whatsapp/webhook_delivery_worker_test.go`,
**Cuando** se ejecuta la suite,
**Entonces** cubre al menos:
- entrega exitosa `2xx`
- respuesta `5xx` con reprogramación
- timeout HTTP
- payload inválido / envelope mal formado
- lock optimista (item ya reclamado)
- shutdown por `context.Context`
**Y** los tests pasan sin flakes.

**AC10 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `cd backend && go build ./...` y `cd backend && go test ./...`,
**Entonces** ambos comandos terminan sin errores ni regresiones.

## Tasks / Subtasks

- [x] **T1 — Worker background de entrega** (AC1-AC8)
  - [x] Crear `backend/internal/whatsapp/webhook_delivery_worker.go` en `package whatsapp`
  - [x] Definir `WebhookDeliveryWorker` con dependencias explícitas (`WebhookStore`, `http.Client`, logger/config, clock opcional para tests)
  - [x] Implementar loop `Run(ctx)` con ticker de `5s`
  - [x] Implementar `processDueItems` / `processItem` con lock optimista y abortable por context
  - [x] Implementar generación de firma HMAC-SHA256 y headers requeridos
  - [x] Implementar tabla de backoff y criterio terminal documentado en código

- [x] **T2 — Ajustes de store y modelo para soportar el worker** (AC4-AC7)
  - [x] Añadir en `backend/internal/domain/webhook.go` el tipo envelope para cola (ej. `WebhookDeliveryEnvelope`)
  - [x] Ajustar `backend/internal/storage/webhook_store.go` para soportar:
    - [x] éxito (`done` + reset de `failure_count` + `last_success_at`)
    - [x] retry pendiente (`pending` + `intentos` + `proximo_intento_at` + `last_error`)
    - [x] fallo terminal (`failed` + incremento de `failure_count` + `last_error`)
  - [x] Reutilizar el schema existente; no crear migraciones nuevas

- [x] **T3 — Integración con StartupTasks** (AC1)
  - [x] Integrar el worker al flujo actual sin romper `buildStartupBootstrap(...)`
  - [x] Si hace falta, componer múltiples startup funcs en `router.go` o `container.go`
  - [x] Arrancar el worker solo si hay `DB`/`WebhookStore` disponible

- [x] **T4 — Tests de integración del worker** (AC9)
  - [x] Crear `backend/internal/whatsapp/webhook_delivery_worker_test.go`
  - [x] Usar `httptest.Server` para `2xx`, `5xx` y handler lento (timeout)
  - [x] Usar SQLite en memoria + función `NOW()` igual al patrón de otros tests del repo
  - [x] Verificar headers `X-Wsapi-Signature`, `X-Wsapi-Event`, `X-Wsapi-Delivery`
  - [x] Verificar cambios en DB (`pending`, `done`, `failed`, `failure_count`, `activo`)

- [x] **T5 — Verificación final** (AC10)
  - [x] `cd backend && go build ./...` sin errores
  - [x] `cd backend && go test ./...` sin regresiones

## Riesgos y edge cases

- **R1 — Crash con items en `sending`:** el esquema actual no tiene `sending_at` ni `updated_at` en la cola; no intentar resolver stale leases en esta story.
- **R2 — Payload sin `event_type`:** si el worker no define explícitamente el envelope, la Story 5.5 no sabrá cómo encolar datos compatibles.
- **R3 — Fuga de secretos en logs:** nunca loggear `secret`, firma o body completo.
- **R4 — Requests colgados:** usar `http.Client` con timeout y `context` propagado.
- **R5 — Cambiar `StartupTasks` y romper bootstrap:** preservar el comportamiento actual de `buildStartupBootstrap(...)`; componer, no reemplazar.
- **R6 — Reintentos inconsistentes:** documentar en código la relación exacta entre `intentos` guardado y el próximo delay aplicado para evitar off-by-one.

## Dependencias y bloqueos

- **Depende de:** Story 5.2 ✅ (schema + store base), Story 5.3 ✅ (alta de webhooks)
- **Bloquea a:** Story 5.5 (emisión real desde whatsmeow)
- **Rama obligatoria:** `feature/integracion-loyo`

## Dev Notes

### Decisión crítica: no cambiar el schema de cola en 5.4

La cola creada en `018_create_webhooks_outbound_queue.up.sql` tiene solo:
- `webhook_id`
- `payload`
- `intentos`
- `proximo_intento_at`
- `estado`
- `last_error`

No existe columna `event_type` ni timestamps de lease. Por eso esta story debe:

1. **usar un envelope JSON dentro de `payload`** para transportar `event_type`
2. **no intentar stale recovery de `sending`**
3. **evitar nuevas migraciones**

Envelope recomendado:

```json
{
  "event_type": "message.received",
  "data": {
    "telefono_id": 10,
    "from": "51999999999",
    "message_id": "wamid-123"
  }
}
```

### Patrón de ubicación del worker

Poner el worker en `backend/internal/whatsapp/`, no en `handlers/` ni en `storage/`.

**Motivo:**
- el proyecto ya tiene lógica de background/workers en `backend/internal/whatsapp/` (`broadcast.go`, `startup_bootstrap.go`)
- esta story es entrega async de eventos relacionados a la integración WhatsApp
- `storage` debe seguir siendo persistencia, no orquestación de red

### Patrón de startup actual a preservar

Hoy `Container.StartupTasks` recibe un solo `func(context.Context)` y `router.go` la ejecuta una sola vez con `sync.Once`.

Archivos críticos a leer antes de implementar:
- `backend/internal/http/container.go`
- `backend/internal/http/router.go`
- `backend/main.go`

El worker **no** debe reemplazar el bootstrap de sesiones. Debe coexistir con él. La forma más segura es componer startup tasks:

```go
func composeStartupTasks(tasks ...func(context.Context)) func(context.Context)
```

Cada task debe respetar cancelación de `ctx`.

### Ajustes mínimos esperables en `WebhookStore`

El store actual de 5.2 fue suficiente para CRUD y cola base, pero todavía no expresa bien los estados del worker.

Métodos actuales a revisar:
- `PollPending`
- `MarkSending`
- `MarkDone`
- `MarkFailed`
- `IncrementFailureCount`
- `Deactivate`
- `GetByID`

**Problema actual:** `MarkFailed` siempre deja el item en `failed`, pero la story necesita distinguir entre:
- fallo temporal → volver a `pending`
- fallo terminal → `failed`

Por eso el dev agent debe extender/refactorizar el store con una API explícita para esos dos caminos. No dejar lógica ambigua en el worker.

### Contrato HTTP de entrega

Implementar el POST con estándar simple y trazable:

```http
POST <webhook.url>
Content-Type: application/json
X-Wsapi-Signature: sha256=<hex_hmac>
X-Wsapi-Event: message.received
X-Wsapi-Delivery: 123
```

- Firma: `HMAC-SHA256(body, webhook.secret)`
- Encoding recomendado: hex
- Body: el JSON de `data`, no el envelope completo

### Seguridad aplicada (golang-security)

- usar `crypto/hmac` + `crypto/sha256`
- usar `http.NewRequestWithContext`
- no usar `http.DefaultClient`
- truncar/normalizar errores antes de persistirlos si son demasiado largos
- no loggear payloads completos ni secretos
- validar que el worker solo use URLs persistidas por 5.3; no volver a permitir HTTP plano

### Logging recomendado

Usar `zerolog` vía `config.GetLogger()` porque el worker corre fuera de un request HTTP y no tendrá logger de middleware.

Patrón sugerido:

```go
logger := config.GetLogger().With().Str("component", "webhook_delivery_worker").Logger()
```

Campos mínimos por intento:
- `queue_id`
- `webhook_id`
- `event_type`
- `attempt`
- `status_code`
- `latency_ms`
- `resultado`

### Testing notes

Reutilizar patrones ya existentes:
- SQLite en memoria como en `startup_bootstrap_test.go`
- `httptest.Server` como en tests HTTP del repo
- función `NOW()` custom para SQLite si se usa `PollPending` tal cual

Casos que suelen romperse si no se prueban:
1. timeout de cliente HTTP
2. envelope inválido sin `event_type`
3. `MarkSending` devuelve 0 filas afectadas
4. éxito debe resetear `failure_count` previo
5. `failure_count` llega a 20 y desactiva webhook
6. cancelación de context mientras el ticker espera o mientras el request está en vuelo

### Learnings de stories previas

**De 5.2:**
- La cola y el store ya existen; extenderlos, no duplicarlos.
- Las migraciones ya están en producción de desarrollo; no reabrir schema salvo bloqueo real.

**De 5.3:**
- Las URLs de webhook ya fueron validadas como HTTPS en alta; el worker no debe reimplementar toda la validación de registro.
- `Webhook.Secret` está protegido con `json:"-"`; mantener esa disciplina y no exponerlo por logs ni respuestas.

### Project Structure Notes

- Módulo Go: `wsapi`
- Imports internos: `wsapi/internal/...`
- Background workers actuales: `backend/internal/whatsapp/`
- Persistencia: `backend/internal/storage/`
- Config / logger: `backend/internal/config/`
- Wiring HTTP/runtime: `backend/internal/http/`

### References

- Epic base: `_bmad-output/planning-artifacts/epic-integracion-loyo.md` (Story 5.4)
- Estado y orden autoritativo: `_bmad-output/implementation-artifacts/sprint-status.yaml`
- Store actual: `backend/internal/storage/webhook_store.go`
- Schema actual: `backend/internal/storage/migrations/017_create_webhooks_outbound.up.sql`
- Schema actual: `backend/internal/storage/migrations/018_create_webhooks_outbound_queue.up.sql`
- Startup bootstrap existente: `backend/internal/http/router.go`
- Entry point del servidor: `backend/main.go`
- Logging estructurado: `backend/internal/config/logger.go`
- Middleware / helpers de logger: `backend/internal/http/middleware.go`
- Worker de referencia por patrón, no por dominio: `backend/internal/whatsapp/broadcast.go`
- Tests de worker/SQLite de referencia: `backend/internal/whatsapp/startup_bootstrap_test.go`
- Reglas del proyecto: `docs/bmad-project-rules.md`
- Contexto técnico global: `_bmad-output/project-context.md`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (bmad-dev-story)

### Debug Log References

- Se implementó `WebhookDeliveryWorker` con polling inmediato + ticker de 5s, firma HMAC, timeout HTTP y cancelación por `context.Context`.
- Se extendió `WebhookStore` con operaciones explícitas para éxito, retry pendiente, fallo terminal y fallo de cola sin webhook.
- Se compusieron `StartupTasks` para que el bootstrap existente y el worker convivan sin bloquearse entre sí.
- Validaciones ejecutadas: `cd backend && go test ./internal/whatsapp/... -run TestWebhookDeliveryWorker`, `cd backend && go build ./...`, `cd backend && go test ./...`.

### Completion Notes List

- AC1–AC10 satisfechos con código y pruebas automatizadas.
- El worker usa envelope `{event_type,data}` para no tocar el schema actual de cola.
- Se definió criterio terminal explícito: tras 6 intentos fallidos ya no se reprograma otro retry; el delay de `24h` queda fuera de uso al no añadirse un séptimo intento.
- Los logs estructurados incluyen `queue_id`, `webhook_id`, `event_type`, `status_code`, `attempt`, `latency_ms` y `resultado`, sin exponer secretos ni payload completo.

### File List

- `backend/internal/whatsapp/webhook_delivery_worker.go` — worker de entrega de webhooks con HMAC, retry y cancelación
- `backend/internal/whatsapp/webhook_delivery_worker_test.go` — tests de éxito, retry, timeout, payload inválido, lock optimista y shutdown
- `backend/internal/storage/webhook_store.go` — operaciones de store para éxito, retry pendiente y fallo terminal
- `backend/internal/domain/webhook.go` — tipo `WebhookDeliveryEnvelope`
- `backend/internal/http/router.go` — composición de múltiples startup tasks
- `backend/internal/http/container.go` — wiring del worker en `StartupTasks`
- `_bmad-output/implementation-artifacts/5-4-worker-envio-retry-exponencial.md` — tracking de story actualizado
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — estado de sprint sincronizado

## Change Log

- 2026-05-22 — Story 5.4 implementada y validada: worker de entrega con retry exponencial, integración al startup y tests de integración.
