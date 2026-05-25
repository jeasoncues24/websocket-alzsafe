# Difusiones (Broadcast)

Envío masivo de mensajes a múltiples destinatarios de forma asíncrona.

**Condición general:** Requieren API Key en `Authorization: Bearer <API_KEY>`.

---

## GET /api/service/v1/difusiones

Lista todas las difusiones creadas desde el teléfono de la API Key.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/difusiones \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "broadcasts": [
      {
        "reference_id": "bc550e8400-...",
        "telefono_id": 5,
        "total": 150,
        "adjuntos": [],         // array vacío si sin adjuntos
        "status": "completed",  // "pending" | "running" | "completed" | "failed"
        "created_at": "2026-05-24T12:00:00Z"
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

---

## GET /api/service/v1/difusiones/{id}

Retorna el detalle y resultados individuales de una difusión.

**Path params:**
- `id` — `reference_id` de la difusión (UUID string).

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/difusiones/bc550e8400-... \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "reference_id": "bc550e8400-...",
    "empresa_id": 1,
    "telefono_id": 5,
    "total": 3,
    "adjuntos": [],
    "status": "completed",
    "results": [
      { "destino": "51987000001", "ok": true, "error": null },
      { "destino": "51987000002", "ok": true, "error": null },
      { "destino": "51987000003", "ok": false, "error": "number not on whatsapp" }
    ],
    "created_at": "2026-05-24T12:00:00Z"
  },
  "meta": { "empresa_id": 1 }
}
```

> `results` puede ser `null` si la difusión aún no comenzó a procesarse.

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `MISSING_BROADCAST_ID` | ID ausente en el path. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `403 Forbidden` — `FORBIDDEN` | La difusión no pertenece al teléfono de esta API Key. |
| `404 Not Found` — `BROADCAST_NOT_FOUND` | No existe difusión con ese ID. |

---

## POST /api/service/v1/difusiones

Crea y encola una nueva difusión de mensajes. El procesamiento es asíncrono; la respuesta es inmediata con `202 Accepted`.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `destinos` | array de strings | Sí | Lista de números destino con código de país. Al menos 1. |
| `mensaje` | string | Sí | Texto del mensaje a enviar a todos. |
| `adjuntos` | array | No | Lista de adjuntos (misma estructura que en mensajes individuales). |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/service/v1/difusiones \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "destinos": ["51987000001", "51987000002"],
    "mensaje": "Oferta especial para ti!"
  }'
```

**Respuesta exitosa `202 Accepted`:**
```json
{
  "ok": true,
  "data": {
    "reference_id": "bc550e8400-...",
    "total": 2,
    "estado": "pending"
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `INVALID_JSON` | JSON inválido. |
| `400 Bad Request` — `MISSING_FIELDS` | `destinos` vacío o ausente. |
| `400 Bad Request` — `SESSION_NOT_ACTIVE` | El teléfono no está activo. |
| `400 Bad Request` — `INVALID_ATTACHMENT` | Adjunto con formato inválido. |
| `400 Bad Request` | Validación de dominio (destinos inválidos, límites excedidos, etc). |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `404 Not Found` — `TELEFONO_NOT_FOUND` | El teléfono de la API Key no existe. |
