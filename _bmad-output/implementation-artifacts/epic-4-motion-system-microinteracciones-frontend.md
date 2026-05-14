# Epic 4: Motion System, Microinteracciones y Pulido UX del Frontend

Status: done

## Objetivo

Elevar la percepción de calidad del panel administrativo incorporando un sistema consistente de animaciones, transiciones suaves y feedback visual significativo en navegación, formularios, tablas, overlays y estados de carga, sin alterar la lógica de negocio existente.

## Evaluación actual

### Resumen ejecutivo

El frontend ya tiene una base visual funcional, pero la capa de motion está fragmentada: existen transiciones aisladas, loaders y algunos efectos en login, pero no hay un sistema de animación compartido ni una gramática consistente de entrada/salida, hover, press, loading y cambios de estado.

### Hallazgos principales

1. **Hay animaciones puntuales, pero no un sistema global de motion.**
   - `frontend/app/login/carousel.tsx` usa `transition-transform duration-500 ease-in-out`.
   - `frontend/components/layout/sidebar.tsx` usa `transition-all duration-300`.
   - `frontend/components/ui/button.tsx` solo define `transition-colors`.
   - `frontend/app/globals.css` no define tokens de duración, easing, reduced-motion ni utilidades semánticas.

2. **La navegación principal sigue sintiéndose rígida.**
   - El shell autenticado en `frontend/components/admin-auth-check.tsx` cambia entre loader/login/app sin transición perceptual.
   - Sidebar y navegación móvil resuelven layout, pero no continuidad espacial entre estados (`sidebar.tsx`, `mobile-nav.tsx`).

3. **Los estados de carga existen, pero son heterogéneos.**
   - Hay spinners ad-hoc en `admin-auth-check.tsx`, `login/page.tsx`, `dashboard/page.tsx`, `qr/page.tsx`.
   - Hay skeletons en algunos módulos, pero no una política consistente por tipo de pantalla.

4. **Existen microinteracciones útiles, pero aisladas y desbalanceadas.**
   - Flecha con hover en `frontend/app/login/page.tsx`.
   - Flash de autofill y panel colapsable en `frontend/components/companies/empresa-form-modal.tsx`.
   - Tabs y badges usan estados visuales, pero sin ritmo compartido entre páginas.

5. **Hay deuda de composición shadcn que afecta el polish general.**
   - Uso extendido de `space-y-*` en páginas y formularios (`dashboard/page.tsx`, `messages/page.tsx`, `empresa-form-modal.tsx`, etc.), contrario a la guía del skill.
   - Formularios con `label` + `div` manual en lugar de `FieldGroup` + `Field`.
   - Estados visuales con colores hardcodeados en badges, por ejemplo `frontend/app/messages/page.tsx` y `frontend/app/broadcasts/page.tsx`.

6. **No hay estrategia visible de accesibilidad para motion.**
   - No se observa soporte explícito para `prefers-reduced-motion` en `globals.css`.
   - Varias animaciones usan `animate-ping`, `animate-spin` o `transition-all` sin una capa semántica reutilizable.

## Oportunidad de mejora

Si se sistematiza la motion layer, el panel puede verse notablemente más moderno sin rediseñar toda la UI. El mayor retorno está en:
- transiciones de navegación y shell,
- feedback de formularios y acciones async,
- estados vacíos/carga,
- estandarización de timing/easing,
- limpieza de patrones shadcn que hoy rompen consistencia.

## Alcance incluido

- Definir tokens/utilidades globales de motion en `frontend/app/globals.css`.
- Estandarizar duraciones, easing y reglas de reduced-motion.
- Mejorar microinteracciones de sidebar, mobile nav, cards, tabs, botones y formularios.
- Unificar loading states, skeletons y feedback de acciones async.
- Pulir login y pantallas principales con animación sutil y consistente.
- Corregir patrones de composición shadcn que impactan UX visual y consistencia.

## Fuera de alcance

- Rebranding completo.
- Reescritura funcional de flujos de negocio backend/frontend.
- Introducir animaciones decorativas pesadas o librerías de motion complejas sin necesidad.
- Refactor completo de todos los componentes custom del proyecto en una sola story.

## Dependencias

- Base actual Next.js + Tailwind v4 + shadcn local.
- Alineación con `frontend/components.json` y alias `@/components/ui`.
- Revisión por página para no romper estados existentes de dashboard, mensajes, empresas, sesiones y QR.

## Riesgos conocidos

- Sobrecargar la UI con motion decorativo en lugar de motion funcional.
- Introducir regresiones visuales por tocar componentes base (`button`, `tabs`, `dialog`, `sheet`).
- Mezclar mejoras de motion con cambios de layout muy grandes y perder foco.

## Criterios de éxito

- Existe un sistema de motion reutilizable con durations/easings/tokens claros.
- Navegación, overlays, tabs, cards y formularios comparten ritmo visual consistente.
- Los estados async importantes muestran feedback inmediato y suave.
- Se respeta `prefers-reduced-motion`.
- `cd frontend && npm run lint` y `cd frontend && npm run build` pasan sin errores.
- El panel mejora perceptiblemente sin cambiar reglas de negocio ni contratos de API.

## Stories propuestas

| ID  | Nombre | Tipo | Prioridad | Estado actual | Nota |
|-----|--------|------|-----------|---------------|------|
| 4-1 | Auditoría UX motion + tokens globales | Frontend | Alta | done | Base semántica de durations, easing, reduced-motion y checklist de adopción |
| 4-2 | Smooth navigation del shell admin | Frontend | Alta | done | Sidebar, mobile nav, transiciones del shell y continuidad entre estados |
| 4-3 | Microinteracciones de formularios y modales | Frontend | Alta | done | Feedback de focus, validación, submit, autofill, dialogs y sheets |
| 4-4 | Loading, empty states y tablas con feedback consistente | Frontend | Media | done | Skeletons, empty states, acciones async y feedback contextual |
| 4-5 | Pulido motion del login y dashboard | Frontend | Media | done | Hero/login, cards métricas, tabs, refresh y entrada visual inicial |
| 4-6 | Mejora visual mobile login — enfoque mobile-first | Frontend | Alta | done | Reenfoca el login móvil como experiencia pensada primero para teléfono |
| 4-7 | Animaciones de apertura y cierre para modales y overlays | Frontend | Alta | done | Unifica el lenguaje de motion para Dialog y Sheet con foco en open/close |

## Orden recomendado de implementación

```text
4-1 → prerequisito para 4-2, 4-3, 4-4, 4-5
4-2 → habilita consistencia del shell y navegación principal
4-3 → habilita consistencia de formularios y overlays
4-4 → extiende feedback al resto del panel operativo
4-5 → remata las pantallas de mayor visibilidad
4-6 → profundiza el login móvil tras el polish general de 4-5
4-7 → refina overlays y cambio de contexto en dialogs/sheets tras el polish general
```

## Condición de cierre del epic

El epic quedó cerrado con las stories definidas en `done`, una capa de motion consistente y accesible en el frontend, y mejoras visibles en login, dashboard, navegación, formularios, tablas y estados de feedback del panel.
