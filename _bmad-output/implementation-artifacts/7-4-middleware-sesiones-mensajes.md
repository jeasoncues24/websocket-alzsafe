---
status: review
type: backend
story_key: 7-4-middleware-sesiones-mensajes
created: 2026-04-15
last_updated: 2026-04-15
---

# Story 7.4: Middleware de protección para endpoints de sesiones y mensajes

## Story

**As a** sistema backend
**I want** proteger los endpoints de sesiones WhatsApp y mensajes con autenticación JWT
**So that** solo usuarios autenticados puedan acceder y operar con estos recursos

## Acceptance Criteria

**Given** una request a /api/sessions/*, /api/message/*, o /api/broadcast/*
**When** no incluye JWT válido
**Then** retorna HTTP 401 con mensaje "Token requerido"

**Given** un JWT válido
**When** intenta acceder a recursos de otra empresa (distinta a su empresa_id en token)
**Then** retorna HTTP 403 Forbidden

**Given** un JWT válido con empresa_id null (super_admin)
**When** accede a endpoints de sesiones o mensajes
**Then** puede operar con todas las empresas
**And** debe incluir header X-Empresa-ID para especificar la empresa objetivo

## Tasks/Subtasks

- [x] 1. Proteger endpoints de mensajes (/api/message, /api/messages, /api/broadcast)
- [x] 2. Proteger endpoints de sesiones (/api/sessions)
- [x] 3. Proteger endpoints admin (/api/admin/messages, /api/admin/broadcasts)
- [x] 4. Tests y verificación de build

## File List

- internal/http/router.go (modificar)

## Change Log

- (2026-04-15) Story creada para middleware de protección sesiones/mensajes
- (2026-04-15) Implementado: protección JWT para endpoints de mensajes y sesiones