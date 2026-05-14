---
story_id: "4.3"
epic: "epic-4"
title: "Microinteracciones de Formularios y Modales"
status: done
estimated_days: 1
priority: high
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/components/ui/dialog.tsx
  - frontend/components/ui/input.tsx
  - frontend/components/ui/select.tsx
  - frontend/components/companies/empresa-form-modal.tsx
  - frontend/components/companies/telefono-form-modal.tsx
---

# Story 4.3: Microinteracciones de Formularios y Modales

## Objetivo

Mejorar el feedback visual y la sensación de suavidad en formularios y modales clave del panel, reforzando foco, validación, loading y transiciones sin cambiar la lógica de negocio.

## Acceptance Criteria

- Los modales de empresa y teléfono tienen encabezado contextual y mejor feedback de estado.
- Inputs y selects usan una base de transición/focus más consistente.
- Los submits muestran estado de guardado más claro.
- No se usan colores hardcodeados para estados funcionales nuevos.
- `npm run lint` y `npm run build` pasan.
