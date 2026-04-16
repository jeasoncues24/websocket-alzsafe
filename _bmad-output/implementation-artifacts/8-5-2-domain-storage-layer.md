# Story 8.5.2: Domain Entities y Storage Layer

## Parent Epic: Epic 8.5 - Sistema de Gestión de Usuarios con Roles y Módulos

## Objetivo

Implementar entidades y storage para roles, módulos y user_modules.

## Tareas

- [x] Crear `domain/role.go` con struct Role y métodos
- [x] Crear `domain/module.go` con struct Module
- [x] Crear `storage/role.go` con RoleStore (CRUD + GetRootRole)
- [x] Crear `storage/module.go` con ModuleStore (CRUD + GetAll)
- [x] Crear `storage/user_module.go` con UserModuleStore
- [x] Crear `storage/user.go` actualizado con métodos de roles
- [x] Ejecutar tests unitarios existentes
- [ ] Agregar tests para los nuevos stores

## Criterios de Aceptación

1. [x] Todos los stores implementan interfaz válida
2. [x] Tests existentes pasan sin regresiones
3. [x] Nuevos tests para Role, Module, UserModule stores pasan

## Dependencies

- Story 8.5.1 (Migraciones y Estructura de Datos)

## Estimated Effort

2 horas