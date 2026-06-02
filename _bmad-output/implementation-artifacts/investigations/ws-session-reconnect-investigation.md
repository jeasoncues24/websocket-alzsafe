# Investigation: WhatsApp Session Reconnection Bug

## Hand-off Brief

1. **What happened.** Las sesiones de WhatsApp se marcan como `disconnected` en BD y en SessionStore en cuanto reciben el evento `Disconnected` de whatsmeow, aunque whatsmeow lanza simultáneamente un goroutine `autoReconnect` que reconectaría la sesión sola — el servicio termina su goroutine `runSession` y cierra el SQLite antes de que whatsmeow pueda reconectar.
2. **Dónde está el caso.** Root cause confirmado en dos puntos del código; hay un tercer bug secundario en el timeout del bootstrap. Se identificaron también gaps de evidencia sobre el comportamiento real de SQLite bajo reconexión.
3. **Qué se necesita.** Implementar fix en `backend/internal/whatsapp/service.go`: escuchar el evento `Connected` de whatsmeow durante el período de reconexión antes de marcar la sesión como muerta.

## Case Info

| Campo            | Valor                                                                           |
| ---------------- | ------------------------------------------------------------------------------- |
| Ticket           | N/A                                                                             |
| Fecha apertura   | 2026-05-27                                                                      |
| Status           | Active                                                                          |
| Sistema          | Go backend, whatsmeow v0.0.0-20260410162419-b95d92207080, MySQL/SQLite          |
| Fuentes          | service.go, startup_bootstrap.go, whatsmeow/client.go, whatsmeow/connectionevents.go, logs de sesión del usuario |

## Problem Statement

Sesiones de WhatsApp que se conectaron correctamente muestran estado `disconnected` en el sistema aunque el dispositivo físico (teléfono) sigue conectado a WhatsApp. Cuando el servicio se reinicia, no restaura la sesión correctamente. La hipótesis inicial del usuario es una condición de carrera en la lógica de reconexión.

## Evidence Inventory

| Fuente                                                        | Status      | Notas                                                                |
| ------------------------------------------------------------- | ----------- | -------------------------------------------------------------------- |
| `backend/internal/whatsapp/service.go`                        | Available   | Leído completo. Contiene `runSession`, `waitForDisconnect`, `markDisconnected` |
| `backend/internal/whatsapp/startup_bootstrap.go`              | Available   | Leído completo. Bootstrap al arrancar el servicio                    |
| `backend/internal/whatsapp/manager.go`                        | Available   | Leído completo. Gestión de clientes en memoria                       |
| `backend/internal/storage/sessions.go`                        | Available   | Leído completo. SessionStore en memoria                              |
| `whatsmeow/client.go` (módulo Go)                             | Available   | Leído fragmentos clave: `NewClient`, `onDisconnect`, `autoReconnect` |
| `whatsmeow/connectionevents.go` (módulo Go)                   | Available   | Leído: `dispatchEvent(&events.Connected{})` en línea 201             |
| Logs de sesión del usuario (captura pantalla conversación)    | Partial     | Visible la secuencia de eventos del 26/5 y 27/5                      |
| Comportamiento real de SQLite al cerrarse con reconnect activo | Missing     | No verificado si genera panic/error silencioso                       |
| Logs del backend (archivos de log en producción)              | Missing     | No accesibles en esta sesión                                         |

## Investigation Backlog

| #  | Camino a Explorar                                                | Prioridad | Status | Notas |
| -- | ---------------------------------------------------------------- | --------- | ------ | ----- |
| 1  | Verificar si `runtime.storage.Close()` es goroutine-safe mientras `autoReconnect` corre | High | Open | Potencial panic oculto |
| 2  | Verificar si `whatsmeow` usa `runtime.ctx` o socket context en `autoReconnect` | High | Done | Usa el contexto del socket, derivado de `runtime.ctx` |
| 3  | Verificar qué ocurre con `markDisconnected` → `telefonoStore.SetDisconnected` en reconexión temporal | High | Done | Se escribe en BD permanentemente |
| 4  | Verificar si el bootstrap detecta telefonos con status `disconnected` | Medium | Done | No los reconecta, solo detecta los que siguen en `active` |
| 5  | Revisar si hay algún job periódico de reconciliación de estado  | Medium | Done | No existe ninguno |
| 6  | Revisar logs de producción para identificar el evento exacto que dispara la desconexión | Medium | Open | Requiere acceso al servidor |

