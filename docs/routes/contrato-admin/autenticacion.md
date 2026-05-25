# Autenticación — Panel Admin

Endpoints de login, logout y refresh para usuarios del panel de administración.

---

## POST /api/auth/login

Autentica a un usuario admin y retorna un JWT de sesión.

**Condición:** Público. Sin autenticación previa requerida.

**Body JSON:**

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `username` | string | Sí | Nombre de usuario del admin. |
| `password` | string | Sí | Contraseña del admin. |

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "mi_password"}'
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login exitoso"
}
```

> El token tiene una expiración definida en la configuración del servidor. Usarlo en `Authorization: Bearer <token>` en todas las rutas admin protegidas.

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | JSON inválido o `username`/`password` ausentes. |
| `401 Unauthorized` | Credenciales incorrectas o usuario inactivo. |
| `500 Internal Server Error` | Error interno al consultar usuario o generar token. |

---

## POST /api/auth/logout

Invalida el JWT actual añadiéndolo a la blacklist. La sesión queda inutilizable inmediatamente.

**Condición:** No requiere autenticación válida (el token puede estar próximo a expirar). Si no hay token, simplemente retorna `ok`.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/auth/logout \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "logged_out": true
}
```

---

## POST /api/auth/refresh

Genera un nuevo JWT a partir de uno válido. El token anterior queda invalidado automáticamente.

**Condición:** Requiere JWT de admin activo y no en blacklist en `Authorization: Bearer <ADMIN_JWT>`.

**Request:**
```bash
curl -X POST https://tu-dominio.com/api/auth/refresh \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Token refrescado"
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | Token ausente, inválido, expirado o en blacklist. |
| `500 Internal Server Error` | Error al generar el nuevo token. |

---

## GET /api/auth/me

Retorna el perfil del usuario admin autenticado.

**Condición:** Requiere JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/auth/me \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@empresa.com",
    "role_id": 2,           // null si no tiene rol asignado
    "is_root": false,
    "activo": true
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `404 Not Found` | Usuario no encontrado (raro, desincronización de datos). |
