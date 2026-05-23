---
story_id: "5.6"
epic: "epic-5"
title: "Documentación de la integración por webhooks"
status: review
estimated_days: 1
priority: medium
branch: "feature/integracion-loyo"
skills: ["bmad-code-review"]
affects:
  - docs/webhooks-integracion.md
  - docs/integracion-b2b.md
  - README.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
---

# Story 5.6: Documentación de la integración por webhooks

Status: review

## Contexto

Las stories 5.3 y 5.4 ya dejaron definido el contrato base de integración:

- alta/listado/eliminación de webhooks vía `POST/GET/DELETE /api/service/v1/webhooks`
- autenticación por API key (`X-API-Key`, `Authorization: ApiKey`, o `Authorization: Bearer`)
- entrega HTTP con headers `X-Wsapi-Signature`, `X-Wsapi-Event` y `X-Wsapi-Delivery`
- reintentos controlados y desactivación automática tras fallos acumulados

Además, `docs/integracion-b2b.md` ya documenta el health endpoint, y el `README.md` principal hoy está orientado a deploy backend. Falta el documento que permita a un integrador externo usar webhooks sin leer el código fuente.

**Hallazgo crítico:** la Story 5.5 todavía está en `ready-for-dev`, por lo que los payloads finales de eventos aún no son fuente de verdad en código. Esta story debe dejar un guardrail explícito: **no inventar ejemplos de payload**; antes de cerrar la documentación, validar la implementación real de 5.5 en código/tests.

## Story

Como integrador B2B que quiere consumir eventos de wsapi por webhooks,
quiero una guía exacta de registro, firma, payloads, reintentos y manejo de fallos,
para integrar mi receptor de forma segura sin depender del código interno del proyecto.

## Objetivo

Crear un documento público y mantenible que explique:

- cómo registrar un webhook
- qué eventos existen
- qué headers y body entrega wsapi
- cómo verificar la firma HMAC
- cómo funciona la política real de reintentos/fallos
- dónde encontrar la documentación desde `README.md` y `docs/integracion-b2b.md`

## Scope

### Incluido
- Nuevo documento `docs/webhooks-integracion.md`.
- Ejemplos de registro/listado/eliminación usando el contrato HTTP real.
- Explicación de headers de entrega y body real enviado al receptor.
- Snippets de verificación HMAC en Node.js y Go usando librerías estándar.
- Política real de retry, fallo terminal y desactivación automática.
- Enlaces desde `README.md` y `docs/integracion-b2b.md`.

### Excluido
- Cambios de backend, schema, rutas o payloads para “acomodar” la documentación.
- Inventar eventos o campos que aún no existan en código.
- Convertir el `README.md` en manual completo de webhooks; solo debe enlazar.
- Duplicar toda la documentación de health endpoint dentro del nuevo doc.

## Acceptance Criteria

**AC1 — Nuevo documento de integración:**
**Dado** el repositorio del proyecto,
**Cuando** se complete la story,
**Entonces** existe `docs/webhooks-integracion.md`
**Y** el documento cubre al menos: overview, autenticación, registro del webhook, eventos, headers de entrega, verificación HMAC, retries/fallos y troubleshooting.

**AC2 — Registro y gestión documentados con el contrato HTTP real:**
**Dado** `backend/internal/http/handlers/v1_webhooks.go`, `v1_webhooks_test.go` y `v1_helpers.go`,
**Cuando** se documenten los endpoints REST,
**Entonces** los ejemplos reflejan el comportamiento real del código:
- `POST /api/service/v1/webhooks` requiere API key, valida HTTPS y responde `201` con `{ok,data:{id,secret}}`
- `GET /api/service/v1/webhooks` lista webhooks sin exponer `secret`
- `DELETE /api/service/v1/webhooks/{id}` responde con `ok/data/meta`
- se documentan los eventos permitidos exactos: `message.received`, `message.status_update`, `session.connected`, `session.disconnected`
- se documenta el límite configurable `WEBHOOKS_MAX_PER_EMPRESA` (default 10)

**AC3 — Los payloads de ejemplo son reales, no inventados:**
**Dado** que el epic exige “payloads reales”,
**Cuando** se escriban ejemplos de eventos,
**Entonces** cada payload se toma de la implementación final y/o tests reales de 5.5
**Y** la fuente de verdad mínima a revisar es `backend/internal/whatsapp/webhook_events.go` y `backend/internal/whatsapp/webhook_events_test.go`
**Y** si esos archivos aún no existen o no dejan claro el contrato final, la story no debe marcarse `done` con ejemplos inventados.

