# Mensajes — Vista Admin

Consulta y reintento de mensajes desde el panel de administración. Permite ver mensajes de todas las empresas.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/mensajes

Lista mensajes con filtros opcionales. Pensado para soporte y diagnóstico.

**Query params:**

| Param | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `empresa_id` | integer | — | Filtra por empresa. |
| `telefono_id` | integer | — | Filtra por teléfono. |
| `estado` | string | — | `"pending"` \| `"sent"` \| `"delivered"` \| `"failed"` |
| `page` | integer | `1` | Página. |
| `limit` | integer | `50` | Resultados por página. Máx `100`. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/admin/mensajes?empresa_id=1&estado=failed&limit=20" \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "messages": [
    {
      "reference_id": "msg_abc123",
      "empresa_id": 1,
      "telefono_id": 5,
      "destino": "51987654321",
      "contenido": "Hola!",
      "adjuntos": [],           // array vacío si sin adjuntos
      "estado": "failed",
      "error_reason": "connection timeout",  // "" si no hubo error
      "retry_count": 2,
      "created_at": "2026-05-24T12:00:00Z",
      "timestamp_sent": null,               // null si no fue enviado
      "last_attempt": "2026-05-24T12:05:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## POST /api/admin/mensajes/{id}

Reintenta el envío de un mensaje fallido desde el panel admin. Funciona igual que el retry del contrato empresa pero sin restricción de API Key.

**Path params:**
- `id` — `reference_id` del mensaje.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/mensajes/msg_abc123 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `202 Accepted` (reintento exitoso):**
```json
{
  "ok": true,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "sent"
  }
}
```

**Respuesta con error `202 Accepted` (falló de nuevo):**
```json
{
  "ok": false,
  "data": {
    "reference_id": "msg_abc123",
    "estado": "failed",
    "error": "connection timeout"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Mensaje no encontrado. |
| `400 Bad Request` | El mensaje ya fue enviado o tiene adjuntos (retry no soportado con media). |
| `500 Internal Server Error` | Error al preparar el reintento. |
