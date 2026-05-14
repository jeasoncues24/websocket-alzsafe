---
story_id: "4.2"
epic: "epic-4"
title: "Smooth Navigation del Shell Admin"
status: done
estimated_days: 1
priority: high
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/components/admin-auth-check.tsx
  - frontend/components/layout/sidebar.tsx
  - frontend/components/layout/mobile-nav.tsx
  - frontend/components/layout/nav-items.ts
---

# Story 4.2: Smooth Navigation del Shell Admin

## Contexto

Tras la story 4.1 ya existe una base de motion global, pero el shell autenticado sigue sintiéndose rígido en cambios de ruta y navegación principal. Sidebar y mobile nav funcionan, aunque carecen de continuidad espacial, indicador activo más claro y una fuente única de verdad para sus items.

## User Story

Como administrador que navega muchas pantallas del panel,
quiero que el shell de navegación se sienta más fluido y coherente,
para ubicarme mejor, percibir continuidad entre vistas y reducir fricción al moverme entre módulos.

## Acceptance Criteria

**AC1 — Fuente única de navegación:**
Sidebar y MobileNav consumen la misma definición de items desde un módulo compartido.

**AC2 — Estado activo más claro:**
El item activo en sidebar y mobile nav tiene un indicador más evidente y consistente, sin usar colores hardcodeados.

**AC3 — Continuidad visual del shell:**
El shell autenticado anima suavemente entrada/cambio de contenido principal y evita sensación de salto brusco.

**AC4 — Mobile nav contextual:**
La topbar móvil muestra el contexto actual de la sección y el sheet se cierra de forma controlada al navegar.

**AC5 — Validación técnica:**
`cd frontend && npm run lint && npm run build` pasan sin errores.
