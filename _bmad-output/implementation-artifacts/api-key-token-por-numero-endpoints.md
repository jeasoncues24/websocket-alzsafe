# Documentación de API Key por Teléfono

## Overview

Este documento explica como usar una **API Key por teléfono** (`api_key_telefono` o `token_por_numero`).

La idea es simple:
- la key identifica a una empresa y a un telefono especifico
- no hace falta mandar `telefono_id` en el body de los endpoints de consumo
- se puede enviar texto, imagenes y otros adjuntos permitidos
- el mismo token sirve para mensajes, difusiones, consulta y seguimiento

## Resumen rapido

- Header principal: `X-API-Key`
- Tambien funciona `Authorization: Bearer <token>` o `Authorization: ApiKey <token>`
- Un solo adjunto por mensaje o difusion
- Tamano maximo por archivo: 5 MB
- Tamano maximo total por request: 20 MB
- Broadcast maximo: 500 destinos
- Audio no lleva caption ni texto junto al adjunto

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

## Adjuntos permitidos

Cuando el mensaje o la difusion necesita una imagen, PDF, audio o documento, usa `adjuntos`.

### Formato del adjunto

```json
{
  "nombre": "foto.jpg",
  "mime_type": "image/jpeg",
  "contenido_base64": "...base64..."
}
```

### Tipos permitidos

- `image/jpeg`
- `image/png`
- `video/mp4`
- `audio/mpeg`
- `audio/mp4`
- `audio/ogg`
- `audio/wav`
- `application/pdf`
- `application/msword`
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`

### Reglas simples

- Solo se permite 1 adjunto por request
- El archivo debe coincidir con su extension y su MIME type
- Audio no puede llevar `mensaje` al mismo tiempo
- Si mandas adjunto, el contenido debe ir en Base64

## Endpoints Disponibles

### 1. GET /api/me

**Propósito:** Obtener información de la API Key, empresa y teléfono asociados.

**Compatibilidad:** También disponible como `GET /api/v1/me`.

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

### 1.1 GET /api/sesion

**Propósito:** Diagnosticar estado de sesión del teléfono de la API key comparando DB vs runtime real.

**Autenticación:** Requiere API Key.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "telefono_id": 5,
    "account_id": "51999999999",
    "status_db": "active",
    "status_runtime": "connected",
    "runtime_connected": true,
    "mismatch": false,
    "mismatch_reason": "",
    "recommended_action": "none"
  }
}
```

**Notas:**
- `mismatch=true` indica inconsistencia entre DB y cliente runtime.
- `recommended_action` puede ser `none`, `reanudar_conexion` o `iniciar_conexion`.

---

### 2. GET /api/mensajes

**Propósito:** Listar mensajes enviados por el teléfono de esta API Key.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Query Parameters:**

| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| limit | int | (Opcional) Límite de resultados (default 50, max 100) |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "messages": [
      {
        "reference_id": "550e8400-e29b-41d4-a716-446655440000",
        "telefono_id": 5,
        "destino": "519888888888",
        "contenido": "Hola, este es un mensaje de prueba",
        "adjuntos": [
          {
            "nombre": "foto.jpg",
            "sha256_hash": "abcd1234",
            "tamano_bytes": 245000
          }
        ],
        "estado": "sent",
        "tiempo": "2026-04-18T10:30:00Z"
      }
    ],
    "total": 1
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Notas:**
- Solo retorna mensajes del teléfono asociado a la API Key.
- El campo `telefono_id` ya no es un parámetro aceptado en el body — está implícito en la API Key.
- Si el mensaje tuvo adjuntos, también se devuelven en `adjuntos`.

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| API_KEY_REQUIRED | 401 | No se proporcionó API Key |
| INVALID_TELEFONO_ID | 400 | telefono_id en query string inválido |

---

### 3. POST /api/mensajes

**Propósito:** Enviar un mensaje directo a un número WhatsApp. El mensaje es enviado en tiempo real via WhatsApp.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Body (JSON):**

