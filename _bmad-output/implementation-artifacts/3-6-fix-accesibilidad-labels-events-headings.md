---
story_id: "3.6"
epic: "epic-3"
title: "Fix Accesibilidad — Labels, Key Events y Headings"
status: backlog
estimated_days: 1
priority: medium
skills: ["bmad-code-review"]
affects:
  - frontend/components/ui/label.tsx
  - frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/api-keys/page.tsx
  - frontend/components/ui/alert.tsx
---

# Story 3.6: Fix Accesibilidad — Labels, Key Events y Headings

## Contexto

React Doctor detectó 28 issues de accesibilidad. Los más impactantes:

1. **26 labels sin `htmlFor`** en `components/ui/label.tsx:7` — el componente base `Label` no propaga `htmlFor` al elemento `<label>`, haciendo que todos los formularios del panel sean inaccesibles por teclado/lector de pantalla.
2. **Click sin key event** en `api-keys/page.tsx:354` — usuario de teclado no puede activar ese control.
3. **Heading sin contenido** en `components/ui/alert.tsx:28` — lector de pantalla no puede anunciar el heading del Alert.

## User Story

Como usuario del panel admin que depende de teclado o lector de pantalla,
quiero que los formularios y controles interactivos sean completamente accesibles por teclado,
para poder operar la plataforma sin necesidad de mouse.

## Scope

### Incluido
- Fix del componente base `Label` en `components/ui/label.tsx`
- Fix del click element en `api-keys/page.tsx:354`
- Fix del heading vacío en `components/ui/alert.tsx:28`

### Excluido
- Auditoría completa de accesibilidad del panel (scope de un epic de UX futuro)
- Cambios en el diseño visual de los componentes
- Tests automatizados de accesibilidad (puede ser story futura)

## Acceptance Criteria

**AC1 — Label base propaga htmlFor:**
**Dado** que `components/ui/label.tsx:7` define el componente `Label` con `React.forwardRef`
**Cuando** se usa el componente con `htmlFor="campo-id"`
**Entonces** el elemento `<label>` renderizado tiene el atributo `for="campo-id"` en el DOM
**Y** ESLint regla `jsx-a11y/label-has-associated-control` no reporta warnings para usos de `Label` con `htmlFor`
**Y** los 26 usos existentes de `Label` en el proyecto funcionan correctamente sin cambios en los call sites

**AC2 — Click element con key event:**
**Dado** que `app/empresas/[empresaId]/telefonos/[telefonoId]/api-keys/page.tsx:354` tiene un elemento con `onClick` sin handler de teclado
**Cuando** se agrega `onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { /* misma acción */ } }}`
**Entonces** el elemento puede activarse con Enter y Space desde el teclado
**Y** ESLint regla `jsx-a11y/click-events-have-key-events` pasa sin warnings

**AC3 — Alert heading con contenido:**
**Dado** que `components/ui/alert.tsx:28` tiene un elemento heading (`<h*>`) vacío o con solo elementos decorativos
**Cuando** se agrega contenido textual accesible (o se usa `aria-label` si el contenido es dinámico)
**Entonces** el heading tiene contenido legible para screen readers
**Y** ESLint regla `jsx-a11y/heading-has-content` pasa sin warnings

**AC4 — Lint pasa sin warnings jsx-a11y:**
**Dado** que todos los fixes de accesibilidad están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint`
**Entonces** no hay warnings de la categoría `jsx-a11y` en los archivos modificados

**AC5 — Build sin errores:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run build`
**Entonces** no hay errores de TypeScript

## Notas de Implementación

- Para `Label`: el componente usa `React.forwardRef` y probablemente `cva`. Verificar que el tipo de props incluya `htmlFor?: string` (debería heredarlo de `React.LabelHTMLAttributes<HTMLLabelElement>`).
- Para el click element en `api-keys/page.tsx:354`: identificar qué elemento es (¿un `<div>`, `<span>`?). Si no es semánticamente un botón, considerar cambiarlo a `<button>` con estilos apropiados — es más correcto que agregar key events a un div.
- Para `Alert`: verificar si el heading es parte del API público del componente (¿el título viene de props?) para determinar el fix correcto.
