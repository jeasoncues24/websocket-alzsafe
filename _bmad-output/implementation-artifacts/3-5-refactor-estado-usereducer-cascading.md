---
story_id: "3.5"
epic: "epic-3"
title: "Refactor Estado — useReducer y Eliminar Cascading setState"
status: backlog
estimated_days: 2
priority: high
skills: ["bmad-code-review"]
affects:
  - frontend/app/empresas/[empresaId]/telefonos/page.tsx
  - frontend/components/api-key-metrics.tsx
---

# Story 3.5: Refactor Estado — useReducer y Eliminar Cascading setState

## Contexto

React Doctor detectó 28 issues de State & Effects, con los más críticos en:

1. **`CompanyPhonesPage`** (`app/empresas/[empresaId]/telefonos/page.tsx:21`): 7 `useState` relacionados entre sí → `useReducer` recomendado.
2. **Cascading setState** (`page.tsx:34`): 8 `setState` dentro de un único `useEffect` → renders intermedios con estado parcial.
3. **Effect chain** (`components/api-key-metrics.tsx:245`): un `useEffect` reacciona a state seteado por otro `useEffect` → render extra por cada link.

## User Story

Como desarrollador del equipo,
quiero que `CompanyPhonesPage` y `api-key-metrics` gestionen su estado de forma consolidada,
para que los renders sean predecibles, los effects no se encadenen y el código sea mantenible.

## Scope

### Incluido
- Refactor de 7 `useState` → `useReducer` en `CompanyPhonesPage`
- Consolidar el cascading setState del `useEffect` en `page.tsx:34`
- Eliminar el effect chain en `api-key-metrics.tsx:245`

### Excluido
- Cambios en la lógica de negocio de telefonos, QR o reconexión
- Cambios en la UI o el diseño visual del componente
- Migrar otros componentes con useState (se abordan en la misma story solo los 2 archivos identificados)

## Acceptance Criteria

**AC1 — useReducer en CompanyPhonesPage:**
**Dado** que `app/empresas/[empresaId]/telefonos/page.tsx:21` tiene 7 `useState` relacionados
**Cuando** se define un tipo `PhonesState` con todos los campos de estado y un `PhonesAction` union type
**Entonces** el componente usa `const [state, dispatch] = useReducer(phonesReducer, initialState)`
**Y** todos los 7 campos se gestionan a través de `dispatch`
**Y** el comportamiento visible del componente (carga, QR, reconexión, modales) es idéntico al anterior
**Y** React Doctor no reporta "UseReducer" warning en `telefonos/page.tsx`

**AC2 — Cascading setState eliminado:**
**Dado** que el `useEffect` en `page.tsx:34` ejecuta 8 `setState` separados
**Cuando** se consolida usando `dispatch` del reducer con una acción que actualiza todos los campos en un solo paso
**Entonces** el componente realiza una sola actualización de estado por ciclo del effect
**Y** no hay renders intermedios con estado parcialmente actualizado
**Y** React Doctor no reporta "Cascading set state" en `telefonos/page.tsx`

**AC3 — Effect chain eliminado en api-key-metrics:**
**Dado** que `components/api-key-metrics.tsx:245` tiene un `useEffect` que reacciona a state seteado por otro `useEffect`
**Cuando** se identifica el valor derivado y se mueve el cómputo al render (si es síncrono) o se consolida en un solo `useEffect`
**Entonces** no hay chain de effects que cause renders extras
**Y** React Doctor no reporta "Effect chain" en `api-key-metrics.tsx`

**AC4 — Sin regresiones funcionales:**
**Dado** que `CompanyPhonesPage` maneja lógica de QR, reconexión y gestión de teléfonos
**Cuando** se verifica manualmente el flujo después del refactor
**Entonces** la carga de teléfonos funciona correctamente
**Y** la generación y visualización de QR funciona correctamente
**Y** el botón de reconexión funciona correctamente
**Y** los modales de crear/editar teléfono funcionan correctamente

**AC5 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores de TypeScript ni ESLint

## Notas de Implementación

- Leer `app/empresas/[empresaId]/telefonos/page.tsx` completo antes de refactorizar — entender el flujo QR y reconexión antes de tocar el estado.
- El `useReducer` debe tener un reducer puro; no hacer side effects en el reducer.
- Para el effect chain en `api-key-metrics.tsx`: identificar qué calcula el segundo effect y si puede derivarse durante el render sin useState adicional.
- Ejecutar prueba manual del flujo completo: navegar a Empresas → Teléfonos → verificar QR → reconectar.
