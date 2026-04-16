# Story 9.2: Payload enriquecido en respuestas de mensajes

## Parent Epic: Epic 9 - Mensajería Enriquecida con Contexto

## Objetivo

Enriquecer los endpoints de mensajes con información contextual del usuario, empresa y sesión.

## Tareas

- [ ] Modificar respuesta de GET /api/message/send para incluir usuario_id, ruc_empresa, empresa_nombre, session_id
- [ ] Modificar respuesta de GET /api/messages para incluir campos enriquecidos en cada mensaje
- [ ] Modificar respuesta de POST /api/broadcast/send para incluir ruc_empresa y empresa_nombre
- [ ] Modificar respuesta de GET /api/broadcast/{id}/results para incluir ruc_empresa en cada resultado
- [ ] Tests de integración para verificar campos enriquecidos

## Criterios de Aceptación

1. [ ] GET /api/message/send retorna campos enriquecidos
2. [ ] GET /api/messages lista incluye campos enriquecidos
3. [ ] POST /api/broadcast/send retorna ruc_empresa y empresa_nombre
4. [ ] GET /api/broadcast/{id}/results incluye ruc_empresa por resultado

## Dependencies

- Story 9.1 (Endpoint /usuario/me)

## Estimated Effort

2 horas