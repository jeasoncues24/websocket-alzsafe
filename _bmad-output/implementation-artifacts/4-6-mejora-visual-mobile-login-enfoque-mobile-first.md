---
story_id: "4.6"
epic: "epic-4"
title: "Mejora Visual Mobile Login — Enfoque Mobile-First"
status: done
estimated_days: 1
priority: high
skills: ["ui-ux-pro-max", "shadcn", "bmad-code-review"]
affects:
  - frontend/app/login/page.tsx
  - frontend/app/login/carousel.tsx
  - frontend/app/globals.css
---

# Story 4.6: Mejora Visual Mobile Login — Enfoque Mobile-First

## Contexto

La vista de login ya mejoró en polish general dentro del Epic 4, pero en móvil todavía se percibe como una adaptación de desktop más que como una experiencia pensada primero para teléfono.

Hoy el login móvil tiene buenas bases funcionales:
- jerarquía razonable,
- formulario claro,
- CTA visible,
- cambio de tema accesible,
- motion básica ya integrada.

Sin embargo, persisten fricciones de enfoque UX en pantallas pequeñas:
1. El bloque principal aún se siente como una **card trasladada** en vez de una pantalla mobile-first.
2. El encabezado móvil comunica poco valor antes del formulario.
3. El contenedor visual compite con el espacio vertical disponible en teléfonos pequeños.
4. El estado “Sistema en línea” aparece, pero todavía no funciona como señal de confianza integrada al flujo.
5. La primera impresión emocional del login en mobile es correcta, pero no memorable ni claramente premium.

## User Story

Como administrador que entra desde un teléfono,
quiero que el login se sienta diseñado específicamente para móvil,
para entender rápido dónde estoy, confiar en el sistema y completar el acceso con menos fricción visual.

## Objetivo UX

Transformar el login mobile desde una composición tipo “card centrada” a una composición **mobile-first basada en flujo vertical**, con mejor jerarquía, más claridad contextual y una sensación visual más ligera, moderna y confiable.

## Scope

### Incluido
- Reorganización visual del layout mobile del login
- Mejora del hero/header mobile
- Ajuste del peso visual de la card/contenedor en mobile
- Integración más fuerte del badge o estado de confianza
- Refinamiento del CTA principal y del espaciado vertical
- Ajustes de copy cortos si mejoran claridad y enfoque

### Excluido
- Cambios del flujo de autenticación
- Nuevos endpoints, validaciones o lógica de negocio
- Reemplazo del login desktop completo
- Rebranding global del producto
- Introducción de librerías nuevas de animación

## Acceptance Criteria

**AC1 — Layout mobile-first más natural:**
**Dado** que el login se usa en pantallas pequeñas
**Cuando** la pantalla se renderiza en viewport móvil
**Entonces** la composición se percibe como una experiencia pensada primero para mobile
**Y** el contenido se organiza como stack vertical claro: identidad, contexto, formulario, acción principal
**Y** se reduce la sensación de “card desktop encajada” en el centro.

**AC2 — Header mobile con mejor contexto y confianza:**
**Dado** que el usuario decide en pocos segundos si está en la vista correcta
**Cuando** entra al login desde móvil
**Entonces** la parte superior comunica con claridad:
- qué producto es,
- qué tipo de acceso está realizando,
- una señal breve de confianza o estado,
**Y** todo ello sin saturar el espacio vertical.

**AC3 — Jerarquía visual más clara en mobile:**
**Dado** que el login actual ya tiene título, subtítulo, estado y formulario
**Cuando** se refine la vista móvil
**Entonces** el orden visual prioriza correctamente:
1. identidad,
2. confianza/contexto,
3. tarea principal,
4. acción primaria,
**Y** los textos secundarios quedan subordinados.

**AC4 — Contenedor visual más ligero en pantallas pequeñas:**
**Dado** que el bloque actual tiene borde, fondo y sombra
**Cuando** se vea en mobile
**Entonces** el contenedor principal reduce peso visual donde convenga
**Y** conserva separación, legibilidad y foco
**Y** no desperdicia altura útil en teléfonos pequeños.

