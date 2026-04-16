# Story 9.5: Dashboard con métricas por empresa

## Parent Epic: Epic 9 - Mensajería Enriquecida con Contexto

## Objetivo

Implementar endpoint de métricas del dashboard con información por empresa.

## Tareas

- [ ] Implementar GET /api/dashboard/metricas con métricas de la empresa del usuario
- [ ] Métricas: total_mensajes, mensajes_hoy, mensajes_semana, mensajes_exitosos, mensajes_fallidos, sesiones_activas, broadcasts_ejecutados
- [ ] Para super_admin: soportar parámetro empresa_id opcional
- [ ] Para super_admin sin param: retornar métricas agregadas de todas las empresas
- [ ] Tests de métricas

## Criterios de Aceptación

1. [ ] GET /api/dashboard/metricas retorna métricas de la empresa del usuario
2. [ ] Métricas calculadas correctamente (hoy, semana, 30 días)
3. [ ] super_admin puede filtrar por empresa_id específico
4. [ ] super_admin sin param retorna métricas agregadas

## Dependencies

- Story 9.4 (Historial de mensajes con filtros)

## Estimated Effort

2 horas