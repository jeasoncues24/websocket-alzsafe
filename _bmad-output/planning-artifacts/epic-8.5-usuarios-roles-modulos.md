---
stepsCompleted:
  - validate-prerequisites
  - design-epics
  - create-stories
  - final-validation
inputDocuments:
  - _bmad-output/project-context.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
  - _bmad-output/implementation-artifacts/future-index.md
status: "ready-for-dev"
---

# wsapi - Epic Breakdown: Usuario Admin, Roles y Modulos

## Overview

Este epic redefine el panel administrativo para que todo el backend de administracion viva bajo `/api/admin/*` y se consuma solo con JWT de empresa. El foco es refactorizar `usuarios` a `usuario_admin`, actualizar el CRUD de roles, exponer `modules` solo en lectura y ajustar la logica de eliminacion segun dependencias reales de base de datos.

## Requirements Inventory

### Functional Requirements

FR1: Todos los endpoints del panel administrativo deben vivir bajo `/api/admin/*` y aceptar solo JWT de empresa.
FR2: El recurso de administracion de usuarios debe renombrarse a `usuario_admin` en rutas, contratos y handlers.
FR3: `usuario_admin` debe soportar CRUD completo alineado con la tabla `admin_users` actual.
FR4: `roles` debe soportar CRUD completo alineado con la tabla `roles` actual, incluyendo `permissions`.
FR5: `modules` debe ser de solo lectura, con endpoints `GET` para catalogo y consulta.
FR6: `user_modules` debe soportar asignacion y reemplazo de modulos por usuario como override de permisos.
FR7: La eliminacion de `usuario_admin` debe borrar solo si no hay dependencias; si existen dependencias no cascada, debe deshabilitarse con `activo = 0`.
FR8: La eliminacion de `roles` debe fallar si el rol esta en uso por cualquier `usuario_admin`.
FR9: Los contratos deben reflejar los campos reales de migracion: `username`, `password_hash`, `email`, `empresa_id`, `rol`, `role_id`, `is_root`, `activo`, `last_login_at`, `permissions`.
FR10: El backend debe incluir pruebas para rutas, contratos, validaciones y reglas de eliminacion.

### NonFunctional Requirements

NFR1: Los contratos administrativos deben ser consistentes y predecibles para el panel.
NFR2: La logica de borrado debe proteger integridad referencial y no romper tablas relacionadas.
NFR3: Los cambios deben ser testeables con cobertura de handlers, servicios y storage.
NFR4: El epic no debe mezclar frontend; la UI administrativa se tratara en un epic separado.

### Additional Requirements

- La ruta `/api/admin/*` es exclusiva para administracion y no debe compartir contrato con endpoints del token por telefono.
- El JWT de telefono no debe poder consumir estos endpoints.
- `rol` existe como columna legacy/operativa; el contrato debe priorizar `role_id` para escritura y devolver ambos campos donde aplique.
- `is_root` y `permissions` deben validarse en el backend y no aceptarse sin control.
- `modules` no tiene CRUD; solo lectura del catalogo.
- `DELETE` de usuarios debe evaluar dependencias reales antes de decidir entre hard delete o deshabilitacion.
- `DELETE` de roles no debe degradar a disable; si esta en uso, se rechaza.

### UX Design Requirements

No UX design document is included for this backend-only epic.

### FR Coverage Map

FR1: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR2: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR3: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR4: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR5: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR6: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR7: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR8: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR9: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos
FR10: Epic 8.5 - Panel Admin de Usuarios, Roles y Modulos

## Epic List

### Epic 8.5: Panel Admin de Usuarios, Roles y Modulos
Permitir gestion completa del panel administrativo con contratos bajo `/api/admin/*`, CRUD de `usuario_admin`, CRUD de roles, catalogo de modulos en solo lectura y reglas de borrado segun uso real en base de datos.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8, FR9, FR10

## Contract Notes

### Base Route

- Todos los endpoints de este epic viven bajo `/api/admin/*`.
- Solo se aceptan requests autenticadas con JWT de empresa.
- No se expone ninguna variante bajo `/api/*` para este dominio.

### Proposed Endpoints

- `GET /api/admin/usuario_admin`
- `GET /api/admin/usuario_admin/:id`
- `POST /api/admin/usuario_admin`
- `PUT /api/admin/usuario_admin/:id`
- `DELETE /api/admin/usuario_admin/:id`
- `GET /api/admin/usuario_admin/:id/modulos`
- `PUT /api/admin/usuario_admin/:id/modulos`
- `GET /api/admin/roles`
- `GET /api/admin/roles/:id`
- `POST /api/admin/roles`
- `PUT /api/admin/roles/:id`
- `DELETE /api/admin/roles/:id`
- `GET /api/admin/modules`
- `GET /api/admin/modules/:id`

