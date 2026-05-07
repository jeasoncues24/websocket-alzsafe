---
title: 'Story 2.1 — Eliminar select de empresa del formulario usuario_admin'
type: 'fix'
created: '2026-05-04'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '69509dc'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** El modal de creación/edición de `usuario_admin` expone un select de empresa que no tiene sentido para este tipo de usuario. La tabla `admin_users` no tiene columna `empresa_id` — el campo viene de una asociación distinta y no aplica al contexto de administración. El selector genera confusión operativa y debe eliminarse.

**Approach:** Remover el select de empresa del formulario `UserFormModal`, limpiar el estado y la carga de empresas asociada en el componente padre, eliminar la columna "Empresa" de la tabla de listado, y dejar de enviar `empresa_id` en los payloads de creación y edición. Cambio exclusivamente frontend; el backend no requiere modificación.

## Boundaries & Constraints

**Always:** eliminar el select del formulario y la columna de empresa de la tabla; limpiar todo estado y efecto relacionado a `empresaId` / `loadCompanies` / `companies` del componente `UsuarioAdminPage`; mantener el resto del formulario y la tabla intactos; ajustar `UserFormModal` para no recibir ni usar el prop `companies`.

**Never:** modificar el backend; tocar otros formularios o páginas que usen empresas; cambiar la lógica de roles, módulos ni otros campos del formulario; eliminar `getEmpresas` ni el tipo `Empresa` de `lib/api.ts` si son usados en otros archivos (solo remover el import en este archivo si queda huérfano).

**Ask First:** si se detecta que `empresa_id` tiene algún uso funcional real en `admin_users` (p.ej. algún guard en backend que lo valide).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Crear usuario_admin | Formulario sin campo empresa | Payload enviado sin `empresa_id`; backend acepta normalmente | N/A |
| Editar usuario_admin | Usuario existente con empresa_id en BD | Formulario no muestra campo empresa; no sobreescribe el valor existente en backend | No enviar `empresa_id: undefined` en el payload de update |
| Tabla de listado | Usuarios con y sin empresa_id | Columna "Empresa" eliminada; columnas restantes sin cambio | N/A |
| Props del modal | `companies` prop eliminada | Modal compila sin el prop; no se llama `getEmpresas` en este módulo | N/A |

</frozen-after-approval>

## Code Map

- `frontend/app/users_admin/page.tsx` — componente principal: `UserFormModal` y `UsuarioAdminPage`; aquí viven la mayoría de los cambios
- `frontend/app/usuario_admin/page.tsx` — re-exporta `users_admin/page`; no requiere cambios
- `frontend/lib/api.ts` — tipos `UserAdminRol`, `CreateUserRequest`, `UpdateUserRequest`; remover `empresa_id` de los tres
- `backend/internal/domain/admin_user.go` — referencia: `AdminUser` struct no tiene `EmpresaID`; `AdminJWTClaims.EmpresaID` existe pero es control de acceso del JWT, no atributo del usuario — no tocar
- `backend/internal/storage/admin_user.go` — referencia: `GetAllByEmpresa` marcado como deprecated; no tocar en esta story

## Tasks & Acceptance

**Execution:**

- [x] `UserFormModal` — eliminar el estado `empresaId`, la variable `companyOptions`, y el bloque JSX del select de empresa (`<div>` con label "Empresa" y `<select>` que itera `companyOptions`); dejar de incluir `empresa_id` en el objeto `data` del `handleSubmit`
- [x] `UserFormModal` — eliminar el prop `companies: Empresa[]` de la interfaz del componente y de todos los lugares donde se pasa al modal
- [x] `UsuarioAdminPage` — eliminar el estado `companies`, el callback `loadCompanies`, y el `useEffect` que lo llama; eliminar el import de `getEmpresas` y `Empresa` si quedan sin uso en este archivo
- [x] `UsuarioAdminPage` — eliminar la columna "Empresa" del `<TableHeader>` y su celda en cada `<TableRow>`; ajustar `colSpan` de filas vacías/cargando de 7 → 6
- [x] `frontend/lib/api.ts` — remover `empresa_id` de los tres tipos: `UserAdminRol`, `CreateUserRequest` y `UpdateUserRequest`