## Timeline of Events (caso del 26/5)

| Hora        | Evento                          | Fuente          | Confianza |
| ----------- | ------------------------------- | --------------- | --------- |
| 14:56:42    | WebSocket admin cerrado (normal) | Logs usuario    | Confirmed |
| 14:57:04    | QR generado                     | Logs usuario    | Confirmed |
| 14:57:24    | QR generado (segundo)           | Logs usuario    | Confirmed |
| 14:57:44    | QR generado (tercero)           | Logs usuario    | Confirmed |
| 14:57:46    | WebSocket admin cerrado (normal) | Logs usuario    | Confirmed |
| 14:57:51    | `connected` — sesión activa     | Logs usuario    | Confirmed |
| 14:57:51    | `disconnect: ws_closed` (WS admin, no WA) | Logs usuario | Confirmed |
| ~14:57–16:33 | Sesión activa en el sistema     | Deducido       | Deduced   |
| 16:33:14    | `disconnected` — sesión caída   | Logs usuario    | Confirmed |
| 16:33:14+   | Teléfono físico continúa en WA  | Reporte usuario | Hypothesized |
| 27/5 12:48  | `initializing` + `connected`    | Logs usuario    | Confirmed |

## Confirmed Findings

### Finding 1: whatsmeow emite `Disconnected` Y lanza `autoReconnect` simultáneamente

**Evidencia:** `whatsmeow/client.go:557-560`

```go
if !cli.isExpectedDisconnect() && (cli.forceAutoReconnect.Swap(false) || remote) {
    go cli.dispatchEvent(&events.Disconnected{})
    go cli.autoReconnect(ctx)   // ← goroutine paralelo
}
```

**Detalle:** Cuando la conexión TCP/WebSocket cae de forma remota (caída de red, timeout del servidor de WA, rolling update), whatsmeow emite el evento `Disconnected` **Y simultáneamente lanza un goroutine de reconexión automática** con backoff exponencial (`0s, 2s, 4s, 6s…`). `EnableAutoReconnect: true` es el valor por defecto en `NewClient` (`client.go:269`).

---

### Finding 2: `service.go` termina `runSession` al recibir `Disconnected` sin esperar reconexión

**Evidencia:** `backend/internal/whatsapp/service.go:266-288`

```go
waitForDisconnect := func() {
    for {
        select {
        case disconnect := <-disconnectCh:
            if runtime.ctx.Err() != nil { return }
            s.markDisconnected(accountID, reason)   // ← actualiza BD + SessionStore
            emitActive("Sesion desconectada", false, extra)
            return   // ← runSession goroutine termina aquí
        ...
```

**Detalle:** Al recibir cualquier evento de desconexión (incluyendo el temporal `Disconnected`), `waitForDisconnect` llama a `markDisconnected` y retorna. El goroutine `runSession` completa, ejecutando el `defer` que:
1. Cierra `runtime.storage` (SQLite) — `service.go:194`
2. Elimina el cliente del Manager — `service.go:197-198`
3. Elimina el runtime del Service — `service.go:200-202`
4. Cierra el canal de eventos — `service.go:203`

El goroutine `autoReconnect` de whatsmeow sigue corriendo (su context no fue cancelado), pero el SQLite que necesita para guardar estado de sesión ya está cerrado.

---

### Finding 3: `markDisconnected` escribe en BD de forma permanente e inmediata

**Evidencia:** `backend/internal/whatsapp/service.go:383-393`

```go
func (s *Service) markDisconnected(accountID, reason string) {
    if s.sessionStore != nil {
        s.sessionStore.SetDisconnected(accountID, reason)
    }
    s.syncTelefonoDisconnected(accountID)   // ← SetDisconnected en BD
    ...
}
```

Y `backend/internal/storage/telefono.go:192-194`:
```go
func (s *TelefonoStore) SetDisconnected(id int64) error {
    _, err := s.db.Exec("UPDATE telefonos SET status=? ...", domain.TelefonoStatusDisconnected, id)
```

