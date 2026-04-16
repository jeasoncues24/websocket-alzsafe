# Story 8.5.4: User Service con Promoción Segura a Root

## Parent Epic: Epic 8.5 - Sistema de Gestión de Usuarios con Roles y Módulos

## Objetivo

Implementar lógica de negocio para gestión de usuarios y promoción segura.

## Tareas

- [x] Crear `domain/user_service.go` con UserService
- [x] Implementar método `CreateUser` con validación de rol
- [x] Implementar método `UpdateUser` con protección de root
- [x] Implementar método `DeleteUser` con validación
- [x] Implementar método `PromoteToRoot` con validación de existente root
- [x] Implementar método `AssignModules` para override de permisos
- [x] Integrar UserService en handlers HTTP
- [x] Tests unitarios para UserService

## Criterios de Aceptación

1. [x] Crear usuario asigna rol correctamente
2. [x] No se puede cambiar is_root directamente en update
3. [x] PromoteToRoot falla si no existe otro root
4. [x] PromoteToRoot requiere que el solicitante sea root
5. [x] Tests pasan

## Dependencies

- Story 8.5.3 (Authorization Service)

## Estimated Effort

2 horas