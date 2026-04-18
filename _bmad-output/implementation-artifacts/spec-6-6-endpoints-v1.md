# S-6.6: Endpoints v1/* + elimina legacy

## Objetivo

Crear handlers HTTP para la API v1/* protegida con JWT empresa, reemplazar endpoints legacy en /api/* con filtros por empresa_id.

## Endpoints a implementar

### Sesiones
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /v1/sessions | Listar sesiones de la empresa |
| POST | /v1/sessions | Crear nueva sesión (generar QR) |
| GET | /v1/sessions/{telefono_id} | Obtener estado de sesión |
| DELETE | /v1/sessions/{telefono_id} | Desconectar sesión |

### Mensajes
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /v1/messages | Listar mensajes (filtrado por empresa) |
| POST | /v1/message | Enviar mensaje |

### Difusiones
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /v1/broadcasts | Listar difusiones |
| POST | /v1/broadcast | Crear difusión |
| GET | /v1/broadcast/{id} | Obtener estado de difusión |

### Métricas
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /v1/metrics | Métricas de la empresa |

### Teléfonos
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /v1/phones | Listar teléfonos registrados |
| POST | /v1/phones/{telefono_id}/qr | Regenerar QR |

## Estructura de directorios

```
internal/http/handlers/
  v1_sessions.go   - handlers para /v1/sessions
  v1_messages.go   - handlers para /v1/messages
  v1_broadcasts.go - handlers para /v1/broadcasts
  v1_metrics.go   - handlers para /v1/metrics
  v1_phones.go     - handlers para /v1/phones
```

## Dependencias

- S-6.0: Schema DB multi-tenant ✅
- S-6.2: Modelo Empresa + Teléfono ✅
- S-6.4: Generación JWT empresa ✅
- S-6.5: Middleware auth JWT + ownership ✅

## Cambios en router.go

1. Registrar handlers v1/* con `empresaAuthMiddleware.RequireEmpresaAuth()`
2. Para endpoints con telefono_id en path, usar adicionalmente `empresaAuthMiddleware.RequireOwnership()`
3. Eliminar o marcar como deprecated los endpoints legacy en /api/*

## Response format

```json
{
  "ok": true,
  "data": { ... },
  "meta": {
    "empresa_id": 123,
    "timestamp": "2026-04-16T12:00:00Z"
  }
}
```

## Errores

```json
{
  "ok": false,
  "error": "CODE",
  "message": "Descripción"
}
```

## Notas

- Todos los endpoints filtran por empresa_id del JWT
- Validation de telefono_id pertenece a la empresa (ownership check)
- Usar los stores existentes: EmpresaStore, TelefonoStore, MessagesRepository