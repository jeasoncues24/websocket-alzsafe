---
status: done
type: backend
story_key: 8-2-sesion-whatsapp-empresa
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 8.2: Asignación de sesión WhatsApp por empresa

## Story

**As a** sistema backend
**I want** asociar sesiones WhatsApp a empresas específicas
**So that** cada empresa gestione su propia sesión de forma aislada

## Acceptance Criteria

**Given** una empresa creada
**When** se solicita iniciar sesión WhatsApp para esa empresa
**Then** la sesión se asocia a esa empresa específico (ruc_empresa)
**And** el WebSocket identifica la empresa por el JWT

**Given** múltiples empresas en el sistema
**When** cada empresa inicia su propia sesión
**Then** el sistema aísla correctamente las sesiones por ruc_empresa
**And** no hay fuga de datos entre empresas

**Given** un super_admin
**When** inicia sesión WhatsApp
**Then** debe especificar empresa_id en el request
**And** la sesión se crea para esa empresa específica

## Tasks/Subtasks

- [x] 1. WebSocket ya usa ruc_empresa como identificador único (implementado)
- [x] 2. Sesiones aisladas por ruc_empresa en whatsapp.Manager (ya existe)
- [x] 3. Validación de empresa en init-session (validar RUC correcto)

## File List

- internal/http/handlers.go (HandleWS - ya usa ruc_empresa)
- internal/whatsapp/manager.go (StartSession usa ruc como accountID)

## Change Log

- (2026-04-15) Story 8-2 - sesiones ya aisladas por ruc_empresa en código existente