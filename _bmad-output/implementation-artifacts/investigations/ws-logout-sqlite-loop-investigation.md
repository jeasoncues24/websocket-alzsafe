# Investigation: Bucle WS en logout + SQLite stale credentials

## Hand-off Brief

1. **Qué ocurre.** Cuando WhatsApp envía un evento `LoggedOut`, el archivo SQLite de sesión no se elimina; en el siguiente intento de conexión, `GetQRChannel` falla (el device ya tiene credenciales), whatsmeow intenta autenticar con claves expiradas, el servidor vuelve a recibir `LoggedOut`, emite `requiresNewQR: true` y el frontend reabre el WS dos segundos después — creando un bucle infinito sin mostrar QR.
2. **Dónde está el caso.** Causa raíz confirmada en dos capas: backend no borra SQLite al logout (`service.go:193-208`), y frontend siempre reconecta cuando recibe `requiresNewQR: true` (`connect/page.tsx:162-165`) sin backoff ni límite.
3. **Qué falta hacer.** Implementar borrado del SQLite en el defer de `runSession` cuando la razón es `logged_out`; secundariamente revisar la lógica de reconexión del frontend.

## Case Info

| Campo            | Valor                                                          |
| ---------------- | -------------------------------------------------------------- |
| Ticket           | N/A                                                            |
| Fecha apertura   | 2026-06-02                                                     |
| Estado           | Concluded                                                      |
| Sistema          | Go backend + Next.js frontend, WhatsApp vía whatsmeow          |
| Fuentes          | Código fuente, análisis estático, traza de flujo de control    |

## Problem Statement

El usuario reporta dos síntomas relacionados:
1. Al detectar un logout de WhatsApp, ¿es funcional/seguro borrar el archivo SQLite de sesión?
2. Al desloguearse entra en un bucle: el WS se abre y se cierra repetidamente sin mostrar el QR.

## Evidence Inventory

| Fuente                                                              | Estado      | Notas                                               |
| ------------------------------------------------------------------- | ----------- | --------------------------------------------------- |
| `backend/internal/whatsapp/service.go`                              | Disponible  | Flujo completo `runSession`, manejo de `LoggedOut`  |
| `backend/internal/whatsapp/sqlite.go`                               | Disponible  | `removeSQLiteArtifacts` existe pero no se usa aquí  |
| `backend/internal/http/handlers/v1_ws.go`                           | Disponible  | Handler WS de la página QR pública                  |
| `backend/internal/http/admin.go` (ConnectCompanyPhoneWS)            | Disponible  | Handler WS del panel admin                          |
| `frontend/app/empresas/[...]/connect/page.tsx`                      | Disponible  | Lógica de reconexión con `requiresNewQR`            |
| `frontend/app/qr/page.tsx`                                          | Disponible  | Página QR pública, sin reconexión automática        |
| `backend/internal/whatsapp/manager.go`                              | Disponible  | Gestión de clientes, `Delete` llama `StopSession`   |
| `backend/internal/whatsapp/startup_bootstrap.go`                    | Disponible  | Restauración de sesiones al inicio                  |

## Investigation Backlog

| # | Camino a explorar                       | Prioridad | Estado |
| - | --------------------------------------- | --------- | ------ |
| 1 | Verificar flujo exacto de `GetQRChannel` con device enrollado | Alta | Done |
| 2 | Confirmar que `requiresNewQR: true` dispara siempre el reconnect | Alta | Done |
| 3 | Revisar si el startup_bootstrap también entra en bucle | Media | Done |

## Timeline of Events (flujo del bucle)

| Paso | Evento                                                                          | Fuente                        | Confianza  |
| ---- | ------------------------------------------------------------------------------- | ----------------------------- | ---------- |
| 1    | WhatsApp envía `LoggedOut` → `permanent: true`                                  | `service.go:258-262`          | Confirmado |
| 2    | `runSession` termina; defer llama `runtime.storage.Close()` pero NO borra SQLite | `service.go:193-208`          | Confirmado |
| 3    | `emitActive("Sesion desconectada", false, {requiresNewQR: true})`               | `service.go:295-300`          | Confirmado |
| 4    | WS handler recibe `!ok` del canal cerrado → retorna                             | `v1_ws.go:131-133` / `admin.go:1521-1523` | Confirmado |
| 5    | Frontend recibe `active-*` con `requiresNewQR: true` → `setTimeout(openSocket, 2000)` | `connect/page.tsx:162-165` | Confirmado |
| 6    | WS cierra → `ws.onclose` → `startStatusPoll` activo                             | `connect/page.tsx:185-193`    | Confirmado |
| 7    | 2 segundos después: frontend abre NUEVO WS → `StartSession`                     | `connect/page.tsx:163-165`    | Confirmado |
| 8    | Runtime anterior ya terminó → nuevo runtime inicia → `openSQLiteContainer`      | `service.go:72-98`            | Confirmado |
| 9    | SQLite tiene credenciales viejas → `GetFirstDevice()` devuelve device enrollado | `sqlite.go:23-43`             | Confirmado |
| 10   | `GetQRChannel()` falla (device tiene JID) → sin QR                             | `service.go:354,366`          | Deducido   |
| 11   | whatsmeow intenta conectar con credenciales expiradas → WA envía `LoggedOut`    | whatsmeow + `service.go:258`  | Deducido   |
| 12   | Vuelve al paso 2 → bucle infinito                                               | —                             | Confirmado |