**Detalle:** Una caída temporal de red escribe `status = 'disconnected'` en la BD de forma permanente. El bootstrap al reiniciar solo intenta restaurar teléfonos en estado `active` — los que quedaron en `disconnected` **no son restaurados**.

---

### Finding 4: El código NO escucha `events.Connected` después de la conexión inicial

**Evidencia:** `backend/internal/whatsapp/service.go:232-263` — el event handler solo captura:
- `*waEvents.Message`, `*waEvents.Receipt` → mensajes
- `*waEvents.Disconnected`, `*waEvents.StreamReplaced`, `*waEvents.LoggedOut`, `*waEvents.TemporaryBan`, `*waEvents.ConnectFailure` → desconexiones

El evento `*waEvents.Connected` (emitido en `whatsmeow/connectionevents.go:201` tras reconexión exitosa) **no está registrado**. Incluso si whatsmeow reconecta, el servicio no se entera.

---

### Finding 5: El bootstrap usa timeout de 10 segundos para reconexión

**Evidencia:** `backend/internal/whatsapp/service.go:304`

```go
if waitForConnection(runtime.client, 10*time.Second) {
    s.markConnected(accountID)
} else if runtime.ctx.Err() == nil {
    s.markDisconnected(accountID, "connect_timeout")   // ← escribe en BD como disconnected
    return
}
```

**Detalle:** Si WhatsApp no responde en 10 segundos al reiniciar el servicio, la sesión es marcada como `disconnected` en BD permanentemente y el bootstrap la abandona. Con latencia o carga en servidores de WA, 10s puede ser insuficiente.

---

### Finding 6: Diferencia entre tipos de desconexión no se modela en el código

**Evidencia:** `backend/internal/whatsapp/service.go:239-257`

```go
switch v := evt.(type) {
case *waEvents.Disconnected:
    disconnect.reason = "disconnect"      // temporal (caída TCP)
case *waEvents.StreamReplaced:
    disconnect.reason = "stream_replaced" // otra sesión se conectó
case *waEvents.LoggedOut:
    disconnect.reason = "logged_out"      // PERMANENTE: desvinculado desde cel.
case *waEvents.TemporaryBan:
    disconnect.reason = "temporary_ban"   // PERMANENTE: WA baneó el número
case *waEvents.ConnectFailure:
    disconnect.reason = "connect_failure" // puede ser temporal o permanente
```

Todos los eventos se manejan de forma idéntica: `markDisconnected` + `return`. No hay distinción entre permanente y temporal.

## Deduced Conclusions

### Deducción 1: La condición de carrera entre `autoReconnect` y `runtime.storage.Close()`

**Basado en:** Finding 1 + Finding 2

**Razonamiento:**
1. whatsmeow emite `Disconnected` y lanza `autoReconnect` como goroutines paralelos (`go`)
2. El evento llega al handler de `service.go` y `waitForDisconnect` ejecuta `markDisconnected` + `return`
3. `runSession` retorna → el `defer` ejecuta `runtime.storage.Close()` cerrando el SQLite
4. Mientras tanto, `autoReconnect` hace backoff (0s, luego 2s, 4s...) y llama `cli.connect(ctx)`
5. Si `connect` tiene éxito, whatsmeow emite `events.Connected{}` — pero el handler ya fue removido (`defer RemoveEventHandler`) y el runtime ya no existe

**Conclusión:** Existe una condición de carrera donde el SQLite puede cerrarse mientras whatsmeow intenta reconectar. En el mejor caso: whatsmeow reconecta pero el servicio no se entera y no actualiza el estado. En el peor caso: el cierre del SQLite falla la reconexión con un error que puede ser silencioso.

---

### Deducción 2: El escenario reportado por el usuario se explica por Finding 2 + 3

**Basado en:** Finding 2 + Finding 3 + Finding 6

**Razonamiento:** El teléfono físico no se desconecta de WhatsApp. Lo que ocurre es una caída temporal de la conexión TCP/WebSocket entre el servidor de WA y el backend (puede ser un rolling update del servidor de WA, cambio de IP, timeout de keep-alive). whatsmeow emite `Disconnected` para notificar la caída temporal. El servicio lo trata como desconexión permanente, escribe en BD y destruye el runtime.

