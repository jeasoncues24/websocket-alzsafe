# Health Check

---

## GET /api/service/v1/health

Verifica que el servicio esté activo y retorna la versión actual. No requiere autenticación. Aplica rate-limit por IP (ventana fija de 60 segundos).

**Condiciones:** Público. Sin autenticación requerida.

**Query params:** Ninguno.

**Request:**
```bash
curl -X GET https://tu-dominio.com/api/service/v1/health
```

**Respuesta exitosa `200 OK`:**
```json
{
  "ok": true,
  "service": "wsapi",
  "version": "1.0.0",
  "timestamp": "2026-05-24T12:00:00Z"
}
```

**Errores posibles:**

| Código | Body | Causa |
|--------|------|-------|
| `429 Too Many Requests` | `{"ok":false,"error":"RATE_LIMITED","message":"Demasiados requests"}` | Se superó el límite de requests por IP en 60 segundos. El header `Retry-After: 60` indica cuándo reintentar. |
