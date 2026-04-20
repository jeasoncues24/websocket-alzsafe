# Epic 12: Backend Mensajeria Real con API Key por Telefono

## Objetivo

Implementar envio real de mensajes y difusiones usando `whatsmeow.Client.SendMessage`, unificar endpoints publicos de mensajeria en rutas en espanol y eliminar la dependencia de `telefono_id` en bodies cuando se usa `api_token` por numero.

## Alcance

- Backend Go unicamente.
- Autenticacion por `api_token` para rutas publicas de mensajeria.
- Envio real y actualizacion de estado de mensajes.
- Documentacion tecnica de endpoints.

## Historias

### 12.1 SendTextMessage utilitario

- Crear `internal/whatsapp/send.go` con:
  - `SendTextMessage(ctx, manager, accountID, destino, contenido) error`
- Resolver cliente WhatsApp desde `Manager` por `accountID`.
- Validar conexion activa.
- Enviar texto via `client.SendMessage()`.

**Criterio de aceptacion**
- Si no existe cliente o no esta conectado retorna error trazable.
- Si el envio es exitoso retorna `nil`.

### 12.2 Worker de difusiones con envio real

- Actualizar `internal/whatsapp/broadcast.go`:
  - Inyectar `*Manager` en `BroadcastWorker`.
  - Agregar `AccountID` en `BroadcastJob`.
  - Conectar `processItem()` a `SendTextMessage`.

**Criterio de aceptacion**
- Cada item del broadcast intenta envio real con retries configurados.
- El resultado por item marca `sent` o `failed`.

### 12.3 POST /api/mensajes real y sin telefono_id

- Actualizar `internal/http/handlers/v1_messages.go`:
  - Body solo `{ destino, contenido }`.
  - `telefono_id` derivado de `ApiKeyClaims`.
  - Persistir `pending` con `msgRepo.Create` (sin ignorar error).
  - Envio real con `SendTextMessage`.
  - Actualizar estado con `msgRepo.UpdateEstado` a `sent` o `failed`.
  - Responder `202 Accepted` con `reference_id`.

**Criterio de aceptacion**
- No acepta `telefono_id` en body.
- No responde exito si falla la persistencia.
- Estado del mensaje queda consistente con resultado de envio.

### 12.4 POST /api/difusiones real y sin telefono_id

- Actualizar `internal/http/handlers/v1_broadcasts.go`:
  - Body solo `{ destinos, mensaje }`.
  - `telefono_id` derivado de `ApiKeyClaims`.
  - Crear job con `reference_id` UUID.
  - Encolar envio real en `BroadcastWorker`.
  - Responder `202 Accepted`.

**Criterio de aceptacion**
- Difusion se procesa de forma asincrona con envio real.
- `reference_id` permite consultar estado luego.

### 12.5 Router unificado y limpieza de duplicados

- Actualizar `internal/http/router.go`:
  - Canonicos: `/api/mensajes`, `/api/difusiones`, `/api/difusiones/{id}`.
  - Solo `apiKeyProtected` para esos endpoints.
  - Eliminar rutas duplicadas legacy y `v1/messages|broadcasts`.

**Criterio de aceptacion**
- Las rutas canonicas funcionan con `api_token`.
- Las rutas duplicadas ya no estan expuestas.

### 12.6 Documentacion de endpoints

- Actualizar `_bmad-output/implementation-artifacts/api-key-token-por-numero-endpoints.md`:
  - Endpoints nuevos canonicos.
  - Sin `telefono_id` en body.
  - Respuestas `202 Accepted` y ejemplos reales.

**Criterio de aceptacion**
- La documentacion coincide con el comportamiento real del backend.

## Riesgos y Mitigaciones

- Sesion WhatsApp inactiva para el telefono del token.
  - Mitigacion: error explicito `SESSION_NOT_ACTIVE` y estado `failed`.
- Destinos con formato invalido.
  - Mitigacion: registrar fallo por item en difusion y mantener trazabilidad.

## Definicion de Done

- Build `go build ./...` exitoso.
- Tests `go test ./...` en verde.
- Endpoints de mensajeria operan solo con API Key por telefono.
- Envio real conectado a WhatsApp para mensajes y difusiones.
