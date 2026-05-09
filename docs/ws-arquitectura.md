# Arquitectura WebSocket — wsapi

Guía de referencia para integradores y desarrolladores que trabajan con el canal WebSocket de wsapi. No se requieren conocimientos de Go.

---

## Tabla de contenidos

1. [Endpoints WS](#1-endpoints-ws)
2. [Autenticación en el endpoint de servicio](#2-autenticación-en-el-endpoint-de-servicio)
3. [Estados de una sesión](#3-estados-de-una-sesión)
4. [Eventos que envía el servidor](#4-eventos-que-envía-el-servidor)
5. [Ciclo de vida de la conexión](#5-ciclo-de-vida-de-la-conexión)
6. [Flujo QR Link — enlace de escaneo sin login](#6-flujo-qr-link--enlace-de-escaneo-sin-login)
7. [Reconexión automática al arrancar el servidor](#7-reconexión-automática-al-arrancar-el-servidor)

---

## 1. Endpoints WS

El sistema expone **dos** endpoints WebSocket con propósitos distintos:

| Endpoint | Ruta | Quién lo usa | Auth requerida |
|----------|------|--------------|----------------|
| **Admin WS** | `GET /api/admin/telefonos/{id}/connect/ws` | Panel administrativo | JWT de administrador (cookie de sesión admin) |
| **Servicio V1 WS** | `GET /api/service/v1/ws` | Integraciones B2B, página `/qr` | JWT de empresa o JWT qr_link |

> El **Admin WS** lo usa exclusivamente el panel web interno. Si estás integrando un sistema externo, usa el **Servicio V1 WS**.

### Cómo conectarse al Servicio V1 WS

El token se puede pasar de dos formas:

```
# Opción A — query string (más fácil desde el navegador)
wss://tudominio.com/api/service/v1/ws?token=TU_JWT

# Opción B — header HTTP (en clientes que lo soporten)
GET /api/service/v1/ws HTTP/1.1
Authorization: Bearer TU_JWT
```

---

## 2. Autenticación en el endpoint de servicio

Existen dos tipos de JWT válidos para el Servicio V1 WS, cada uno con un flujo distinto:

### Flujo A — JWT de empresa (integración B2B)

Usado por sistemas externos que tienen un JWT de empresa permanente.

```
Cliente                          Servidor
  |                                  |
  |--- Conectar WS con token ------->|  (JWT empresa, scope vacío)
  |                                  |
  |--- Enviar mensaje subscribe ----->|
  |    {                             |
  |      "type": "subscribe",        |
  |      "data": { "phone_id": 42 }  |
  |    }                             |
  |                                  |
  |<-- Eventos de la sesión ---------|  (qr, connected, disconnected, ping...)
```

**Reglas:**
- El primer mensaje que envíe el cliente **debe** ser el `subscribe`. Si no llega o tiene formato incorrecto, el servidor cierra la conexión con un evento `error`.
- El `phone_id` debe pertenecer a la empresa del JWT. Si no, el servidor rechaza con `"forbidden"`.

### Flujo B — JWT qr_link (escaneo sin login)

Usado por la página `/qr` cuando un operador recibe un enlace para escanear el QR de WhatsApp.

```
Cliente                          Servidor
  |                                  |
  |--- Conectar WS con token ------->|  (JWT qr_link, scope="qr_link")
  |                                  |
  |<-- El servidor auto-suscribe ----|  (usa el phone_id del propio token)
  |                                  |
  |<-- Eventos de la sesión ---------|  (qr, connected, disconnected, ping...)
```

**Diferencias clave:**
- No se envía ningún mensaje `subscribe` — el servidor lo hace automáticamente.
- El token vence en **10 minutos** (el servidor cerrará la conexión al expirar).
- El `phone_id` viene embebido en el token; el cliente no elige el teléfono.

---

## 3. Estados de una sesión

Cada teléfono tiene un estado que el servidor actualiza conforme avanza el proceso de conexión con WhatsApp:

```
  ┌─────────────────────────────────────────────────────────────────┐
  │                                                                 │
  │  [inicio]                                                       │
  │     │                                                           │
  │     ▼                                                           │
  │  initializing  ←── Se abre la conexión WS / arranca el server  │
  │     │                                                           │
  │     ▼                                                           │
  │  qr_pending    ←── El servidor generó un QR, esperando escaneo │
  │     │                                                           │
  │     ▼                                                           │
  │  active        ←── El operador escaneó el QR con éxito         │
  │     │                                                           │
  │     ▼                                                           │
  │  disconnected  ←── WhatsApp cerró la sesión (cierre, ban, etc) │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘
```

| Estado | Significado |
|--------|-------------|
| `initializing` | El servidor está estableciendo la conexión interna con WhatsApp |
| `qr_pending` | El servidor generó un QR; el operador debe escanearlo |
| `active` | Sesión activa — el teléfono está conectado y puede enviar/recibir mensajes |
| `disconnected` | La sesión se cerró por algún motivo (ver razones en eventos) |

---

## 4. Eventos que envía el servidor

El servidor envía mensajes JSON por el WebSocket. Todos tienen el campo `type` y, cuando aplica, un campo `data`.

### Formato general

```json
{
  "type": "nombre-del-evento",
  "data": { ... }
}
```

### Tabla de eventos

| `type` | Cuándo ocurre | Tiene `data` |
|--------|--------------|--------------|
| `qr` | El servidor generó un nuevo código QR | Sí |
| `connected` | La sesión cambió de estado (activa o no) | Sí |
| `disconnected` | La sesión se desconectó | Sí |
| `ping` | Keepalive cada 25 segundos | No |
| `error` | Error fatal — el servidor cierra la conexión | Sí |

---

### Evento `qr`

El servidor envió un nuevo código QR para escanear.

```json
{
  "type": "qr",
  "data": {
    "qr_string": "2@abc123xyz...",
    "expires_in": 60,
    "message": "Escanee el codigo QR para iniciar sesion."
  }
}
```

- `qr_string`: string que se debe convertir a imagen QR para mostrar al operador.
- `expires_in`: segundos antes de que este QR expire. Cuando expira, el servidor envía un nuevo `qr` automáticamente.

---

### Evento `connected`

Cambio de estado de la sesión. Puede indicar que está activa **o** que se desconectó.

**Sesión activa:**
```json
{
  "type": "connected",
  "data": {
    "isActive": true,
    "message": "Sesion activa"
  }
}
```

**Sesión no activa (requiere nuevo QR):**
```json
{
  "type": "connected",
  "data": {
    "isActive": false,
    "reason": "qr_timeout",
    "requiresNewQR": true
  }
}
```

> **Nota:** cuando `isActive` es `false`, el servidor puede cerrar la conexión WS poco después, dependiendo de la razón.

---

### Evento `disconnected`

La sesión se desconectó de WhatsApp.

```json
{
  "type": "disconnected",
  "data": {
    "isActive": false,
    "reason": "logged_out",
    "requiresNewQR": true
  }
}
```

**Razones posibles de desconexión:**

| `reason` | Causa |
|----------|-------|
| `disconnect` | Desconexión genérica |
| `stream_replaced` | WhatsApp abrió la sesión en otro dispositivo |
| `logged_out` | El número fue desvinculado desde WhatsApp |
| `temporary_ban` | WhatsApp bloqueó temporalmente la cuenta |
| `connect_failure` | Falló el intento de conexión |
| `qr_timeout` | El QR expiró sin ser escaneado |
| `qr_error` | Error al generar el QR |
| `qr_channel_closed` | El canal interno de QR se cerró inesperadamente |
| `connect_timeout` | Tiempo de espera de conexión agotado |
| `connect_error` | Error general de conexión |

---

### Evento `ping`

Mensaje de keepalive que el servidor envía cada 25 segundos para mantener la conexión activa.

```json
{
  "type": "ping"
}
```

No contiene campo `data`. El cliente no necesita responder.

---

### Evento `error`

Error fatal. El servidor cierra la conexión WebSocket inmediatamente después.

```json
{
  "type": "error",
  "data": {
    "message": "phone_id requerido"
  }
}
```

---

## 5. Ciclo de vida de la conexión

### Lo que pasa cuando el cliente conecta

1. El cliente abre la conexión WebSocket con su JWT.
2. El servidor valida el token. Si es inválido, rechaza la conexión con HTTP 401 antes de abrir el WS.
3. Según el tipo de token, el servidor espera el mensaje `subscribe` (JWT empresa) o auto-suscribe (JWT qr_link).
4. El servidor inicia internamente la sesión de WhatsApp para ese teléfono.
5. El servidor empieza a enviar eventos al cliente.

### Lo que pasa cuando el cliente desconecta

Cuando el cliente cierra el WebSocket (o pierde conectividad), el servidor limpia la sesión internamente:

- Si la sesión estaba en estado `initializing` o `qr_pending` (el QR nunca fue escaneado), **el servidor termina la sesión de WhatsApp** para no dejar recursos colgados.
- Si la sesión ya estaba `active`, **el servidor la mantiene** — el teléfono sigue conectado a WhatsApp aunque el cliente WS se haya ido.

Esto significa que:
- Un cliente puede reconectarse al WS y retomar una sesión `active` ya existente.
- Si el cliente cierra antes de escanear el QR, hay que volver a conectar y escanearlo de nuevo.

### Keepalive

El servidor envía un `ping` cada 25 segundos. Si el cliente necesita mantener la conexión activa desde su lado, puede implementar su propio heartbeat, pero no es estrictamente necesario.

---

## 6. Flujo QR Link — enlace de escaneo sin login

El QR Link permite que un operador (sin credenciales de admin) abra un enlace en su navegador y escanee el QR para activar un teléfono.

### Paso a paso

```
Admin                    Sistema                  Operador
  |                         |                         |
  |-- POST /api/admin/      |                         |
  |   telefonos/{id}/qr-link|                         |
  |                         |-- genera JWT (10 min) --|
  |<-- { token, phone_id,   |                         |
  |      expires_in: 600 }  |                         |
  |                         |                         |
  | construye URL:          |                         |
  | {origin}/qr?token=TOKEN |                         |
  |                         |                         |
  |--- envía enlace ------->|-------> al operador ----|
  |                         |                         |
  |                         |         abre el enlace  |
  |                         |<-- WS con JWT qr_link --|
  |                         |                         |
  |                         |-- auto-suscribe ------->|
  |                         |<-- eventos (qr...) -----|
  |                         |                         |
  |                         |      escanea el QR      |
  |                         |<-- connected isActive:true
```

### Detalles de la API

**Generar el token:**

```
POST /api/admin/telefonos/{id}/qr-link
Authorization: Bearer <JWT admin>
```

Respuesta:
```json
{
  "ok": true,
  "token": "eyJhbGci...",
  "phone_id": 42,
  "expires_in": 600
}
```

**Construir el enlace:**

```
https://tudominio.com/qr?token=eyJhbGci...
```

La página `/qr` se conecta automáticamente al WS con ese token. El operador solo abre el enlace, ve el QR en pantalla y lo escanea con WhatsApp.

**Consideraciones:**
- El token dura **10 minutos** (`expires_in: 600`).
- Si el operador no escanea a tiempo, hay que generar un nuevo token.
- El enlace es de un solo uso conceptual — si el operador cierra y reabre antes de que expire, se puede reconectar con el mismo token.

---

## 7. Reconexión automática al arrancar el servidor

Cuando el servidor de wsapi se reinicia, no todas las sesiones de WhatsApp se recuperan automáticamente. El servidor ejecuta un proceso de reconexión al arrancar.

### Qué hace

Al iniciar (si está habilitado en la configuración):

1. El servidor revisa todos los teléfonos que en la base de datos tienen estado `active`.
2. Verifica cuáles de esos teléfonos ya tienen una sesión de WhatsApp activa en memoria (normalmente ninguna, porque el servidor acaba de arrancar).
3. Para los que están en DB como `active` pero sin sesión en memoria, intenta reconectarlos automáticamente.
4. Máximo **4 reconexiones en paralelo** al mismo tiempo.
5. Si falla un intento, reintenta hasta **2 veces más** con 1.2 segundos de pausa entre intentos.
6. Si después de los reintentos sigue fallando, el teléfono queda marcado como `disconnected` en la DB.

### Resultado en el log del servidor

```
[INFO] startup bootstrap sesiones: total=10 activos_db=7 runtime_activos=0 mismatches=7 intentos_start=8 errores_start=1 duracion=4s
```

| Campo | Significado |
|-------|-------------|
| `total` | Total de teléfonos en el sistema |
| `activos_db` | Cuántos estaban marcados como `active` en DB |
| `runtime_activos` | Cuántos ya tenían sesión activa en memoria (normalmente 0 al arrancar) |
| `mismatches` | Teléfonos que debían reconectarse |
| `intentos_start` | Total de intentos de conexión (incluyendo reintentos) |
| `errores_start` | Cuántos no pudieron reconectarse tras todos los intentos |
| `duracion` | Tiempo total que tomó el proceso |

### Lo que esto implica para los integradores

- Si el servidor se reinicia, los teléfonos activos se reconectan solos en segundos o pocos minutos.
- Un cliente WS que tenía conexión abierta perderá la conexión al reiniciar el servidor y deberá reconectarse.
- Una vez reconectado el teléfono (estado `active`), el cliente WS puede volver a suscribirse y seguir recibiendo eventos normalmente.

---

## Apéndice — Resumen rápido para integradores

```
1. Obtener JWT de empresa (vía API REST de auth)
2. Abrir WebSocket: wss://host/api/service/v1/ws?token=JWT
3. Enviar mensaje subscribe: {"type":"subscribe","data":{"phone_id":N}}
4. Recibir eventos:
   - "qr"          → mostrar código QR al operador
   - "connected"   → sesión activa (isActive:true) o caída (isActive:false)
   - "disconnected"→ sesión cerrada por WhatsApp
   - "ping"        → keepalive, ignorar
   - "error"       → algo falló, el servidor cerrará la conexión
5. Cuando la sesión está "active", el teléfono puede enviar y recibir mensajes vía REST
```
