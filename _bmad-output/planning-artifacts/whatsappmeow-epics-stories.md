# WhatsAppMeow - Epics y Stories

**Fuente:** `_bmad-output/implementation-artifacts/party-mode-whatsappmeow-analysis.md`

## Estado actual del repo

- Ya existe `internal/http/handlers.go` con flujo WS de `init-session`, `session-ready`, `session-disconnected` y `session-logout`.
- Ya existe `internal/http/handlers/v1_sessions.go` y `internal/http/handlers/v1_phones.go` con endpoints de estado/QR.
- `internal/whatsapp/manager.go`, `client.go` y `qr.go` todavía están en modo base/skeleton.
- El frontend de teléfonos existe en `frontend/app/empresas/[empresaId]/telefonos/page.tsx`.

## Epic 11: Núcleo WhatsAppMeow Backend

Objetivo: dejar el backend con lifecycle real de WhatsAppMeow, estados consistentes y QR usable desde el flujo actual.

### Story 11.1: Cliente WhatsApp real y lifecycle de sesión

**Objetivo:** crear y administrar el cliente WhatsApp real por teléfono/empresa, con inicio, conexión y cierre controlados.

**AC:**
- Al iniciar una sesión, el backend crea o recupera el cliente correcto.
- Si ya existe una sesión activa, el backend no duplica conexiones.
- El estado de sesión se refleja como `disconnected`, `qr_pending` o `active`.

### Story 11.2: QR, estados y limpieza de conexiones

**Objetivo:** publicar QR y estados sin fugas de conexiones ni WebSockets colgados.

**AC:**
- El QR se publica cuando WhatsApp lo genera.
- Al cerrar una sesión, se limpia el cliente y el estado asociado.
- Los WebSockets se cierran limpiamente al terminar la conexión.

### Story 11.3: Share link temporal de 5 minutos

**Objetivo:** generar una URL temporal para que soporte o cliente pueda escanear/vincular desde otro dispositivo.

**AC:**
- La URL expira a los 5 minutos.
- El link se puede revocar.
- Si el link expiró, el backend responde con error claro.

## Epic 12: UX de Teléfonos y Vinculación

Objetivo: llevar el flujo técnico a una experiencia simple dentro de la vista de teléfonos.

### Story 12.1: Modal de conectar con QR e instrucciones

**Objetivo:** abrir un modal al pulsar conectar con QR, temporizador e instrucciones tipo WhatsApp Web.

**AC:**
- El modal muestra QR y estado.
- Las instrucciones explican qué debe hacer el cliente en su teléfono.
- El copy es claro para soporte técnico y administración.

### Story 12.2: Validación de estado antes de conectar

**Objetivo:** impedir conectar si el teléfono ya está vinculado o si el estado no lo permite.

**AC:**
- Si el teléfono ya está `active`, el botón conectar no inicia otro flujo.
- Si está `qr_pending`, el UI muestra el QR vigente.
- Si está `disconnected`, permite reiniciar la vinculación.

### Story 12.3: Botón de share link al cliente

**Objetivo:** permitir copiar/compartir la URL temporal desde el modal de conexión.

**AC:**
- El usuario puede generar el link desde el modal.
- El link se muestra con countdown.
- Se muestra un estado claro cuando expira.

## Epic 13: QA y Pruebas de Mensajería

Objetivo: validar que el backend no deja fugas y que el flujo de mensajes se puede probar manualmente.

### Story 13.1: Tests de lifecycle y fuga de WebSocket

**Objetivo:** cubrir apertura/cierre de WS y limpieza de sesiones.

**AC:**
- El ciclo de vida de WS no deja conexiones activas al cerrar.
- Los tests concurrentes no rompen el manager.

### Story 13.2: Guía paso a paso para enviar un mensaje de prueba

**Objetivo:** dejar documentado el flujo mínimo para validar envío de mensajes desde frontend.

**AC:**
- Existe un paso a paso reproducible.
- El flujo incluye un número de prueba controlado.
- El procedimiento indica qué validar antes y después del envío.

## Dependencia abierta

`whatsmeow` documenta `sqlstore` con soporte completo para SQLite/Postgres. MariaDB no es la opción natural para ese store.

**Decisión pendiente para implementar:**
- Opción A: usar un SQLite local separado solo para WhatsAppMeow.
- Opción B: mantener MariaDB y construir una capa de persistencia propia.

Sin esa decisión, la implementación real del cliente queda bloqueada.
