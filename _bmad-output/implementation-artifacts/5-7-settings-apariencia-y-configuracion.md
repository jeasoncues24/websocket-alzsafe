# Story 5.7: Settings - Apariencia y Configuración

Status: ready-for-dev

## Story

As a operations manager,
I want personalizar la apariencia del panel,
So that pueda trabajar cómodamente con mi theme preferido.

## Acceptance Criteria

1. **Given** `/settings`, **When** carga, **Then** muestra tabs: Apariencia, General, Acerca de.

2. **Given** tema, **When** usuario toggla light/dark, **Then** cambio instantáneo + persistencia en localStorage via Zustand.

3. **Given** densidad UI, **When** usuario selecciona, **Then** opciones: Compact, Default, Spacious - afecta spacing general.

4. **Given** General, **When** usuario configura, **Then** opciones: auto-refresh interval, items por página, notificaciones.

5. **Given** preferencias, **When** cierro navegador, **Then** mantienen al volver.

## Tasks

- [ ] Page `app/settings/page.tsx`
- [ ] Tabs: Apariencia, General, Acerca de (shadcn Tabs)
- [ ] Theme toggle (light/dark/system)
- [ ] Selector densidad UI
- [ ] Ociones auto-refresh, items por página
- [ ] Sección Acerca de con versión
- [ ] Persistir todo via Zustand + localStorage