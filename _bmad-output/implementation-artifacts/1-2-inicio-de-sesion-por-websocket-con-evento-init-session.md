# Story 1.2: Inicio de sesion por WebSocket con evento init-session

Status: in-progress

## Story

As a operador de empresa,
I want iniciar sesion enviando init-session con mi ruc_empresa,
so that pueda establecer conectividad de WhatsApp para mi empresa.

## Acceptance Criteria

1. Dado un websocket conectado y una empresa habilitada, cuando se recibe init-session, entonces se dispara la inicializacion del cliente de esa empresa.
2. Dado una solicitud invalida o no autorizada, cuando se procesa init-session, entonces se retorna error-event claro.

## Tasks / Subtasks

- [x] Implementar endpoint WebSocket base en servidor HTTP (AC: 1)
- [x] Parsear payload con evento init-session y data.ruc_empresa (AC: 1)
- [x] Validar ruc_empresa y emitir error-event en casos invalidos (AC: 2)
- [x] Integrar inicio de sesion al Session Manager multiempresa (AC: 1)
- [ ] Agregar verificacion de empresa habilitada contra fuente persistente (AC: 1, 2)
- [ ] Incluir pruebas unitarias especificas para flujos de error-event y evento desconocido (AC: 2)

## Dev Notes

- Esta implementacion cubre el baseline de transporte y validacion minima.
- Queda pendiente conectar autorizacion/habilitacion de empresa con capa de datos.
- La inicializacion real de cliente WhatsApp se completara en stories siguientes del Epic 1.

### Technical Requirements

- Endpoint WebSocket en /ws.
- Formato de eventos: { event, data }.
- Compatibilidad de errores: error-event.
- Mantener aislamiento por ruc_empresa mediante manager.

### Files Implemented

- [internal/http/handlers.go](internal/http/handlers.go)
- [internal/http/router.go](internal/http/router.go)
- [main.go](main.go)
- [internal/whatsapp/client.go](internal/whatsapp/client.go)

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L86)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md#L76)
- Project context: [\_bmad-output/project-context.md](_bmad-output/project-context.md#L40)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- N/A

### Completion Notes List

- Handler WebSocket implementado con lectura continua de mensajes.
- Soporte de init-session implementado.
- Manejo de error-event implementado para payload invalido/evento no soportado.

### File List

- \_bmad-output/implementation-artifacts/1-2-inicio-de-sesion-por-websocket-con-evento-init-session.md
