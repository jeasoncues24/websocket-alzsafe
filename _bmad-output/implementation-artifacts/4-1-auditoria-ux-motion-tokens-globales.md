---
story_id: "4.1"
epic: "epic-4"
title: "Auditoría UX Motion + Tokens Globales"
status: done
estimated_days: 1
priority: high
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/app/globals.css
  - frontend/components/ui/button.tsx
  - frontend/components/layout/sidebar.tsx
  - frontend/components/layout/mobile-nav.tsx
  - frontend/app/login/page.tsx
  - frontend/app/login/carousel.tsx
  - frontend/app/dashboard/page.tsx
  - frontend/components/admin-auth-check.tsx
---

# Story 4.1: Auditoría UX Motion + Tokens Globales

## Contexto

El frontend del panel ya incluye animaciones aisladas, pero no existe un sistema de motion compartido. Hoy se observan patrones dispersos:

- `frontend/app/login/carousel.tsx` usa `transition-transform duration-500 ease-in-out`.
- `frontend/components/layout/sidebar.tsx` usa `transition-all duration-300` para colapsar el sidebar.
- `frontend/components/ui/button.tsx` solo define `transition-colors`.
- `frontend/components/admin-auth-check.tsx` usa un loader ad-hoc sin transición de entrada/salida.
- `frontend/app/globals.css` no define tokens semánticos de duración, easing, focus motion ni reglas explícitas para `prefers-reduced-motion`.

Además, la base actual presenta deuda de consistencia visual relacionada con shadcn:

- Uso extendido de `space-y-*` donde el proyecto debería migrar a `flex flex-col gap-*`.
- Formularios manuales fuera del patrón `FieldGroup` + `Field`.
- Estados visuales y colores hardcodeados en algunas páginas (`messages`, `broadcasts`).

Esta story no busca rehacer toda la UI. Su objetivo es **crear la base técnica y de criterios** sobre la que se apoyarán las siguientes stories del Epic 4.

## User Story

Como equipo de frontend,
quiero una auditoría clara de motion y una capa global de tokens reutilizables,
para poder mejorar animaciones y microinteracciones del panel de forma consistente, accesible y sin improvisación componente por componente.

## Scope

### Incluido
- Auditoría documentada de los patrones actuales de motion y feedback visual del panel
- Definición de tokens globales de motion en `frontend/app/globals.css`
- Definición de una política base de `prefers-reduced-motion`
- Estandarización mínima en componentes base/pantallas clave como referencia inicial
- Checklist de adopción para las stories 4-2, 4-3, 4-4 y 4-5

### Excluido
- Reanimar todas las páginas del panel en esta story
- Agregar librerías nuevas como Framer Motion sin justificación adicional
- Refactor completo de formularios al sistema `FieldGroup` + `Field`
- Reemplazo total de todos los `space-y-*` del proyecto en esta iteración

## Acceptance Criteria

**AC1 — Auditoría motion documentada:**
**Dado** que el panel tiene motion fragmentado entre login, navegación, loaders, cards y tabs
**Cuando** se revisan las pantallas y componentes principales del frontend
**Entonces** queda documentado en esta story un inventario mínimo de:
- patrones ya existentes,
- gaps de consistencia,
- componentes prioritarios,
- anti-patrones a evitar,
- prioridad de adopción para el resto del epic.

**AC2 — Tokens globales definidos en globals.css:**
**Dado** que `frontend/app/globals.css` es el archivo global de Tailwind v4 y tokens del proyecto
**Cuando** se implementa la base de motion global
**Entonces** ese archivo define tokens reutilizables para al menos:
- duraciones cortas, medias y largas,
- easing estándar de entrada/salida,
- transición de focus/press/hover,
- una utilidad o convención clara de motion consistente,
**Y** no se crea un archivo CSS global paralelo.

