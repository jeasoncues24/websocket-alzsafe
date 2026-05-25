# API Keys — Gestión Admin

Creación, consulta, rotación, revocación y auditoría de API Keys por teléfono.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/telefonos/{id}/api-keys

Lista todas las API Keys asociadas a un teléfono.

**Path params:**
- `id` — ID del teléfono.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/telefonos/5/api-keys \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "api_keys": [
    {
      "id": 3,
      "empresa_id": 1,
      "telefono_id": 5,
      "nombre": "Key produccion",
      "key_prefix": "wsk_abc",
      "scopes": ["send_message"],
      "activo": true,
      "expires_at": null,                    // null si no expira
      "created_by_user_id": 1,               // null si fue creada por el sistema
      "rotated_from_id": null,               // null si no es rotación de otra key
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado a ese teléfono. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al consultar base de datos. |

---

## POST /api/admin/telefonos/{id}/api-keys

Crea una nueva API Key para un teléfono. La respuesta incluye el `secret` completo, que solo se muestra una vez.

**Path params:**
- `id` — ID del teléfono.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `nombre` | string | No | Nombre descriptivo. Default: `"Key <numero>"`. |
| `scopes` | array de strings | No | Permisos de la key. |
| `expires_at` | string (RFC3339) | No | Fecha de expiración. Ej: `"2027-01-01T00:00:00Z"`. `null` = sin expiración. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/telefonos/5/api-keys \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Key produccion", "scopes": ["send_message"]}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "api_key": {
    "id": 3,
    "empresa_id": 1,
    "telefono_id": 5,
    "nombre": "Key produccion",
    "key_prefix": "wsk_abc",
    "scopes": ["send_message"],
    "activo": true,
    "expires_at": null,
    "created_at": "2026-05-24T12:00:00Z"
  },
  "secret": "wsk_abc.xyzXYZ1234...",
  "message": "API key creada exitosamente. Guárdala en un lugar seguro."
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido, ID inválido o `expires_at` con formato incorrecto. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado a ese teléfono. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al generar o guardar la key. |

---

## GET /api/admin/api-keys/{id}

Retorna el detalle de una API Key por ID (sin el secret).

**Path params:**
- `id` — ID numérico de la API Key.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/api-keys/3 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "api_key": {
    "id": 3,
    "empresa_id": 1,
    "telefono_id": 5,
    "nombre": "Key produccion",
    "key_prefix": "wsk_abc",
    "scopes": ["send_message"],
    "activo": true,
    "expires_at": null,
    "created_by_user_id": 1,
    "rotated_from_id": null,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | API Key no encontrada. |

---

## POST /api/admin/api-keys/{id}/rotate

Rota una API Key: crea una nueva (con el mismo nombre y scopes) e invalida la anterior. La respuesta incluye el nuevo `secret`.

**Path params:**
- `id` — ID de la API Key a rotar.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/api-keys/3/rotate \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "api_key": {
    "id": 4,
    "nombre": "Key produccion",
    "key_prefix": "wsk_xyz",
    "activo": true,
    "rotated_from_id": 3,   // ID de la key anterior
    "..."
  },
  "secret": "wsk_xyz.nuevoSecret...",
  "message": "API key rotada exitosamente. La anterior quedó revocada."
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | API Key no encontrada. |
| `500 Internal Server Error` | Error al generar o guardar la nueva key. |

---

## POST /api/admin/api-keys/{id}/revoke

Revoca una API Key, dejándola inactiva inmediatamente.

**Path params:**
- `id` — ID de la API Key.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/api-keys/3/revoke \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "api_key": { "id": 3, "activo": false, "..." }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | API Key no encontrada. |
| `500 Internal Server Error` | Error al revocar. |

---

## GET /api/admin/api-keys/{id}/audit

Retorna el historial de eventos de auditoría de una API Key (creación, rotación, revocación, usos).

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/api-keys/3/audit \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "audit": [
    {
      "id": 10,
      "api_key_id": 3,
      "empresa_id": 1,
      "telefono_id": 5,
      "action": "created",          // "created" | "rotated" | "revoked" | "used"
      "actor_user_id": 1,           // null si fue acción del sistema
      "metadata": { "nombre": "Key produccion" },
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | API Key no encontrada. |
| `500 Internal Server Error` | Error al consultar. |

---

## GET /api/admin/api-keys/{id}/usage/stats

Retorna estadísticas de uso agregadas de una API Key.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/api-keys/3/usage/stats \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "stats": {
    "total_requests": 1250,
    "requests_today": 45,
    "last_used_at": "2026-05-24T11:55:00Z"  // null si nunca usada
  }
}
```

---

## GET /api/admin/api-keys/{id}/usage/timeseries

Retorna el uso de la API Key en series de tiempo (por día/hora).

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/api-keys/3/usage/timeseries \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "timeseries": [
    { "period": "2026-05-24", "requests": 45 },
    { "period": "2026-05-23", "requests": 120 }
  ]
}
```

---

## GET /api/admin/api-keys/{id}/audit/stats

Retorna estadísticas agregadas de eventos de auditoría de una API Key.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/api-keys/3/audit/stats \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "stats": {
    "total_events": 5,
    "by_action": {
      "created": 1,
      "rotated": 1,
      "revoked": 0,
      "used": 3
    }
  }
}
```
