# WebSocket — Eventos en Tiempo Real

Conexión WebSocket para recibir eventos de sesión WhatsApp en tiempo real (QR, conexión, desconexión, mensajes).

---

## GET /api/service/v1/ws

Abre una conexión WebSocket autenticada. Soporta dos modos según el tipo de token.

**Condición:** Requiere JWT de empresa (token de larga duración) **o** token provisional QR-link. Se pasa como query param o como `Authorization: Bearer`.

**Query params:**

| Param | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `token` | string | Sí* | JWT de empresa o token QR-link. Alternativa al header `Authorization`. |

*Si no viene en query param, se acepta vía `Authorization: Bearer <token>`.

**Request (conexión):**
```bash
# Usando query param
wscat -c "wss://tu-dominio.com/api/service/v1/ws?token=$EMPRESA_JWT"

# Usando curl (upgrade a WS no disponible con curl estándar)
# Usar un cliente WebSocket como wscat, websocat o desde el frontend
```

### Flujo — JWT de empresa regular

Tras conectar, el cliente **debe** enviar el mensaje `subscribe` para indicar a qué teléfono suscribirse:

```json
{
  "type": "subscribe",
  "data": {
    "phone_id": 5
  }
}
```

Si el `phone_id` no pertenece a la empresa, el servidor cierra la conexión con un evento `error`.

### Flujo — Token QR-link (`scope: "qr_link"`)

No se requiere mensaje `subscribe`. El servidor se suscribe automáticamente al teléfono incluido en el token.

---

### Eventos recibidos del servidor

| `type` | Descripción |
|--------|-------------|
| `qr` | QR disponible. `data` contiene el string QR para mostrar al usuario. |
| `connected` | El teléfono se conectó exitosamente a WhatsApp. |
| `disconnected` | El teléfono se desconectó. |
| `ping` | Heartbeat enviado cada 25 segundos para mantener la conexión viva. |
| `error` | Error de autenticación, suscripción inválida u otro fallo. `data.message` describe el error. |

**Ejemplo de evento QR:**
```json
{
  "type": "qr",
  "data": {
    "qrString": "2@ABC123...",
    "isQR": true
  }
}
```

**Ejemplo de evento connected:**
```json
{
  "type": "connected",
  "data": {
    "isActive": true
  }
}
```

**Ejemplo de ping:**
```json
{
  "type": "ping"
}
```

**Ejemplo de error:**
```json
{
  "type": "error",
  "data": {
    "message": "forbidden"
  }
}
```

---

### Errores de conexión (HTTP antes de upgrade)

| Código | Causa |
|--------|-------|
| `401 Unauthorized` — `TOKEN_REQUIRED` | Token ausente. |
| `401 Unauthorized` — `INVALID_TOKEN` | Token inválido o expirado. |
