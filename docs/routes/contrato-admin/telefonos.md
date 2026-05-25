# Teléfonos — Gestión Admin

CRUD de teléfonos WhatsApp por empresa y gestión de sus sesiones.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/empresas/{id}/telefonos

Lista todos los teléfonos registrados para una empresa.

**Path params:**
- `id` — ID de la empresa.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/empresas/1/telefonos \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "phones": [
    {
      "id": 5,
      "empresa_id": 1,
      "numero": "999000111",
      "codigo_pais": "51",
      "numeroCompleto": "51999000111",
      "status": "active",
      "lastConnected": "2026-05-24T10:00:00Z"  // null si nunca conectado
    }
  ]
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado a esa empresa. |
| `404 Not Found` | Empresa no encontrada. |

---

## POST /api/admin/empresas/{id}/telefonos

Agrega un nuevo teléfono a una empresa.

**Path params:**
- `id` — ID de la empresa.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `codigo_pais` | string | Sí | Código de país sin `+`. Ej: `"51"` |
| `numero` | string | Sí | Número sin código de país. |
| `status` | string | No | Estado inicial. Default `"qr_pending"`. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/empresas/1/telefonos \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"codigo_pais": "51", "numero": "999000111"}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "phone": {
    "id": 5,
    "empresa_id": 1,
    "numero": "999000111",
    "codigo_pais": "51",
    "numeroCompleto": "51999000111",
    "status": "qr_pending"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o campos requeridos ausentes. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Empresa no encontrada. |
| `500 Internal Server Error` | Error al guardar en base de datos. |

---

## GET /api/admin/telefonos/{id}

Retorna el detalle de un teléfono por ID.

**Path params:**
- `id` — ID numérico del teléfono.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/telefonos/5 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "phone": {
    "id": 5,
    "empresa_id": 1,
    "numero": "999000111",
    "codigo_pais": "51",
    "numeroCompleto": "51999000111",
    "status": "active",
    "lastConnected": "2026-05-24T10:00:00Z"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Teléfono no encontrado. |

---

## PUT /api/admin/telefonos/{id}

Actualiza datos de un teléfono.

**Path params:**
- `id` — ID numérico del teléfono.

**Body JSON:** Mismos campos que `POST` (todos opcionales).

**Request:**
```bash
curl -X PUT https://tu-dominio.com/api/admin/telefonos/5 \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"status": "disconnected"}'
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "phone": { "id": 5, "status": "disconnected", "..." } }
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al actualizar. |

---

## DELETE /api/admin/telefonos/{id}

Elimina un teléfono y desconecta su sesión WhatsApp.

**Request:**
```bash
curl -X DELETE https://tu-dominio.com/api/admin/telefonos/5 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true }
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al eliminar. |

---

## POST /api/admin/telefonos/{id}/connect

Inicia la conexión WhatsApp de un teléfono desde el panel admin.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/telefonos/5/connect \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "phone_id": 5,
  "status": "initializing"
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al iniciar la sesión. |

---

## GET /api/admin/telefonos/{id}/connect/ws

Abre una conexión WebSocket admin para recibir eventos de sesión (QR, conexión) en tiempo real. No requiere mensaje `subscribe`.

**Condición:** Se autentica con el JWT de admin como query param `?token=` o header `Authorization: Bearer`.

**Request:**
```bash
wscat -c "wss://tu-dominio.com/api/admin/telefonos/5/connect/ws?token=$ADMIN_JWT"
```

> Emite los mismos eventos que el WS B2B (`qr`, `connected`, `disconnected`, `ping`, `error`).

---

## GET /api/admin/telefonos/{id}/webhooks

Lista todos los webhooks registrados para un teléfono específico.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/telefonos/5/webhooks \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "webhooks": [
    {
      "id": 12,
      "telefono_id": 5,
      "api_key_id": 3,
      "url": "https://mi-servidor.com/webhook",
      "eventos": ["message_received"],
      "activo": true,
      "failure_count": 0,
      "last_error": null,
      "last_success_at": null,
      "created_at": "2026-05-20T09:00:00Z"
    }
  ]
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado a este teléfono. |
| `404 Not Found` | Teléfono no encontrado. |
