# Story 6-1: Corregir Theme Light/Dark y globals.css

## Parent Epic: Epic 6 - Frontend shadcn Refactor

## Objetivo

Corregir el globals.css para que el theme light/dark funcione correctamente con Tailwind 4 y shadcn/ui.

## Tareas

- [ ] globals.css usa `@theme` directive de Tailwind 4
- [ ] Variables CSS con `--color-*` prefix para shadcn
- [ ] Dark mode con clase `.dark` en lugar de media query
- [ ] ThemeProvider con `attribute="class"` en todos los layouts
- [ ] Toggle theme en sidebar funciona

## Criterios de Aceptación

1. Al hacer click en toggle de theme, los colores cambian correctamente
2. Las variables CSS se reflejan en todos los componentes shadcn
3. No hay flash de contenido sin estilos (no flash of unstyled content)
4. El theme se persiste en localStorage

## Dependencies

- Ninguna (story independiente)

## Estimated Effort

1 hour