**AC4 — La documentación explica correctamente qué firma wsapi y qué body recibe el integrador:**
**Dado** `backend/internal/whatsapp/webhook_delivery_worker.go`,
**Cuando** se documente la entrega HTTP,
**Entonces** queda explícito que:
- `X-Wsapi-Signature` usa formato `sha256=<hex>`
- la firma se calcula sobre el **raw body** enviado al receptor
- el body enviado es `data` del envelope, **no** el envelope completo `{event_type,data}`
- el tipo de evento llega en el header `X-Wsapi-Event`
- `X-Wsapi-Delivery` contiene el `queue.id` y debe usarse como clave de idempotencia del lado receptor

**AC5 — Snippets de verificación HMAC correctos y seguros:**
**Dado** que la guía debe incluir ejemplos en Node.js y Go,
**Cuando** se escriban los snippets,
**Entonces** ambos usan librerías estándar y comparación en tiempo constante:
- Node.js: `crypto.createHmac(...)` + `crypto.timingSafeEqual(...)`
- Go: `crypto/hmac` + `sha256` + `hmac.Equal(...)`
**Y** ambos snippets advierten que la verificación debe hacerse con el body crudo antes de parsear JSON.

**AC6 — Retry/backoff y fallos reflejan la implementación real, no el borrador del epic:**
**Dado** `backend/internal/whatsapp/webhook_delivery_worker.go` y la story 5.4,
**Cuando** se documente la política de entrega,
**Entonces** la guía refleja el comportamiento real actual:
- polling cada 5s
- hasta 6 intentos totales
- delays de retry: `1m`, `5m`, `30m`, `2h`, `6h`
- tras agotar intentos, el item queda `failed`
- si `failure_count >= 20`, el webhook se desactiva automáticamente
**Y** no se documenta un retry de `24h` si el código actual no lo ejecuta.

**AC7 — Navegación y discoverability:**
**Dado** que el integrador puede entrar por `README.md` o por `docs/integracion-b2b.md`,
**Cuando** se agreguen enlaces,
**Entonces**:
- `README.md` enlaza a `docs/webhooks-integracion.md` desde la sección de documentación complementaria
- `docs/integracion-b2b.md` enlaza al nuevo documento como guía detallada de webhooks
**Y** se preserva el enfoque actual del README (deploy primero, documentación detallada aparte).

**AC8 — Estilo y verificación final:**
**Dado** que la documentación es parte del entregable del producto,
**Cuando** se cierre la story,
**Entonces** el texto queda en español técnico consistente, con tablas y ejemplos concisos
**Y** se verifica manualmente que los endpoints, headers, eventos y links coinciden con el código vigente
**Y** si solo se tocan archivos Markdown, no se requieren cambios de backend/frontend para marcar la story como completada.

## Tasks / Subtasks

- [x] **T1 — Validar fuentes de verdad antes de redactar** (AC2, AC3, AC4, AC6)
  - [x] Leer `backend/internal/http/handlers/v1_webhooks.go`
  - [x] Leer `backend/internal/http/handlers/v1_webhooks_test.go`
  - [x] Leer `backend/internal/http/handlers/v1_helpers.go`
  - [x] Leer `backend/internal/http/middleware/api_key_auth.go`
  - [x] Leer `backend/internal/domain/webhook.go`
  - [x] Leer `backend/internal/whatsapp/webhook_delivery_worker.go`
  - [x] Verificar si ya existen `backend/internal/whatsapp/webhook_events.go` y `backend/internal/whatsapp/webhook_events_test.go`
  - [x] Si 5.5 no está implementada todavía, registrar el bloqueo y no inventar payloads finales

- [x] **T2 — Crear `docs/webhooks-integracion.md`** (AC1-AC6)
  - [x] Añadir overview de la integración y casos de uso serverless/B2B
  - [x] Documentar autenticación para registrar webhooks
  - [x] Incluir ejemplos `cURL` reales para `POST`, `GET` y `DELETE`
  - [x] Añadir tabla de eventos disponibles
  - [x] Añadir payloads de ejemplo reales tomados de código/tests de 5.5
  - [x] Añadir tabla de headers enviados por wsapi
  - [x] Añadir snippets de verificación HMAC en Node.js y Go
  - [x] Documentar retry, backoff, fallo terminal y desactivación automática
  - [x] Añadir sección breve de troubleshooting/idempotencia

- [x] **T3 — Enlaces de navegación** (AC7)
  - [x] Actualizar `README.md` con enlace breve al nuevo documento
  - [x] Actualizar `docs/integracion-b2b.md` con referencia al nuevo documento de webhooks

- [x] **T4 — Revisión de exactitud** (AC2-AC8)
  - [x] Confirmar que los ejemplos HTTP coinciden con el código actual
  - [x] Confirmar que el formato de firma coincide con `buildWebhookSignature(...)`
  - [x] Confirmar que el body documentado es `data`, no el envelope completo
  - [x] Confirmar que la política de reintentos coincide con `defaultWebhookRetrySchedule` y `defaultWebhookMaxAttempts`
  - [x] Confirmar que todos los links Markdown funcionan y apuntan a rutas existentes

