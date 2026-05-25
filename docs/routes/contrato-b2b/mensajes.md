# Mensajes

Envío, consulta, edición y reintento de mensajes WhatsApp individuales.

**Condición general:** Requieren API Key en el header `Authorization: Bearer <API_KEY>`.

---

## GET /api/service/v1/mensajes

Lista los mensajes enviados desde el teléfono asociado a la API Key. Limitado a los mensajes del teléfono de la key.

**Query params:**

| Param | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `telefono_id` | integer | No | Filtra por teléfono (debe coincidir con el teléfono de la API Key, de lo contrario retorna `403`). |
| `limit` | integer | No | Cantidad de resultados. Mín `1`, máx `100`. Default `50`. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/service/v1/mensajes?limit=20" \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "messages": [
      {
        "reference_id": "msg_abc123",
        "telefono_id": 5,
        "destino": "51987654321",
        "contenido": "Hola, ¿cómo estás?",
        "adjuntos": [],              // array vacío si no hay adjuntos
        "estado": "sent",            // "pending" | "sent" | "delivered" | "failed"
        "tiempo": "2026-05-24T12:00:00Z"
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
| `400 Bad Request` — `INVALID_TELEFONO_ID` | `telefono_id` no es numérico. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `403 Forbidden` — `FORBIDDEN` | `telefono_id` no coincide con el teléfono de la API Key. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## POST /api/service/v1/mensajes

Envía un mensaje de texto (con o sin adjuntos) al destinatario indicado. La respuesta se devuelve con `202 Accepted` independientemente de si el envío fue exitoso o falló, verificable por el campo `ok`.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `destino` | string | Sí | Número de teléfono destino con código de país. Ej: `"51987654321"` |
| `contenido` | string | Sí | Texto del mensaje. |
| `adjuntos` | array | No | Lista de adjuntos (ver estructura abajo). |

**Estructura de un adjunto:**
```json
{
  "tipo": "imagen",      // "imagen" | "video" | "audio" | "documento"
  "url": "https://...",  // URL pública del archivo
  "nombre": "foto.jpg"   // nombre del archivo (opcional)
}
```

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/service/v1/mensajes \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"destino": "51987654321", "contenido": "Hola!"}'
```

**Respuesta exitosa `202 Accepted` (enviado):**
```json
{
  "ok": true,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "sent"
  },
  "meta": { "empresa_id": 1 }
}
```

**Respuesta con error de envío `202 Accepted` (fallido):**
```json
{
  "ok": false,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "failed",
    "error": "connection timeout"
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `INVALID_JSON` | Cuerpo no es JSON válido. |
| `400 Bad Request` — `SESSION_NOT_ACTIVE` | El teléfono de la API Key no está activo. |
| `400 Bad Request` — `INVALID_ATTACHMENT` | El adjunto tiene formato inválido. |
| `400 Bad Request` — `DESTINO_REQUERIDO` o similar | Validación de dominio fallida (`destino` vacío, número inválido, etc). |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `404 Not Found` — `TELEFONO_NOT_FOUND` | El teléfono de la API Key no existe. |
| `500 Internal Server Error` | Error al registrar el mensaje en base de datos. |

---

## GET /api/service/v1/mensajes/{id}

Retorna el detalle completo de un mensaje por su `reference_id`.

**Path params:**
- `id` — `reference_id` del mensaje (string UUID-like).

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/mensajes/msg_abc123 \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "message": {
      "reference_id": "msg_abc123",
      "telefono_id": 5,
      "destino": "51987654321",
      "contenido": "Hola!",
      "adjuntos": [],                   // array vacío si sin adjuntos
      "estado": "sent",                 // "pending" | "sent" | "delivered" | "failed"
      "error_reason": "",               // "" o null si no hubo error
      "retry_count": 0,
      "created_at": "2026-05-24T12:00:00Z",
      "timestamp_sent": "2026-05-24T12:00:01Z",  // null si aún no enviado
      "last_attempt": "2026-05-24T12:00:01Z"      // null si no hubo intento
    }
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `MISSING_REFERENCE_ID` | El path no contiene un `reference_id`. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `403 Forbidden` — `FORBIDDEN` | El mensaje no pertenece al teléfono de esta API Key. |
| `404 Not Found` — `MESSAGE_NOT_FOUND` | No existe mensaje con ese `reference_id`. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## PATCH /api/service/v1/mensajes/{id}

Edita el contenido o destino de un mensaje que aún no fue enviado o entregado. No es posible editar mensajes en estado `sent` o `delivered`.

**Path params:**
- `id` — `reference_id` del mensaje.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `contenido` | string | No* | Nuevo texto. |
| `destino` | string | No* | Nuevo número destino. |

*Al menos uno de los dos campos debe estar presente.

**Request:**
```bash
curl -X PATCH https://tu-dominio.com/api/service/v1/mensajes/msg_abc123 \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"contenido": "Texto corregido"}'
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "ok": true,
    "reference_id": "msg_abc123",
    "message": "Mensaje actualizado"
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `INVALID_JSON` | JSON inválido. |
| `400 Bad Request` — `MISSING_FIELDS` | Ni `contenido` ni `destino` presentes. |
| `400 Bad Request` — `INVALID_STATE` | El mensaje ya fue enviado o entregado (no editable). |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `403 Forbidden` — `FORBIDDEN` | El mensaje no pertenece a esta API Key. |
| `404 Not Found` — `MESSAGE_NOT_FOUND` | No existe mensaje con ese `reference_id`. |
| `500 Internal Server Error` — `UPDATE_ERROR` | Error al actualizar en base de datos. |

---

## POST /api/service/v1/mensajes/{id}

Reintenta el envío de un mensaje fallido. No soporta mensajes con adjuntos.

**Path params:**
- `id` — `reference_id` del mensaje.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/service/v1/mensajes/msg_abc123 \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `202 Accepted` (reintento exitoso):**
```json
{
  "ok": true,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "sent"
  },
  "meta": { "empresa_id": 1 }
}
```

**Respuesta con error de reintento `202 Accepted` (falló de nuevo):**
```json
{
  "ok": false,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "failed",
    "error": "connection timeout"
  },
  "meta": { "empresa_id": 1 }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` — `MISSING_REFERENCE_ID` | ID ausente en el path. |
| `400 Bad Request` — `INVALID_STATE` | El mensaje ya fue enviado o entregado. |
| `400 Bad Request` — `MEDIA_RETRY_UNSUPPORTED` | El mensaje tiene adjuntos; el reintento no está soportado para este caso. |
| `400 Bad Request` — `SESSION_NOT_ACTIVE` | El teléfono no está activo. |
| `401 Unauthorized` — `API_KEY_REQUIRED` | API Key ausente o inválida. |
| `403 Forbidden` — `FORBIDDEN` | El mensaje no pertenece a esta API Key. |
| `404 Not Found` — `MESSAGE_NOT_FOUND` | No existe mensaje con ese `reference_id`. |
| `404 Not Found` — `TELEFONO_NOT_FOUND` | Teléfono no encontrado. |
| `500 Internal Server Error` — `RETRY_ERROR` | Error al preparar el reintento. |