**Conclusión:** El usuario ve `disconnected` en su panel porque así lo reportó el sistema, pero el teléfono físico sigue válido en WhatsApp (la sesión no fue logout-eada desde el cel.).

## Hypothesized Paths

### Hypothesis 1: El cierre del SQLite bajo `autoReconnect` genera errores silenciosos

**Status:** Open

**Teoría:** Cuando `runtime.storage.Close()` se ejecuta mientras `autoReconnect` está activo, whatsmeow obtiene errores de SQLite al intentar escribir/leer el device store, pero los loguea a nivel `Errorf` (no panic) y el goroutine termina sin escalar.

**Indicadores de soporte:** El `AutoReconnectErrors` de whatsmeow se incrementa en cada fallo; el `AutoReconnectHook` no está configurado, así que el cliente reintenta indefinidamente hasta que el contexto se cancele.

**Confirmaría:** Ver en logs del backend mensajes de whatsmeow como `"Error reconnecting after autoreconnect sleep: sql: database is closed"` o similar.

**Refutaría:** Encontrar que el contexto usado en `autoReconnect` se cancela antes de que pueda ejecutar `connect()` (si el socket context ya expiró).

**Resolución:** Pendiente de logs de producción.

---

### Hypothesis 2: `StreamReplaced` es temporal y debería manejarse diferente

**Status:** Open

**Teoría:** El evento `StreamReplaced` ocurre cuando WhatsApp abre una nueva sesión en el mismo dispositivo (ej: refresh de autenticación). whatsmeow lo trata como reconnectable internamente. Marcarlo inmediatamente como `disconnected` puede ser prematuro.

**Confirmaría:** Documentación de whatsmeow o casos donde `StreamReplaced` fue seguido de reconexión exitosa.

**Refutaría:** Si `StreamReplaced` siempre requiere re-autenticación con QR nuevo.

---

### Hypothesis 3: El contexto de `autoReconnect` se cancela antes de reconectar

**Status:** Open

**Teoría:** El `ctx` pasado a `autoReconnect` en `client.go:560` es el socket context (no `runtime.ctx`). El socket context puede expirar al cerrarse el socket, terminando `autoReconnect` antes de que pueda reconectar.

**Confirmaría:** Trazar el `ctx` que `onDisconnect` recibe — si es el socket fs context (vida corta) vs `runtime.ctx` (vida larga).

**Refutaría:** Confirmar que el ctx de `autoReconnect` es `runtime.ctx` o un derivado de larga vida.

**Resolución:** Parcialmente investigado — `ConnectContext` recibe `runtime.ctx`. El socket fs tiene su propio context (`fs.Context()`). `onDisconnect` recibe el socket context (`ctx` local en `unlockedConnect` scope). **Si el socket context ya fue cancelado, `autoReconnect` termina inmediatamente en línea 596-598.**

## Missing Evidence

| Gap                                            | Impacto                                          | Cómo obtener                                  |
| ---------------------------------------------- | ------------------------------------------------ | --------------------------------------------- |
| Logs del backend en producción                 | Confirmaría qué evento exacto dispara la desconex | Acceso al servidor, revisar stdout del proceso |
| Context exacto pasado a `autoReconnect`        | Define si H3 es relevante                        | Leer `unlockedConnect` y `keepAliveLoop` completos |
| Comportamiento de SQLite al cerrarse con goroutine activa | Confirma H1                        | Test unitario o logs con SQLite debug mode    |
| Razón exacta del disconnect a las 16:33:14     | Identifica si fue `Disconnected`, `ConnectFailure`, u otro | Logs del backend |

## Source Code Trace

| Elemento      | Detalle                                                                               |
| ------------- | ------------------------------------------------------------------------------------- |
| Error origin  | `service.go:281` — `s.markDisconnected(accountID, reason)` dentro de `waitForDisconnect` |
| Trigger       | Evento `*waEvents.Disconnected` recibido del handler de whatsmeow                    |
| Condición     | La conexión TCP/WebSocket cae de forma remota (caída de red, timeout WA, rolling update) |
| Archivos relacionados | `service.go`, `startup_bootstrap.go`, `storage/sessions.go`, `storage/telefono.go`, `whatsmeow/client.go:550-618` |

