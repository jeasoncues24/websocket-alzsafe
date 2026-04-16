# Historia: 9.2 - Payload enriquecido en respuestas de mensajes

**Epic:** 9 - Mensajería Enriquecida
**Estado:** done
**Story ID:** 9-2-payload-mensajes-enriquecido

## Objetivo

Enriquecer los endpoints de mensajes con información contextual del usuario, empresa y sesión.

## Tareas Requeridas

- [x] Modificar respuesta de POST /api/message para incluir usuario_id, ruc_empresa, empresa_nombre, session_id
- [x] Modificar respuesta de GET /api/messages para incluir campos enriquecidos en cada mensaje
- [x] Modificar respuesta de POST /api/broadcast para incluir ruc_empresa y empresa_nombre
- [x] Modificar respuesta de GET /api/broadcast/{id} para incluir ruc_empresa en cada resultado
- [ ] Tests de integración para verificar campos enriquecidos

## Criterios de Aceptación

1. [x] POST /api/message retorna campos enriquecidos
2. [x] GET /api/messages lista incluye campos enriquecidos
3. [x] POST /api/broadcast retorna ruc_empresa y empresa_nombre
4. [x] GET /api/broadcast/{id} incluye ruc_empresa por resultado

## Cambios Realizados

### Backend (Go)

1. **internal/domain/message.go**
   - MessageResponse: Agregados campos `UsuarioID`, `RUCEmpresa`, `EmpresaNombre`, `SessionID`
   - MessagesListResponse: Agregados campos `UsuarioID`, `RUCEmpresa`, `EmpresaNombre`

2. **internal/domain/broadcast.go**
   - BroadcastResponse: Agregados campos `RUCEmpresa`, `EmpresaNombre`
   - BroadcastResult: Agregado campo `RUCEmpresa`

3. **internal/storage/sessions.go**
   - SessionState: Agregado campo `SessionID`

4. **internal/http/handlers.go**
   - HandlePostMessage: Extrae claims del JWT y retorna campos enriquecidos
   - HandleGetMessages: Extrae claims del JWT y retorna campos enriquecidos
   - HandlePostBroadcast: Extrae claims del JWT y retorna campos enriquecidos + agrega RUCEmpresa a cada resultado
   - HandleGetBroadcast: Retorna resultados con RUCEmpresa (ya enriquecidos en el worker)

## Notas

- Los campos enriquecidos se obtienen del JWT claims (user_id, empresa_ruc, empresa_nombre)
- SessionID se obtiene del estado de la sesión en memoria
- Los resultados de broadcast ya incluyen ruc_empresa porque se agrega al crear el resultado en el worker
