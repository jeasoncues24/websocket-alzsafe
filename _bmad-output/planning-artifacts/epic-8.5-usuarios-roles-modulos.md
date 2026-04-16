# Epic 8.5: Sistema de Gestión de Usuarios con Roles y Módulos

## Overview

| Campo | Valor |
|-------|-------|
| **Nombre** | Sistema de Gestión de Usuarios con Roles y Módulos |
| **Tipo** | Backend + Frontend + DB |
| **Estado** | Planning |
| **Story Count** | 5 |
| **Effort Estimate** | 8-12 horas |

## Objetivo

Implementar un sistema completo de gestión de usuarios con roles y permisos por módulos para el panel administrativo de wsapi.

- Crear estructura de tablas para roles, módulos y permisos usuario-módulo
- Implementar sistema de autorización centralizado con bypass para root
- Proteger el flag de root con múltiples capas de seguridad
- Permitir promoción a root solo cuando existe otro root activo
- CRUD de usuarios con asignación de roles y módulos
- Frontend para gestionar usuarios, roles y módulos

## Modules del Sistema (definidos)

| Módulo | Slug | Descripción |
|--------|------|-------------|
| Dashboard | dashboard | Panel de métricas |
| Empresas | companies | Gestión de empresas |
| Mensajes | messages | Historial de mensajes |
| Sesiones | sessions | Gestión de sesiones WhatsApp |
| Difusiones | broadcasts | Envío masivo |
| Configuración | settings | Settings del sistema |

## Arquitectura de Datos

### Diagrama de Relaciones

```
┌─────────────────┐       ┌─────────────────┐
│    roles        │       │    modules      │
├─────────────────┤       ├─────────────────┤
│ id (PK)         │       │ id (PK)         │
│ name (UNIQUE)   │       │ name (UNIQUE)   │
│ description     │       │ description     │
│ is_root         │       │ slug            │
│ created_at      │       │ created_at      │
│ updated_at      │       └─────────────────┘
└─────────────────┐              ▲
        │                        │
        │              ┌─────────────────┐
        │              │  user_modules   │
        │              ├─────────────────┤
        │              │ user_id (FK)    │
        │              │ module_id (FK)  │
        │              │ created_at      │
        │              └─────────────────┘
        │
        ▼
┌─────────────────┐
│  admin_users    │
├─────────────────┤
│ id (PK)         │
│ username        │
│ password_hash  │
│ email           │
│ empresa_id      │
│ role_id (FK)    │
│ is_root         │
│ is_active       │
│ created_at      │
│ updated_at      │
└─────────────────┘
```

## Reglas de Seguridad

### 1. Protección de Root

- El flag `is_root` está en la tabla `roles`, no en `admin_users`
- Trigger en DB previene modificación directa de `is_root`
- Storage layer valida que no se pueda modificar rol root
- Solo un usuario root puede promover a otro usuario a root
- **Requiere al menos un root activo existente** para poder promover

### 2. Acceso a Módulos

- Root tiene acceso a TODOS los módulos sin verificar permisos
- Resolución de permisos: **override usuario → rol → denegado**
- Si usuario tiene entradas en `user_modules`, esas prevalecen
- Si no hay override, usar módulos del rol del usuario

### 3. Drop/Recreate para Pruebas

```sql
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS user_modules;
DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS roles;
SET FOREIGN_KEY_CHECKS = 1;
```

## Criterios de Aceptación

1. Tablas creadas con migraciones y seeds iniciales
2. CRUD de usuarios funcional con roles asignados
3. Sistema de autorización bloquea acceso a módulos no permitidos
4. Root tiene acceso a todo sin hardcoding en handlers
5. Solo root puede cambiar rol de usuario a root (con otro root presente)
6. Trigger de protección en DB para evitar cambios directos a is_root
7. Frontend permite gestionar usuarios y asignar roles
8. Tests de integración pasan con drop/recreate

---

## Stories

### 8.5.1: Migraciones y Estructura de Datos

**Objetivo:** Crear tablas de roles, módulos y user_modules con seeds iniciales

**Tareas:**
- [ ] Crear migración `001_create_roles.up.sql` con seeds (root, admin, operador, viewer)
- [ ] Crear migración `001_create_roles.down.sql` (drop table)
- [ ] Crear migración `002_create_modules.up.sql` con seeds (6 módulos del sistema)
- [ ] Crear migración `002_create_modules.down.sql` (drop table)
- [ ] Crear migración `003_create_user_modules.up.sql`
- [ ] Crear migración `003_create_user_modules.down.sql` (drop table)
- [ ] Crear migración `004_alter_admin_users_add_role_id.up.sql`
- [ ] Crear migración `004_alter_admin_users_add_role_id.down.sql`
- [ ] Crear trigger para protección de is_root en roles
- [ ] Ejecutar migraciones en entorno de desarrollo
- [ ] Verificar que las tablas y datos seed existen

