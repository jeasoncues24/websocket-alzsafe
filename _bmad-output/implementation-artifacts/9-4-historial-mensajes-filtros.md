# Story 9.4: Historial de mensajes por empresa con filtros

## Parent Epic: Epic 9 - Mensajería Enriquecida con Contexto

## Objetivo

Implementar endpoint de historial de mensajes con filtros por empresa, fecha y teléfono.

## Tareas

- [x] Implementar GET /api/messages con filtros: empresa_id, desde, hasta, telefono
- [x] Implementar paginación: page, limit con metadatos (total, page, limit, total_pages)
- [x] Agregar filtros por rango de fechas (desde/hasta)
- [x] Agregar filtro por teléfono (búsqueda parcial)
- [x] Tests de filtros y paginación

## Criterios de Aceptación

1. [x] GET /api/messages filtra por empresa_id (si no es super_admin ignora el param)
2. [x] Filtros por rango de fechas funcionan
3. [x] Filtro por teléfono con búsqueda parcial funciona
4. [x] Paginación retorna metadatos correctos (total, page, limit, total_pages)

## Dependencies

- Story 9.3 (Aislamiento de datos por empresa)

## Estimated Effort

2 horas

## Implementación

### Cambios realizados:

1. **domain/message.go**: Agregado campo `TotalPages` a `MessagesListResponse`

2. **storage/messages.go**: 
   - Actualizada interfaz `MessagesRepository` para incluir filtro `telefono`
   - Actualizados métodos `GetByEmpresa` y `GetByEmpresaAndDateRange` para soportar filtro por teléfono
   - Implementado filtro parcial por teléfono usando LIKE

3. **http/handlers.go**: 
   - Agregados filtros `desde`, `hasta`, `telefono`, `empresa_id`
   - El filtro `empresa_id` solo es procesado para usuarios root
   - Agregado cálculo de `TotalPages` en la respuesta
   - Mantenida compatibilidad con filtros `start_date`/`end_date` existentes

4. **http/handlers_test.go**:
   - Actualizado mock para nueva firma de métodos
   - Agregado test de cálculo de paginación `TestGetMessagesTotalPagesCalculation`