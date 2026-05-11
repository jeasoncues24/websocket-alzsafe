---
story_id: "3.8"
epic: "epic-3"
title: "Tailwind Shortcuts, Font Headings y useRouter Destructuring"
status: backlog
estimated_days: 2
priority: normal
skills: ["bmad-code-review"]
affects:
  - frontend/ (132+ archivos para Tailwind, 12 para font, 19 para useRouter)
---

# Story 3.8: Tailwind Shortcuts, Font Headings y useRouter Destructuring

## Contexto

React Doctor detectó **180 issues de Architecture** — la mayoría son aplicables con scripts automatizados:

- **132 casos** de `w-N h-N` → `size-N` (Tailwind v3.4+ shorthand, mismo valor en ambos ejes)
- **12 headings** con `font-bold` → `font-semibold` (mejor legibilidad tipográfica a display sizes)
- **19 componentes** con `useRouter()` sin destructuring → `const { push } = useRouter()`

Esta story puede y debe automatizarse parcialmente con `sed` o scripts para los 132 casos de Tailwind.

## User Story

Como desarrollador del equipo,
quiero que el código Tailwind y los patrones de hooks sigan las convenciones modernas recomendadas,
para que el código sea consistente, más legible y compatible con el React Compiler.

## Scope

### Incluido
- Script de reemplazo para `w-N h-N` → `size-N` en todos los archivos afectados
- Fix de `font-bold` → `font-semibold` en headings
- Refactor de `useRouter()` → `const { push } = useRouter()` en los 19 casos

### Excluido
- Casos donde `w-N` y `h-N` tienen valores diferentes (solo aplica `size-N` cuando ambos son iguales)
- Cambios en el comportamiento visual (todos son equivalentes semánticos)
- Otros issues de Architecture no listados en esta story

## Acceptance Criteria

**AC1 — Tailwind size-N aplicado:**
**Dado** que existen 132 casos de `w-4 h-4`, `w-6 h-6`, `w-8 h-8`, etc. en múltiples archivos
**Cuando** se ejecuta un script de reemplazo (ej. `sed -i 's/w-\([0-9]*\) h-\1/size-\1/g'` en los archivos TSX)
**Entonces** todos los casos donde ambos valores son iguales usan `size-N`
**Y** la UI renderiza exactamente igual (size-N es equivalente a w-N h-N con mismo valor)
**Y** `npm run build` pasa sin errores

**AC2 — Font headings actualizados:**
**Dado** que 12 elementos heading usan `font-bold` (700 weight)
**Cuando** se identifican con react-doctor verbose y se reemplazan por `font-semibold` (600 weight)
**Entonces** los headings usan peso semibold
**Y** React Doctor no reporta "Design no bold heading" en los archivos modificados

**AC3 — useRouter destructurado:**
**Dado** que 19 componentes usan `const router = useRouter()` y luego `router.push(...)`
**Cuando** se refactorizan a `const { push } = useRouter()` y se reemplazan los `router.push` por `push`
**Entonces** cada componente declara explícitamente qué métodos del router necesita
**Y** React Doctor no reporta "React compiler destructure method" en los archivos modificados
**Y** TypeScript no reporta errores en los archivos modificados

**AC4 — Score React Doctor ≥ 90:**
**Dado** que esta es la última story del epic
**Cuando** se ejecuta `npx react-doctor@latest . --verbose`
**Entonces** el score reportado es ≥ 90/100
**Y** el número de issues es significativamente menor a los 312 del baseline

**AC5 — Build y lint pasan:**
**Dado** que todos los cambios están aplicados
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** no hay errores de TypeScript ni ESLint

## Notas de Implementación

- **Para Tailwind (132 casos)**: Crear un script simple antes de ejecutar manualmente:
  ```bash
  # Reemplaza w-N h-N con size-N para valores iguales (1-96)
  find frontend/app frontend/components -name "*.tsx" | xargs sed -i -E 's/\bw-([0-9]+) h-\1\b/size-\1/g'
  find frontend/app frontend/components -name "*.tsx" | xargs sed -i -E 's/\bh-([0-9]+) w-\1\b/size-\1/g'
  ```
  Revisar el diff completo antes de commit. Verificar visualmente que el panel se ve igual.

- **Para font-bold en headings**: Los 12 casos están en headings (`<h1>`, `<h2>`, etc.) — no en otros elementos. No reemplazar `font-bold` globalmente, solo en contexto de heading.

- **Para useRouter**: Atención con componentes que usan múltiples métodos del router (`push` y `replace` por ejemplo). En ese caso, destructurar ambos: `const { push, replace } = useRouter()`.
