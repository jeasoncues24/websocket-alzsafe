# Sesiones — Vista Admin

Consulta, diagnóstico y gestión de sesiones WhatsApp desde el panel de administración.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/sesiones

Lista todas las sesiones activas o registradas en el sistema.

**Query params:**

| Param | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `empresa_id` | integer | — | Filtra por empresa. |
| `page` | integer | `1` | Página. |
| `limit` | integer | `20` | Resultados por página. |

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/sesiones \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "sessions": [
    {
      "telefono_id": 5,
      "empresa_id": 1,
      "numeroCompleto": "51999000111",
      "status": "active",
      "lastConnected": "2026-05-24T10:00:00Z"  // null si nunca conectado
    }
  ],
  "total": 1
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## POST /api/admin/sesiones

Crea una nueva sesión (teléfono) para una empresa desde el panel admin.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `empresa_id` | integer | Sí | ID de la empresa a la que pertenecerá. |
| `codigo_pais` | string | Sí | Código de país sin `+`. |
| `numero` | string | Sí | Número sin código de país. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/sesiones \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"empresa_id": 1, "codigo_pais": "51", "numero": "999000111"}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "session": {
    "telefono_id": 5,
    "empresa_id": 1,
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
| `404 Not Found` | Empresa no encontrada. |
| `500 Internal Server Error` | Error al crear. |

---

## GET /api/admin/sesiones/diagnostico

Retorna el diagnóstico de todas las sesiones: compara el estado en base de datos vs el estado en tiempo de ejecución del manager WhatsApp. Útil para detectar desincronizaciones.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/sesiones/diagnostico \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "diagnostics": [
    {
      "telefono_id": 5,
      "empresa_id": 1,
      "account_id": "51999000111@s.whatsapp.net",
      "status_db": "active",
      "status_runtime": "disconnected",
      "runtime_connected": false,
      "mismatch": true,
      "mismatch_reason": "db_active_runtime_disconnected",
      "recommended_action": "reanudar_conexion"
    }
  ]
}
```

> `mismatch_reason` posibles valores:
> - `"db_active_runtime_disconnected"` — DB dice activo pero el manager no tiene conexión.
> - `"db_not_active_runtime_connected"` — DB dice inactivo pero el manager tiene conexión.
> - `""` — sin desincronización.

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |

---

## POST /api/admin/telefonos/{id}/qr-link

Genera un token QR-link provisional (JWT de corta duración) para que el usuario cliente pueda escanear el QR sin tener acceso al JWT de empresa.

**Path params:**
- `id` — ID del teléfono.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/telefonos/5/qr-link \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "qr_link_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 300,
  "ws_url": "wss://tu-dominio.com/api/service/v1/ws?token=<qr_link_token>"
}
```

> El cliente usa `ws_url` para conectarse al WebSocket y recibir el QR sin credenciales de empresa.

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Teléfono no encontrado. |
| `500 Internal Server Error` | Error al generar el token. |
