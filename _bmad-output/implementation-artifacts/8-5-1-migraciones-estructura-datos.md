# Story 8.5.1: Migraciones y Estructura de Datos

## Parent Epic: Epic 8.5 - Sistema de Gestión de Usuarios con Roles y Módulos

## Objetivo

Crear tablas de roles, módulos y user_modules con seeds iniciales y trigger de protección.

## Tareas

- [x] Crear migración `009_create_roles.up.sql` con seeds (root, admin, operador, viewer)
- [x] Crear migración `009_create_roles.down.sql` (drop table)
- [x] Crear migración `010_create_modules.up.sql` con seeds (6 módulos del sistema)
- [x] Crear migración `010_create_modules.down.sql` (drop table)
- [x] Crear migración `011_create_user_modules.up.sql`
- [x] Crear migración `011_create_user_modules.down.sql` (drop table)
- [x] Crear migración `012_alter_admin_users_add_role.up.sql`
- [x] Crear migración `012_alter_admin_users_add_role.down.sql`
- [x] Crear trigger para protección de is_root en roles
- [x] Ejecutar migraciones en entorno de desarrollo
- [x] Verificar que las tablas y datos seed existen

## Criterios de Aceptación

1. [x] Tabla `roles` tiene 4 registros seed (root, admin, operador, viewer)
2. [x] Tabla `modules` tiene 6 registros seed (dashboard, companies, messages, sessions, broadcasts, settings)
3. [x] Trigger previene UPDATE de is_root en roles
4. [x] FK entre admin_users y roles funciona

## Dependencies

- Ninguna

## Estimated Effort

2 horas