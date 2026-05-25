# Empresas — Gestión Admin

CRUD completo de empresas y gestión de sus tokens JWT de largo plazo.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`. Los endpoints de generación y revocación de token requieren rol `super_admin`.

---

## GET /api/admin/empresas

Lista todas las empresas con paginación y filtros opcionales.

**Query params:**

| Param | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `page` | integer | `1` | Página a retornar. |
| `limit` | integer | `50` | Resultados por página. Máx `100`. |
| `busqueda` | string | — | Búsqueda por nombre o RUC. |
| `estado` | string | — | `"activo"` o `"inactivo"` para filtrar. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/admin/empresas?page=1&limit=20&estado=activo" \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "empresas": [
    {
      "id": 1,
      "ruc": "20123456789",
      "nombre": "Mi Empresa SAC",
      "nombre_comercial": "Mi Empresa",  // "" si no definido
      "telefono": "+51999000111",
      "direccion": null,                 // null si no definida
      "activo": true,
      "token_version": 1,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 42,
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

## GET /api/admin/empresas/{id}

Retorna el detalle de una empresa por ID.

**Path params:**
- `id` — ID numérico de la empresa.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/empresas/1 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "empresa": {
    "id": 1,
    "ruc": "20123456789",
    "nombre": "Mi Empresa SAC",
    "nombre_comercial": "Mi Empresa",
    "telefono": "+51999000111",
    "direccion": null,
    "activo": true,
    "token_version": 1,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido en el path. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | El usuario admin no tiene acceso a esta empresa. |
| `404 Not Found` | Empresa no encontrada. |
| `500 Internal Server Error` | Error al consultar la base de datos. |

---

## POST /api/admin/empresas

Crea una nueva empresa.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `ruc` | string | Sí | RUC único de la empresa. |
| `nombre` | string | Sí | Razón social. |
| `nombre_comercial` | string | No | Nombre comercial. |
| `telefono` | string | No | Teléfono de contacto. |
| `direccion` | string | No | Dirección física. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/empresas \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"ruc": "20123456789", "nombre": "Nueva Empresa SAC"}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "empresa": {
    "id": 5,
    "ruc": "20123456789",
    "nombre": "Nueva Empresa SAC",
    "nombre_comercial": "",
    "telefono": "",
    "direccion": null,
    "activo": true,
    "token_version": 0,
    "created_at": "2026-05-24T12:00:00Z"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o `ruc`/`nombre` ausentes. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `409 Conflict` | Ya existe una empresa con ese RUC. |
| `500 Internal Server Error` | Error al guardar en base de datos. |

---

## PUT /api/admin/empresas/{id}

Actualiza los datos de una empresa.

**Path params:**
- `id` — ID numérico de la empresa.

**Body JSON:** Mismos campos opcionales que `POST`, excepto `ruc` que no aplica aquí.

**Request:**
```bash
curl -X PUT https://tu-dominio.com/api/admin/empresas/1 \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Empresa Actualizada SAC", "telefono": "+51999001122"}'
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "empresa": { "id": 1, "nombre": "Empresa Actualizada SAC", "..." }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado a esta empresa. |
| `404 Not Found` | Empresa no encontrada. |
| `500 Internal Server Error` | Error al actualizar. |

---

## DELETE /api/admin/empresas/{id}

Desactiva lógicamente una empresa (soft delete). No se puede eliminar si tiene sesiones WhatsApp activas o ya está inactiva.

**Path params:**
- `id` — ID numérico de la empresa.

**Request:**
```bash
curl -X DELETE https://tu-dominio.com/api/admin/empresas/1 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true }
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Empresa no encontrada. |
| `409 Conflict` | La empresa tiene sesiones activas o ya está inactiva. |
| `500 Internal Server Error` | Error al desactivar. |

---

## POST /api/admin/empresas/{id}/restore

Reactiva una empresa previamente desactivada.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/empresas/1/restore \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "empresa": { "id": 1, "activo": true, "..." } }
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | Acceso denegado. |
| `404 Not Found` | Empresa no encontrada. |
| `500 Internal Server Error` | Error al reactivar. |

---

## POST /api/admin/empresas/{id}/token

Genera un JWT de larga duración (5 años) para la empresa. La empresa usa este token para autenticarse en el contrato B2B.

**Condición:** Solo `super_admin`.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/empresas/1/token \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "JWT de empresa generado exitosamente. Guárdalo en un lugar seguro."
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | El usuario no tiene rol `super_admin`. |
| `404 Not Found` | Empresa no encontrada. |
| `409 Conflict` | La empresa está inactiva. |
| `500 Internal Server Error` | Error al generar el JWT. |

---

## POST /api/admin/empresas/{id}/token/revoke

Revoca todos los JWT activos de la empresa incrementando su `token_version`. Útil ante una filtración del token.

**Condición:** Solo `super_admin`.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/empresas/1/token/revoke \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "token_version": 2,
  "message": "Todos los JWT de empresa han sido revocados"
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | ID inválido. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `403 Forbidden` | El usuario no tiene rol `super_admin`. |
| `500 Internal Server Error` | Error al incrementar `token_version`. |
