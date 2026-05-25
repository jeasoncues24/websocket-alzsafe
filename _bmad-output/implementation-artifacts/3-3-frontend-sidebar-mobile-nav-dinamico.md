# Story 3-3: Frontend — Sidebar y mobile-nav dinámico

**Estado:** review
**Epic padre:** Epic 3 — Módulos Dinámicos desde BD y Perfil de Usuario
**Story ID:** 3.3

---

## Story

Como **Usuario Autenticado**,
quiero **que el menú de navegación (sidebar y móvil) se filtre dinámicamente y se apliquen guardas de ruta**,
para **que solo pueda ver y navegar hacia las secciones autorizadas para mi cuenta en el panel**.

---

## Acceptance Criteria

**AC1 — El Sidebar renderiza dinámicamente los módulos autorizados**
Dado un usuario autenticado cuyos módulos permitidos en el store Zustand son `["dashboard", "companies", "broadcasts"]`,
cuando visualiza el panel en desktop,
entonces el Sidebar muestra únicamente los botones para Dashboard, Empresas y Difusiones. Módulos como Usuarios, Roles o Sesiones no se muestran bajo ninguna circunstancia.

**AC2 — El Mobile Nav renderiza dinámicamente en resoluciones móviles**
Dado el mismo usuario en una pantalla móvil,
cuando abre el menú deslizable lateral móvil,
entonces se muestran únicamente los botones autorizados, manteniendo el mismo filtro que en desktop.

**AC3 — Guarda de navegación en frontend bloquea accesos directos**
Dado un usuario no-root que intenta acceder manualmente escribiendo una URL en el navegador (ej. `/users` o `/roles`),
cuando la sección/slug correspondiente a esa URL no está en su listado de `allowedModules`,
entonces el sistema lo intercepta en el cliente y lo redirige automáticamente a `/dashboard` (opcionalmente mostrando una notificación visual de acceso no autorizado).

**AC4 — Usuario root tiene acceso irrestricto**
Dado un usuario con `is_root = true` o cuyos permisos contienen el comodín especial,
cuando navega por el panel,
entonces visualiza la totalidad de los 9 módulos y puede acceder a cualquiera sin ser redirigido.

**AC5 — Fallback de navegación seguro**
Dado un usuario con un perfil inválido o vacío de permisos (o cuyo rol no tenga módulos asignados),
cuando carga el panel de control,
entonces visualiza únicamente la sección `/dashboard` como fallback mínimo obligatorio.

---

## Tasks / Subtasks

- [ ] **T1 — Modificar el menú lateral Sidebar**
  - **Archivo:** `frontend/components/layout/sidebar.tsx`
  - [ ] Consumir `allowedModules` y el indicador `user` (para validar `is_root`) de `useAppStore`.
  - [ ] Filtrar el array de `navItems` importado de `nav-items.ts` comparando su campo `slug` o `id` con la lista de `allowedModules` (salvo que sea `is_root` o el slug sea `"dashboard"`).
  - [ ] Asegurar transiciones visuales fluidas al renderizar el menú.

- [ ] **T2 — Modificar el menú móvil MobileNav**
  - **Archivo:** `frontend/components/layout/mobile-nav.tsx`
  - [ ] Consumir el estado del store y aplicar exactamente el mismo filtro dinámico a los ítems del menú.
  - [ ] Verificar que al pulsar sobre un ítem se cierre el menú correctamente y no cause parpadeos de carga.

- [ ] **T3 — Desarrollar la Guarda de Rutas en el Cliente**
  - **Archivos:** `frontend/app/layout.tsx`, `frontend/components/layout/shell.tsx` (o un wrapper de layout que proteja las páginas admin)
  - [ ] Identificar el layout o shell común de todas las páginas protegidas.
  - [ ] En un hook `useEffect` o mediante el router de Next.js, obtener la ruta actual (pathname) y resolver a qué slug de módulo corresponde.
  - [ ] Si la ruta requiere un módulo específico y dicho módulo no está en `allowedModules` (y `is_root` es falso), realizar un `router.push('/dashboard')` de inmediato.
  - [ ] Proteger contra parpadeos visuales (ej. ocultar el contenido de la página protegida mostrando un spinner de carga en el milisegundo en que se decide si se redirige o no).

---

## Dev Notes

- **Diseño Responsivo y UX Pro Max:** Los menús de navegación deben lucir premium, con sutiles efectos hover (transiciones de opacidad y color de fondo), íconos perfectamente alineados y atajos de teclado o colapsabilidad fluida.
- **Rutas a proteger:**
  - `/dashboard` -> slug `dashboard` (siempre accesible).
  - `/companies` -> slug `companies`.
  - `/users` -> slug `users`.
  - `/roles` -> slug `roles`.
  - `/modules` -> slug `modules`.
  - `/sessions` -> slug `sessions`.
  - `/messages` -> slug `messages`.
  - `/broadcasts` -> slug `broadcasts`.

---

## Dev Agent Record

### Agent Model Used
Gemini 1.5 Pro (Antigravity Coordinator)

### Debug Log References

### Completion Notes List

### File List
