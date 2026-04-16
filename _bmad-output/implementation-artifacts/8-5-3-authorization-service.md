# Story 8.5.3: Authorization Service

## Parent Epic: Epic 8.5 - Sistema de Gestión de Usuarios con Roles y Módulos

## Objetivo

Implementar servicio centralizado de autorización con bypass para root.

## Tareas

- [x] Crear `domain/authorization.go` con AuthorizationService
- [x] Implementar `CanAccess(ctx, moduleSlug)` con lógica de bypass root
- [x] Implementar `GetUserModules(ctx)` con lógica de resolución
- [x] Crear middleware de autorización para rutas
- [x] Integrar middleware en router
- [x] Tests unitarios para AuthorizationService

## Criterios de Aceptación

1. [x] `CanAccess` retorna true para root sin verificar módulos
2. [x] `CanAccess` verifica permisos para usuarios no-root
3. [x] `GetUserModules` devuelve todos los módulos para root
4. [x] Middlego intercepta requests y valida acceso

## Dependencies

- Story 8.5.2 (Domain Entities y Storage Layer)

## Estimated Effort

2 horas