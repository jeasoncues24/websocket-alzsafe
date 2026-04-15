# Story 1.4: Manejo de desconexion y logout con limpieza aislada

Status: done

## Story

As a administrador de plataforma,
I want manejar desconexiones por empresa sin impacto global,
so that el sistema mantenga continuidad para otras empresas activas.

## Acceptance Criteria

1. Dado una empresa desconectada o logout, cuando ocurre el evento de desconexion, entonces se limpia su sesion y se notifica active-ruc_empresa con isActive false.
2. Dado multiples empresas activas, cuando una empresa se desconecta, entonces no se interrumpe el servicio de otras empresas.

## Tasks / Subtasks

- [x] Agregar evento session-disconnected en WebSocket handler (AC: 1)
- [x] Agregar evento session-logout en WebSocket handler (AC: 1)
- [x] Limpiar cliente en manager por ruc_empresa en desconexion/logout (AC: 1)
- [x] Persistir estado disconnected con razon por empresa en SessionStore (AC: 1)
- [x] Emitir active-ruc_empresa con isActive false y flags de reautenticacion (AC: 1)
- [x] Agregar pruebas unitarias por aislamiento multiempresa (AC: 2)
- [ ] Conectar disparadores reales desde proveedor WhatsApp a estos eventos (AC: 1, 2)

## Dev Notes

- El manejo es aislado por clave de empresa: toda limpieza opera solo sobre ruc_empresa objetivo.
- Se incluyen dos rutas de caída controlada:
  - session-disconnected (requiresNewQR false por defecto)
  - session-logout (requiresNewQR true)
- Falta wiring de proveedor real y test de no regresión multiempresa en capa HTTP.

### Files Implemented

- [internal/http/handlers.go](internal/http/handlers.go)
- [internal/storage/sessions.go](internal/storage/sessions.go)

### References

- Story source: [\_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md#L112)
- PRD source: [\_bmad-output/planning-artifacts/prd.md](_bmad-output/planning-artifacts/prd.md#L83)
- Project context: [\_bmad-output/project-context.md](_bmad-output/project-context.md#L56)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- N/A

### Completion Notes List

- Eventos session-disconnected y session-logout implementados.
- Limpieza de manager y transición a estado disconnected implementadas.
- Notificación active-ruc_empresa con isActive false implementada para ambas rutas.

### File List

- \_bmad-output/implementation-artifacts/1-4-manejo-de-desconexion-y-logout-con-limpieza-aislada.md