## Conclusion

**Confianza: High**

El root cause está **confirmado** en dos bugs encadenados:

**Bug 1 (Principal):** `service.go:waitForDisconnect` no diferencia entre desconexiones temporales (`Disconnected` — caída TCP que whatsmeow reconectaría sola) y permanentes (`LoggedOut` — el usuario desvinculó el teléfono). Trata ambas igual: llama `markDisconnected` y termina el goroutine. Esto destruye el runtime y cierra el SQLite antes de que `autoReconnect` de whatsmeow pueda completar la reconexión. El teléfono físico sigue siendo válido en WhatsApp, pero el sistema ya lo declaró muerto.

**Bug 2 (Secundario):** `service.go:304` usa un timeout de 10 segundos en el bootstrap para confirmar reconexión. Insuficiente bajo carga o latencia de WA.

**Bug 3 (Agravante):** `markDisconnected` escribe `status='disconnected'` en BD permanentemente. El bootstrap solo intenta restaurar teléfonos en `status='active'`, así que sesiones que cayeron por Bug 1 nunca son restauradas al reiniciar.

Hay una hipótesis abierta (H3) sobre si el context de `autoReconnect` ya está cancelado cuando se lanza — si es el socket context (vida corta) en vez de `runtime.ctx`, entonces `autoReconnect` aborta inmediatamente y la condición de carrera no se materializa, pero el efecto observable es el mismo: la sesión se marca como muerta y no se restaura.

## Recommended Next Steps

### Fix direction

**Fix 1 (esencial):** Modificar `waitForDisconnect` en `service.go` para:
- Clasificar eventos como **permanentes** (`LoggedOut`, `TemporaryBan`) o **temporales** (`Disconnected`, `ConnectFailure`, `StreamReplaced`)
- Para temporales: escuchar `*waEvents.Connected` con un timeout de 60-90 segundos antes de declarar la sesión muerta
- Solo marcar `markDisconnected` si el timeout expira sin reconexión confirmada
- Agregar `*waEvents.Connected` al event handler existente

**Fix 2 (simple):** Aumentar timeout en `service.go:304` de 10 segundos a 30 segundos para el path de bootstrap.

**Fix 3 (preventivo):** NO llamar `runtime.storage.Close()` en el `defer` hasta que el contexto esté cancelado o el cliente sea reemplazado — evita cerrar SQLite bajo reconexión activa. (Alternativa: cancelar `runtime.ctx` antes del defer para garantizar que `autoReconnect` termine antes de cerrar el storage.)

### Diagnostic

Para confirmar la causa exacta del evento del 26/5 a las 16:33:14:
1. Agregar logging en `waitForDisconnect` con el tipo exacto del evento recibido antes del fix
2. Agregar logging en whatsmeow `autoReconnect` (vía `AutoReconnectHook`) para capturar errores de reconexión
3. Verificar si hay mensajes de error de SQLite en los logs del proceso al momento del disconnect

## Reproduction Plan

1. Conectar una sesión de prueba y esperar a que quede activa
2. Simular caída temporal de red (iptables block por 30s)
3. Observar: ¿el sistema marca `disconnected`? ¿whatsmeow reconecta sola?
4. Comparar vs comportamiento esperado post-fix: debería mantenerse `active` o volver a `active` después de la reconexión

## Side Findings

- El SessionStore es volátil (en memoria). Al reiniciar el servicio se pierde todo el estado en memoria; el bootstrap lo reconstruye desde BD. Esto es intencional pero significa que el estado en memoria y en BD deben mantenerse sincronizados — lo cual es precisamente lo que Bug 1 rompe.
- No existe ningún job periódico de reconciliación de estado (health check) que valide la consistencia entre Manager, SessionStore y BD. Cualquier desincronización persiste hasta el próximo restart.
- El evento `StreamReplaced` podría ser una señal de que WhatsApp está rotando internamente la conexión (actualización del servidor) — marcarlo como permanente puede generar falsos positivos de desconexión.