**AC3 — Reduced motion soportado:**
**Dado** que las mejoras del epic introducirán más motion en la UI
**Cuando** el usuario tenga activo `prefers-reduced-motion`
**Entonces** las animaciones no esenciales se reducen o desactivan mediante la capa global definida en `globals.css`
**Y** spinners críticos o feedback funcional siguen siendo comprensibles sin depender de motion compleja.

**AC4 — Base de referencia aplicada a puntos visibles del sistema:**
**Dado** que esta story debe probar que los tokens globales son utilizables
**Cuando** se ajustan algunos puntos representativos del frontend
**Entonces** al menos estas áreas usan el nuevo criterio global o quedan alineadas con él:
- `components/ui/button.tsx` para hover/press/focus consistentes,
- `components/layout/sidebar.tsx` o `mobile-nav.tsx` para navegación,
- `app/login/page.tsx` o `app/login/carousel.tsx` como caso visible de motion refinado,
**Y** los cambios siguen siendo sutiles, funcionales y coherentes con shadcn.

**AC5 — Anti-patrones explícitamente prohibidos:**
**Dado** que la meta es mejorar perceived quality sin degradar performance ni accesibilidad
**Cuando** se deje lista la base del sistema motion
**Entonces** quedan explícitamente prohibidos para stories futuras:
- `transition-all` salvo caso justificado,
- animar `width/height/top/left` cuando pueda usarse `transform/opacity`,
- duraciones > 500ms para microinteracciones,
- animación decorativa sin significado,
- ignorar `prefers-reduced-motion`.

**AC6 — Validación técnica:**
**Dado** que la story modifica estilos globales y componentes base visibles
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** ambos comandos pasan sin errores nuevos.

## Tareas técnicas sugeridas

- [ ] Auditar login, dashboard, navegación, formularios y loaders para mapear motion actual
- [ ] Definir tokens globales en `frontend/app/globals.css` usando la capa existente del proyecto
- [ ] Añadir política base de `prefers-reduced-motion`
- [ ] Aplicar la base a 2–4 puntos visibles de referencia
- [ ] Documentar reglas de uso y anti-patrones dentro de esta story
- [ ] Ejecutar lint y build del frontend

## Archivos o áreas probablemente afectadas

- `frontend/app/globals.css`
- `frontend/components/ui/button.tsx`
- `frontend/components/layout/sidebar.tsx`
- `frontend/components/layout/mobile-nav.tsx`
- `frontend/app/login/page.tsx`
- `frontend/app/login/carousel.tsx`
- `frontend/app/dashboard/page.tsx`
- `frontend/components/admin-auth-check.tsx`

## Pruebas requeridas

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- Validación manual visual en:
  - `/login`
  - `/dashboard`
  - navegación desktop sidebar
  - navegación móvil sheet/topbar
- Verificación manual con `prefers-reduced-motion` activado

## Riesgos y edge cases

- Tocar `globals.css` puede afectar toda la app si los tokens quedan mal nombrados o muy agresivos.
- Ajustar `button.tsx` puede propagar cambios visuales a todas las acciones del panel.
- Usar motion exagerada puede deteriorar performance o hacer que la UI se sienta menos profesional.
- Aplicar animación solo en light mode o sin revisar dark mode puede romper consistencia visual.

## Dependencias y bloqueos

- No depende de backend.
- Sirve como prerequisito conceptual para `4-2`, `4-3`, `4-4` y `4-5`.
- Debe respetar las reglas del skill shadcn: semántica de tokens, no hardcodear colores, no usar overrides visuales arbitrarios.

## Notas de implementación

- Preferir tokens semánticos y utilidades globales sobre clases ad-hoc repetidas en cada pantalla.
- En componentes base, priorizar `transition-[color,box-shadow,transform,opacity]` o equivalente específico en vez de `transition-all`.
- Mantener las animaciones de microinteracción en un rango aproximado de 150–300ms.
- Usar motion para comunicar estado: hover, press, focus, apertura/cierre, loading y cambio de contexto.
- Esta story puede dejar algunos hallazgos documentados como follow-up directo para las stories posteriores del epic.