**Acceptance Criteria:**

- Given el formulario de nuevo usuario_admin, when se abre, then no existe ningún campo ni label de "Empresa".
- Given el formulario de editar usuario_admin, when se abre, then el modal no muestra campo empresa y el payload de update no incluye `empresa_id`.
- Given la tabla de listado de usuario_admin, when se visualiza, then no aparece la columna "Empresa".
- Given los tipos de `lib/api.ts`, when se revisa `UserAdminRol`, `CreateUserRequest` y `UpdateUserRequest`, then ninguno declara `empresa_id`.
- Given la compilación del proyecto, when se ejecuta `npm run build` en `frontend/`, then no hay errores de TypeScript relacionados con los cambios.

## Spec Change Log

_(vacío al crear)_

## Design Notes

El campo `empresa_id` en la tabla `admin_users` no existe a nivel de schema (migration 004). El frontend lo exponía a través de un JOIN o campo calculado que el backend devuelve pero no persiste desde `CreateUser` / `UpdateUser`. Al dejar de enviarlo, el comportamiento del backend no cambia.

## Verification

**Commands:**
- `grep -rn "empresaId\|loadCompanies\|companyOptions" frontend/app/users_admin/` — expected: sin resultados
- `grep -n "empresa_id" frontend/app/users_admin/page.tsx frontend/lib/api.ts` — expected: sin resultados en ninguno de los dos archivos
- `cd frontend && npm run build 2>&1 | grep -i "error"` — expected: sin errores de TS

## Suggested Review Order

**Formulario:**
- Confirmar que el select de empresa ya no aparece y que el payload de save no incluye `empresa_id`.
  [`frontend/app/users_admin/page.tsx`](../../frontend/app/users_admin/page.tsx)

**Tipos API:**
- Confirmar que `CreateUserRequest` y `UpdateUserRequest` ya no declaran `empresa_id`.
  [`frontend/lib/api.ts`](../../frontend/lib/api.ts)

---

## Dev Agent Record

### Implementation Plan

Eliminación del select de empresa del formulario de usuario_admin:
1. UserFormModal: eliminar estado `empresaId`, prop `companies`, y JSX del select
2. UsuarioAdminPage: eliminar estado `companies`, `loadCompanies`, useEffect, imports, columna de tabla
3. api.ts: remover `empresa_id` de CreateUserRequest y UpdateUserRequest

### Completion Notes

- ✅ Estado `empresaId` eliminado de UserFormModal
- ✅ Prop `companies` eliminada de UserFormModal
- ✅ Select de empresa removido del JSX
- ✅ Payload ya no incluye `empresa_id`
- ✅ Estado y funciones de companies eliminados de UsuarioAdminPage
- ✅ Columna "Empresa" eliminada de la tabla
- ✅ colSpan ajustado de 7 → 6
- ✅ Tipos API actualizados en lib/api.ts
- ✅ Build exitoso sin errores de TypeScript

## File List

- `frontend/app/users_admin/page.tsx` — modificado
- `frontend/lib/api.ts` — modificado

## Change Log

- 2026-05-04: Implementación completa de story 2-1 — eliminar select empresa de usuario_admin

### Review Findings

- [x] [Review][Patch] `UserAdminRol` aún declara `empresa_id?: number` — viola AC4 [`frontend/lib/api.ts:559`] — fixed 2026-05-05
- [x] [Review][Defer] `restoreEmpresa` añadida en `api.ts` sin endpoint backend — pertenece a story 2-2, no a 2-1 [`frontend/lib/api.ts:292`] — deferred, pre-existing
