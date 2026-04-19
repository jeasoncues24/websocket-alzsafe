# Story 11.1: Núcleo WhatsAppMeow con SQLite y lifecycle real

Status: review

## Story

As a operador de empresa,
I want iniciar y mantener sesiones WhatsApp reales con SQLite local,
so that pueda conectar, ver QR y enviar mensajes sin fugas de WebSocket.

## Acceptance Criteria

1. Dado un teléfono sin sesión activa, cuando se inicia conexión, entonces el backend crea un cliente whatsmeow real con SQLite local.
2. Dado un QR generado por WhatsApp, cuando el backend lo recibe, entonces lo publica por WS y lo persiste en `telefonos.qr_string` y `SessionStore`.
3. Dado un teléfono ya conectado, cuando se consulta o se intenta conectar otra vez, entonces el backend devuelve estado activo sin duplicar clientes.
4. Dado un cierre de WebSocket o desconexión, cuando el flujo termina, entonces se limpia el runtime y no quedan conexiones colgadas.
5. Dado el flujo de conexión, cuando se ejecutan tests, entonces pasan las pruebas unitarias del fallback y la sanitización del store SQLite.

## Tasks / Subtasks

- [x] Agregar dependencia SQLite pure-Go y configuración de directorio local (AC: 1)
- [x] Crear runtime WhatsAppMeow con `sqlstore` por teléfono (AC: 1, 3)
- [x] Integrar `StartSession` con eventos QR/active reales y fallback compatible (AC: 1, 2, 3)
- [x] Conectar el WS de `init-session` al stream de eventos real (AC: 2, 4)
- [x] Agregar endpoint `POST /api/admin/telefonos/{id}/connect` para iniciar conexión desde la vista de teléfonos del panel admin (AC: 1, 3)
- [x] Limpiar runtime al cerrar sesión o cortar el WS (AC: 4)
- [x] Agregar tests unitarios del helper SQLite y fallback de `StartSession` (AC: 5)

## Dev Notes

- La persistencia de WhatsAppMeow usa SQLite local por teléfono en `sessions/whatsappmeow`.
- El driver elegido es `modernc.org/sqlite`, así que no hace falta instalar SQLite nativo en el sistema.
- El runtime real emite eventos `qr-<id>` y `active-<id>` para mantener compatibilidad con el WS actual.
- El endpoint nuevo de conexión permite que el frontend admin arranque el vínculo sin inventar otra ruta.

### File List

- `internal/config/config.go`
- `internal/whatsapp/manager.go`
- `internal/whatsapp/client.go`
- `internal/whatsapp/service.go`
- `internal/whatsapp/sqlite.go`
- `internal/whatsapp/sqlite_test.go`
- `internal/http/handlers.go`
- `internal/http/handlers/v1_sessions.go`
- `internal/http/router.go`
- `go.mod`
- `go.sum`

## Dev Agent Record

### Agent Model Used

gpt-5.4-mini

### Debug Log References

- `go test ./internal/whatsapp ./internal/http/...`
- `go test ./...`

### Completion Notes List

- Se añadió un runtime real de WhatsAppMeow con SQLite local y directorio configurable.
- El WS de `init-session` ahora consume eventos reales de QR/estado.
- Se creó `POST /api/admin/telefonos/{id}/connect` para iniciar la conexión desde la vista de teléfonos del panel admin.
- Se agregaron tests unitarios para la sanitización del archivo SQLite y el fallback de `StartSession`.

## Change Log

- 2026-04-18: Implementado runtime WhatsAppMeow con SQLite local, eventos QR/estado y cleanup de sesión.
