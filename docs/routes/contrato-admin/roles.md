# Roles y Módulos — Gestión Admin

CRUD de roles de acceso y listado de módulos disponibles.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/roles

Lista todos los roles disponibles.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/roles \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "roles": [
    {
      "id": 1,
      "nombre": "super_admin",
      "descripcion": "Acceso total al sistema",
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "nombre": "admin",
      "descripcion": "Administrador estándar",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

## GET /api/admin/roles/{id}

Retorna el detalle de un rol por ID.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/roles/1 \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "role": {
    "id": 1,
    "nombre": "super_admin",
    "descripcion": "Acceso total al sistema"
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Rol no encontrado. |

---

## POST /api/admin/roles

Crea un nuevo rol.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `nombre` | string | Sí | Nombre único del rol. |
| `descripcion` | string | No | Descripción del rol. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/admin/roles \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "soporte", "descripcion": "Acceso de solo lectura"}'
```

**Respuesta exitosa `201 Created`:**
```json
{
  "ok": true,
  "role": { "id": 3, "nombre": "soporte", "descripcion": "Acceso de solo lectura" }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o `nombre` ausente. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `409 Conflict` | Ya existe un rol con ese nombre. |
| `500 Internal Server Error` | Error al crear. |

---

## PUT /api/admin/roles/{id}

Actualiza un rol existente.

**Request:**
```bash
curl -X PUT https://tu-dominio.com/api/admin/roles/3 \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"descripcion": "Acceso de lectura a reportes"}'
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "role": { "id": 3, "nombre": "soporte", "..." } }
```

---

## DELETE /api/admin/roles/{id}

Elimina un rol. No se puede eliminar si está asignado a usuarios activos.

**Request:**
```bash
curl -X DELETE https://tu-dominio.com/api/admin/roles/3 \
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
| `404 Not Found` | Rol no encontrado. |
| `409 Conflict` | El rol está en uso por usuarios activos. |
| `500 Internal Server Error` | Error al eliminar. |

---

## GET /api/admin/modules

Lista todos los módulos de acceso disponibles para asignar a usuarios.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/modules \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "modules": [
    { "id": 1, "nombre": "mensajes", "descripcion": "Gestión de mensajes" },
    { "id": 2, "nombre": "empresas", "descripcion": "Gestión de empresas" },
    { "id": 3, "nombre": "reportes", "descripcion": "Acceso a reportes" }
  ]
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
