# Métricas y Clientes — Vista Admin

Endpoints de dashboard y búsqueda de clientes para el panel de administración.

**Condición general:** Requieren JWT de admin en `Authorization: Bearer <ADMIN_JWT>`.

---

## GET /api/admin/metricas

Retorna métricas globales del sistema para el dashboard admin: mensajes, empresas activas, teléfonos conectados, etc.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/admin/metricas \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "metrics": {
    "total_empresas": 25,
    "empresas_activas": 22,
    "total_telefonos": 48,
    "telefonos_conectados": 31,
    "mensajes_hoy": 1240,
    "mensajes_total": 95000,
    "mensajes_fallidos_hoy": 12,
    "tasa_exito": 99.03       // porcentaje, 0 si no hay mensajes
  }
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `500 Internal Server Error` | Error al calcular métricas. |

---

## GET /api/admin/difusiones

Lista todas las difusiones del sistema con filtros opcionales. Solo lectura para diagnóstico admin.

**Query params:**

| Param | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `empresa_id` | integer | — | Filtra por empresa. |
| `status` | string | — | `"pending"` \| `"running"` \| `"completed"` \| `"failed"` |
| `page` | integer | `1` | Página. |
| `limit` | integer | `50` | Resultados por página. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/admin/difusiones?status=failed" \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "broadcasts": [
    {
      "reference_id": "bc550e8400-...",
      "empresa_id": 1,
      "telefono_id": 5,
      "total": 100,
      "status": "failed",
      "created_at": "2026-05-24T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 50
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `401 Unauthorized` | JWT ausente o inválido. |
| `500 Internal Server Error` | Error al consultar. |

---

## GET /api/admin/clientes/buscar

Busca contactos de clientes por número de teléfono o nombre dentro del historial de mensajes.

**Query params:**

| Param | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `q` | string | Sí | Término de búsqueda (número o nombre). |
| `empresa_id` | integer | No | Limita la búsqueda a una empresa. |

**Request:**
```bash
curl -X GET "https://tu-dominio.com/api/admin/clientes/buscar?q=51987654321" \
  -H "Authorization: Bearer $ADMIN_JWT"
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "clientes": [
    {
      "numero": "51987654321",
      "nombre": null,                    // null si no se guardó nombre
      "ultimo_mensaje": "2026-05-24T11:00:00Z",
      "empresa_id": 1,
      "telefono_id": 5
    }
  ],
  "total": 1
}
```

**Errores posibles:**

| Código | Causa |
|--------|-------|
| `400 Bad Request` | Parámetro `q` ausente. |
| `401 Unauthorized` | JWT ausente o inválido. |
| `500 Internal Server Error` | Error al buscar. |

---

## GET /metrics

Endpoint de métricas internas en formato Prometheus. Sin autenticación.

**Condición:** Público (restringir a red interna en producción).

**Request:**
```bash
curl -X GET https://tu-dominio.com/metrics
```

**Respuesta:** Texto en formato Prometheus con contadores y gauges del servidor.

---

## GET /health

Healthcheck simple a nivel de servidor (diferente del healthcheck B2B). Retorna `200 OK` si el proceso está corriendo.

**Request:**
```bash
curl -X GET https://tu-dominio.com/health
```

**Respuesta exitosa `200 OK`:**
```json
{ "ok": true, "status": "healthy" }
```
