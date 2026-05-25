# Me — Información de la API Key

Endpoints para que una integración consulte su propia identidad y el estado de sesión en tiempo de ejecución.

**Condición general:** Requieren API Key en `Authorization: Bearer <API_KEY>`.

---

## GET /api/service/v1/me

Retorna el detalle completo de la API Key activa, la empresa y el teléfono al que pertenece, junto con el estado de sesión en tiempo real.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/me \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "api_key": {
    "id": 3,
    "nombre": "Key produccion",
    "key_prefix": "wsk_abc",
    "empresa_id": 1,
    "telefono_id": 5,
    "scopes": ["send_message"],
    "activo": true,
    "expires_at": null,          // null si no tiene fecha de expiración
    "created_at": "2026-01-01T00:00:00Z"
  },
  "empresa": {
    "id": 1,
    "ruc": "20123456789",
    "nombre": "Mi Empresa SAC",
    "activo": true
  },
  "telefono": {
    "id": 5,
    "numero": "999000111",
    "codigo_pais": "51",
    "numeroCompleto": "51999000111",
    "status": "active"
  },
  "session_runtime": {
    "telefono_id": 5,
    "account_id": "51999000111@s.whatsapp.net",
    "status_db": "active",
    "status_runtime": "connected",   // "connected" | "disconnected"
    "runtime_connected": true,
    "mismatch": false,               // true si DB y runtime difieren
    "mismatch_reason": "",           // "db_active_runtime_disconnected" | "db_not_active_runtime_connected" | ""
    "recommended_action": "none"     // "none" | "reanudar_conexion" | "iniciar_conexion"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | API Key ausente o inválida. |
| `404 Not Found` | API Key, teléfono o empresa no encontrados (desincronización de datos). |

---

## GET /api/service/v1/sesion

Retorna únicamente el estado de sesión en tiempo real del teléfono asociado a la API Key.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/sesion \
  -H "Authorization: Bearer $API_KEY"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "data": {
    "telefono_id": 5,
    "account_id": "51999000111@s.whatsapp.net",
    "status_db": "active",
    "status_runtime": "connected",
    "runtime_connected": true,
    "mismatch": false,
    "mismatch_reason": "",
    "recommended_action": "none"
  }
}
```

> Usar `recommended_action` para decidir si es necesario reconectar antes de enviar mensajes.

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | API Key ausente o inválida. |
| `404 Not Found` | Teléfono no encontrado. |