```json
{
  "destino": "519888888888",
  "contenido": "Hola, este es un mensaje de prueba",
  "adjuntos": [
    {
      "nombre": "foto.jpg",
      "mime_type": "image/jpeg",
      "contenido_base64": "...base64..."
    }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| destino | string | Sí | Número de destino con código de país (sin '+') |
| contenido | string | Condicional | Texto del mensaje. Obligatorio si no envias adjunto. |
| adjuntos | array | No | Un solo archivo adjunto codificado en Base64 |

> **Nota:** `telefono_id` ya no es un campo válido en el body. El teléfono emisor está implícito en la API Key.
>
> **Importante:** si el adjunto es audio, no mandes `contenido` en el mismo request.

**Respuesta exitosa (202 Accepted):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440000",
    "estado": "sent"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Respuesta cuando el envío falla (202 Accepted con ok=false):**

```json
{
  "ok": false,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440000",
    "estado": "failed",
    "error": "cliente WhatsApp no conectado para este número"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELD | 400 | Falta destino |
| EMPTY_MESSAGE | 400 | No enviaste texto ni adjunto |
| TELEFONO_NOT_FOUND | 404 | Teléfono de la API Key no encontrado |
| SESSION_NOT_ACTIVE | 400 | El teléfono no está activo (sin sesión WhatsApp) |
| INVALID_JSON | 400 | JSON inválido |
| VALIDATION_ERROR | 400 | El adjunto/texto no cumple una regla de negocio |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |
| INTERNAL_ERROR | 500 | Error al registrar el mensaje en la base de datos |

---

### 4. GET /api/difusiones

**Propósito:** Listar difusiones (broadcasts) enviadas por el teléfono de esta API Key.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "broadcasts": [
      {
        "reference_id": "550e8400-e29b-41d4-a716-446655440001",
        "telefono_id": 5,
        "total": 100,
        "adjuntos": [
          {
            "nombre": "catalogo.pdf",
            "sha256_hash": "abcd1234",
            "tamano_bytes": 300000
          }
        ],
        "status": "completed",
        "created_at": "2026-04-18T09:00:00Z"
      }
    ],
    "total": 1
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Notas:**
- Solo retorna difusiones del teléfono asociado a la API Key.
- Si la difusion incluye adjuntos, tambien se devuelven en `adjuntos`.

---

### 5. POST /api/difusiones

**Propósito:** Crear una difusión masiva a múltiples destinos. Los mensajes son enviados de forma asíncrona via WhatsApp.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Body (JSON):**

```json
{
  "destinos": ["519888888888", "519777777777", "519666666666"],
  "mensaje": "Esta es una difusión masiva",
  "adjuntos": [
    {
      "nombre": "catalogo.pdf",
      "mime_type": "application/pdf",
      "contenido_base64": "...base64..."
    }
  ]

}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| destinos | array | Sí | Lista de números destino con código de país (sin '+') |
| mensaje | string | Condicional | Contenido del mensaje. Obligatorio si no envias adjunto. |
| adjuntos | array | No | Un solo archivo para toda la difusión |

> **Nota:** `telefono_id` ya no es un campo válido en el body. El teléfono emisor está implícito en la API Key.
>
> **Importante:** si el adjunto es audio, la difusión no debe llevar `mensaje`.

**Respuesta exitosa (202 Accepted):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440001",
    "total": 3,
    "estado": "pending"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Notas:**
- El envío es **asíncrono**: la respuesta 202 confirma que la difusión fue registrada.
- El estado evoluciona a `completed` o `failed` al terminar el procesamiento.
- Consulta `GET /api/difusiones/{reference_id}` para conocer el estado final.

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELDS | 400 | Faltan destinos |
| EMPTY_MESSAGE | 400 | No enviaste texto ni adjunto válido |
| TELEFONO_NOT_FOUND | 404 | Teléfono de la API Key no encontrado |
| SESSION_NOT_ACTIVE | 400 | El teléfono no está activo |
| INVALID_JSON | 400 | JSON inválido |
| VALIDATION_ERROR | 400 | El adjunto/texto no cumple una regla de negocio |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |

---

### 6. GET /api/difusiones/{reference_id}

**Propósito:** Consultar el estado y resultados de una difusión específica.

**Autenticación:** Requiere API Key.

**URL Parameters:**

| Parámetro | Descripción |
|-----------|-------------|
| reference_id | UUID de la difusión retornado por POST /api/difusiones |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440001",
    "empresa_id": 10,
    "telefono_id": 5,
    "total": 3,
    "adjuntos": [
      {
        "nombre": "catalogo.pdf",
        "sha256_hash": "abcd1234",
        "tamano_bytes": 300000
      }
    ],
    "status": "completed",
    "results": [
      {"index": 0, "destino": "519888888888", "state": "sent", "timestamp": "2026-04-18T09:00:01Z"},
      {"index": 1, "destino": "519777777777", "state": "sent", "timestamp": "2026-04-18T09:00:02Z"},
      {"index": 2, "destino": "519666666666", "state": "failed", "error": "número inválido", "timestamp": "2026-04-18T09:00:03Z"}
    ],
    "created_at": "2026-04-18T09:00:00Z"
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_BROADCAST_ID | 400 | No se proporcionó reference_id |
| BROADCAST_NOT_FOUND | 404 | Difusión no encontrada |
| FORBIDDEN | 403 | La difusión no pertenece a esta API Key |

---

## Uso de API Keys por Teléfono

### Características del token_por_numero

1. **Scoped por teléfono:** La API Key está vinculada a un teléfono específico.
2. **Sin selección de teléfono en body:** `telefono_id` no se acepta en el body de ningún endpoint — el teléfono emisor está implícito en la API Key.
3. **Restricción de acceso:** Solo puede acceder a datos del teléfono asociado.
4. **Envío real:** Los mensajes son enviados efectivamente via WhatsApp usando la sesión activa del teléfono.
5. **Adjuntos soportados:** imágenes, PDF, audio y documentos, con las reglas de tamaño y formato descritas arriba.

### Ejemplo de uso

```bash
# Obtener información de la API Key
curl -X GET https://api.tuservidor.com/api/me \
  -H "X-API-Key: wapi_abc123DEF456"

# Enviar mensaje (sin telefono_id — implícito en la API Key)
curl -X POST https://api.tuservidor.com/api/mensajes \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destino": "519888888888", "contenido": "Hola mundo", "adjuntos": [{"nombre":"foto.jpg","mime_type":"image/jpeg","contenido_base64":"..."}]}'

# Crear difusión (sin telefono_id — implícito en la API Key)
curl -X POST https://api.tuservidor.com/api/difusiones \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destinos": ["519888888888", "519777777777"], "mensaje": "Difusión", "adjuntos": [{"nombre":"catalogo.pdf","mime_type":"application/pdf","contenido_base64":"..."}]}'

# Consultar estado de difusión
curl -X GET https://api.tuservidor.com/api/difusiones/550e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: wapi_abc123DEF456"
```

---

## Reintentar Mensaje (Retry)

### API Key: POST /api/mensajes/{reference_id}/reintentar

**Propósito:** Reintentar el envío de un mensaje que falló o está pendiente.

**Autenticación:** Requiere API Key.

**Path:** `POST /api/mensajes/{reference_id}/reintentar`

**Request:** No requiere body.

**Nota importante:** si el mensaje tiene `adjuntos`, este reintento no aplica todavia. En ese caso, hay que crear un envio nuevo.

**Respuesta exitosa (202):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "msg-123-uuid",
    "estado": "sent"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Respuesta fallida (400/404/500):**

```json
{
  "ok": false,
  "data": {
    "reference_id": "msg-123-uuid",
    "estado": "failed",
    "error": "cliente WhatsApp no conectado"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Errores posibles:**

| Código | HTTP | Causa |
|--------|------|-------|
| MISSING_REFERENCE_ID | 400 | Falta el reference_id en el path |
| MESSAGE_NOT_FOUND | 404 | No existe mensaje con ese reference_id |
| FORBIDDEN | 403 | El mensaje no pertenece al teléfono de la API Key |
| INVALID_STATE | 400 | El mensaje ya fue enviado (sent/delivered) |
| SESSION_NOT_ACTIVE | 400 | La sesión de WhatsPhone no está activa |
| RETRY_ERROR | 500 | Error al reintentar el envío |

---

## Editar Mensaje (PATCH)

### API Key: PATCH /api/mensajes/{reference_id}

**Propósito:** Editar el contenido de un mensaje que aún no ha sido enviado.

**Autenticación:** Requiere API Key.

**Path:** `PATCH /api/mensajes/{reference_id}`

**Request:**

```json
{
  "contenido": "Nuevo contenido del mensaje"
}
```

**Nota:** Solo se puede editar si el estado es `pending`, `failed` o `rejected`.
**Nota adicional:** este endpoint solo cambia el texto; no reemplaza ni agrega adjuntos.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "msg-123-uuid",
    "message": "Mensaje actualizado"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

---

## Admin: Reintentar Mensaje

### Admin: POST /api/admin/mensajes/{reference_id}/reintentar

**Propósito:** Reintentar el envío de un mensaje desde el panel de administración.

**Autenticación:** Requiere JWT de admin (`Authorization: Bearer <token>`).

**Path:** `POST /api/admin/mensajes/{reference_id}/reintentar`

**Request:** No requiere body.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "reference_id": "msg-123-uuid",
  "estado": "sent"
}
```

**Errores posibles:**

| Código | HTTP | Causa |
|--------|------|-------|
| missing reference_id | 400 | Falta el reference_id en el path |
| message not found | 404 | No existe mensaje con ese reference_id |
| forbidden | 403 | No tienes acceso a este mensaje |
| message already sent | 400 | El mensaje ya fue enviado |
| session not active | 400 | La sesión de WhatsApp no está activa |

---

## Códigos de Error Comunes

| Código | HTTP | Causa |
|--------|------|-------|
| API_KEY_REQUIRED | 401 | No se envió header de API Key |
| INVALID_API_KEY | 401 | La key no existe o está inactiva/expirada |
| TELEFONO_NOT_FOUND | 401/404 | El teléfono asociado no existe |
| FORBIDDEN | 403 | Teléfono no coincide con la API Key |
| EMPRESA_INACTIVE | 403 | La empresa está deshabilitada |
| MISSING_FIELDS | 400 | Faltan campos requeridos en el request |
| INVALID_JSON | 400 | El body no es JSON válido |
| SESSION_NOT_ACTIVE | 400 | El teléfono no tiene sesión WhatsApp activa |
| MEDIA_RETRY_UNSUPPORTED | 400 | El mensaje tiene adjuntos y no se puede reintentar todavía |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |
| INTERNAL_ERROR | 500 | Error interno al persistir el mensaje |
# Documentación de API Key por Teléfono

## Overview

Este documento explica como usar una **API Key por teléfono** (`api_key_telefono` o `token_por_numero`).

La idea es simple:
- la key identifica a una empresa y a un telefono especifico
- no hace falta mandar `telefono_id` en el body de los endpoints de consumo
- se puede enviar texto, imagenes y otros adjuntos permitidos
- el mismo token sirve para mensajes, difusiones, consulta y seguimiento

## Resumen rapido

- Header principal: `X-API-Key`
- Tambien funciona `Authorization: Bearer <token>` o `Authorization: ApiKey <token>`
- Un solo adjunto por mensaje o difusion
- Tamano maximo por archivo: 5 MB
- Tamano maximo total por request: 20 MB
- Broadcast maximo: 500 destinos
- Audio no lleva caption ni texto junto al adjunto

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

## Adjuntos permitidos

Cuando el mensaje o la difusion necesita una imagen, PDF, audio o documento, usa `adjuntos`.

### Formato del adjunto

```json
{
  "nombre": "foto.jpg",
  "mime_type": "image/jpeg",
  "contenido_base64": "...base64..."
}
```

### Tipos permitidos

- `image/jpeg`
- `image/png`
- `video/mp4`
- `audio/mpeg`
- `audio/mp4`
- `audio/ogg`
- `audio/wav`
- `application/pdf`
- `application/msword`
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`

### Reglas simples

- Solo se permite 1 adjunto por request
- El archivo debe coincidir con su extension y su MIME type
- Audio no puede llevar `mensaje` al mismo tiempo
- Si mandas adjunto, el contenido debe ir en Base64

## Endpoints Disponibles

### 1. GET /api/me

**Propósito:** Obtener información de la API Key, empresa y teléfono asociados.

**Compatibilidad:** También disponible como `GET /api/v1/me`.

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

### 1.1 GET /api/sesion

**Propósito:** Diagnosticar estado de sesión del teléfono de la API key comparando DB vs runtime real.

**Autenticación:** Requiere API Key.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "telefono_id": 5,
    "account_id": "51999999999",
    "status_db": "active",
    "status_runtime": "connected",
    "runtime_connected": true,
    "mismatch": false,
    "mismatch_reason": "",
    "recommended_action": "none"
  }
}
```

**Notas:**
- `mismatch=true` indica inconsistencia entre DB y cliente runtime.
- `recommended_action` puede ser `none`, `reanudar_conexion` o `iniciar_conexion`.

---

### 2. GET /api/mensajes

**Propósito:** Listar mensajes enviados por el teléfono de esta API Key.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Query Parameters:**

| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| limit | int | (Opcional) Límite de resultados (default 50, max 100) |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "messages": [
      {
        "reference_id": "550e8400-e29b-41d4-a716-446655440000",
        "telefono_id": 5,
        "destino": "519888888888",
        "contenido": "Hola, este es un mensaje de prueba",
        "adjuntos": [
          {
            "nombre": "foto.jpg",
            "sha256_hash": "abcd1234",
            "tamano_bytes": 245000
          }
        ],
        "estado": "sent",
        "tiempo": "2026-04-18T10:30:00Z"
      }
    ],
    "total": 1
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Notas:**
- Solo retorna mensajes del teléfono asociado a la API Key.
- El campo `telefono_id` ya no es un parámetro aceptado en el body — está implícito en la API Key.
- Si el mensaje tuvo adjuntos, también se devuelven en `adjuntos`.

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| API_KEY_REQUIRED | 401 | No se proporcionó API Key |
| INVALID_TELEFONO_ID | 400 | telefono_id en query string inválido |

---

### 3. POST /api/mensajes

**Propósito:** Enviar un mensaje directo a un número WhatsApp. El mensaje es enviado en tiempo real via WhatsApp.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Body (JSON):**

```json
{
  "destino": "519888888888",
  "contenido": "Hola, este es un mensaje de prueba",
  "adjuntos": [
    {
      "nombre": "foto.jpg",
      "mime_type": "image/jpeg",
      "contenido_base64": "...base64..."
    }
  ]
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| destino | string | Sí | Número de destino con código de país (sin '+') |
| contenido | string | Condicional | Texto del mensaje. Obligatorio si no envias adjunto. |
| adjuntos | array | No | Un solo archivo adjunto codificado en Base64 |

> **Nota:** `telefono_id` ya no es un campo válido en el body. El teléfono emisor está implícito en la API Key.
>
> **Importante:** si el adjunto es audio, no mandes `contenido` en el mismo request.

**Respuesta exitosa (202 Accepted):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440000",
    "estado": "sent"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Respuesta cuando el envío falla (202 Accepted con ok=false):**

```json
{
  "ok": false,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440000",
    "estado": "failed",
    "error": "cliente WhatsApp no conectado para este número"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELD | 400 | Falta destino |
| EMPTY_MESSAGE | 400 | No enviaste texto ni adjunto |
| TELEFONO_NOT_FOUND | 404 | Teléfono de la API Key no encontrado |
| SESSION_NOT_ACTIVE | 400 | El teléfono no está activo (sin sesión WhatsApp) |
| INVALID_JSON | 400 | JSON inválido |
| VALIDATION_ERROR | 400 | El adjunto/texto no cumple una regla de negocio |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |
| INTERNAL_ERROR | 500 | Error al registrar el mensaje en la base de datos |

---

### 4. GET /api/difusiones

**Propósito:** Listar difusiones (broadcasts) enviadas por el teléfono de esta API Key.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "broadcasts": [
      {
        "reference_id": "550e8400-e29b-41d4-a716-446655440001",
        "telefono_id": 5,
        "total": 100,
        "adjuntos": [
          {
            "nombre": "catalogo.pdf",
            "sha256_hash": "abcd1234",
            "tamano_bytes": 300000
          }
        ],
        "status": "completed",
        "created_at": "2026-04-18T09:00:00Z"
      }
    ],
    "total": 1
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Notas:**
- Solo retorna difusiones del teléfono asociado a la API Key.
- Si la difusion incluye adjuntos, tambien se devuelven en `adjuntos`.

---

### 5. POST /api/difusiones

**Propósito:** Crear una difusión masiva a múltiples destinos. Los mensajes son enviados de forma asíncrona via WhatsApp.

**Autenticación:** Requiere API Key. **Solo acepta `api_token`** — no acepta JWT de empresa.

**Body (JSON):**

```json
{
  "destinos": ["519888888888", "519777777777", "519666666666"],
  "mensaje": "Esta es una difusión masiva",
  "adjuntos": [
    {
      "nombre": "catalogo.pdf",
      "mime_type": "application/pdf",
      "contenido_base64": "...base64..."
    }
  ]

}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| destinos | array | Sí | Lista de números destino con código de país (sin '+') |
| mensaje | string | Condicional | Contenido del mensaje. Obligatorio si no envias adjunto. |
| adjuntos | array | No | Un solo archivo para toda la difusión |

> **Nota:** `telefono_id` ya no es un campo válido en el body. El teléfono emisor está implícito en la API Key.
>
> **Importante:** si el adjunto es audio, la difusión no debe llevar `mensaje`.

**Respuesta exitosa (202 Accepted):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440001",
    "total": 3,
    "estado": "pending"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Notas:**
- El envío es **asíncrono**: la respuesta 202 confirma que la difusión fue registrada.
- El estado evoluciona a `completed` o `failed` al terminar el procesamiento.
- Consulta `GET /api/difusiones/{reference_id}` para conocer el estado final.

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_FIELDS | 400 | Faltan destinos |
| EMPTY_MESSAGE | 400 | No enviaste texto ni adjunto válido |
| TELEFONO_NOT_FOUND | 404 | Teléfono de la API Key no encontrado |
| SESSION_NOT_ACTIVE | 400 | El teléfono no está activo |
| INVALID_JSON | 400 | JSON inválido |
| VALIDATION_ERROR | 400 | El adjunto/texto no cumple una regla de negocio |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |

---

### 6. GET /api/difusiones/{reference_id}

**Propósito:** Consultar el estado y resultados de una difusión específica.

**Autenticación:** Requiere API Key.

**URL Parameters:**

| Parámetro | Descripción |
|-----------|-------------|
| reference_id | UUID de la difusión retornado por POST /api/difusiones |

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "550e8400-e29b-41d4-a716-446655440001",
    "empresa_id": 10,
    "telefono_id": 5,
    "total": 3,
    "adjuntos": [
      {
        "nombre": "catalogo.pdf",
        "sha256_hash": "abcd1234",
        "tamano_bytes": 300000
      }
    ],
    "status": "completed",
    "results": [
      {"index": 0, "destino": "519888888888", "state": "sent", "timestamp": "2026-04-18T09:00:01Z"},
      {"index": 1, "destino": "519777777777", "state": "sent", "timestamp": "2026-04-18T09:00:02Z"},
      {"index": 2, "destino": "519666666666", "state": "failed", "error": "número inválido", "timestamp": "2026-04-18T09:00:03Z"}
    ],
    "created_at": "2026-04-18T09:00:00Z"
  },
  "meta": {
    "empresa_id": 10,
    "timestamp": "2026-04-18T10:31:00Z"
  }
}
```

**Errores posibles:**

| Código | HTTP | Descripción |
|--------|------|-------------|
| MISSING_BROADCAST_ID | 400 | No se proporcionó reference_id |
| BROADCAST_NOT_FOUND | 404 | Difusión no encontrada |
| FORBIDDEN | 403 | La difusión no pertenece a esta API Key |

---

## Uso de API Keys por Teléfono

### Características del token_por_numero

1. **Scoped por teléfono:** La API Key está vinculada a un teléfono específico.
2. **Sin selección de teléfono en body:** `telefono_id` no se acepta en el body de ningún endpoint — el teléfono emisor está implícito en la API Key.
3. **Restricción de acceso:** Solo puede acceder a datos del teléfono asociado.
4. **Envío real:** Los mensajes son enviados efectivamente via WhatsApp usando la sesión activa del teléfono.
5. **Adjuntos soportados:** imágenes, PDF, audio y documentos, con las reglas de tamaño y formato descritas arriba.

### Ejemplo de uso

```bash
# Obtener información de la API Key
curl -X GET https://api.tuservidor.com/api/me \
  -H "X-API-Key: wapi_abc123DEF456"