## Riesgos y edge cases

- **R1 — Story 5.5 aún no implementada:** el mayor riesgo es escribir payloads “plausibles” pero no reales. Si `webhook_events.go` no existe o cambió el contrato, detenerse y actualizar la documentación solo cuando el código sea fuente de verdad.
- **R2 — Confundir envelope interno con body entregado:** internamente la cola usa `{event_type,data}`, pero el receptor HTTP ve solo `data`; el evento viaja por header. Documentarlo mal rompe integraciones.
- **R3 — Documentar el retry del epic y no el del código:** el borrador de planificación menciona `24h`, pero el worker actual no hace un séptimo intento. La guía debe seguir el código.
- **R4 — Exponer secretos por accidente:** el doc puede mostrar el `secret` solo en el contexto de respuesta de creación y aclarar que luego no vuelve a listarse.
- **R5 — Inflar el README:** el README principal es de deploy. Debe mantener ese foco y enlazar, no duplicar el manual entero.

## Dependencias y bloqueos

- **Depende de:** Story 5.3 (contrato REST) ✅
- **Depende de:** Story 5.4 (worker de delivery) ✅
- **Depende operativamente de:** Story 5.5 para obtener payloads finales reales ⚠️
- **Rama recomendada:** `feature/integracion-loyo` (actualmente activa)

## Dev Notes

### Fuente de verdad real hoy

Para esta story la documentación no debe partir del epic solamente. Las fuentes de verdad reales están repartidas así:

- **REST de registro/listado/eliminación:** `backend/internal/http/handlers/v1_webhooks.go`
- **Shape exacto de respuestas:** `backend/internal/http/handlers/v1_webhooks.go` + `backend/internal/http/handlers/v1_helpers.go` + `v1_webhooks_test.go`
- **Eventos válidos:** `backend/internal/domain/webhook.go`
- **Headers, firma y retries:** `backend/internal/whatsapp/webhook_delivery_worker.go`
- **Payloads finales de eventos:** implementación real de 5.5 (`webhook_events.go` y tests), no el borrador del epic

### Diferencia importante entre POST y GET/DELETE

No asumir que todas las respuestas V1 usan el mismo helper:

- `POST /api/service/v1/webhooks` responde manualmente con `201` y body:
  ```json
  {
    "ok": true,
    "data": {
      "id": 123,
      "secret": "..."
    }
  }
  ```
- `GET` y `DELETE` usan `writeV1Success(...)`, por lo que incluyen `meta` con `empresa_id` y `timestamp`.

La documentación debe mostrar esta diferencia tal cual.

### Autenticación real aceptada por middleware

`backend/internal/http/middleware/api_key_auth.go` acepta cualquiera de estas formas:

- `X-API-Key: <raw-key>`
- `Authorization: ApiKey <raw-key>`
- `Authorization: Bearer <raw-key>`

Para la guía pública, elegir una forma principal (recomendado: `X-API-Key`) y mencionar las otras dos como alternativas soportadas.

### Firma HMAC real

El worker usa:

```go
req.Header.Set("X-Wsapi-Signature", buildWebhookSignature(body, webhook.Secret))
```

Y `buildWebhookSignature(...)` devuelve:

```text
sha256=<hex>
```

**Clave:** `body` es `envelope.Data`, no el envelope completo. El receptor debe verificar el raw body exacto antes de parsear JSON.

### Política real de retry

El código actual define:

- polling: `5s`
- max attempts: `6`
- retry schedule: `1m`, `5m`, `30m`, `2h`, `6h`
- desactivación automática: `failure_count >= 20`

La guía debe explicar que el `X-Wsapi-Delivery` puede usarse para idempotencia y deduplicación del lado receptor.

### Estado actual de la documentación del repo

- `README.md` está centrado en deploy backend y operación con PM2.
- `docs/integracion-b2b.md` ya documenta el health endpoint.
- No existe todavía `docs/webhooks-integracion.md`.

La implementación más limpia es:

1. crear `docs/webhooks-integracion.md` como manual detallado
2. agregar un link corto desde `README.md`
3. agregar un link corto desde `docs/integracion-b2b.md`

### Estilo recomendado para el nuevo doc

Seguir el patrón ya usado en `docs/integracion-b2b.md`:

- títulos directos
- tablas Markdown para contratos
- ejemplos `curl` concretos
- lenguaje técnico en español
- advertencias cortas para edge cases importantes

### Verificación sugerida

Si la story toca solo Markdown, la validación principal debe ser de exactitud documental, por ejemplo:

