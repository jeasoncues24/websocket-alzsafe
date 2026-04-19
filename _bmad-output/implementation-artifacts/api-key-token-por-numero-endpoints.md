# Documentación de Endpoints con API Key

## Overview

Este documento describe los endpoints disponibles para consumo mediante **API Keys por teléfono** (`token_por_numero`). Las API Keys permiten autenticación programmatic para enviar mensajes y difusiones sin usar JWT de empresa.

## Autenticación

Las API Keys se autentican enviando el header:

```
X-API-Key: <tu_api_key>
```

O usando el header `Authorization`:

```
Authorization: Bearer <tu_api_key>
Authorization: ApiKey <tu_api_key>
```

## Endpoints Disponibles

### 1. GET /api/v1/me

**Propósito:** Obtener información de la API Key, empresa y teléfono asociados.

**Autenticación:** Requiere API Key.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "api_key": {
    "id": 1,
    "key_prefix": "wapi_abc123",
    "empresa_id": 10,
    "telefono_id": 5,
    "scopes": ["message:send", "broadcast:send"],
    "activo": true,
    "created_at": "2026-04-15T10:00:00Z",
    "expires_at": "2027-04-15T10:00:00Z"
  },
  "empresa": {
    "id": 10,
    "ruc": "20123456789",
    "nombre": "Mi Empresa SAC",
    "activo": true
  },
  "telefono": {
    "id": 5,
    "numero": "+519999999999",
    "status": "active"
  }
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| API_KEY_REQUIRED | 401 | No se proporcionó API Key |
| INVALID_API_KEY | 401 | API Key inválida o expirada |
| TELEFONO_NOT_FOUND | 401 | Teléfono asociado no encontrado |
| EMPRESA_NOT_FOUND | 401 | Empresa no encontrada |
| EMPRESA_INACTIVE | 403 | Empresa inactiva |

---

### 2. GET /api/v1/messages

**Propósito:** Listar mensajes enviados por la empresa/teléfono.

**Autenticación:** Requiere API Key.

**Query Parameters:**

| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| telefono_id | int64 | (Opcional) Filtrar por teléfono específico |
| limit | int | (Opcional) Límite de resultados (default 50, max 100) |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "messages": [
    {
      "id": 100,
      "telefono_id": 5,
      "destino": "+519888888888",
      "contenido": "Hola, este es un mensaje de prueba",
      "estado": "sent",
      "tiempo": "2026-04-18T10:30:00Z"
    }
  ],
  "total": 1
}
```

**Notas:**
- Si la API Key está asociada a un teléfono específico, solo retorna mensajes de ese teléfono.
- El campo `telefono_id` es obligatorio para API Keys con scope de teléfono.

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| FORBIDDEN | 403 | La API key solo puede usarse con su teléfono asignado |
| INVALID_TELEFONO_ID | 400 | telefono_id inválido |

---

### 3. POST /api/v1/messages

**Propósito:** Enviar un mensaje directo a un número WhatsApp.

**Autenticación:** Requiere API Key.

**Body (JSON):**

```json
{
  "telefono_id": 5,
  "destino": "+519888888888",
  "contenido": "Hola, este es un mensaje de prueba"
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| telefono_id | int64 | Opcional* | ID del teléfono a usar. Si es 0, usa el teléfono de la API Key |
| destino | string | Sí | Número de destino con código de país |
| contenido | string | Sí | Texto del mensaje |

*Para API Keys de teléfono específico, `telefono_id` debe ser 0 o coincidir con el teléfono asignado.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "status": "sent"
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELDS | 400 | Faltan campos requeridos |
| FORBIDDEN | 403 | La API key no corresponde al teléfono |
| TELEFONO_NOT_FOUND | 404 | Teléfono no encontrado |
| SESSION_NOT_ACTIVE | 400 | El teléfono no está activo |
| INVALID_JSON | 400 | JSON inválido |

---

### 4. GET /api/v1/broadcasts

**Propósito:** Listar difusiones (broadcasts) enviadas.

**Autenticación:** Requiere API Key.

**Query Parameters:**

| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| telefono_id | int64 | (Opcional) Filtrar por teléfono específico |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "broadcasts": [
    {
      "reference_id": "BC_10_100",
      "telefono_id": 5,
      "total": 100,
      "status": "completed",
      "created_at": "2026-04-18T09:00:00Z"
    }
  ],
  "total": 1
}
```

**Notas:**
- Si la API Key está asociada a un teléfono específico, solo retorna difusiones de ese teléfono.

---

### 5. POST /api/v1/broadcasts

**Propósito:** Crear una difusión masiva a múltiples destinos.

**Autenticación:** Requiere API Key.

**Body (JSON):**

```json
{
  "telefono_id": 5,
  "destinos": ["+519888888888", "+519777777777", "+519666666666"],
  "mensaje": "Esta es una difusión masiva"
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| telefono_id | int64 | Opcional* | Teléfono a usar para la difusión |
| destinos | array | Sí | Lista de números destino |
| mensaje | string | Sí | Contenido del mensaje |

*Para API Keys de teléfono específico, `telefono_id` debe ser 0 o coincidir.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "reference_id": "BC_10_3",
  "total": 3,
  "status": "pending"
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELDS | 400 | Faltan campos requeridos |
| FORBIDDEN | 403 | Teléfono no corresponde a la API Key |
| TELEFONO_NOT_FOUND | 404 | Teléfono no encontrado |

---

## Uso de API Keys por Teléfono

### Características del token_por_numero

1. **Scoped por teléfono:** La API Key está vinculada a un teléfono específico.
2. **Sin selección de teléfono:** El parámetro `telefono_id` es opcional; si no se envía, usa el teléfono de la API Key.
3. **Restricción de acceso:** Solo puede acceder a datos del teléfono asociado.

### Ejemplo de uso

```bash
# Obtener información de la API Key
curl -X GET https://api.tuservidor.com/api/v1/me \
  -H "X-API-Key: wapi_abc123DEF456"

# Enviar mensaje (telefono_id opcional, usa el de la API Key)
curl -X POST https://api.tuservidor.com/api/v1/messages \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destino": "+519888888888", "contenido": "Hola mundo"}'

# Crear difusión
curl -X POST https://api.tuservidor.com/api/v1/broadcasts \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destinos": ["+519888888888", "+519777777777"], "mensaje": "Difusión"}'
```

---

## Códigos de Error Comunes

| Código | HTTP | Causa |
|--------|------|-------|
| API_KEY_REQUIRED | 401 | No se envió header de API Key |
| INVALID_API_KEY | 401 | La key no existe o está inactiva/ expirada |
| TELEFONO_NOT_FOUND | 401 | El teléfono asociado no existe |
| FORBIDDEN | 403 | Teléfono no coincide con la API Key |
| EMPRESA_INACTIVE | 403 | La empresa está deshabilitada |
| MISSING_FIELDS | 400 | Faltan campos requeridos en el request |
| INVALID_JSON | 400 | El body no es JSON válido |