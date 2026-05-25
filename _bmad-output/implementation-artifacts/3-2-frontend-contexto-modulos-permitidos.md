# Story 3-2: Frontend — Contexto de módulos permitidos

**Estado:** review
**Epic padre:** Epic 3 — Módulos Dinámicos desde BD y Perfil de Usuario
**Story ID:** 3.2

---

## Story

Como **Usuario Autenticado**,
quiero **que el frontend almacene y exponga de forma global los módulos asignados a mi cuenta**,
para **que el panel de control pueda ajustar de forma dinámica la navegación y el acceso a páginas**.

---

## Acceptance Criteria

**AC1 — El store de Zustand almacena `allowedModules`**
Dado un usuario no autenticado o recién autenticado,
cuando se consulta el store global `useAppStore`,
entonces existe un estado `allowedModules: string[]` inicializado por defecto como un array vacío `[]` (o en su defecto `["dashboard"]`).

**AC2 — El login de usuario inicializa `allowedModules`**
Dado un usuario que inicia sesión exitosamente en `/login`,
cuando el endpoint retorna los datos de usuario incluyendo `allowed_modules`,
entonces el store `useAppStore` se actualiza guardando exactamente esa lista de slugs de módulos permitidos.

**AC3 — La revalidación de sesión (`GET /api/auth/me`) actualiza `allowedModules`**
Dado un usuario con un token válido en `localStorage` que recarga la página,
cuando se realiza la revalidación automática llamando a `GET /api/auth/me`,
entonces la respuesta que contiene `allowed_modules` se sincroniza con el store `useAppStore`, sobreescribiendo `allowedModules`.

**AC4 — Compilación y tipado correctos**
El tipado de TypeScript del usuario en el frontend incluye de forma opcional u obligatoria `allowed_modules?: string[]`, y la aplicación de frontend compila sin errores.

---

## Tasks / Subtasks

- [ ] **T1 — Modificar el store Zustand global**
  - **Archivo:** `frontend/stores/useAppStore.ts` (o el store de auth respectivo)
  - [ ] Agregar el estado `allowedModules: string[]` al struct del estado.
  - [ ] Agregar la acción `setAllowedModules: (modules: string[]) => void`.
  - [ ] Inicializar `allowedModules` como `[]`.

- [ ] **T2 — Actualizar el cliente API de frontend y tipados**
  - **Archivo:** `frontend/lib/api.ts` (o archivos de tipos compartidos de usuario)
  - [ ] Actualizar la interfaz `User` (o el tipo del objeto de usuario retornado por `/api/auth/me`) para incluir `allowed_modules: string[]`.
  - [ ] Asegurar que las llamadas de Axios/fetch que invocan login o revalidación retornen el usuario tipado con esta nueva propiedad.

- [ ] **T3 — Sincronizar el estado durante el ciclo de vida de la sesión**
  - **Archivos:** Componentes de layouts principales, proveedores de sesión o `frontend/app/layout.tsx`
  - [ ] En la llamada de inicio de sesión, despachar `setAllowedModules` con los módulos recibidos del backend.
  - [ ] En la llamada de revalidación inicial (ej. en un `useEffect` que carga el perfil si hay un token de JWT presente), despachar `setAllowedModules` con el resultado.
  - [ ] Asegurar que al cerrar sesión (`logout`), se limpie `allowedModules` devolviéndolo a `[]`.

---

## Dev Notes

- **Concurrencia y React Hydration:** Zustand en Next.js App Router debe manejarse con cuidado para evitar errores de hidratación. El store generalmente se sincroniza en el cliente (`useEffect` o similar).
- **Consumo:** El estado de `allowedModules` será consumido en la Story 3-3 por el sidebar y el mobile nav para filtrar la navegación, y por un middleware o layout superior para bloquear rutas.

---

## Dev Agent Record

### Agent Model Used
Gemini 1.5 Pro (Antigravity Coordinator)

### Debug Log References

### Completion Notes List

### File List
