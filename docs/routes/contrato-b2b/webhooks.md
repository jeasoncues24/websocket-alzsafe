# Webhooks

Registro y gestión de webhooks para recibir eventos de WhatsApp en tu servidor.

Los eventos disponibles son:
- `message_received` — mensaje entrante recibido.
- `message_status` — cambio de estado de un mensaje enviado.
- `session_connected` — sesión WhatsApp conectada.
- `session_disconnected` — sesión WhatsApp desconectada.

---

## POST /api/service/v1/webhooks

Registra un nuevo webhook para la API Key activa. Solo acepta URLs HTTPS. La respuesta incluye el `secret` para verificar la firma HMAC de los eventos recibidos; guárdalo en un lugar seguro porque no se vuelve a mostrar.

**Condición:** Requiere API Key en `Authorization: Bearer <API_KEY>`.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `url` | string | Sí | URL HTTPS de tu endpoint receptor. |
| `eventos` | array de strings | Sí | Al menos un evento de la lista válida. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/service/v1/webhooks \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://mi-servidor.com/webhook",
    "eventos": ["message_received", "message_status"]
  }'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "data": {
    "id": 12,
    "secret": "a3f7b2c1d4e5..."  // secreto HMAC, solo se muestra una vez
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `INVALID_JSON` | JSON inválido. |
| `400 Bad Request` — `INVALID_URL` | URL no es HTTPS o tiene formato incorrecto. |
| `400 Bad Request` — `INVALID_EVENTOS` | Array vacío o contiene un evento no reconocido. |
| `400 Bad Request` — `MAX_WEBHOOKS_REACHED` | Se alcanzó el límite máximo de webhooks activos para esta API Key. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `500 Internal Server Error` | Error al generar el secret o al guardar en base de datos. |

---

## GET /api/service/v1/webhooks

Lista todos los webhooks registrados para la API Key activa.

**Condición:** Requiere API Key en `Authorization: Bearer <API_KEY>`.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/webhooks \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "webhooks": [
      {
        "id": 12,
        "telefono_id": 5,
        "api_key_id": 3,
        "url": "https://mi-servidor.com/webhook",
        "eventos": ["message_received", "message_status"],
        "activo": true,
        "failure_count": 0,
        "last_error": null,           // null si no hubo errores previos
        "last_success_at": "2026-05-24T11:00:00Z",  // null si nunca se entregó
        "created_at": "2026-05-20T09:00:00Z",
        "updated_at": "2026-05-20T09:00:00Z"
      }
    ],
    "total": 1
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## DELETE /api/service/v1/webhooks/{id}

Elimina un webhook. Solo puede eliminar webhooks creados por la misma API Key.

**Condición:** Requiere API Key en `Authorization: Bearer <API_KEY>`.

**Path params:**
- `id` — ID numérico del webhook.

**Request:**
```bash
curl -X DELETE https://tu-dominio.com/api/service/v1/webhooks/12 \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "deleted": true
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `INVALID_ID` | El ID del path no es un número válido. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `404 Not Found` — `NOT_FOUND` | Webhook no encontrado o no pertenece a esta API Key. |
| `500 Internal Server Error` | Error al eliminar el webhook. |