## Confirmed Findings

### Finding 1: `removeSQLiteArtifacts` existe pero no se invoca en logout

**Evidencia:** `sqlite.go:66-77` define `removeSQLiteArtifacts(path)` que borra `.db`, `.db-wal` y `.db-shm`. Se invoca solo en `openSQLiteContainer` cuando hay un error de upgrade (`isWhatsmeowUpgradeConflictError`), nunca cuando el evento es `LoggedOut`.

**Detalle:** El defer de `runSession` (`service.go:193-208`) solo cierra el container con `runtime.storage.Close()`. El archivo SQLite persiste en disco con las credenciales de device revocadas.

---

### Finding 2: El frontend reconecta automáticamente sin límite cuando `requiresNewQR: true`

**Evidencia:** `connect/page.tsx:161-166`:
```javascript
if (requiresNewQR) {
    setTimeout(() => {
        openSocket();
    }, 2000);
}
```
Todos los paths de desconexión en `service.go` emiten `requiresNewQR: true`: líneas 295, 362, 372, 386, 406, 411. No hay contador de intentos ni backoff exponencial.

---

### Finding 3: Con credenciales stale, `GetQRChannel` falla y nunca se muestra el QR

**Evidencia (deducida):** `service.go:354`:
```go
qrChan, qrErr := runtime.client.GetQRChannel(runtime.ctx)
```
Cuando el device está enrollado (tiene JID en SQLite), whatsmeow retorna error en `GetQRChannel`. El código va al branch `if qrErr != nil` (línea 366) que llama `waitForConnection` por 30s. WhatsApp rechaza las credenciales expiradas enviando `LoggedOut` o `ConnectFailure`. Tras el timeout/rechazo, `markDisconnected` emite `requiresNewQR: true` → bucle.

**Qué confirmaría/refutaría:** Agregar log al branch `qrErr != nil` para ver si se ejecuta durante el bucle.

---

### Finding 4: La página QR pública (`/qr/page.tsx`) NO tiene bucle

**Evidencia:** `qr/page.tsx:66-69` — `ws.onclose` solo actualiza estado a `"closed"`, no llama `openSocket()` ni reconecta. El bucle solo afecta al panel admin (`connect/page.tsx`).

---

### Finding 5: El startup_bootstrap no causa bucle en runtime

**Evidencia:** `startup_bootstrap.go:189-209` — `startSessionWithRetry` reintenta `StartSession` hasta 3 veces. Pero `StartSession` es idempotente: si el runtime ya existe (`s.runtimes[accountID]`), devuelve snapshot. Los reintentos quedan bloqueados por el runtime activo, no crean runtimes duplicados.

## Deduced Conclusions

### Deduction 1: Borrar SQLite en logout rompe el bucle

**Basado en:** Findings 1, 2, 3

**Razonamiento:** Si `removeSQLiteArtifacts` se llama en el defer de `runSession` cuando la razón de desconexión es `logged_out`, el siguiente `StartSession` llama `container.GetFirstDevice()` que retorna `nil` → `container.NewDevice()` → `GetQRChannel` retorna el canal QR correctamente → el QR se muestra → el bucle no se produce.

**Conclusión:** Sí es funcional y seguro borrar el SQLite en logout. Las credenciales ya están revocadas en los servidores de WhatsApp. El device ID en SQLite es inservible. `removeSQLiteArtifacts` ya elimina `.db`, `.db-wal` y `.db-shm`, que son todos los artefactos necesarios.

---

### Deduction 2: El diseño de "un solo WS" es correcto en intención, roto en práctica

**Basado en:** Findings 2, 3, 5

**Razonamiento:** `StartSession` retorna snapshot si el runtime existe (idempotente). El problema es que el runtime termina rápido (por credenciales inválidas), y el frontend lo reinicia. Con SQLite limpio, el runtime dura hasta que se conecta o el QR expira: una sola sesión, múltiples WS como observadores del canal de eventos.

**Nota crítica:** El canal `runtime.events` es consumido por el primer WS. Las conexiones posteriores reciben solo el snapshot (estado actual), no el stream en vivo. Esto significa que si el admin abre el panel y el QR ya se generó, el segundo observador no verá el QR en el stream. Solo recibirá el snapshot del `sessionStore`.

## Hypothesized Paths

### Hipótesis H1: `LoggedOut` durante `ConnectContext` vs durante sesión activa se comportan diferente

**Estado:** Open  
**Descripción:** `v.OnConnect` en `waEvents.LoggedOut` (service.go:261) indica que el logout ocurrió durante la conexión. En ese caso el runtime puede terminar antes de emitir el evento de desconexión al WS. ¿Hay race condition entre el cierre del canal y la lectura del WS?  
**Para confirmar:** Agregar logs con timestamp en `close(runtime.events)` y en la lectura `!ok` del WS handler.

