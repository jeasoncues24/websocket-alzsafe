# Story 1.3: Emision de QR y confirmacion de estado activo por empresa

Status: in-progress

## Story

As a operador de empresa,
I want recibir qr-ruc_empresa y active-ruc_empresa,
so that pueda autenticar y confirmar que mi sesion quedo lista.

## Acceptance Criteria

1. Dado una sesion en inicializacion, cuando el proveedor emite QR, entonces el backend publica evento qr-ruc_empresa.
2. Dado una sesion autenticada/ready, cuando se confirma estado activo, entonces el backend publica active-ruc_empresa.
3. Dado cambios de estado de sesion por empresa, cuando ocurren, entonces se persiste estado operativo por empresa.

## Tasks / Subtasks

- [x] Emitir evento qr-ruc_empresa durante init-session (AC: 1)
- [x] Emitir evento active-ruc_empresa con isActive false en fase de inicializacion (AC: 2)
- [x] Agregar transicion de estado activo con evento session-ready -> active-ruc_empresa (AC: 2)
- [x] Persistir estado de sesion por empresa en SessionStore (initializing, qr_pending, active) (AC: 3)
- [ ] Integrar evento de ready real desde proveedor WhatsApp (AC: 2)
- [ ] Persistir estado en capa durable (DB) para reinicio de proceso (AC: 3)

## Dev Notes

- La emision de QR y active ya esta operativa con contrato de eventos compatible.
- En esta fase, el estado de sesion se persiste en memoria por proceso.
- El evento session-ready es puente temporal para completar wiring con proveedor real en stories siguientes.

### Files Implemented

- [internal/http/handlers.go](internal/http/handlers.go)
- [internal/http/router.go](internal/http/router.go)
- [internal/whatsapp/qr.go](internal/whatsapp/qr.go)
- [internal/storage/sessions.go](internal/storage/sessions.go)

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L99)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md#L76)
- Project context: [\_bmad-output/project-context.md](_bmad-output/project-context.md#L40)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- N/A

### Completion Notes List

- Emision de qr-ruc_empresa implementada.
- Emision de active-ruc_empresa implementada para estado inicial y estado activo.
- SessionStore agregado para persistencia de estado por empresa en runtime.

### File List

- \_bmad-output/implementation-artifacts/1-3-emision-de-qr-y-confirmacion-de-estado-activo-por-empresa.md
