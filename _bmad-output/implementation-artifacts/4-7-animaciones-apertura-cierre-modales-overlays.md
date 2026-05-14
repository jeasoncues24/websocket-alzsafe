---
story_id: "4.7"
epic: "epic-4"
title: "Animaciones de Apertura y Cierre para Modales y Overlays"
status: done
estimated_days: 1
priority: high
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/components/ui/dialog.tsx
  - frontend/components/ui/sheet.tsx
  - frontend/components/companies/empresa-form-modal.tsx
  - frontend/components/companies/telefono-form-modal.tsx
---

# Story 4.7: Animaciones de Apertura y Cierre para Modales y Overlays

## Contexto

El frontend ya avanzó en motion, navegación, formularios y polish general, pero la experiencia de overlays todavía puede sentirse algo mecánica al abrir o cerrar modales. Aunque `Dialog` y `Sheet` ya usan clases de animación base, falta una intención UX más clara y consistente en la forma en que aparecen y desaparecen:

- la entrada no siempre comunica suficiente continuidad espacial,
- el cierre puede sentirse abrupto,
- el overlay no siempre acompaña visualmente el foco del usuario,
- no hay una regla explícita de ritmo para open/close de modales del sistema.

En términos de percepción, los modales son momentos de cambio de contexto. Si ese cambio entra y sale con mejor cadencia, el panel se siente más premium, más claro y menos rígido.

## User Story

Como usuario del panel,
quiero que los modales y overlays abran y cierren con animaciones suaves y coherentes,
para entender mejor el cambio de foco, percibir continuidad visual y sentir la interfaz más pulida.

## Objetivo UX

Definir y aplicar una animación consistente para apertura/cierre de modales, dialogs y sheets, reforzando:
- foco visual,
- continuidad espacial,
- claridad del cambio de contexto,
- y una sensación de UI más elegante sin caer en animación decorativa.

## Scope

### Incluido
- Ajuste de animaciones base en `Dialog` y `Sheet`
- Revisión del ritmo de overlay + content
- Refinamiento de duración/easing de open/close
- Consistencia entre modal central y panel lateral/sheet
- Validación visual en modales clave del login/admin si aplica

### Excluido
- Reescribir lógica de apertura/cierre
- Introducir librerías nuevas de motion
- Animaciones exageradas, elásticas o decorativas
- Rehacer todos los componentes overlay del sistema fuera de `Dialog` y `Sheet`

## Acceptance Criteria

**AC1 — Apertura con continuidad espacial:**
**Dado** que el usuario abre un modal o sheet
**Cuando** el overlay aparece
**Entonces** la entrada del contenido comunica claramente que un nuevo foco tomó prioridad
**Y** la transición se siente suave, legible y consistente con el sistema visual.

**AC2 — Cierre menos abrupto:**
**Dado** que el usuario cierra un modal o sheet
**Cuando** el contenido desaparece
**Entonces** el cierre se percibe más natural que en la versión actual
**Y** la duración de salida es ligeramente más rápida que la de entrada para mantener sensación de respuesta.

**AC3 — Overlay acompaña el cambio de contexto:**
**Dado** que el fondo queda despriorizado al abrir un modal
**Cuando** el overlay entra o sale
**Entonces** su transición acompaña el foco sin competir con el contenido principal
**Y** no produce sensación de parpadeo o corte duro.

**AC4 — Consistencia entre Dialog y Sheet:**
**Dado** que el sistema usa modal central y panel lateral
**Cuando** se comparan sus animaciones
**Entonces** ambos comparten una misma lógica de motion
**Y** solo cambia la dirección/contexto espacial, no el lenguaje visual base.

**AC5 — Respeto por reduced motion:**
**Dado** que el sistema ya tiene una base de motion accesible
**Cuando** el usuario tenga `prefers-reduced-motion`
**Entonces** las animaciones de open/close se reducen o simplifican apropiadamente.

**AC6 — Sin regresión funcional:**
**Dado** que los modales son críticos para CRUD y navegación contextual
**Cuando** se implementen los cambios de motion
**Entonces** abrir, cerrar, overlay click, Escape y foco siguen funcionando igual que antes.

**AC7 — Validación técnica:**
**Dado** que la story toca componentes base reutilizados por todo el frontend
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** ambos comandos pasan sin errores.

## Propuesta UX concreta

### Dirección recomendada
- Entrada: **fade + scale/translate sutil**
- Salida: **misma lógica, un poco más corta**
- Overlay: **fade progresivo breve**
- Sheet: **slide contextual con easing consistente**
- Nada de rebotes, overshoot ni animaciones teatrales

### Principios
1. El usuario debe sentir cambio de contexto, no espectáculo.
2. El contenido debe entrar como capa prioritaria.
3. La salida debe sentirse ágil.
4. Overlay y panel deben moverse como una sola escena.

## Tareas técnicas sugeridas

- [ ] Auditar `frontend/components/ui/dialog.tsx`
- [ ] Auditar `frontend/components/ui/sheet.tsx`
- [ ] Ajustar durations/easing de open/close
- [ ] Refinar relación overlay/content
- [ ] Validar modales de empresa y teléfono como casos reales
- [ ] Verificar reduced motion
- [ ] Ejecutar lint y build

## Archivos o áreas probablemente afectadas

- `frontend/components/ui/dialog.tsx`
- `frontend/components/ui/sheet.tsx`
- `frontend/components/companies/empresa-form-modal.tsx`
- `frontend/components/companies/telefono-form-modal.tsx`

## Pruebas requeridas

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- Revisión manual de:
  - apertura/cierre con click en trigger
  - cierre con overlay
  - cierre con Escape
  - apertura/cierre de Sheet mobile nav
  - apertura/cierre de modales CRUD

## Riesgos y edge cases

- Afinar demasiado la duración puede hacer la UI lenta.
- Animación excesiva en overlays puede sentirse pesada.
- Cambios en componentes base pueden impactar múltiples pantallas si no se validan bien.
- Diferencias de timing entre Dialog y Sheet pueden romper la percepción de sistema unificado.

## Dependencias y bloqueos

- Depende conceptualmente de las stories previas del Epic 4.
- No depende de backend.
- Debe respetar los tokens de motion ya definidos en `globals.css`.

## Notas de implementación

- Priorizar `transform` y `opacity` sobre propiedades con reflow.
- Usar salida un poco más rápida que entrada.
- Mantener la animación suficientemente visible para comunicar contexto, pero suficientemente sutil para no distraer.
