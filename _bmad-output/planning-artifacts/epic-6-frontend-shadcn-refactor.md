# Epic 6: Refactor Frontend - shadcn/ui Full Adoption

## Overview

Refactorizar el frontend para adoptar completamente shadcn/ui como sistema de componentes. Eliminar CSS manual, usar todos los componentes disponibles de shadcn, y asegurar que el theme light/dark funcione correctamente.

## Status: draft

## Problema

El frontend actual tiene:
- CSS manual en componentes que debería usar shadcn
- Theme light/dark no funciona correctamente
- Componentes shadcn no usados (tabs, select, dialog, sheet)
- Sidebar con estilos propios en lugar de variantes shadcn

## Objetivo

Adopción completa de shadcn/ui:
- Usar todos los componentes shadcn disponibles
- Reemplazar CSS manual con componentes shadcn
- Theme light/dark funcionando con variables CSS correctas
- Animaciones y transiciones de shadcn

## Scope

### Componentes shadcn a integrar completamente:
- `Button` - Todos los variants (default, destructive, outline, secondary, ghost, link)
- `Input` - Con estilos properios de shadcn
- `Card` - Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
- `Table` - Table, TableHeader, TableRow, TableCell, TableBody
- `Tabs` - Tabs, TabsList, TabsTrigger, TabsContent
- `Select` - Select, SelectTrigger, SelectValue, SelectContent, SelectItem
- `Dialog` - Dialog, DialogTrigger, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter
- `Sheet` - Sheet, SheetTrigger, SheetContent, SheetHeader, SheetTitle, SheetDescription
- `Badge` - Todos los variants
- `Skeleton` - Para estados de carga

### Componentes a refactorizar:
1. **Sidebar** - Usar Button de shadcn con variants completos
2. **Login Page** - Card, Input, Button de shadcn
3. **Dashboard** - Cards, Tabs para métricas
4. **Companies** - Table con shadcn, Select para filtros
5. **Messages** - Table, Tabs para filtros de estado
6. **Sessions** - Cards para sesiones, Dialog para acciones
7. **Broadcasts** - Table, Sheet para resultados
8. **Settings** - Tabs para categorías

### Theme:
- Corregir globals.css para Tailwind 4 + shadcn
- Variables CSS con fallback para dark mode
- ThemeProvider configurado correctamente
- Toggle theme en sidebar funcionando

## Métricas de Éxito

- [ ] Theme light/dark funciona correctamente
- [ ] Todos los componentes shadcn usados en lugar de CSS manual
- [ ] Animaciones de shadcn presentes (transitions, hover)
- [ ] No hay CSS inline o estilos manuales en componentes
- [ ] UI consistente en todas las páginas

## Technical Notes

- Tailwind 4 usa `@theme` directive
- shadcn usa CSS variables con `--color-*` prefix
- Theme toggle usa `next-themes` con `attribute="class"`
- Probar en desarrollo con `npm run dev`