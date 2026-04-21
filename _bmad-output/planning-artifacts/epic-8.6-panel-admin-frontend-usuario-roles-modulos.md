---
status: "draft"
inputDocuments:
  - _bmad-output/planning-artifacts/epic-8.5-usuarios-roles-modulos.md
  - _bmad-output/project-context.md
  - frontend/app/users_admin/page.tsx
  - frontend/app/roles/page.tsx
  - frontend/app/modules/page.tsx
  - frontend/components/layout/sidebar.tsx
  - frontend/lib/api.ts
---

# Epic 8.6: Panel Admin Frontend de Usuario Admin, Roles y Modulos

## Overview

Refactorizar la experiencia frontend del panel administrativo para que el manejo de `usuario_admin`, roles y módulos quede alineado con el nuevo backend bajo `/api/admin/*`, usando únicamente el JWT de empresa.

El epic parte de la UI existente y la convierte en una experiencia más clara, densa y segura para administración: listas rápidas, formularios consistentes, estados de carga explícitos, confirmaciones para acciones destructivas y feedback visible cuando una eliminación se convierte en deshabilitación por dependencias.

## Objetivo

Permitir que el administrador:

- gestione usuarios admin con el nuevo contrato `usuario_admin`
- cree, actualice y elimine roles desde la UI
- consulte módulos en modo solo lectura
- asigne o reemplace módulos por usuario
- entienda cuándo una acción no borra realmente sino que deshabilita

## Requisitos Funcionales

- FR-01: El panel frontend debe consumir exclusivamente `/api/admin/*` para este dominio.
- FR-02: La vista de usuarios debe migrar de `users` a `usuario_admin` en labels, contratos y acciones.
- FR-03: La vista de usuarios debe permitir buscar, crear, editar, deshabilitar y eliminar usuarios admin.
- FR-04: La vista de usuarios debe permitir asignar módulos por usuario como override.
- FR-05: La vista de roles debe permitir listar, crear, editar y eliminar roles.
- FR-06: La vista de roles debe mostrar cuándo un rol está en uso y no puede eliminarse.
- FR-07: La vista de módulos debe ser de solo lectura y funcionar como catálogo de referencia.
- FR-08: La navegación del panel debe dejar claro qué secciones pertenecen a administración.
- FR-09: El frontend debe manejar errores de integridad referencial sin romper la experiencia.
- FR-10: El panel debe seguir usando el JWT de empresa guardado como `admin_token`.

## Requisitos No Funcionales

- NFR-01: La interfaz debe ser rápida para uso operativo diario.
- NFR-02: Las tablas deben ser densas, legibles y útiles en desktop sin perder respuesta en mobile.
- NFR-03: Los formularios deben ser consistentes, con validación visible y errores accionables.
- NFR-04: Las acciones destructivas deben requerir confirmación explícita.
- NFR-05: El diseño debe aprovechar shadcn/ui y mantener coherencia visual con el panel actual.
- NFR-06: Los cambios deben ser testeables en componentes y flujos críticos.

## UX Requirements

- UX-01: La sección de usuarios admin debe priorizar lectura rápida: nombre, email, empresa, rol, root, estado y acciones.
- UX-02: Las acciones de borrar deben mostrar un copy claro: borrar, deshabilitar o bloquear por uso.
- UX-03: Los roles deben mostrar badges de root y uso activo para que el riesgo sea visible antes de tocar nada.
- UX-04: Los módulos deben verse como catálogo de permisos, no como pantalla de edición.
- UX-05: Los formularios de usuario y rol deben abrirse en modal o drawer, no en pantallas separadas, para reducir fricción.
- UX-06: Los estados vacíos deben guiar la acción siguiente, no solo informar ausencia.

## FR Coverage Map

- FR-01 -> Epic 8.6
- FR-02 -> Epic 8.6
- FR-03 -> Epic 8.6
- FR-04 -> Epic 8.6
- FR-05 -> Epic 8.6
- FR-06 -> Epic 8.6
- FR-07 -> Epic 8.6
- FR-08 -> Epic 8.6
- FR-09 -> Epic 8.6
- FR-10 -> Epic 8.6

## Epic List

### Epic 8.6: Panel Admin Frontend de Usuario Admin, Roles y Modulos
Convertir la experiencia actual del panel admin en una UI consistente y eficiente para gestionar usuarios admin, roles y módulos con el nuevo backend.
**FRs covered:** FR-01, FR-02, FR-03, FR-04, FR-05, FR-06, FR-07, FR-08, FR-09, FR-10

## Contract Notes

