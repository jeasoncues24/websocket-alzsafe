# Story 9.3: Aislamiento de datos por empresa en queries

## Parent Epic: Epic 9 - Mensajería Enriquecida con Contexto

## Objetivo

Implementar filtro automático por empresa_id del token JWT en todas las queries de mensajes y sesiones.

## Tareas

- [x] Modificar queries de mensajes para filtrar por empresa_id del token
- [x] Modificar queries de sesiones WhatsApp para filtrar por empresa_id
- [x] Implementar lógica para super_admin (empresa_id null) puede ver todas
- [x] Implementar soporte para header X-Empresa-ID en super_admin
- [x] Tests de aislamiento de datos

## Criterios de Aceptación

1. [x] Usuario con empresa_id específica solo ve datos de su empresa
2. [x] super_admin sin empresa_id ve todas las empresas
3. [x] super_admin puede usar X-Empresa-ID para filtrar por empresa específica
4. [x] Tests de aislamiento pasan

## Dependencies

- Story 9.2 (Payload enriquecido)

## Estimated Effort

2 horas

## Implementation Notes

### Archivos modificados/creados:
- `internal/domain/empresa_filter.go` - Nuevo archivo con lógica de aislamiento
- `internal/domain/empresa_filter_test.go` - Tests unitarios
- `internal/http/handlers.go` - Modificado para usar filtro de empresa
- `internal/http/handlers/auth.go` - Actualizado para usar interfaz
- `internal/http/handlers/companies.go` - Actualizado para usar interfaz
- `internal/http/router.go` - Actualizado para inyectar empresaStore

### Endpoints protegidos:
- POST /api/message
- GET /api/messages
- POST /api/broadcast
- GET /api/broadcast/

### Lógica de aislamiento:
1. Extraer claims del token JWT
2. Si es root (is_root=true) y no tiene empresa_id: puede ver todas
3. Si es root y pasa X-Empresa-ID header: filtra por esa empresa
4. Si no es root: solo puede ver su empresa del token

## Status

**DONE** - 2026-04-16