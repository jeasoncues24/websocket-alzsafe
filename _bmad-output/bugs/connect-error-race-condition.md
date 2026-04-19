# Bug: connect_error falso por race condition con ConnectContext

**Fecha detectado:** 2026-04-18  
**Archivo afectado:** `internal/whatsapp/service.go`  
**Severidad:** Alta — impide toda sesión nueva de WhatsApp  
**Estado:** Documentado, pendiente fix

---

## Síntoma

El WebSocket emite `connect_error` y `requiresNewQR: true` en cada intento de conectar un teléfono nuevo, incluso cuando la red funciona y WhatsApp responde con QR codes válidos.

```json
{"event":"active-+51961669652","data":{"isActive":false,"message":"No se pudo conectar","reason":"connect_error","requiresNewQR":true}}
```

---

## Evidencia en logs

Patrón repetido en cada intento (extracto del primer ciclo):

```
13:11:00.986  Dialing wss://web.whatsapp.com/ws/chat          ← dial OK
13:11:01.549  Frame websocket read pump starting               ← WS establecido
13:11:01.682  ConnectContext retornó — err: <nil>              ← RETORNA ANTES DE QR
13:11:01.682  IsConnected: true                                ← sigue conectado
13:11:01.682  IsLoggedIn: false                                ← no logueado (normal)
13:11:01.972  Received QR code event, starting to emit codes   ← QR llega 290ms TARDE
13:11:01.972  Emitting QR code 2@fu6Gth9EG7...                ← QR válido generado
```

**ConnectContext retorna a los ~700ms del dial. El QR llega ~290ms después.**  
Esto ocurre exactamente igual en los 6 intentos registrados.

---

## Causa raíz

### Supuesto incorrecto en el diseño

El código en `service.go:182-183` asume que `ConnectContext` es una **llamada bloqueante hasta que la sesión termina**:

```go
// línea 182-183
go func() {
    connectErrCh <- runtime.client.ConnectContext(runtime.ctx)
}()
```

Y el select loop (línea 206+) trata cualquier retorno de `connectErrCh` como "sesión terminada":

```go
case connectErr := <-connectErrCh:  // se asume: sesión cayó
    // → emite connect_error si !activeEmitted
```

### Comportamiento real de ConnectContext (whatsmeow v0.0.0-20260410162419)

`ConnectContext` **NO bloquea hasta el fin de la sesión**. Retorna después del handshake inicial (cuando el WebSocket queda establecido con `s.whatsapp.net`). Esto se confirma con:

- `err: <nil>` — sin error
- `IsConnected: true` — la conexión sigue viva cuando retorna
- `IsLoggedIn: false` — aún no completó el login (esperado: el QR todavía no fue escaneado)

La API oficial de whatsmeow lo confirma en su ejemplo de documentación:

```go
// Ejemplo oficial: Connect() retorna rápido, luego lees el canal de QR
err = client.Connect()      // retorna al establecer el WS
if err != nil { panic(err) }
for evt := range qrChan {   // los QR llegan DESPUÉS, async
    ...
}
// El programa sigue vivo con signal.Notify — no espera en ConnectContext
```

### La race condition

```
Goroutine A (runSession):          Goroutine B (ConnectContext):
  select { qrChan | connectErrCh }   ConnectContext() establece WS
                                      ← retorna nil (~700ms)
  connectErrCh ← nil                ← select lo recibe AQUÍ
  activeEmitted = false
  → emite connect_error ← BUG
  → runSession() termina
  
  (290ms después, qrChan tenía datos, pero ya nadie escucha)
```

---

## Lo que NO es el problema

- ❌ Red o firewall — WhatsApp responde correctamente, el WS se establece, los QR codes son válidos
- ❌ Número bloqueado — se reciben `pair-device` IQs con múltiples refs
- ❌ SQLite o driver — el store funciona
- ❌ Versión de whatsmeow — se comporta exactamente como documenta

---

## Fix requerido

Cambiar la arquitectura de `runSession()` para que coincida con el patrón oficial:

1. **Llamar `Connect()` síncronamente** (no en goroutine) y verificar error de handshake inicial
2. **Iterar qrChan** normalmente (ya funciona, el problema es que muere antes)
3. **Detectar desconexión real** usando `client.AddEventHandler()` con `*events.Disconnected` o manteniendo un canal de lifecycle separado que solo se cierra cuando whatsmeow cierra el WS internamente

El `connectErrCh` actual no sirve como señal de "sesión terminada" — debe eliminarse del flujo QR.

---

## Archivos a modificar

- `internal/whatsapp/service.go` — función `runSession()` líneas 179-259
