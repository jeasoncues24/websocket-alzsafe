# Integración B2B — wsapi

Este documento describe las capacidades de wsapi diseñadas para integradores externos (SaaS, plataformas serverless, dashboards de terceros). Cualquier integrador puede usar estas APIs, no solo proyectos específicos.

---

## Health Endpoint

Permite validar que una URL apunta efectivamente a un wsapi vivo antes de solicitar credenciales al usuario final.

### Especificación

| Campo | Valor |
|-------|-------|
| Método | `GET` |
| URL | `/api/service/v1/health` |
| Autenticación | Ninguna (endpoint público) |
| Rate limit | 60 requests/min por IP (configurable con `HEALTH_RATE_LIMIT_PER_MIN`) |

### Ejemplo de request

```bash
curl -i https://tu-wsapi.example.com/api/service/v1/health
```

### Ejemplo de respuesta exitosa (200)

```json
{
  "ok": true,
  "service": "wsapi",
  "version": "v1.2.3",
  "timestamp": "2026-05-17T14:23:01Z"
}
```

### Campos de la respuesta

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `ok` | boolean | `true` si el servicio está operativo |
| `service` | string | Siempre `"wsapi"` — identifica la plataforma |
| `version` | string | Versión del binario (configurable con `APP_VERSION`, default `"dev"`) |
| `timestamp` | string | Hora actual del servidor en formato RFC3339 UTC |

### Respuesta en caso de rate limit (429)

```json
{
  "ok": false,
  "error": "RATE_LIMITED",
  "message": "Demasiados requests"
}
```

Header adicional: `Retry-After: 60`

### Política de rate limit

- Ventana fija de 60 segundos por IP de origen.
- Por defecto: 60 requests por ventana. Configurable con la variable de entorno `HEALTH_RATE_LIMIT_PER_MIN`.
- Si el servidor está detrás de un reverse proxy, se respeta el header `X-Forwarded-For`.

### Configuración de versión

La versión reportada se toma de la variable de entorno `APP_VERSION`. Si no está definida, el valor es `"dev"`.

```bash
APP_VERSION=v1.2.3 ./wsapi
```

Para inyectarla en build time con Docker:

```dockerfile
ARG VERSION=dev
RUN go build -ldflags "-X 'wsapi/internal/config.versionLDFlag=${VERSION}'" -o wsapi .
```

> **Nota**: Este endpoint es un *liveness probe* de identidad — confirma que el proceso responde y se identifica como wsapi. No verifica el estado de la base de datos, sesiones de WhatsApp ni otros subsistemas.