# Enviar mensaje (sin telefono_id — implícito en la API Key)
curl -X POST https://api.tuservidor.com/api/mensajes \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destino": "519888888888", "contenido": "Hola mundo", "adjuntos": [{"nombre":"foto.jpg","mime_type":"image/jpeg","contenido_base64":"..."}]}'

# Crear difusión (sin telefono_id — implícito en la API Key)
curl -X POST https://api.tuservidor.com/api/difusiones \
  -H "X-API-Key: wapi_abc123DEF456" \
  -H "Content-Type: application/json" \
  -d '{"destinos": ["519888888888", "519777777777"], "mensaje": "Difusión", "adjuntos": [{"nombre":"catalogo.pdf","mime_type":"application/pdf","contenido_base64":"..."}]}'

# Consultar estado de difusión
curl -X GET https://api.tuservidor.com/api/difusiones/550e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: wapi_abc123DEF456"
```

---

## Reintentar Mensaje (Retry)

### API Key: POST /api/mensajes/{reference_id}/reintentar

**Propósito:** Reintentar el envío de un mensaje que falló o está pendiente.

**Autenticación:** Requiere API Key.

**Path:** `POST /api/mensajes/{reference_id}/reintentar`

**Request:** No requiere body.

**Nota importante:** si el mensaje tiene `adjuntos`, este reintento no aplica todavia. En ese caso, hay que crear un envio nuevo.

**Respuesta exitosa (202):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "msg-123-uuid",
    "estado": "sent"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Respuesta fallida (400/404/500):**

```json
{
  "ok": false,
  "data": {
    "reference_id": "msg-123-uuid",
    "estado": "failed",
    "error": "cliente WhatsApp no conectado"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

**Errores posibles:**

| Código | HTTP | Causa |
|--------|------|-------|
| MISSING_REFERENCE_ID | 400 | Falta el reference_id en el path |
| MESSAGE_NOT_FOUND | 404 | No existe mensaje con ese reference_id |
| FORBIDDEN | 403 | El mensaje no pertenece al teléfono de la API Key |
| INVALID_STATE | 400 | El mensaje ya fue enviado (sent/delivered) |
| SESSION_NOT_ACTIVE | 400 | La sesión de WhatsPhone no está activa |
| RETRY_ERROR | 500 | Error al reintentar el envío |

---

## Editar Mensaje (PATCH)

### API Key: PATCH /api/mensajes/{reference_id}

**Propósito:** Editar el contenido de un mensaje que aún no ha sido enviado.

**Autenticación:** Requiere API Key.

**Path:** `PATCH /api/mensajes/{reference_id}`

**Request:**

```json
{
  "contenido": "Nuevo contenido del mensaje"
}
```

**Nota:** Solo se puede editar si el estado es `pending`, `failed` o `rejected`.
**Nota adicional:** este endpoint solo cambia el texto; no reemplaza ni agrega adjuntos.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "data": {
    "reference_id": "msg-123-uuid",
    "message": "Mensaje actualizado"
  },
  "meta": {
    "empresa_id": 10
  }
}
```

---

## Admin: Reintentar Mensaje

### Admin: POST /api/admin/mensajes/{reference_id}/reintentar

**Propósito:** Reintentar el envío de un mensaje desde el panel de administración.

**Autenticación:** Requiere JWT de admin (`Authorization: Bearer <token>`).

**Path:** `POST /api/admin/mensajes/{reference_id}/reintentar`

**Request:** No requiere body.

**Respuesta exitosa (200):**

```json
{
  "ok": true,
  "reference_id": "msg-123-uuid",
  "estado": "sent"
}
```

**Errores posibles:**

| Código | HTTP | Causa |
|--------|------|-------|
| missing reference_id | 400 | Falta el reference_id en el path |
| message not found | 404 | No existe mensaje con ese reference_id |
| forbidden | 403 | No tienes acceso a este mensaje |
| message already sent | 400 | El mensaje ya fue enviado |
| session not active | 400 | La sesión de WhatsApp no está activa |

---

## Códigos de Error Comunes

| Código | HTTP | Causa |
|--------|------|-------|
| API_KEY_REQUIRED | 401 | No se envió header de API Key |
| INVALID_API_KEY | 401 | La key no existe o está inactiva/expirada |
| TELEFONO_NOT_FOUND | 401/404 | El teléfono asociado no existe |
| FORBIDDEN | 403 | Teléfono no coincide con la API Key |
| EMPRESA_INACTIVE | 403 | La empresa está deshabilitada |
| MISSING_FIELDS | 400 | Faltan campos requeridos en el request |
| INVALID_JSON | 400 | El body no es JSON válido |
| SESSION_NOT_ACTIVE | 400 | El teléfono no tiene sesión WhatsApp activa |
| MEDIA_RETRY_UNSUPPORTED | 400 | El mensaje tiene adjuntos y no se puede reintentar todavía |
| ATTACHMENT_COUNT_EXCEEDED | 400 | Se envió más de un adjunto |
| ATTACHMENT_SIZE_EXCEEDED | 400 | El archivo supera el tamaño permitido |
| ATTACHMENT_TYPE_NOT_ALLOWED | 400 | El tipo de archivo no está permitido |
| INVALID_ATTACHMENT_FORMAT | 400 | El adjunto no tiene un formato válido |
| INVALID_ATTACHMENT_NAME | 400 | El nombre del archivo no es válido |
| INTERNAL_ERROR | 500 | Error interno al persistir el mensaje |