### Fuentes de verdad

- `frontend/lib/api.ts` debe ser la capa que normaliza el acceso a `/api/admin/*`.
- `admin_token` sigue siendo el token de sesión del panel administrativo.
- Este epic no toca rutas del token por teléfono.

### Vistas afectadas

- `frontend/app/users_admin/page.tsx`
- `frontend/app/roles/page.tsx`
- `frontend/app/modules/page.tsx`
- `frontend/components/layout/sidebar.tsx`
- `frontend/lib/api.ts`

### UX de navegación

- Mantener el acceso a usuarios, roles y módulos desde la sidebar.
- Renombrar labels a términos del nuevo dominio cuando ayude a reducir ambigüedad.
- Mantener estados activos claros para que el admin no se pierda entre secciones.

## Stories

### Story 8.6.1: Migración de API client y navegación admin

**Objetivo:** alinear la capa frontend de acceso al backend con los nuevos endpoints `/api/admin/*` y el lenguaje `usuario_admin`.

**Acceptance Criteria:**

- `Given` el frontend usa `admin_token` `When` consulta usuarios, roles o módulos `Then` lo hace contra `/api/admin/*`.
- `Given` la sidebar se renderiza `When` el admin navega `Then` las secciones de usuarios, roles y módulos quedan visibles y consistentes.
- `Given` un endpoint legacy de usuarios `When` se revisa la capa API `Then` ya no depende de `/api/admin/users`.

### Story 8.6.2: Usuario Admin workspace

**Objetivo:** refactorizar la vista actual de usuarios admin para soportar el nuevo contrato y una experiencia más clara.

**Acceptance Criteria:**

- `Given` la página de usuarios carga `When` recibe datos `Then` muestra columnas para username, email, empresa, rol, root y estado.
- `Given` un usuario admin existe `When` se edita `Then` el formulario respeta los campos válidos del backend.
- `Given` un usuario admin tiene dependencias `When` se elimina `Then` la UI muestra que se deshabilitó y no que desapareció.
- `Given` se asignan módulos `When` se guarda `Then` el set final reemplaza el anterior.

### Story 8.6.3: Roles workspace

**Objetivo:** crear una experiencia completa para administrar roles sin salir del panel.

**Acceptance Criteria:**

- `Given` la página de roles carga `When` recibe datos `Then` muestra nombre, descripción, root y uso.
- `Given` un rol no está en uso `When` se elimina `Then` la acción se confirma y desaparece de la lista.
- `Given` un rol está en uso `When` se intenta eliminar `Then` la UI muestra el bloqueo de forma clara.
- `Given` se crea o actualiza un rol `When` se guarda `Then` se valida `name`, `description`, `is_root` y `permissions`.

### Story 8.6.4: Modules catalog y permisos visibles

**Objetivo:** dejar `modules` como catálogo de referencia y soporte visual para permisos.

**Acceptance Criteria:**

- `Given` la página de módulos carga `When` obtiene datos `Then` muestra nombre, slug y descripción.
- `Given` el usuario revisa módulos `When` interactúa `Then` no encuentra acciones de crear, editar o borrar.
- `Given` un formulario necesita ayuda visual `When` se asignan permisos `Then` los módulos sirven como referencia clara.

### Story 8.6.5: Estados críticos y feedback de riesgo

**Objetivo:** hacer visible el comportamiento real de las operaciones peligrosas.

**Acceptance Criteria:**

- `Given` una operación falla por integridad `When` ocurre en UI `Then` se muestra el motivo real y no un error genérico.
- `Given` una acción es destructiva `When` el usuario hace click `Then` aparece confirmación explícita.
- `Given` una entidad está inactiva o bloqueada `When` se muestra en tabla `Then` el badge lo deja evidente.

### Story 8.6.6: Tests frontend y accesibilidad

**Objetivo:** asegurar que la refactorización de la UI no rompa la operación diaria.

**Acceptance Criteria:**

- `Given` la UI refactorizada `When` se ejecutan tests `Then` las páginas admin principales siguen montando.
- `Given` un formulario de usuario o rol `When` se valida `Then` los errores de inputs aparecen correctamente.
- `Given` la navegación admin `When` se prueba `Then` los links y estados activos funcionan.
- `Given` se revisa accesibilidad básica `When` se navega con teclado `Then` los controles críticos siguen siendo utilizables.

## Out of Scope

- Cambios al backend, que ya quedan cubiertos por el epic 8.5.
- Rutas o contratos del token por teléfono.
- Nuevo diseño global del resto del panel fuera de usuarios, roles y módulos.
