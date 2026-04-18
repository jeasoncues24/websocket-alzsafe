# S-6.7: WebSocket /v1/ws + eventos

## Objetivo

Implementar WebSocket endpoint `/v1/ws` protegido con JWT empresa para tiempo real.

## Endpoint

```
GET /v1/ws
```

## Autenticación

- Header: `Authorization: Bearer <JWT_EMPRESA>`
- Query: `?token=<JWT_EMPRESA>` (fallback)

## Eventos a enviar

| Evento | Payload |
|-------|---------|
| qr | `{"type": "qr", "phone_id": 1, "qr": "string"}` |
| connected | `{"type": "connected", "phone_id": 1}` |
| disconnected | `{"type": "disconnected", "phone_id": 1}` |
| message_status | `{"type": "message_status", "message_id": "abc", "status": "sent|delivered|read|failed"}` |
| service_status | `{"type": "service_status", "status": "up|down"}` |

## Implementación

1. Upgrade HTTP a WebSocket en `/v1/ws`
2. Validar JWT empresa del query param o header
3. Suscribir a eventos del teléfono
4. Broadcast de eventos a cliente

## Librería

Usar `github.com/coder/websocket` (ya integrada en proyecto)