**Criterios de Aceptación:**
- [ ] Tabla `roles` tiene 4 registros种子 (root, admin, operador, viewer)
- [ ] Tabla `modules` tiene 6 registros种子 (dashboard, companies, messages, sessions, broadcasts, settings)
- [ ] Trigger previene UPDATE de is_root en roles
- [ ] FK entre admin_users y roles funciona

**Dependencies:** Ninguna

**Estimated Effort:** 2 horas

---

### 8.5.2: Domain Entities y Storage Layer

**Objetivo:** Implementar entidades y storage para roles, módulos y user_modules

**Tareas:**
- [ ] Crear `domain/role.go` con struct Role y métodos
- [ ] Crear `domain/module.go` con struct Module
- [ ] Crear `storage/role.go` con RoleStore (CRUD + GetRootRole)
- [ ] Crear `storage/module.go` con ModuleStore (CRUD + GetAll)
- [ ] Crear `storage/user_module.go` con UserModuleStore
- [ ] Crear `storage/user.go` actualizado con métodos de roles
- [ ] Ejecutar tests unitarios existentes
- [ ] Agregar tests para los nuevos stores

**Criterios de Aceptación:**
- [ ] Todos los stores implementan interfaz válida
- [ ] Tests existentes pasan sin regresiones
- [ ] Nuevos tests para Role, Module, UserModule stores pasan

**Dependencies:** Story 8.5.1

**Estimated Effort:** 2 horas

---

### 8.5.3: Authorization Service

**Objetivo:** Implementar servicio centralizado de autorización con bypass para root

**Tareas:**
- [ ] Crear `domain/authorization.go` con AuthorizationService
- [ ] Implementar `CanAccess(ctx, moduleSlug)` con lógica de bypass root
- [ ] Implementar `GetUserModules(ctx)` con lógica de resolución
- [ ] Crear middleware de autorización para rutas
- [ ] Integrar middleware en router
- [ ] Tests unitarios para AuthorizationService

**Criterios de Aceptación:**
- [ ] `CanAccess` retorna true para root sin verificar módulos
- [ ] `CanAccess` verifica permisos para usuarios no-root
- [ ] `GetUserModules` devuelve todos los módulos para root
- [ ] Middlego intercepta requests y valida acceso

**Dependencies:** Story 8.5.2

**Estimated Effort:** 2 horas

---

### 8.5.4: User Service con Promoción Segura a Root

**Objetivo:** Implementar lógica de negocio para gestión de usuarios y promoción segura

**Tareas:**
- [ ] Crear `domain/user_service.go` con UserService
- [ ] Implementar método `CreateUser` con validación de rol
- [ ] Implementar método `UpdateUser` con protección de root
- [ ] Implementar método `DeleteUser` con validación
- [ ] Implementar método `PromoteToRoot` con validación de existente root
- [ ] Implementar método `AssignModules` para override de permisos
- [ ] Integrar UserService en handlers HTTP
- [ ] Tests unitarios para UserService

**Criterios de Aceptación:**
- [ ] Crear usuario asigna rol correctamente
- [ ] No se puede cambiar is_root directamente en update
- [ ] PromoteToRoot falla si no existe otro root
- [ ] PromoteToRoot requiere que el solicitante sea root
- [ ] Tests pasan

**Dependencies:** Story 8.5.3

**Estimated Effort:** 2 horas

---

### 8.5.5: Frontend - Gestión de Usuarios

**Objetivo:** Crear interfaz para gestionar usuarios, roles y módulos

**Tareas:**
- [ ] Crear página `/users` con tabla de usuarios
- [ ] Crear modal/form para crear/editar usuario
- [ ] Agregar selector de roles en formulario
- [ ] Agregar selector de módulos (checkbox) para override
- [ ] Crear página `/roles` (opcional, solo listar)
- [ ] Crear página `/modules` (opcional, solo listar)
- [ ] Integrar con API de usuarios
- [ ] Tests E2E para flujos de usuario

**Criterios de Aceptación:**
- [ ] Lista de usuarios muestra username, email, rol, estado
- [ ] Crear usuario con rol funciona
- [ ] Editar usuario funciona
- [ ] Asignar módulos adicionales funciona
- [ ] Botón de eliminar funciona

**Dependencies:** Story 8.5.4

**Estimated Effort:** 2-3 horas

---

## Dependencias Externas

- MariaDB/MySQL para persistencia
- Frontend Next.js existente
- shadcn/ui componentes (ya instalados)

## Notas de Implementación

1. **No modificar lógica de módulos existentes** - La autorización se integra como middleware, no dentro de cada handler
2. **JWT incluye is_root** - Para evitar consulta DB en hot path
3. **Seed inicial tiene root** - Crear primer usuario root manualmente después de migraciones
4. **Tests con drop/recreate** - Cada test de integración ejecuta script de limpieza