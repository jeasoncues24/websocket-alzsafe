---
story_id: "3.7"
epic: "epic-3"
title: "Limpieza de Dead Code"
status: backlog
estimated_days: 1
priority: normal
skills: ["bmad-code-review"]
affects:
  - frontend/ (múltiples archivos — ver react-doctor report)
---

# Story 3.7: Limpieza de Dead Code

## Contexto

React Doctor detectó **38 issues de Dead Code** en el frontend:
- **26 exports sin usar**: símbolos exportados que ningún módulo importa
- **10 types sin usar**: TypeScript interfaces/types definidos pero nunca referenciados
- **2 archivos sin usar**: archivos que no son importados por ningún otro módulo

Correr con `--verbose` para obtener la lista exacta de archivos y símbolos.

## User Story

Como desarrollador del equipo,
quiero que el proyecto esté libre de exports, types y archivos sin usar detectados por React Doctor,
para mantener el bundle limpio y reducir superficie de confusión en el código base.

## Scope

### Incluido
- Eliminar o hacer privados los 26 exports sin usar
- Eliminar los 10 TypeScript types sin usar
- Eliminar los 2 archivos sin usar (previa verificación)

### Excluido
- Refactors adicionales en los archivos afectados
- Dead code en el backend Go (scope separado)

## Acceptance Criteria

**AC1 — Exports auditados y resueltos:**
**Dado** que `npx react-doctor@latest . --verbose` lista los 26 exports sin usar
**Cuando** se audita cada export
**Entonces** si el export es interno y no necesita ser público → se elimina `export` (queda como declaración interna)
**Y** si el export nunca se usa en ningún contexto → se elimina completamente
**Y** React Doctor no reporta "Unused export" en los archivos modificados

**AC2 — Types sin usar eliminados:**
**Dado** que React Doctor lista los 10 TypeScript types sin usar
**Cuando** se verifican con TypeScript compiler que no tienen referencias
**Entonces** los types se eliminan de los archivos correspondientes
**Y** `npm run build` (type-check incluido) pasa sin errores

**AC3 — Archivos sin usar verificados y eliminados:**
**Dado** que React Doctor detectó 2 archivos no importados por ningún módulo
**Cuando** se verifica que:
  - No son entry points de Next.js (páginas en `app/`, archivos especiales como `layout.tsx`, `error.tsx`)
  - No son archivos de configuración implícita
  - No son usados por scripts externos al proyecto
**Entonces** se eliminan del repositorio
**Y** el build y lint pasan sin referencias rotas

**AC4 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores de TypeScript ni ESLint

## Notas de Implementación

- **Primer paso obligatorio**: ejecutar `npx react-doctor@latest . --verbose` y obtener la lista exacta. Los 38 issues tienen ubicaciones específicas de archivo y línea.
- Para exports: buscar con `grep -r "import.*{NombreExport}" frontend/` para confirmar que realmente no hay imports antes de eliminar.
- Para archivos: verificar las convenciones de Next.js App Router — algunos archivos son implícitamente usados por el framework aunque no haya imports explícitos.
- Para types: TypeScript strict mode ayudará a detectar referencias rotas si se elimina algo que sí se usaba.
