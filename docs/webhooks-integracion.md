# Integración por webhooks — wsapi

Guía detallada para integradores B2B, plataformas serverless y sistemas externos que necesitan recibir eventos de wsapi sin depender de WebSocket persistente.

---

## Tabla de contenidos

1. [Overview](#overview)
2. [Autenticación para registrar webhooks](#autenticación-para-registrar-webhooks)
3. [Endpoints de gestión](#endpoints-de-gestión)
4. [Eventos disponibles](#eventos-disponibles)
5. [Entrega HTTP del webhook](#entrega-http-del-webhook)
6. [Payloads reales de ejemplo](#payloads-reales-de-ejemplo)
7. [Verificación de firma HMAC](#verificación-de-firma-hmac)
8. [Retries, fallos e idempotencia](#retries-fallos-e-idempotencia)
9. [Troubleshooting](#troubleshooting)

---

## Overview

wsapi permite registrar webhooks **por API key (cada API key está atada a un teléfono)** para recibir eventos operativos de WhatsApp en tiempo real. El grain de ownership es el contrato del integrador: cada teléfono/número tiene su propia configuración de webhooks, independiente del resto de teléfonos de la misma empresa.

> **Modelo de ownership:** un webhook pertenece a una `api_key`, que a su vez está atada a un `telefono`. Si tu empresa tiene tres números con tres API keys distintas, cada uno puede tener sus propios webhooks; los eventos de un número nunca se entregan a webhooks de otro. Para soporte interno, el panel admin expone una vista read-only por empresa (`GET /api/service/v1/empresas/webhooks`).

Casos de uso típicos:

- aplicaciones serverless que no pueden mantener conexiones WebSocket persistentes
- paneles externos que necesitan reaccionar a mensajes entrantes
- integraciones B2B que quieren auditar estado de sesión y confirmaciones de mensajes

El flujo general es:

1. Tu sistema registra un webhook vía API REST autenticada con API key.
2. wsapi encola eventos internos en `webhooks_outbound_queue`.
3. Un worker background entrega esos eventos por `POST` firmado.
4. Tu receptor verifica la firma, procesa el body y responde `2xx`.

---

## Autenticación para registrar webhooks

Los endpoints de gestión de webhooks requieren **API key**.

### Forma recomendada

```http
X-API-Key: TU_API_KEY
```

### Alternativas soportadas

```http
Authorization: ApiKey TU_API_KEY
Authorization: Bearer TU_API_KEY
```

> Recomendación: usa `X-API-Key` para evitar ambigüedad con otros mecanismos auth.

---

## Endpoints de gestión

### Resumen

| Método | URL | Auth | Descripción |
|---|---|---|---|
| `POST` | `/api/service/v1/webhooks` | API key | Crea un webhook para la API key (teléfono) actual |
| `GET` | `/api/service/v1/webhooks` | API key | Lista los webhooks **de la API key actual** (no de toda la empresa) |
| `DELETE` | `/api/service/v1/webhooks/{id}` | API key | Elimina un webhook propio (404 si pertenece a otra API key) |
| `GET` | `/api/service/v1/empresas/webhooks` | JWT empresa (admin) | **Read-only**: lista todos los webhooks de la empresa para soporte |

### Restricciones

- La URL del webhook debe ser **HTTPS**.
- Debe enviarse al menos un evento.
- Eventos permitidos:
  - `message.received`
  - `message.status_update`
  - `session.connected`
  - `session.disconnected`
- Máximo por **API key**: `WEBHOOKS_MAX_PER_EMPRESA` (la env var conserva el nombre histórico; el límite se aplica por API key/teléfono, no por la empresa completa).
  - valor por defecto: `10`

---

### POST `/api/service/v1/webhooks`

Registra un nuevo webhook para la empresa autenticada.

#### Ejemplo de request

```bash
curl -X POST https://tu-wsapi.example.com/api/service/v1/webhooks \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: TU_API_KEY' \
  -d '{
    "url": "https://mi-sistema.example.com/hooks/wsapi",
    "eventos": ["message.received", "session.connected"]
  }'
```

#### Body esperado

```json
{
  "url": "https://mi-sistema.example.com/hooks/wsapi",
  "eventos": ["message.received", "session.connected"]
}
```

#### Respuesta exitosa (`201`)

```json
{
  "ok": true,
  "data": {
    "id": 123,
    "secret": "2d8e6d5d3f5c0d9e7a1b4c8e3f0a1d2b4e6c8a0d1f2b3c4d5e6f7a8b9c0d1e2"
  }
}
```

#### Notas importantes

- `secret` se muestra **solo una vez** al crear el webhook.
- Guarda ese secret de forma segura: lo necesitarás para verificar `X-Wsapi-Signature`.

#### Errores típicos

| Status | `error` | Motivo |
|---|---|---|
| `400` | `INVALID_JSON` | JSON mal formado |
| `400` | `INVALID_URL` | La URL no es HTTPS o es inválida |
| `400` | `INVALID_EVENTOS` | Lista de eventos vacía o con valores no permitidos |
| `400` | `MAX_WEBHOOKS_REACHED` | La empresa alcanzó el límite configurado |
| `401` | `API_KEY_REQUIRED` | Falta API key |

---

### GET `/api/service/v1/webhooks`

Lista todos los webhooks de la empresa autenticada.

#### Ejemplo de request

```bash
curl https://tu-wsapi.example.com/api/service/v1/webhooks \
  -H 'X-API-Key: TU_API_KEY'
```

#### Respuesta exitosa (`200`)

```json
{
  "ok": true,
  "data": {
    "webhooks": [
      {
        "id": 123,
        "url": "https://mi-sistema.example.com/hooks/wsapi",
        "eventos": ["message.received", "session.connected"],
        "activo": true,
        "failure_count": 0,
        "last_error": null,
        "last_success_at": null,
        "created_at": "2026-05-22T09:00:00Z",
        "updated_at": "2026-05-22T09:00:00Z"
      }
    ],
    "total": 1
  },
  "meta": {
    "empresa_id": 7,
    "timestamp": "2026-05-22T09:10:00Z"
  }
}
```

> El campo `secret` **no** aparece en listados.

---

### DELETE `/api/service/v1/webhooks/{id}`

Elimina un webhook perteneciente a la empresa autenticada.

#### Ejemplo de request

```bash
curl -X DELETE https://tu-wsapi.example.com/api/service/v1/webhooks/123 \
  -H 'X-API-Key: TU_API_KEY'
```

#### Respuesta exitosa (`200`)

```json
{
  "ok": true,
  "data": {
    "deleted": true
  },
  "meta": {
    "empresa_id": 7,
    "timestamp": "2026-05-22T09:15:00Z"
  }
}
```

#### Error típico

| Status | `error` | Motivo |
|---|---|---|
| `404` | `NOT_FOUND` | El webhook no existe o no pertenece a la empresa |

---

## Eventos disponibles

| Evento | Cuándo se emite | Fuente real |
|---|---|---|
| `message.received` | Cuando entra un mensaje real a la sesión WhatsApp | `*events.Message` |
| `message.status_update` | Cuando llega un receipt sobre un mensaje saliente | `*events.Receipt` |
| `session.connected` | Cuando wsapi marca la sesión como conectada | `markConnected(...)` |
| `session.disconnected` | Cuando wsapi marca la sesión como desconectada | `markDisconnected(...)` |

### Notas sobre `message.status_update`

- `message_id` corresponde al identificador del proveedor WhatsApp.
- `reference_id` es **opcional**.
- wsapi solo incluye `reference_id` cuando puede resolverlo de forma confiable.
- Si no existe correlación determinista, el evento sale con `message_id`, `status` y `timestamp`, sin inventar relaciones falsas.

---

## Entrega HTTP del webhook

Cada evento se entrega mediante `POST` al `url` registrado.

### Headers enviados por wsapi

| Header | Valor | Descripción |
|---|---|---|
| `Content-Type` | `application/json` | Siempre JSON |
| `X-Wsapi-Signature` | `sha256=<hex>` | Firma HMAC-SHA256 del body crudo |
| `X-Wsapi-Event` | nombre del evento | Ej. `message.received` |
| `X-Wsapi-Delivery` | ID numérico | ID de entrega, útil para idempotencia |

### Importante: envelope interno vs body real

Internamente wsapi guarda en cola un envelope con esta forma:

```json
{
  "event_type": "message.received",
  "data": { ... }
}
```

Pero el receptor HTTP **no** recibe ese envelope completo.

- El body enviado por HTTP es **solo `data`**.
- El nombre del evento llega en `X-Wsapi-Event`.
- La firma se calcula sobre el **raw body** real enviado al receptor.

---

## Payloads reales de ejemplo

Los siguientes ejemplos se basan en la implementación actual de `backend/internal/whatsapp/webhook_events.go` y `backend/internal/whatsapp/webhook_events_test.go`.

### `message.received`

Header:

```http
X-Wsapi-Event: message.received
```

Body:

```json
{
  "telefono_id": 12,
  "from": "51944455566",
  "message_id": "wamid-123",
  "content": "hola webhook",
  "type": "text",
  "timestamp": "2026-05-22T15:04:05Z"
}
```

### `session.connected`

Header:

```http
X-Wsapi-Event: session.connected
```

Body:

```json
{
  "telefono_id": 10,
  "phone": "51999888777",
  "timestamp": "2026-05-22T14:20:00Z"
}
```

### `session.disconnected`

Header:

```http
X-Wsapi-Event: session.disconnected
```

Body:

```json
{
  "telefono_id": 11,
  "phone": "51999000111",
  "reason": "logged_out",
  "timestamp": "2026-05-22T14:21:00Z"
}
```

### `message.status_update`

Header:

```http
X-Wsapi-Event: message.status_update
```

Body con correlación resuelta:

```json
{
  "telefono_id": 13,
  "message_id": "wamid-out-1",
  "reference_id": "ref-123",
  "status": "read",
  "timestamp": "2026-05-22T16:00:00Z"
}
```

Body cuando `reference_id` no puede resolverse de forma confiable:

```json
{
  "telefono_id": 13,
  "message_id": "wamid-out-1",
  "status": "delivered",
  "timestamp": "2026-05-22T16:00:00Z"
}
```

### Valores típicos de `status` para `message.status_update`

| Valor | Significado |
|---|---|
| `sent` | confirmado por receipt tipo sender |
| `delivered` | entregado al dispositivo destino |
| `read` | leído por el destinatario |
| `played` | reproducido/visto en media view-once |
| `retry` | el destinatario pidió reintento |

---

## Verificación de firma HMAC

La firma usa HMAC-SHA256 y formato:

```text
sha256=<hex>
```

### Reglas

1. Lee el **raw body** exacto antes de parsear JSON.
2. Recalcula `HMAC_SHA256(raw_body, secret)`.
3. Compara contra `X-Wsapi-Signature` en tiempo constante.
4. Solo después parsea el JSON.

---

### Ejemplo en Node.js

```js
import crypto from "node:crypto";
import express from "express";

const app = express();

app.post(
  "/hooks/wsapi",
  express.raw({ type: "application/json" }),
  (req, res) => {
    const signature = req.header("X-Wsapi-Signature") || "";
    const secret = process.env.WSAPI_WEBHOOK_SECRET;

    if (!signature.startsWith("sha256=")) {
      return res.status(401).json({ ok: false, error: "invalid_signature_format" });
    }

    const expected = crypto
      .createHmac("sha256", secret)
      .update(req.body)
      .digest("hex");

    const receivedHex = signature.slice("sha256=".length);
    const received = Buffer.from(receivedHex, "hex");
    const expectedBuf = Buffer.from(expected, "hex");

    if (
      received.length !== expectedBuf.length ||
      !crypto.timingSafeEqual(received, expectedBuf)
    ) {
      return res.status(401).json({ ok: false, error: "invalid_signature" });
    }

    const eventType = req.header("X-Wsapi-Event");
    const deliveryId = req.header("X-Wsapi-Delivery");
    const payload = JSON.parse(req.body.toString("utf8"));

    console.log({ eventType, deliveryId, payload });
    return res.status(200).json({ ok: true });
  }
);
```

---

### Ejemplo en Go

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func wsapiWebhookHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		signature := strings.TrimSpace(r.Header.Get("X-Wsapi-Signature"))
		if !strings.HasPrefix(signature, "sha256=") {
			http.Error(w, "invalid signature format", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedHex := hex.EncodeToString(mac.Sum(nil))

		receivedHex := strings.TrimPrefix(signature, "sha256=")
		received, err := hex.DecodeString(receivedHex)
		if err != nil {
			http.Error(w, "invalid signature encoding", http.StatusUnauthorized)
			return
		}
		expected, err := hex.DecodeString(expectedHex)
		if err != nil {
			http.Error(w, "cannot build expected signature", http.StatusInternalServerError)
			return
		}

		if len(received) != len(expected) || !hmac.Equal(received, expected) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}
}
```

---

## Retries, fallos e idempotencia

### Política real de entrega

| Parámetro | Valor actual |
|---|---|
| Polling del worker | cada `5s` |
| Timeout por request | `10s` |
| Intentos máximos totales | `6` |
| Backoff | `1m`, `5m`, `30m`, `2h`, `6h` |
| Estado terminal del item | `failed` |
| Desactivación automática del webhook | `failure_count >= 20` |

### Qué se considera éxito

- cualquier respuesta `2xx`
- el item de cola pasa a `done`
- `failure_count` del webhook vuelve a `0`

### Qué se considera retry

- errores de red
- timeouts
- respuestas `5xx`

### Qué se considera fallo terminal

- payload inválido en cola
- agotar los intentos permitidos

### Idempotencia recomendada

Usa `X-Wsapi-Delivery` como clave de deduplicación en tu receptor.

Ejemplo de estrategia:

- si ya procesaste `delivery_id = 12345`, responde `200` sin reprocesar
- si es nuevo, guarda el ID y procesa normalmente

---

## Troubleshooting

| Síntoma | Causa probable | Qué revisar |
|---|---|---|
| No llega ningún webhook | No hay suscriptores activos para ese evento | `GET /api/service/v1/webhooks`, campo `activo` y lista `eventos` |
| Responde `INVALID_URL` al crear | URL no HTTPS o mal formada | usa `https://...` con host válido |
| La firma no coincide | Se verificó contra JSON parseado y no contra raw body | usa body crudo antes de `JSON.parse` / `json.Unmarshal` |
| Llega `message.status_update` sin `reference_id` | No hubo correlación determinista disponible | usa `message_id` como identificador fuerte del proveedor |
| El webhook deja de recibir eventos | `failure_count` alcanzó umbral de desactivación | lista webhooks y revisa `activo`, `failure_count`, `last_error` |
| Recibes duplicados | Retry tras timeout o `5xx` | implementa idempotencia con `X-Wsapi-Delivery` |

---

## Resumen rápido para integradores

```text
1. Registrar webhook con POST /api/service/v1/webhooks
2. Guardar el secret devuelto una sola vez
3. Escuchar POSTs JSON en tu endpoint HTTPS
4. Verificar X-Wsapi-Signature sobre el raw body
5. Leer el tipo de evento desde X-Wsapi-Event
6. Usar X-Wsapi-Delivery para idempotencia
7. Responder 2xx cuando el evento ya fue aceptado/procesado
```
