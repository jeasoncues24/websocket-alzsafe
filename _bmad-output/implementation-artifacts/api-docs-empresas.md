# API de Empresas - Contrato Real

## Overview

No hay aliases de backend. La convención es:

- `/api/admin/*` para el panel administrativo con JWT admin.
- `/api/*` para el contrato de empresa con JWT de empresa.

En esta documentación se cubre el contrato de empresas del panel admin y la base de rutas empresa que ya quedó estandarizada.

## Autenticación

- Todos los endpoints están protegidos con `Authorization: Bearer <token>`.
- El backend no usa cookie `auth_token` para este contrato.
- `super_admin` tiene acceso completo.
- Un usuario no `super_admin` solo ve/modifica su propia empresa si su JWT incluye `empresa_id`.

## Modelo Empresa

```json
{
  "id": 1,
  "ruc": "20100000001",
  "nombre": "Empresa Demo S.A.C.",
  "nombre_comercial": "DemoCorp",
  "telefono": "+51999999999",
  "direccion": "Av. Principal 123",
  "token_version": 1,
  "permissions": [],
  "activo": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Endpoints

### GET `/api/admin/empresas`

Lista empresas.

#### Query params

| Param | Tipo | Default | Descripción |
|---|---|---:|---|
| `page` | int | 1 | Página actual |
| `limit` | int | 50 | Elementos por página (máx. 100) |
| `busqueda` | string | - | Filtro por nombre o RUC |
| `estado` | string | - | `activo` o `inactivo` |

#### Respuesta 200

```json
{
  "ok": true,
  "empresas": [],
  "total": 0,
  "page": 1,
  "limit": 50
}
```

#### Comportamiento de acceso

- `super_admin`: ve el listado completo.
- no `super_admin`: recibe solo su empresa.

### GET `/api/admin/empresas/{id}`

Obtiene una empresa por ID.

#### Respuesta 200

```json
{
  "ok": true,
  "empresa": { ... }
}
```

#### Errores

- `400` si el ID es inválido
- `403` si no tiene acceso a esa empresa
- `404` si no existe

### POST `/api/admin/empresas`

Crea una empresa.

#### Request

```json
{
  "ruc": "20100000001",
  "nombre": "Empresa Demo S.A.C.",
  "nombre_comercial": "DemoCorp",
  "telefono": "+51999999999",
  "direccion": "Av. Principal 123"
}
```

#### Reglas

- `ruc` y `nombre` son requeridos.
- `ruc` debe ser único.

#### Respuesta 201

```json
{
  "ok": true,
  "empresa": { ... }
}
```

### PUT `/api/admin/empresas/{id}`

Actualiza una empresa.

#### Request

```json
{
  "nombre": "Nuevo Nombre",
  "nombre_comercial": "Nuevo Comercial",
  "telefono": "+51988888888",
  "direccion": "Nueva Dirección"
}
```

#### Reglas

- `ruc` no se actualiza en este contrato.
- Si el usuario no es `super_admin`, solo puede actualizar su propia empresa.

### DELETE `/api/admin/empresas/{id}`

Soft delete de una empresa.

#### Reglas

- La empresa debe estar activa.
- Si tiene sesiones WhatsApp activas, el backend responde `409`.
- No existe el alias `companies` en el router actual.

#### Respuesta 200

```json
{ "ok": true }
```

### POST `/api/admin/empresas/{id}/token`

Genera un JWT de empresa de larga duración.

#### Reglas

- Solo `super_admin`.
- La empresa debe estar activa.

#### Respuesta 200

```json
{
  "ok": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Token generado exitosamente. Guárdalo en un lugar seguro."
}
```

### POST `/api/admin/empresas/{id}/token/revoke`

Revoca tokens incrementando `token_version`.

#### Reglas

- Solo `super_admin`.

#### Respuesta 200

```json
{
  "ok": true,
  "token_version": 2,
  "message": "Todos los tokens de empresa han sido revocados"
}
```

### GET `/api/empresas`

Obtiene la empresa autenticada con JWT de empresa.

### PUT `/api/empresas`

Actualiza la empresa autenticada con JWT de empresa.

#### Reglas

- `ruc` es de solo lectura.
- Solo actualiza la empresa del token actual.

## Contrato Empresa `/api/*`

Rutas actuales del panel empresa con JWT de empresa:

- `GET /api/empresas`
- `PUT /api/empresas`
- `POST /api/auth/empresa/validate`
- `GET /api/telefonos`
- `POST /api/telefonos`
- `GET /api/telefonos/{id}`
- `DELETE /api/telefonos/{id}`
- `GET /api/mensajes`
- `POST /api/mensajes`
- `GET /api/difusiones`
- `POST /api/difusiones`
- `GET /api/difusiones/{reference_id}`
- `GET /api/metricas`

## Errores comunes

| Código | Caso |
|---|---|
| `400` | JSON inválido o ID inválido |
| `401` | Token faltante o inválido |
| `403` | Sin permisos para la empresa pedida |
| `404` | Empresa no encontrada |
| `409` | RUC duplicado, empresa inactiva o sesión activa |
| `500` | Error interno |

## Nota operativa

El frontend admin consume este contrato desde `frontend/lib/api.ts`.

Para el flujo de conexión de integraciones externas con API keys por teléfono, ver `spec-8-7-conexion-api-externa.md`.
Para el detalle del envío de mensajes vía API, ver `spec-8-8-envio-mensajes-api.md`.

Documento actualizado: 2026-04-17