### Payload Rules

- `usuario_admin` create/update: `username`, `password` (create only), `email`, `empresa_id`, `role_id`, `activo`, `is_root` only when backend validation allows it.
- `roles` create/update: `name`, `description`, `is_root`, `permissions`.
- `permissions` debe validar JSON y referenciar slugs reales de `modules`.
- `user_modules` debe reemplazar el set completo de modulos asignados al usuario para evitar estados parciales.

## Story Draft Status

La estructura del epic ya esta definida para backend solamente. El siguiente paso fuera de este epic sera crear un epic separado para frontend con apoyo de `bmad-agent-ux-designer`.

## Stories

### Story 8.5.1: Rutas admin y contratos base

**Objetivo:** mover todo el panel administrativo a `/api/admin/*` y validar que solo use JWT de empresa.

**Acceptance Criteria:**

- `Given` una request sin JWT de empresa `When` accede a `/api/admin/*` `Then` retorna 401.
- `Given` un JWT por telefono `When` intenta consumir `/api/admin/*` `Then` se rechaza.
- `Given` una ruta legacy de usuarios `When` se revisa el router `Then` ya no expone `usuarios` fuera de `usuario_admin`.
- `Given` una request valida `When` consulta el endpoint base `usuario_admin` `Then` la respuesta usa el contrato nuevo.

### Story 8.5.2: CRUD de usuario_admin

**Objetivo:** implementar create, read, update y delete para usuarios administrativos usando la tabla `admin_users`.

**Acceptance Criteria:**

- `Given` datos validos `When` se crea un `usuario_admin` `Then` se persiste `username`, `email`, `empresa_id`, `role_id` y password hash.
- `Given` un usuario existente `When` se actualiza `Then` se modifican solo los campos permitidos y se conserva `last_login_at`.
- `Given` un usuario sin dependencias bloqueantes `When` se elimina `Then` el registro desaparece.
- `Given` un usuario con relaciones en tablas no cascada `When` se elimina `Then` se deshabilita con `activo = 0`.
- `Given` un usuario con override de modulos `When` se reemplaza su set `Then` se actualiza `user_modules` de forma atomica.

### Story 8.5.3: CRUD de roles

**Objetivo:** habilitar administracion completa de roles con validacion de uso real y permisos por modulo.

**Acceptance Criteria:**

- `Given` un rol valido `When` se crea `Then` guarda `name`, `description`, `is_root` y `permissions`.
- `Given` un rol existente `When` se actualiza `Then` se reflejan los cambios sin romper usuarios asociados.
- `Given` un rol usado por al menos un `usuario_admin` `When` se intenta eliminar `Then` retorna conflicto y no borra nada.
- `Given` un rol no usado `When` se elimina `Then` se borra de forma permanente.
- `Given` permissions con slugs inexistentes `When` se valida `Then` retorna error de validacion.

### Story 8.5.4: Catalogo de modules y override por usuario

**Objetivo:** exponer `modules` como catalogo solo lectura y permitir asignacion de modulos por usuario.

**Acceptance Criteria:**

- `Given` una request valida `When` consulta `/api/admin/modules` `Then` retorna el catalogo completo.
- `Given` un slug invalido `When` se intenta asignar a un usuario `Then` se rechaza.
- `Given` una asignacion existente `When` se reemplaza `Then` el estado final coincide con la lista enviada.
- `Given` la tabla `modules` `When` se revisa el contrato `Then` no existe CRUD de escritura.

### Story 8.5.5: Tests de integridad y contratos

**Objetivo:** cubrir la refactorizacion con pruebas de backend para evitar regresiones.

**Acceptance Criteria:**

- `Given` el router actualizado `When` se ejecutan tests `Then` las rutas admin responden bajo `/api/admin/*`.
- `Given` un delete de usuario con dependencias `When` se prueba `Then` el sistema lo deshabilita en lugar de borrarlo.
- `Given` un delete de rol en uso `When` se prueba `Then` falla con el error esperado.
- `Given` el catalogo de modules `When` se prueba `Then` solo hay lecturas.
- `Given` contratos actualizados `When` se ejecutan pruebas de handlers/storage `Then` pasan sin romper el schema actual.

## Out of Scope

- Frontend del panel administrativo.
- Rutas de usuarios clientes o cualquier dominio fuera de `/api/admin/*`.
- Endpoints del token por telefono.
