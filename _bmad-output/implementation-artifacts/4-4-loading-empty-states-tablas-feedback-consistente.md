---
story_id: "4.4"
epic: "epic-4"
title: "Loading, Empty States y Tablas con Feedback Consistente"
status: done
estimated_days: 1
priority: medium
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/app/messages/page.tsx
  - frontend/app/broadcasts/page.tsx
  - frontend/components/feedback/data-empty-state.tsx
  - frontend/components/feedback/table-loading-rows.tsx
---

# Story 4.4: Loading, Empty States y Tablas con Feedback Consistente

## Objetivo

Unificar el feedback visual de cargas, tablas y estados vacíos en vistas operativas del panel para que mensajes y broadcasts compartan el mismo lenguaje de interacción.

## Acceptance Criteria

- Messages y Broadcasts comparten un patrón consistente de empty state.
- Messages y Broadcasts comparten un patrón consistente de loading rows.
- Status badges usan variantes semánticas o estilos consistentes sin hardcodes innecesarios.
- Las tablas se sienten más uniformes en spacing y feedback de acciones.
- `npm run lint` y `npm run build` pasan.
