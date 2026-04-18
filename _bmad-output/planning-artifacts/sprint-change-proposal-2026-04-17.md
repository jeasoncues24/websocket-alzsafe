# Sprint Change Proposal - Epic 8 Revalidation

## 1. Issue Summary

Se detectó que `Epic 8` estaba marcado como `done` en el tracking, pero al revalidarlo aparecieron inconsistencias entre estado, documentación y comportamiento real. La revisión se disparó al intentar confiar en `S-8.1` a `S-8.3` como cerradas sin tener una base viva de contexto para el epic.

Evidencia:
- `sprint-status.yaml` mostraba `S-8.1`, `S-8.2` y `S-8.3` como `done`.
- `internal/http/handlers/companies.go` aún no valida sesiones activas antes del soft delete.
- `frontend/components/companies/empresa-detail-modal.tsx` no muestra sesiones WhatsApp ni mensajes recientes.
- `frontend/lib/api.ts` tenía un bug de rutas para `PUT`/`DELETE`, ya corregido.

## 2. Impact Analysis

### Epic Impact
- El epic no debe cerrarse todavía.
- Debe tratarse como `in-progress` con un contexto vivo propio.
- El objetivo cambia de “cerrado” a “revalidado y estabilizado”.

### Story Impact
- `S-8.1` pasa a ser la story de validación/gate del CRUD backend.
- `S-8.2` queda como documentación viva de endpoints.
- `S-8.3` queda como verificación de implementación real contra el contrato.
- `S-8.4+` continúan en backlog.

### Artifact Conflicts
- `sprint-status.yaml` necesitaba enlace explícito al contexto del epic.
- `project-context.md` seguía reflejando Epic 8 como cerrado sin matiz.
- `bmad-sprint-status` necesitaba resolver contexto por epic.

### Technical Impact
- El frontend de empresas ya usa rutas correctas para `PUT`/`DELETE`.
- Falta completar el borrado con chequeo de sesiones activas.
- Falta enriquecer el modal de detalle con sesión y últimos mensajes.

## 3. Recommended Approach

**Direct Adjustment**.

No hace falta rollback ni redefinición del MVP. El cambio es un reset de validación y trazabilidad para corregir deriva entre tracking y código.

Riesgo: medio.
Esfuerzo: medio.
Impacto en timeline: bajo a medio.

## 4. Detailed Change Proposals

### Sprint Tracking

OLD:
- `epic-8-gestion-empresas.status = done`
- `S-8.1 = done`
- `S-8.2 = done`
- `S-8.3 = done`

NEW:
- `epic-8-gestion-empresas.status = in-progress`
- `epic-8-gestion-empresas.context = _bmad-output/implementation-artifacts/epic-8-context.md`
- `S-8.1 = in-progress`
- `S-8.2 = ready-for-dev`
- `S-8.3 = ready-for-dev`

Rationale: mantener un epic vivo y secuenciar la revalidación desde la story de control.

### Epic Context

NEW FILE:
- `_bmad-output/implementation-artifacts/epic-8-context.md`

Rationale: tener una fuente de verdad viva para Epic 8 y evitar depender solo del YAML.

### Sprint Status Workflow

OLD:
- El workflow solo leía `sprint-status.yaml`.

NEW:
- El workflow ahora puede resolver el contexto del epic recomendado.

Rationale: identificación y gestión más rápida del contexto actual.

### Frontend API

OLD:
- `PUT /api/companies/?id={id}`
- `DELETE /api/companies/?id={id}`

NEW:
- `PUT /api/companies/{id}`
- `DELETE /api/companies/{id}`

Rationale: alineación con el router real.

### Backend Companies Handler

OLD:
- El listado podía caer a acceso ambiguo si faltaba `empresa_id` en claims.

NEW:
- El listado deniega acceso si un usuario no super_admin no trae `empresa_id`.

Rationale: cerrar acceso implícito y reforzar aislamiento.

## 5. Implementation Handoff

### Scope
- Clasificación: `Moderate`

### Handoff
- **Developer agent**: cerrar los gaps funcionales restantes de Epic 8.
- **Sprint status agent**: usar `epic-8-context.md` como fuente viva para seguimiento.

### Success Criteria
- Epic 8 queda rastreable desde un contexto vivo.
- `S-8.1` valida el CRUD real, no solo el estado.
- `S-8.2` y `S-8.3` quedan listos para ejecución secuencial.
- El tracking no vuelve a marcar como `done` algo que todavía requiere revalidación.