**AC5 — CTA principal más protagonista:**
**Dado** que el objetivo principal de la pantalla es autenticarse
**Cuando** el usuario llegue al final del formulario
**Entonces** el botón principal se siente más decisivo y natural en móvil
**Y** su copy, spacing y jerarquía visual apoyan la acción principal sin ambigüedad.

**AC6 — Footer y elementos secundarios más discretos:**
**Dado** que en mobile cada línea compite por atención
**Cuando** se rendericen footer, mensajes secundarios o ayudas menores
**Entonces** esos elementos mantienen utilidad informativa
**Y** no compiten con el formulario ni con el CTA principal.

**AC7 — Sin regresión en desktop:**
**Dado** que la mejora está orientada a mobile
**Cuando** se visualice el login en desktop
**Entonces** la experiencia existente no se degrada
**Y** el carrusel lateral y la composición desktop siguen viéndose coherentes.

**AC8 — Validación técnica:**
**Dado** que la story modifica la pantalla más visible del frontend
**Cuando** se ejecuta `cd frontend && npm run lint && npm run build`
**Entonces** ambos comandos pasan sin errores.

## Propuesta UX concreta

### Dirección recomendada
- **Minimalismo cálido**
- **Confianza operativa**
- **Menos caja, más flujo vertical**
- **Motion sutil, no decorativa**
- **Jerarquía clara para pantallas pequeñas**

### Estructura mobile sugerida
1. **Topbar compacta**
   - logo
   - toggle de tema
2. **Micro-hero mobile**
   - badge/estado de confianza
   - título corto y claro
   - subtítulo de una línea
3. **Formulario**
   - campos con mejor respiración vertical
   - feedback de error integrado sin romper el ritmo
4. **CTA principal**
   - ancho completo
   - copy más directo si aplica
5. **Pie mínimo**
   - acceso restringido / nota secundaria mucho más discreta

### Copy recomendado de referencia
- Título: `Accede al panel`
- Subtítulo: `Gestiona empresas, sesiones y mensajes desde un solo lugar.`
- Señal de confianza: `Sistema en línea` o `Acceso seguro para administradores`
- CTA: `Entrar al panel`

## Tareas técnicas sugeridas

- [ ] Revisar composición mobile actual de `frontend/app/login/page.tsx`
- [ ] Redefinir la jerarquía visual para viewport `< lg`
- [ ] Ajustar card/contenedor para que en mobile se sienta más ligera
- [ ] Reforzar el bloque superior de identidad y confianza
- [ ] Refinar CTA principal, spacing y footer
- [ ] Verificar que desktop mantenga coherencia con el carrusel existente
- [ ] Ejecutar lint y build

## Archivos o áreas probablemente afectadas

- `frontend/app/login/page.tsx`
- `frontend/app/login/carousel.tsx` (solo si se requiere coherencia leve de copy/tono)
- `frontend/app/globals.css` (solo si hiciera falta un ajuste menor de token/utilidad)

## Pruebas requeridas

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- Revisión visual manual en:
  - móvil pequeño (~375px)
  - móvil medio (~390–430px)
  - desktop
- Verificación en dark mode y light mode

## Riesgos y edge cases

- Sobrecargar la parte superior con demasiados elementos en pantallas bajas.
- Hacer el contenedor demasiado liviano y perder estructura visual.
- Mejorar mobile a costa de romper el balance desktop.
- Alargar demasiado copies y provocar salto de líneas innecesario.

## Dependencias y bloqueos

- Depende conceptualmente del sistema visual y motion ya trabajado en stories 4-1 a 4-5.
- No depende de backend.
- Debe respetar la dirección visual actual del proyecto y los componentes existentes de shadcn.

## Notas de implementación

- Priorizar decisiones específicas para mobile usando breakpoints y composición responsive, no solo escalado proporcional.
- Evitar que el login móvil parezca una simple versión reducida del desktop.
- Si hay que elegir entre “más decoración” o “más claridad”, elegir claridad.
- El éxito de esta story se mide por foco, confianza y naturalidad del flujo, no por cantidad de efectos.