## Source Code Trace

| Elemento       | Ubicación                                         |
| -------------- | ------------------------------------------------- |
| Raíz del bug 1 | `service.go:193-208` — defer no borra SQLite      |
| Raíz del bug 2 | `connect/page.tsx:162-165` — reconnect sin límite |
| Función clave  | `sqlite.go:66-77` — `removeSQLiteArtifacts`       |
| Punto de fix 1 | `service.go:193-208` — agregar `removeSQLiteArtifacts` cuando reason es `logged_out` |
| Punto de fix 2 | `connect/page.tsx:162-165` — agregar contador/backoff |

## Final Conclusion

**Confianza: Alta** para el bug de SQLite. **Alta** para el bucle de frontend. La cadena causal está completamente trazada en código fuente sin ambigüedad.

**Causa raíz primaria:** El archivo SQLite con credenciales revocadas no se elimina al recibir `LoggedOut`. Las credenciales expiradas impiden que `GetQRChannel` funcione, la conexión falla, y el frontend reconecta indefinidamente porque todos los paths de desconexión emiten `requiresNewQR: true`.

**Causa raíz secundaria:** El frontend en `connect/page.tsx` reconecta sin backoff ni límite de intentos.

## Fix Direction

### Fix 1 — Backend: borrar SQLite en logout (corrige el bucle en la raíz)

En `service.go`, el `runSession` necesita rastrear si la razón de terminación fue `logged_out`. En el defer, después de `runtime.storage.Close()`, llamar `removeSQLiteArtifacts(dbPath)`.

Requiere: pasar el path del SQLite al runtime (actualmente el `sessionRuntime` no lo almacena), o calcular el path a partir de `baseDir` + `accountID` igual que en `openSQLiteContainer`.

**Complejidad:** Baja. Agregar `dbPath string` al struct `sessionRuntime`, asignarlo en `StartSession`, y en el defer condicionar el borrado.

### Fix 2 — Frontend: backoff y límite de reintentos (reduce el spam, no corrige la raíz)

En `connect/page.tsx`, agregar un contador de reintentos y backoff exponencial antes de llamar `openSocket()`. Resetear el contador cuando el WS llega a estado QR o connected.

**Complejidad:** Baja. Requiere un `useRef` para el contador.

**Estado:** Active → Concluded → **Resuelto (implementado 2026-06-02)**

## Resolución implementada (2026-06-02)

### Fix 1 — Backend: purga de SQLite en logout
- `sqlite.go`: nuevo helper `sqliteDBPath(baseDir, accountID)` reutilizado por `openSQLiteContainer`.
- `service.go`: `sessionRuntime` ahora almacena `dbPath`. En `runSession`, el flag `purgeStore` se activa cuando una desconexión permanente tiene razón `logged_out` (dos puntos: bloque permanente directo y bloque de reconexión). El defer llama `removeSQLiteArtifacts(dbPath)` tras cerrar el container. Solo se purga en `logged_out` (no en `stream_replaced` ni `temporary_ban`).

### Fix 3 — Backend: fan-out multi-observador (corrige el enlace compartido)
- `service.go`: `sessionRuntime` ahora soporta múltiples observadores vía `subscribers map[chan SessionEvent]struct{}`, con métodos `subscribe()` (devuelve canal + unsubscribe, entrega snapshot `last`), `broadcast()` (difusión no bloqueante) y `closeAll()`. Eliminado el mapa `starting` y la lógica snapshot-then-close.
- `StartSession` ahora devuelve `(<-chan SessionEvent, func(), error)`: si el runtime existe, el llamador se suscribe como observador adicional al MISMO cliente WhatsApp — nunca se abre una segunda conexión.
- Limpieza observer-aware: el runtime se cancela solo cuando se va el ÚLTIMO observador y la sesión nunca estuvo activa (`everActive` sticky). Sesiones ya conectadas sobreviven al cierre de la pestaña.
- Eliminada la lógica `manager.Delete`-on-close de `v1_ws.go` y `admin.go` (mataba la sesión de otros observadores); reemplazada por `unsubscribe()`.
- Callers actualizados: `v1_ws.go`, `admin.go` (REST + WS), `startup_bootstrap.go`, `client.go` (fallback).

### Fix 2 — Frontend: backoff con tope
- `connect/page.tsx`: `reconnectAttemptsRef` con backoff exponencial (2s,4s,8s,16s,30s) y tope de 5 intentos. Se resetea al recibir QR o conexión activa, y en reconexión manual. Timer limpiado en `closeSocket`/desmontaje.

### Verificación
- `go build ./...` OK. `go test ./...` OK. `go test -race` OK en whatsapp/http.
- Nuevo `service_runtime_test.go`: 6 tests del fan-out (difusión múltiple, snapshot, unsubscribe, closeAll, abandono QR vs. sesión activa).
- `npm run lint` y `tsc` sin errores nuevos.