```bash
rg -n "POST /api/service/v1/webhooks|GET /api/service/v1/webhooks|DELETE /api/service/v1/webhooks/\{id\}" backend/internal/http/handlers/v1_webhooks.go backend/internal/http/routes_api.go
rg -n "X-Wsapi-Signature|X-Wsapi-Event|X-Wsapi-Delivery|defaultWebhookRetrySchedule|defaultWebhookMaxAttempts|defaultWebhookDeactivateThreshold" backend/internal/whatsapp/webhook_delivery_worker.go
rg -n "WebhookEventMessageReceived|WebhookEventMessageStatus|WebhookEventSessionConnected|WebhookEventSessionDisconnected" backend/internal/domain/webhook.go
rg -n "webhooks-integracion" README.md docs/integracion-b2b.md
```

Y, antes de cerrar la story como `done`:

```bash
test -f backend/internal/whatsapp/webhook_events.go
test -f backend/internal/whatsapp/webhook_events_test.go
```

Si esos archivos no existen, los ejemplos de payload siguen bloqueados.

## Project Structure Notes

- Documentación principal del proyecto: `README.md`
- Documentación técnica/b2b: `docs/`
- Contrato REST de webhooks: `backend/internal/http/handlers/v1_webhooks.go`
- Middleware de API key: `backend/internal/http/middleware/api_key_auth.go`
- Contrato de dominio de eventos: `backend/internal/domain/webhook.go`
- Worker de delivery: `backend/internal/whatsapp/webhook_delivery_worker.go`
- Story previa de eventos: `_bmad-output/implementation-artifacts/5-5-emision-eventos-whatsmeow.md`

## References

- Epic origen: `_bmad-output/planning-artifacts/epic-integracion-loyo.md`
- Reglas del proyecto: `docs/bmad-project-rules.md`
- Contexto técnico global: `_bmad-output/project-context.md`
- Health endpoint ya documentado: `docs/integracion-b2b.md`
- README principal actual: `README.md`
- Handler REST de webhooks: `backend/internal/http/handlers/v1_webhooks.go`
- Tests del handler REST: `backend/internal/http/handlers/v1_webhooks_test.go`
- Helpers de respuesta V1: `backend/internal/http/handlers/v1_helpers.go`
- Middleware de API key: `backend/internal/http/middleware/api_key_auth.go`
- Eventos de dominio: `backend/internal/domain/webhook.go`
- Worker de delivery: `backend/internal/whatsapp/webhook_delivery_worker.go`
- Story 5.4: `_bmad-output/implementation-artifacts/5-4-worker-envio-retry-exponencial.md`
- Story 5.5: `_bmad-output/implementation-artifacts/5-5-emision-eventos-whatsmeow.md`

## Dev Agent Record

### Agent Model Used

gpt-5

### Debug Log References

- Se validaron como fuentes de verdad `v1_webhooks.go`, `v1_webhooks_test.go`, `v1_helpers.go`, `api_key_auth.go`, `domain/webhook.go`, `webhook_delivery_worker.go`, `webhook_events.go` y `webhook_events_test.go`.
- Se confirmó que la rama activa coincide con el epic: `feature/integracion-loyo`.
- Se confirmó una diferencia importante entre la planificación inicial y el código real: el worker actual usa 6 intentos máximos y backoff `1m`, `5m`, `30m`, `2h`, `6h`; no ejecuta un retry adicional de 24h.
- Se confirmó que la entrega HTTP firma y envía solo `data`, mientras `event_type` viaja en `X-Wsapi-Event`.
- Como la 5.5 ya está implementada, los payloads de ejemplo del nuevo documento se tomaron del código y tests reales, sin inventar campos.
- Se validó la documentación con checks por `rg` sobre endpoints, headers, eventos, links y existencia de `webhook_events.go` / `webhook_events_test.go`.

### Completion Notes List

- Se creó `docs/webhooks-integracion.md` con contrato REST, eventos, payloads reales, firma HMAC, retries, idempotencia y troubleshooting.
- Se agregó discoverability desde `README.md` y `docs/integracion-b2b.md` sin romper el enfoque principal de deploy del README.
- La guía documenta el comportamiento real del código: `POST` devuelve `201` manualmente, mientras `GET` y `DELETE` usan envelope con `meta`.
- Los ejemplos de `message.received`, `session.connected`, `session.disconnected` y `message.status_update` quedaron alineados con `webhook_events.go` y `webhook_events_test.go`.
- La story quedó lista para `code-review` sin requerir cambios adicionales de backend/frontend.

### File List

- `_bmad-output/implementation-artifacts/5-6-documentacion-integracion-webhooks.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `README.md`
- `docs/integracion-b2b.md`
- `docs/webhooks-integracion.md`

## Change Log

- 2026-05-22 — Implementada la Story 5.6: documentación completa de webhooks, payloads reales, firma HMAC, retries e integración discoverable desde README/docs.