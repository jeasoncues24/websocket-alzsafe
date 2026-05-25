# Usuarios — Gestión Admin

CRUD de usuarios del panel admin (`/users`) y usuarios admin B2B (`/usuario_admin`).

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/users

Lista todos los usuarios del panel admin con paginación.

**Query params:**

| Param | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `page` | integer | `1` | Página a retornar. |
| `limit` | integer | `20` | Resultados por página. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/admin/users?page=1&limit=20" \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "users": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@empresa.com",
      "role_id": 2,           // null si sin rol
      "role_name": "admin",   // null si sin rol
      "is_root": false,
      "activo": true,
      "last_login": "2026-05-24T10:00:00Z"  // null si nunca inició sesión
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

---

## GET /api/admin/users/{id}

Retorna el detalle de un usuario por ID.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/users/1 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "user": {
    "id": 1, "username": "admin", "email": "admin@empresa.com",
    "role_id": 2, "is_root": false, "activo": true
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Usuario no encontrado. |

---

## POST /api/admin/users

Crea un nuevo usuario admin.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `username` | string | Sí | Nombre de usuario único. |
| `password` | string | Sí | Contraseña (será hasheada con bcrypt). |
| `email` | string | No | Email del usuario. |
| `role_id` | integer | No | ID del rol a asignar. `null` sin rol. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/users \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"username": "operador1", "password": "pass123", "email": "op@empresa.com"}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "user": { "id": 2, "username": "operador1", "activo": true, "..." }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o campos requeridos ausentes. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `409 Conflict` | Ya existe un usuario con ese `username`. |
| `500 Internal Server Error` | Error al crear. |

---

## PUT /api/admin/users/{id}

Actualiza datos de un usuario admin.

**Request:**
```bash
curl -X PUT https://tu-dominio.com/api/admin/users/2 \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"email": "nuevo@empresa.com"}'
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "user": { "id": 2, "..." } }
```

---

## DELETE /api/admin/users/{id}

Desactiva lógicamente un usuario.

**Request:**
```bash
curl -X DELETE https://tu-dominio.com/api/admin/users/2 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true }
```

---

## POST /api/admin/users/{id}/promote

Promueve un usuario a un rol superior.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/users/2/promote \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "user": { "id": 2, "role_name": "super_admin", "..." } }
```

---

## GET /api/admin/users/{id}/modulos

Lista los módulos asignados a un usuario.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/users/2/modulos \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "modules": [
    { "id": 1, "nombre": "mensajes", "descripcion": "Gestión de mensajes" }
  ]
}
```

---

## PUT /api/admin/users/{id}/modulos

Reemplaza completamente los módulos asignados a un usuario.

**Body JSON:**
```json
{ "modulo_ids": [1, 2, 3] }
```

**Request:**
```bash
curl -X PUT https://tu-dominio.com/api/admin/users/2/modulos \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"modulo_ids": [1, 2]}'
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true }
```

---

## Endpoints de UsuarioAdmin (`/usuario_admin`)

Los endpoints bajo `/api/admin/usuario_admin` siguen exactamente la misma estructura y parámetros que `/api/admin/users`, pero gestionan usuarios de tipo B2B (admins de empresa cliente) en lugar de usuarios del panel interno. Se listan a continuación:

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/admin/usuario_admin` | Lista usuarios admin B2B (con `?page` y `?limit`) |
| `GET` | `/api/admin/usuario_admin/{id}` | Detalle de un usuario admin B2B |
| `POST` | `/api/admin/usuario_admin` | Crea usuario admin B2B |
| `PUT` | `/api/admin/usuario_admin/{id}` | Actualiza usuario admin B2B |
| `DELETE` | `/api/admin/usuario_admin/{id}` | Desactiva usuario admin B2B |
| `POST` | `/api/admin/usuario_admin/{id}/promote` | Promueve usuario admin B2B |
| `GET` | `/api/admin/usuario_admin/{id}/modulos` | Lista módulos del usuario admin B2B |
| `PUT` | `/api/admin/usuario_admin/{id}/modulos` | Asigna módulos al usuario admin B2B |
