# Epic 3: Módulos Dinámicos desde BD y Perfil de Usuario

Status: in-progress

## Objetivo

Eliminar la dependencia de módulos hardcodeados en el panel y habilitar una sección de gestión de cuenta personal para todos los usuarios autenticados, conectando el frontend con la infraestructura de permisos que ya existe en base de datos.

## Evaluación actual

### Resumen ejecutivo

El backend ya tiene la infraestructura completa: tablas `modules`, `roles` (con campo `permissions` JSON), `user_modules` y endpoints CRUD para administrar todo esto desde el panel. Sin embargo, el frontend ignora completamente esa información: el sidebar muestra siempre los 9 módulos hardcodeados en `nav-items.ts` sin importar el rol ni los módulos asignados al usuario. Adicionalmente, no existe ninguna sección donde el usuario autenticado pueda gestionar su propia cuenta (email, contraseña, username).

### Hallazgos principales

1. **Módulos hardcodeados en el frontend.**
   - `frontend/components/layout/nav-items.ts`: array estático con 9 entradas fijas.
   - `frontend/components/layout/sidebar.tsx` y `mobile-nav.tsx`: mapean directamente sobre `navItems` sin ningún filtro de permisos.
   - `frontend/stores/useAppStore.ts`: no tiene estado de módulos ni permisos del usuario.

2. **La infraestructura de permisos en BD está completa pero desconectada del frontend.**
   - Tabla `modules`: 8 módulos con slug canónico.
   - Tabla `roles`: campo `permissions` (JSON array de slugs); `is_root` para acceso total.
   - Tabla `user_modules`: override por usuario sobre el rol.
   - Endpoints existentes: `GET /api/admin/modules`, `GET /api/admin/roles`, `GET/PUT /api/admin/usuario_admin/{id}/modulos`.

3. **`/api/auth/me` no retorna los módulos efectivos del usuario.**
   - `backend/internal/http/handlers/auth.go` líneas 189-221: devuelve solo `id`, `username`, `email`, `role_id`, `is_root`, `activo`.
   - Para construir el menú dinámico, el frontend necesita los slugs de módulos permitidos en la misma llamada de login/revalidación.

4. **No existe endpoint de auto-gestión de cuenta.**
   - `PUT /api/admin/usuario_admin/{id}` actualiza email, role_id e is_active, pero lo opera un admin sobre otro usuario.
   - No hay `PUT /api/auth/me` para que el propio usuario actualice sus datos.
   - No hay endpoint de cambio de contraseña; `UpdateUserRequest` no incluye `password`.

5. **Módulo de Configuraciones sin sección de cuenta personal.**
   - `frontend/app/settings/page.tsx`: 3 tabs (Apariencia, General, Acerca de), ninguna relacionada con la cuenta del usuario.
   - No hay ruta `/profile` ni tab "Mi cuenta" en ninguna parte del panel.

## Alcance incluido

- Extender `GET /api/auth/me` para retornar `allowed_modules` (slugs efectivos según rol + overrides de usuario).
- Crear `PUT /api/auth/me` para que el usuario actualice su propio email y username.
- Crear `PUT /api/auth/me/password` para cambio de contraseña autenticado (requiere contraseña actual).
- Hook/contexto en frontend que cargue los módulos permitidos al autenticar y los exponga globalmente.
- Sidebar y MobileNav filtrados dinámicamente por los módulos permitidos del usuario.
- Redirección a `/dashboard` si el usuario navega a una ruta de módulo no permitido.
- Sección "Mi cuenta" dentro de `settings/page.tsx` (nueva tab) con formularios de datos personales y cambio de contraseña, responsive web + mobile.

## Fuera de alcance

- Rediseño del sistema de roles o módulos (el CRUD de roles/módulos ya existe).
- Cambio del modelo de JWT (se mantiene `AdminJWTClaims` sin modificar).
- Permisos granulares dentro de un módulo (lectura/escritura por recurso).
- Avatar o foto de perfil.

## Dependencias

- Backend: Go, dominio en `internal/domain`, handlers en `internal/http/handlers/auth.go`.
- Frontend: Next.js + Tailwind v4 + shadcn local, Zustand (`useAppStore`), `localStorage` para el token.
- La lógica de módulos efectivos sigue este orden de precedencia: `is_root → todos` | `user_modules override → esos slugs` | `role.permissions → esos slugs`.

## Riesgos conocidos

- Si un usuario tiene `user_modules` vacío y su rol tampoco tiene permisos, quedaría sin módulos visibles; debe definirse un fallback mínimo (solo `dashboard`).
- Cambiar el contrato de `/api/auth/me` puede afectar código que ya consume ese endpoint; revisar todos los consumers en frontend antes de modificar.

## Criterios de éxito

- El sidebar y mobile nav muestran únicamente los módulos que el rol/usuario tiene asignados en BD, sin ningún valor hardcodeado.
- Un usuario root ve todos los módulos; un usuario de soporte solo ve los de su rol.
- El usuario autenticado puede cambiar su email, username y contraseña desde Settings > Mi cuenta.
- `cd backend && go test ./...` y `cd backend && go build ./...` pasan sin errores.
- `cd frontend && npm run lint` y `cd frontend && npm run build` pasan sin errores.

## Stories propuestas

| ID  | Nombre | Tipo | Prioridad | Estado |
|-----|--------|------|-----------|--------|
| 3-1 | Backend: módulos efectivos en /api/auth/me | Backend | Alta | backlog |
| 3-2 | Frontend: contexto de módulos permitidos | Frontend | Alta | backlog |
| 3-3 | Frontend: sidebar y mobile-nav dinámico | Frontend | Alta | backlog |
| 3-4 | Backend: endpoints de auto-gestión de cuenta | Backend | Media | backlog |
| 3-5 | Frontend: sección "Mi cuenta" en Settings | Frontend | Media | backlog |

## Skills por story

Invocar únicamente las skills indicadas al crear o implementar cada story. No aplicar todas a todas.

| Story | Skills a invocar | Motivo |
|---|---|---|
| 3-1 | `golang-security`, `sql-optimization` | El endpoint extiende `/api/auth/me` con un JOIN sobre `roles`/`user_modules`; revisar trust boundary del JWT y evaluar el query con EXPLAIN |
| 3-2 | — | Lógica de store Zustand sin SQL ni superficie de ataque nueva |
| 3-3 | `bmad-agent-ux-designer`, `ui-ux-pro-max` | El sidebar dinámico y el filtrado de rutas afectan directamente la navegación; diseño UX crítico para mobile y desktop |
| 3-4 | `golang-security` | Cambio de contraseña y email requieren validación estricta en trust boundary; riesgo de timing attack en comparación de passwords |
| 3-5 | `bmad-agent-ux-designer`, `ui-ux-pro-max` | Pantalla de "Mi cuenta" es nueva; necesita diseño completo de formularios, estados de error, feedback async y responsive web+mobile |

## Orden recomendado de implementación

```text
3-1 → prerequisito para 3-2 y 3-3 (extiende contrato de /api/auth/me)
3-2 → prerequisito para 3-3 (expone módulos via store/context)
3-3 → consume 3-1 y 3-2 (sidebar y nav dinámico)
3-4 → independiente, puede ir en paralelo con 3-1/3-2/3-3
3-5 → depende de 3-4 (consume los nuevos endpoints de auto-gestión)
```

## Condición de cierre del epic

El epic se cierra cuando el sidebar no tenga ningún valor hardcodeado de módulos, los módulos visibles correspondan exactamente a la configuración en BD para cada usuario/rol, y exista una sección funcional de "Mi cuenta" en Settings accesible para todos los usuarios del panel.
