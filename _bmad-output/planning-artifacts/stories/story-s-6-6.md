# Story S-6.6: Endpoints Admin (/api/*)

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P1

## Estado
pending

## Overview

Implementar endpoints de administración para gestionar empresas y sus tokens.

## Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/empresas` | Listar empresas |
| POST | `/api/empresas` | Crear empresa |
| GET | `/api/empresas/{id}` | Ver empresa |
| PUT | `/api/empresas/{id}` | Actualizar empresa |
| DELETE | `/api/empresas/{id}` | Eliminar empresa |
| GET | `/api/empresas/{id}/telefonos` | Ver teléfonos |
| POST | `/api/empresas/{id}/telefonos` | Agregar teléfono |
| POST | `/api/empresas/{id}/token` | Generar nuevo token |
| POST | `/api/empresas/{id}/revoke` | Revocar token |
| GET | `/api/empresas/{id}/metrics` | Métricas de la empresa |

## Ejemplo: Crear Empresa

```json
// POST /api/empresas
{
  "ruc": "20123456789",
  "nombre": "Mi Empresa SAC",
  "permissions": ["send", "broadcast", "sessions"],
  "telefonos": ["+519999999999"]
}
```

## Ejemplo: Revocar Token

```json
// POST /api/empresas/{id}/revoke
{}

Response: {
  "token_version": 2,
  "message": "Token revoked. All previous tokens are now invalid."
}
```

## Dependencias
- S-6.5 (endpoints empresa)

## Estimated Effort
2 days