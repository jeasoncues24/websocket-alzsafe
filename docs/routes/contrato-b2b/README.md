# Contrato B2B — Referencia de API

Endpoints disponibles para integraciones B2B con wsapi.

Base URL: `https://tu-dominio.com`

---

## Índice

| Archivo | Endpoints cubiertos |
|---------|---------------------|
| [health.md](health.md) | `GET /api/service/v1/health` |
| [me.md](me.md) | `GET /api/service/v1/me`, `GET /api/service/v1/sesion` |
| [mensajes.md](mensajes.md) | `GET/POST /api/service/v1/mensajes`, `GET/PATCH/POST /api/service/v1/mensajes/{id}` |
| [difusiones.md](difusiones.md) | `GET/POST /api/service/v1/difusiones`, `GET /api/service/v1/difusiones/{id}` |
| [webhooks.md](webhooks.md) | `POST/GET /api/service/v1/webhooks`, `DELETE /api/service/v1/webhooks/{id}` |
| [websocket.md](websocket.md) | `GET /api/service/v1/ws` |

---

## Autenticación

Todas las rutas (a excepción de `/health` y `/ws` que usa token QR-link temporal) requieren autenticación por **API Key**:

Se puede enviar en las cabeceras de dos formas:

1. Cabecera `X-API-Key`:
```http
X-API-Key: wsk_abc.secretoCompleto...
```

2. Cabecera `Authorization`:
```http
Authorization: Bearer wsk_abc.secretoCompleto...
```

Las API Keys se configuran y generan para cada número de teléfono específico desde el panel de administración.
